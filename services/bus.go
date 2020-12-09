package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Stream struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewBus(c Checkpoint) Bus {
	b := Bus{
		c:       c,
		streams: map[string]Stream{},
	}
	return b
}

type Bus struct {
	streams map[string]Stream
	c       Checkpoint
}

func (b Bus) Create(streamName string) (Stream, error) {
	if _, ok := b.streams[streamName]; ok {
		return Stream{}, fmt.Errorf("stream: '%s' already exists", streamName)
	}
	s := Stream{
		Name:      streamName,
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
	}
	b.streams[streamName] = s
	return s, nil
}

func (b Bus) Delete(streamName string) error {
	if _, ok := b.streams[streamName]; !ok {
		return fmt.Errorf("stream: '%s' not found", streamName)
	}
	delete(b.streams, streamName)
	return nil
}
