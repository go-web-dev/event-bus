package services

import (
	"container/list"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type DB interface {
}

func NewBus(db DB) *Bus {
	b := &Bus{
		c:       NewCheckpoint(),
		db:      db,
		streams: map[string]Stream{},
	}
	return b
}

type Bus struct {
	c       Checkpoint
	mu      sync.Mutex
	db      DB
	streams map[string]Stream
}

func (b *Bus) CreateStream(streamName string) (Stream, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.streams[streamName]; ok {
		return Stream{}, fmt.Errorf("stream: '%s' already exists", streamName)
	}
	s := Stream{
		Name:      streamName,
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
		History:   map[string]Message{},
		queue:     list.New(),
	}
	messages := []Message{
		{
			CreatedAt: time.Now().Add(time.Second * 1),
			Body:      []byte(`{"event": "message 1"}`),
		},
		{
			CreatedAt: time.Now().Add(time.Second * 2),
			Body:      []byte(`{"event": "message 2"}`),
		},
		{
			CreatedAt: time.Now().Add(time.Second * 3),
			Body:      []byte(`{"event": "message 3"}`),
		},
	}
	s.queue.PushBack(messages[0])
	s.queue.PushBack(messages[1])
	s.queue.PushBack(messages[2])
	b.streams[streamName] = s
	return s, nil
}

func (b *Bus) DeleteStream(streamName string) error {
	// delete all messages associated with the stream
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.streams[streamName]; !ok {
		return fmt.Errorf("stream: '%s' not found", streamName)
	}
	delete(b.streams, streamName)
	return nil
}

func (b *Bus) WriteMessage(streamName string, body json.RawMessage) error {
	// decides regarding Message vs Event
	b.mu.Lock()
	defer b.mu.Unlock()
	stream, ok := b.streams[streamName]
	if !ok {
		return fmt.Errorf("stream: '%s' not found", streamName)
	}
	return stream.WriteMessage(body)
}

type MessageProcessor interface {
	Process(messages Message) error
}

// load streams and events
// save unprocessed messages to queue
// save records with ttl
// add snapshot operation
// add possibility reprocess a message
// when processing events: 1) remove old key 2) add same key with prefix `processed` for easier retrieval

func (b *Bus) ProcessMessages(streamName string, processor MessageProcessor) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	stream, ok := b.streams[streamName]
	if !ok {
		return fmt.Errorf("stream: '%s' not found", streamName)
	}
	for {
		element := stream.queue.Front()
		if element == nil {
			return nil
		}
		msg := element.Value.(Message)
		err := processor.Process(msg)
		if err != nil {
			return err
		}
		msg.Processed = true
		stream.History[msg.ID] = msg
		stream.queue.Remove(element)
	}
}
