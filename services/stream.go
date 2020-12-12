package services

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v2/pb"
	"time"

	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

type Stream struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	eventsQueue *list.List
	eventsMap   map[string]*list.Element
}

func (s Stream) WriteEvent(db DB, evt Event) error {
	logger := logging.Logger
	err := db.Update(func(txn *badger.Txn) error {
		if evt.Processed {
			err := txn.Delete(evt.unprocessedKey())
			if err != nil {
				logger.Error("could not delete old unprocessed event", zap.Error(err))
				return err
			}
		}
		oneMonth := uint64(evt.CreatedAt.Add(time.Hour * 720).UnixNano())
		entry := &badger.Entry{
			Key:       evt.key(),
			Value:     evt.value(),
			ExpiresAt: oneMonth,
		}
		return txn.SetEntry(entry)
	})
	if err != nil {
		logging.Logger.Debug("could not save stream to db", zap.Error(err))
		return err
	}

	// To avoid filling the queue on event update
	_, found := s.eventsMap[string(evt.key())]
	if !evt.Processed && !found {
		s.eventsQueue.PushBack(evt)
		s.eventsMap[string(evt.key())] = s.eventsQueue.Back()
	}
	return nil
}

func (s Stream) indexEvents(db DB) error {
	logger := logging.Logger
	key := fmt.Sprintf("event:%s:%d", s.ID, eventUnprocessedStatus)
	logger.Info("events indexing started")
	eventRes, err := s.fetchEvents(db, key)
	if err != nil {
		logger.Error("could not index events from db", zap.Error(err))
		return err
	}

	// ensure correct events order
	for i := len(eventRes) - 1; i >= 0; i-- {
		item := eventRes[i]
		_, found := s.eventsMap[string(item.key)]
		if found {
			continue
		}
		s.eventsQueue.PushFront(item.event)
	}
	logger.Info("events indexing finished")
	return nil
}

type eventResult struct {
	event Event
	key   []byte
}

func (s Stream) fetchEvents(db DB, key string) ([]eventResult, error) {
	logger := logging.Logger
	eventRes := make([]eventResult, 0)
	logger.Info("fetching events started")
	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(key)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(value []byte) error {
				var evt Event
				err := json.Unmarshal(value, &evt)
				if err != nil {
					logger.Error("could not unmarshal event value from db", zap.Error(err))
					return err
				}
				eventRes = append(eventRes, eventResult{event: evt, key: item.Key()})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Error("could not fetch events from db", zap.Error(err))
		return []eventResult{}, err
	}
	logger.Info("fetching events finished")
	return eventRes, nil
}

func (s Stream) key() []byte {
	return []byte(fmt.Sprintf("stream:%s", s.ID))
}

func (s Stream) value() []byte {
	bs, err := json.Marshal(s)
	if err != nil {
		logging.Logger.Debug("could not marshal stream", zap.Error(err))
		return []byte{}
	}
	return bs
}

func (s *Stream) streamEvents(db DB, processor EventProcessor) error {
	prefix := fmt.Sprintf("event:%s:%d",s.ID, eventUnprocessedStatus)
	stream := db.NewStream()
	stream.NumGo = 16
	stream.Prefix = []byte(prefix)

	stream.Send = func(list *pb.KVList) error {
		for _, v := range list.Kv {
			var evt Event
			err := json.Unmarshal(v.Value, &evt)
			if err != nil {
				return err
			}
			err = processor.Process(evt)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if err := stream.Orchestrate(context.Background()); err != nil {
		return err
	}
	return nil
}
