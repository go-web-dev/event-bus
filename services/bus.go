package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

func NewBus(d DB) *Bus {
	b := &Bus{
		db:      db{d},
		streams: map[string]Stream{},
	}
	return b
}

type Bus struct {
	mu      sync.Mutex
	db      db
	streams map[string]Stream
}

func (b *Bus) Init() error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	logger.Info("initializing event bus with streams")
	var streamVar Stream
	err := b.db.fetch("stream:", &streamVar, func(res fetchResult) {
		stream := res.item.(*Stream)
		b.streams[stream.Name] = *stream
	})
	if err != nil {
		logger.Error("could not fetch streams from db", zap.Error(err))
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
	stream := Stream{
		Name:      streamName,
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
	}
	err := b.db.set(stream.key(), stream.value(), 0)
	if err != nil {
		logger.Debug("could not save stream to db", zap.Error(err))
		return Stream{}, err
	}
	b.streams[streamName] = stream
	logger.Info("successfully created stream", zap.String("stream_id", stream.ID))
	return stream, nil
}

func (b *Bus) DeleteStream(streamName string) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("event:%s", stream.ID)
	keysToDelete := make([][]byte, 0)
	var eventVar Event
	err = b.db.fetch(key, &eventVar, func(res fetchResult) {
		keysToDelete = append(keysToDelete, res.key)
	})
	if err != nil {
		logger.Error("could not fetch events from db", zap.Error(err))
		return err
	}

	err = b.db.delete(append(keysToDelete, stream.key())...)
	if err != nil {
		logger.Debug("could not delete stream from db", zap.Error(err))
		return err
	}

	delete(b.streams, streamName)
	logger.Info("successfully deleted stream", zap.String("stream_id", stream.ID))
	return nil
}

func (b *Bus) GetStreamInfo(streamName string) (Stream, error) {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return Stream{}, err
	}

	logger.Info("successfully got stream info", zap.String("stream_id", stream.ID))
	return stream, nil
}

func (b *Bus) WriteEvent(streamName string, body json.RawMessage) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return err
	}

	evt := Event{
		ID:        uuid.New().String(),
		StreamID:  stream.ID,
		CreatedAt: time.Now().UTC(),
		Body:      body,
	}
	err = stream.WriteEvent(b.db, evt)
	if err != nil {
		return err
	}

	logger.Info("successfully wrote event", zap.String("event_id", evt.ID))
	return nil
}

type EventProcessor interface {
	Process(event Event) error
}

// add possibility reprocess a message
// add more custom errors to hide internals
// add backups to a directory + name files with timestamp
// add possibility to snapshot individual streams

func (b *Bus) ProcessEvents(streamName string, processor EventProcessor) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return err
	}

	logger.Info("events processing started", zap.String("stream_id", stream.ID))
	err = stream.processEvents(b.db, processor)
	if err != nil {
		logger.Error("could not process events", zap.Error(err))
		return err
	}
	logger.Info("events processing finished", zap.String("stream_id", stream.ID))
	return nil
}

func (b *Bus) streamLookup(streamName string) (Stream, error) {
	stream, ok := b.streams[streamName]
	if !ok {
		return Stream{}, fmt.Errorf("stream '%s' not found", streamName)
	}
	return stream, nil
}
