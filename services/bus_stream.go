package services

import (
	"container/list"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

type Stream struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	events    *list.List
}

func (s Stream) WriteEvent(db DB, body json.RawMessage) error {
	evt := Event{
		ID:        uuid.New().String(),
		StreamID:  s.ID,
		CreatedAt: time.Now().UTC(),
		Body:      body,
	}
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Set(evt.key(), evt.value())
	})
	if err != nil {
		logging.Logger.Debug("could not save stream to db", zap.Error(err))
		return err
	}
	s.events.PushBack(evt)
	return nil
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
