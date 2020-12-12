package services

import (
	"container/list"
	"encoding/json"
	"fmt"
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
	eventsMap   map[string]struct{}
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
		return txn.Set(evt.key(), evt.value())
	})
	if err != nil {
		logging.Logger.Debug("could not save stream to db", zap.Error(err))
		return err
	}

	// To avoid filling the queue on event update
	_, found := s.eventsMap[string(evt.key())]
	if !evt.Processed && !found {
		s.eventsQueue.PushBack(evt)
		s.eventsMap[string(evt.key())] = struct{}{}
	}
	return nil
}

func (s Stream) indexEvents(db DB) error {
	logger := logging.Logger
	events := make([]Event, 0)
	return db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		key := fmt.Sprintf("event:%s:%d", s.ID, eventUnprocessedStatus)
		prefix := []byte(key)
		logger.Info("events indexing started")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			_, found := s.eventsMap[string(item.Key())]
			if found {
				continue
			}
			err := item.Value(func(value []byte) error {
				var evt Event
				err := json.Unmarshal(value, &evt)
				if err != nil {
					logger.Error("could not unmarshal event value from db", zap.Error(err))
					return err
				}
				events = append(events, evt)
				return nil
			})
			if err != nil {
				return err
			}
		}
		// ensure correct events order
		for i := len(events) - 1; i >= 0; i-- {
			s.eventsQueue.PushFront(events[i])
		}
		logger.Info("events indexing finished")
		return nil
	})
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
