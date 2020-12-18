package models

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/go-web-dev/event-bus/logging"
)

// Stream represents the Event Bus stream that stores certain events
type Stream struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Key generates stream specific key to be stored inside the database
func (s Stream) Key() []byte {
	return []byte(fmt.Sprintf("stream:%s", s.ID))
}

// Value generates stream specific value to be stored inside the database
func (s Stream) Value() []byte {
	bs, err := json.Marshal(s)
	if err != nil {
		logging.Logger.Error("could not marshal stream", zap.Error(err))
		return []byte{}
	}
	return bs
}
