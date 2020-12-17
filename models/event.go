package models

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

// Event statuses
const (
	EventUnprocessedStatus = uint8(0)
	EventProcessedStatus   = uint8(1)
	EventRetryStatus       = uint8(2)
)

// Event represents the the event structure in the Event Bus
type Event struct {
	ID        string          `json:"id"`
	StreamID  string          `json:"stream_id"`
	Status    uint8           `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	Body      json.RawMessage `json:"body"`
}

// Key generates the event specific key for storing in database
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

// Value generates the event specific value for storing in database
func (e Event) Value() []byte {
	bs, err := json.Marshal(e)
	if err != nil {
		logging.Logger.Debug("could not marshal event", zap.Error(err))
		return []byte{}
	}
	return bs
}

// ExpiresAt represents the TTL for an event to be stored inside the database
func (e Event) ExpiresAt() uint64 {
	return uint64(e.CreatedAt.Add(time.Hour * 720).UnixNano())
}
