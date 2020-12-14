package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v2/pb"
	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

type Stream struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (s Stream) WriteEvent(d db, evt Event) error {
	logger := logging.Logger
	err := d.set(evt.key(eventUnprocessedStatus), evt.value(), evt.expiresAt())
	if err != nil {
		logger.Debug("could not write event to db", zap.Error(err))
		return err
	}
	return nil
}

func (s Stream) key() []byte {
	return []byte(fmt.Sprintf("stream:%s", s.ID))
}

func (s Stream) value() []byte {
	bs, err := json.Marshal(s)
	if err != nil {
		logging.Logger.Debug("could not marshal stream", zap.Error(err))
		return []byte{}
	}
	return bs
}

func (s *Stream) processEvents(d db, retry bool) ([]Event, error) {
	logger := logging.Logger
	status := eventUnprocessedStatus
	if retry {
		status = eventRetryStatus
	}
	prefix := fmt.Sprintf("event:%s:%d", s.ID, status)
	stream := d.NewStream()
	stream.NumGo = 16
	stream.Prefix = []byte(prefix)
	events := make([]Event, 0)

	stream.Send = func(list *pb.KVList) error {
		for _, v := range list.Kv {
			var evt Event
			err := json.Unmarshal(v.Value, &evt)
			if err != nil {
				logger.Debug("could not unmarshal message", zap.Error(err))
				continue
			}
			events = append(events, evt)
		}
		return nil
	}

	if err := stream.Orchestrate(context.Background()); err != nil {
		return []Event{}, err
	}
	return events, nil
}
