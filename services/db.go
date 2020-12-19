package services

import (
	"context"
	"encoding/json"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/pb"
	"go.uber.org/zap"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
)

// DB represents the file database that stores streams and events
type DB interface {
	NewTransaction(update bool) *badger.Txn
	NewStream() *badger.Stream
	DropAll() error
	Close() error
}

type db struct {
	DB
}

type fetchResult struct {
	item interface{}
	key  []byte
}

type transactionFunc func(txn *badger.Txn) error

func (d db) txn(update bool, transactions ...transactionFunc) error {
	transaction := d.NewTransaction(update)
	for _, txn := range transactions {
		err := txn(transaction)
		if err != nil {
			transaction.Discard()
			return err
		}
	}
	return transaction.Commit()
}

func (d db) fetch(key string, obj interface{}, cb func(fetchResult)) transactionFunc {
	logger := logging.Logger
	return func(txn *badger.Txn) error {
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
	}
}

func (d db) delete(keys ...[]byte) transactionFunc {
	return func(txn *badger.Txn) error {
		for _, key := range keys {
			err := txn.Delete(key)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (d db) set(key, value []byte, expiresAt uint64) transactionFunc {
	return func(txn *badger.Txn) error {
		entry := &badger.Entry{
			Key:       key,
			Value:     value,
			ExpiresAt: expiresAt,
		}
		return txn.SetEntry(entry)
	}
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
				logger.Error("could not unmarshal message", zap.Error(err))
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
