package services

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

const (
	eventUnprocessedStatus = 0
	eventProcessedStatus   = 1
	eventRetryStatus       = 2
)

type Event struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	StreamID  string          `json:"stream_id"`
	Body      json.RawMessage `json:"body"`
}

func (e Event) key(status int) []byte {
	key := fmt.Sprintf(
		"event:%s:%d:%s:%s",
		e.StreamID,
		status,
		e.CreatedAt.Format(time.RFC3339),
		e.ID,
	)
	return []byte(key)
}

func (e Event) value() []byte {
	bs, err := json.Marshal(e)
	if err != nil {
		logging.Logger.Debug("could not marshal event", zap.Error(err))
		return []byte{}
	}
	return bs
}

func (e Event) expiresAt() uint64 {
	return uint64(e.CreatedAt.Add(time.Hour * 720).UnixNano())
}
