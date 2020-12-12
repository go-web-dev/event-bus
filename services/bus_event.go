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
	Processed bool            `json:"processed"`
	Body      json.RawMessage `json:"body"`
}

func (e Event) key() []byte {
	status := eventUnprocessedStatus
	if e.Processed {
		status = eventProcessedStatus
	}
	return []byte(fmt.Sprintf("event:%s:%d:%s:%s", e.StreamID, status, e.CreatedAt.Format(time.RFC3339), e.ID))
}

func (e Event) unprocessedKey() []byte {
	return []byte(fmt.Sprintf("event:%s:%d:%s:%s", e.StreamID, eventUnprocessedStatus, e.CreatedAt.Format(time.RFC3339), e.ID))
}

func (e Event) value() []byte {
	bs, err := json.Marshal(e)
	if err != nil {
		logging.Logger.Debug("could not marshal event", zap.Error(err))
		return []byte{}
	}
	return bs
}
