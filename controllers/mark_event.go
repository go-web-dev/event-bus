package controllers

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

type eventMarker interface {
	MarkEvent(eventID string, status uint8) error
}

type markEventRequest struct {
	EventID string `json:"event_id" type:"string"`
	Status  uint8  `json:"status" type:"uint8"`
}

func (router Router) markEvent(bus eventMarker) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body markEventRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, markEventOperation, err)
			return err
		}
		statuses := map[uint8]string{
			models.EventUnprocessedStatus: "unprocessed",
			models.EventProcessedStatus:   "processed",
			models.EventRetryStatus:       "retry",
		}
		if _, ok := statuses[body.Status]; !ok {
			fields := make([]string, 0)
			for k, v := range statuses {
				fields = append(fields, fmt.Sprintf("'%d - %s'", k, v))
			}
			e := fmt.Errorf("status must be one of: %s", strings.Join(fields, ","))
			transport.SendError(w, markEventOperation, e)
			return e
		}

		err = bus.MarkEvent(body.EventID, body.Status)
		if err != nil {
			transport.SendError(w, markEventOperation, err)
			return err
		}

		transport.SendJSON(w, markEventOperation, nil)
		return nil
	}
}
