package services

import (
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	StreamID  string          `json:"stream_id"`
	Processed bool            `json:"processed"`
	Body      json.RawMessage `json:"body"`
}

func (m Message) key() string {
	return fmt.Sprintf("message:%s:%s", m.StreamID, m.ID)
}
