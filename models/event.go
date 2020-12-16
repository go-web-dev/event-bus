package models

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

const (
	EventUnprocessedStatus = uint8(0)
	EventProcessedStatus   = uint8(1)
	EventRetryStatus       = uint8(2)
)

type Event struct {
	ID        string          `json:"id"`
	StreamID  string          `json:"stream_id"`
	Status    uint8           `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	Body      json.RawMessage `json:"body"`
}

func (e Event) Key(status uint8) []byte {
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
