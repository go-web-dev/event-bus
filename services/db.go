package services

import (
	"encoding/json"
	"io"

	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
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
			err := item.Value(func(value []byte) error {
				err := json.Unmarshal(value, obj)
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
