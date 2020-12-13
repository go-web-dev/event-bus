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

func (s *Stream) processEvents(d db, processor EventProcessor) error {
	prefix := fmt.Sprintf("event:%s:%d", s.ID, eventUnprocessedStatus)
	stream := d.NewStream()
	stream.NumGo = 16
	stream.Prefix = []byte(prefix)

	stream.Send = func(list *pb.KVList) error {
		for _, v := range list.Kv {
			var evt Event
			err := json.Unmarshal(v.Value, &evt)
			if err != nil {
				return err
			}
			err = processor.Process(evt)
			if err != nil {
				// mark event as retry
				e := d.set(evt.key(eventRetryStatus), evt.value(), evt.expiresAt())
				if e != nil {
					return e
				}
				return err
			} else {
				// mark event as processed
				e := d.set(evt.key(eventProcessedStatus), evt.value(), evt.expiresAt())
				if e != nil {
					return e
				}
			}
			// remove old unprocessed event
			e := d.delete(evt.key(eventUnprocessedStatus))
			if e != nil {
				return e
			}
		}
		return nil
	}

	if err := stream.Orchestrate(context.Background()); err != nil {
		return err
	}
	return nil
}
