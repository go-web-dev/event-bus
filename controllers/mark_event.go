package controllers

import (
	"io"

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
		if _, ok := models.AllowedEventStatus[body.Status]; !ok {
			e := models.InvalidEventStatusError{}
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
