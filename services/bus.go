package services

import (
	"container/list"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

type DB interface {
	View(func(txn *badger.Txn) error) error
	Update(func(txn *badger.Txn) error) error
}

func NewBus(db DB) *Bus {
	b := &Bus{
		c:       NewCheckpoint(),
		db:      db,
		streams: map[string]Stream{},
	}
	return b
}

type Bus struct {
	c       Checkpoint
	mu      sync.Mutex
	db      DB
	streams map[string]Stream
}

func (b *Bus) Init() error {
	logger := logging.Logger
	logger.Info("initializing event bus with streams")
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("stream:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(value []byte) error {
				var stream Stream
				err := json.Unmarshal(value, &stream)
				if err != nil {
					logger.Error("could not unmarshal stream value from db", zap.Error(err))
					return err
				}
				stream.eventsQueue = list.New()
				stream.eventsMap = map[string]struct{}{}
				b.streams[stream.Name] = stream
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Debug("could initialize event bus with streams", zap.Error(err))
		return err
	}
	logger.Info("successfully initialized event bus with streams")
	return nil
}

func (b *Bus) CreateStream(streamName string) (Stream, error) {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.streams[streamName]; ok {
		return Stream{}, fmt.Errorf("stream: '%s' already exists", streamName)
	}
	s := Stream{
		Name:      streamName,
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
		eventsQueue:     list.New(),
		eventsMap: map[string]struct{}{},
	}
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(s.key(), s.value())
	})
	if err != nil {
		logger.Debug("could not save stream to db", zap.Error(err))
		return Stream{}, err
	}
	b.streams[streamName] = s
	logger.Info("successfully created stream", zap.String("stream_id", s.ID))
	return s, nil
}

func (b *Bus) DeleteStream(streamName string) error {
	// delete all messages associated with the stream
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	s, ok := b.streams[streamName]
	if !ok {
		return fmt.Errorf("stream: '%s' not found", streamName)
	}

	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(s.key())
	})
	if err != nil {
		logger.Debug("could not delete stream from db", zap.Error(err))
		return err
	}
	delete(b.streams, streamName)
	logger.Info("successfully deleted stream", zap.String("stream_id", s.ID))
	return nil
}

func (b *Bus) GetStreamInfo(streamName string) (Stream, error) {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	s, ok := b.streams[streamName]
	if !ok {
		return Stream{}, fmt.Errorf("stream: '%s' not found", streamName)
	}
	logger.Info("successfully got stream info", zap.String("stream_id", s.ID))
	return s, nil
}

func (b *Bus) WriteEvent(streamName string, body json.RawMessage) error {
	// add expiration when saving to db
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()
	stream, ok := b.streams[streamName]
	if !ok {
		return fmt.Errorf("stream: '%s' not found", streamName)
	}
	evt := Event{
		ID:        uuid.New().String(),
		StreamID:  stream.ID,
		CreatedAt: time.Now().UTC(),
		Body:      body,
	}
	err := stream.WriteEvent(b.db, evt)
	if err != nil {
		return err
	}
	logger.Info("successfully wrote event", zap.String("event_id", evt.ID))
	return nil
}

type EventProcessor interface {
	Process(events Event) error
}

// save records with ttl
// add snapshot operation
// add possibility reprocess a message
// when processing events: 1) remove old key 2) add same key with prefix `processed` for easier retrieval
// add more custom errors to hide internals

func (b *Bus) ProcessEvents(streamName string, processor EventProcessor) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, ok := b.streams[streamName]
	if !ok {
		return fmt.Errorf("stream: '%s' not found", streamName)
	}

	err := stream.indexEvents(b.db)
	if err != nil {
		logger.Debug("could not index stream events", zap.Error(err))
		return err
	}

	logger.Info("events processing started")
	for stream.eventsQueue.Len() > 0 {
		element := stream.eventsQueue.Front()
		if element == nil {
			logger.Info("successfully processed all events")
			break
		}

		evt := element.Value.(Event)
		evt.Processed = true
		err := processor.Process(evt)
		if err != nil {
			// mark event as failed ==> need to retry later
			return err
		}

		stream.eventsQueue.Remove(element)
		delete(stream.eventsMap, string(evt.unprocessedKey()))
		err = stream.WriteEvent(b.db, evt)
		if err != nil {
			return err
		}
	}
	logger.Info("events processing finished")
	return nil
}
