package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
)

// NewBus creates a new Event Bus service
func NewBus(d DB) *Bus {
	b := &Bus{
		db:      db{d},
		streams: map[string]models.Stream{},
	}
	return b
}

// Bus represents the Event Bus service
type Bus struct {
	mu      sync.RWMutex
	db      db
	streams map[string]models.Stream
}

// Init initializes the event bus with helper data such as streams
func (b *Bus) Init() error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	logger.Info("initializing event bus with streams")
	var streamVar models.Stream
	txn := b.db.fetch("stream:", &streamVar, func(res fetchResult) {
		stream := res.item.(*models.Stream)
		b.streams[stream.Name] = *stream
	})
	err := b.db.txn(false, txn)
	if err != nil {
		logger.Error("could not fetch streams from db", zap.Error(err))
		return err
	}

	logger.Info("successfully initialized event bus with streams")
	return nil
}

// CreateStream creates a new stream for messaging in the Event Bus
func (b *Bus) CreateStream(streamName string) (models.Stream, error) {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.streams[streamName]; ok {
		return models.Stream{}, fmt.Errorf("stream: '%s' already exists", streamName)
	}
	stream := models.Stream{
		Name:      streamName,
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
	}
	txn := b.db.set(stream.Key(), stream.Value(), 0)
	err := b.db.txn(true, txn)
	if err != nil {
		logger.Debug("could not save stream to db", zap.Error(err))
		return models.Stream{}, err
	}
	b.streams[streamName] = stream
	logger.Info("successfully created stream", zap.String("stream_id", stream.ID))
	return stream, nil
}

// DeleteStream deletes a stream and its associated messages
func (b *Bus) DeleteStream(streamName string) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("event:%s", stream.ID)
	keysToDelete := make([][]byte, 0)
	var eventVar models.Event
	fetchTxn := b.db.fetch(key, &eventVar, func(res fetchResult) {
		keysToDelete = append(keysToDelete, res.key)
	})
	err = b.db.txn(false, fetchTxn)
	if err != nil {
		logger.Debug("could not fetch stream events", zap.Error(err))
		return err
	}

	deleteTxn := b.db.delete(append(keysToDelete, stream.Key())...)
	err = b.db.txn(true, deleteTxn)
	if err != nil {
		logger.Debug("could not delete stream and its events", zap.Error(err))
		return err
	}

	delete(b.streams, streamName)
	logger.Info("successfully deleted stream", zap.String("stream_id", stream.ID))
	return nil
}

// GetStreamInfo gets a stream's short information
func (b *Bus) GetStreamInfo(streamName string) (models.Stream, error) {
	logger := logging.Logger
	b.mu.RLock()
	defer b.mu.RUnlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return models.Stream{}, err
	}

	logger.Info("successfully got stream info", zap.String("stream_id", stream.ID))
	return stream, nil
}

// GetStreamEvents gets all events for a certain stream
func (b *Bus) GetStreamEvents(streamName string) ([]models.Event, error) {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return []models.Event{}, err
	}

	logger.Info("getting all stream events", zap.String("stream_id", stream.ID))
	key := fmt.Sprintf("event:%s", stream.ID)
	events, err := b.db.streamEvents(key, nil)
	if err != nil {
		logger.Error("could not get stream events", zap.Error(err))
		return []models.Event{}, err
	}
	logger.Info("successfully got all stream events", zap.String("stream_id", stream.ID))
	return events, nil
}

// WriteEvent writes an event to a certain stream in the Event Bus
func (b *Bus) WriteEvent(streamName string, body json.RawMessage) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return err
	}

	evt := models.Event{
		ID:        uuid.New().String(),
		StreamID:  stream.ID,
		CreatedAt: time.Now().UTC(),
		Body:      body,
		Status:    models.EventUnprocessedStatus,
	}
	txn := b.db.set(evt.Key(models.EventUnprocessedStatus), evt.Value(), evt.ExpiresAt())
	err = b.db.txn(true, txn)
	if err != nil {
		logger.Debug("could not write event to db", zap.Error(err))
		return err
	}

	logger.Info("successfully wrote event", zap.String("event_id", evt.ID))
	return nil
}

// MarkEvent marks an event / changes it's status
func (b *Bus) MarkEvent(eventID string, status uint8) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	key := "event:"
	chooseKeyFunc := func(item *badger.Item) bool {
		keys := strings.Split(string(item.Key()), ":")
		return keys[len(keys)-1] == eventID
	}
	events, err := b.db.streamEvents(key, chooseKeyFunc)
	if err != nil {
		logger.Error("could not stream events", zap.Error(err))
		return err
	}
	if len(events) == 0 {
		return fmt.Errorf("event '%s' not found", eventID)
	}
	evt := events[0]
	evt.Status = status

	keys := [][]byte{
		evt.Key(models.EventUnprocessedStatus),
		evt.Key(models.EventProcessedStatus),
		evt.Key(models.EventRetryStatus),
	}
	deleteKeysTxn := b.db.delete(keys...)
	markEventTxn := b.db.set(evt.Key(status), evt.Value(), evt.ExpiresAt())
	err = b.db.txn(true, deleteKeysTxn, markEventTxn)
	if err != nil {
		logger.Error(
			"could not mark event",
			zap.Error(err),
			zap.String("event_id", eventID),
		)
		return err
	}

	logger.Info("successfully marked event", zap.String("event_id", eventID))
	return nil
}

// ProcessEvents processes/retries all available events in the queue
func (b *Bus) ProcessEvents(streamName string, retry bool) ([]models.Event, error) {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, err := b.streamLookup(streamName)
	if err != nil {
		return []models.Event{}, err
	}

	logger.Info("events processing started", zap.String("stream_id", stream.ID))
	status := models.EventUnprocessedStatus
	if retry {
		status = models.EventRetryStatus
	}
	key := fmt.Sprintf("event:%s:%d", stream.ID, status)
	events, err := b.db.streamEvents(key, nil)
	if err != nil {
		logger.Error("could not process events", zap.Error(err))
		return []models.Event{}, err
	}
	logger.Info("events processing finished", zap.String("stream_id", stream.ID))
	return events, nil
}

func (b *Bus) streamLookup(streamName string) (models.Stream, error) {
	stream, ok := b.streams[streamName]
	if !ok {
		return models.Stream{}, fmt.Errorf("stream '%s' not found", streamName)
	}
	return stream, nil
}
