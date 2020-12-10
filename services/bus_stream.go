package services

import (
	"container/list"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Stream struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	CreatedAt time.Time          `json:"created_at"`
	History   map[string]Message `json:"history"` // no need
	queue     *list.List // rename to something like events
}

func (s Stream) WriteMessage(body json.RawMessage) error {
	msg := Message{
		ID:        uuid.New().String(),
		StreamID:  s.ID,
		CreatedAt: time.Now().UTC(),
		Body:      body,
	}
	s.History[msg.ID] = msg
	s.queue.PushBack(msg)
	return nil
}

func (s Stream) key() string {
	return fmt.Sprintf("stream:%s", s.ID)
}
