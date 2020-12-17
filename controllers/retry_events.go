package controllers

import (
	"io"

	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

type retryEventsRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type retryEventsResponse struct {
	Events []models.Event `json:"events"`
}

func (router Router) retryEvents(bus eventProcessor) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body retryEventsRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, retryEventsOperation, err)
			return err
		}

		events, err := bus.ProcessEvents(body.StreamName, true)
		if err != nil {
			transport.SendError(w, retryEventsOperation, err)
			return err
		}

		res := retryEventsResponse{
			Events: events,
		}
		transport.SendJSON(w, retryEventsOperation, res)
		return nil
	}
}
