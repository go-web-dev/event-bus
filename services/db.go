package services

import (
	"context"
	"encoding/json"
	"io"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/pb"
	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
	"github.com/chill-and-code/event-bus/models"
)

type DB interface {
	View(func(txn *badger.Txn) error) error
	Update(func(txn *badger.Txn) error) error
	Backup(w io.Writer, since uint64) (uint64, error)
	NewStream() *badger.Stream
}

type db struct {
	DB
}

type fetchResult struct {
	item interface{}
	key  []byte
}

func (d db) fetch(key string, obj interface{}, cb func(fetchResult)) error {
	logger := logging.Logger
	return d.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(key)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(stream []byte) error {
				err := json.Unmarshal(stream, obj)
				if err != nil {
					logger.Error("could not unmarshal value from db", zap.Error(err))
					return err
				}
				res := fetchResult{
					key:  item.Key(),
					item: obj,
				}
				cb(res)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (d db) delete(keys ...[]byte) error {
	return d.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			err := txn.Delete(key)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (d db) set(key, value []byte, expiresAt uint64) error {
	return d.Update(func(txn *badger.Txn) error {
		entry := &badger.Entry{
			Key:       key,
			Value:     value,
			ExpiresAt: expiresAt,
		}
		return txn.SetEntry(entry)
	})
}

func (d db) streamEvents(key string, chooseKeyFunc func(item *badger.Item) bool) ([]models.Event, error) {
	logger := logging.Logger
	events := make([]models.Event, 0)
	stream := d.NewStream()
	stream.NumGo = 16
	stream.Prefix = []byte(key)
	stream.ChooseKey = chooseKeyFunc

	stream.Send = func(list *pb.KVList) error {
		for _, v := range list.Kv {
			var evt models.Event
			err := json.Unmarshal(v.Value, &evt)
			if err != nil {
				logger.Debug("could not unmarshal message", zap.Error(err))
				continue
			}
			events = append(events, evt)
		}
		return nil
	}

	if err := stream.Orchestrate(context.Background()); err != nil {
		return []models.Event{}, err
	}
	return events, nil
}
