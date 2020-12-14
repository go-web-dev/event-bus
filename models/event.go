package models

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

const (
	EventUnprocessedStatus = 0
	EventProcessedStatus   = 1
	EventRetryStatus       = 2
)

type Event struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	StreamID  string          `json:"stream_id"`
	Body      json.RawMessage `json:"body"`
}

func (e Event) Key(status int) []byte {
	key := fmt.Sprintf(
		"event:%s:%d:%s:%s",
		e.StreamID,
		status,
		e.CreatedAt.Format(time.RFC3339),
		e.ID,
	)
	return []byte(key)
}

func (e Event) Value() []byte {
	bs, err := json.Marshal(e)
	if err != nil {
		logging.Logger.Debug("could not marshal event", zap.Error(err))
		return []byte{}
	}
	return bs
}

func (e Event) ExpiresAt() uint64 {
	return uint64(e.CreatedAt.Add(time.Hour * 720).UnixNano())
}
