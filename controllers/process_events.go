package controllers

import (
	"io"

	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

type eventProcessor interface {
	ProcessEvents(streamName string, retry bool) ([]models.Event, error)
}

type processEventsRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type processEventsResponse struct {
	Events []models.Event `json:"events"`
}

func (router Router) processEvents(bus eventProcessor) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body processEventsRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, processEventsOperation, err)
			return err
		}

		events, err := bus.ProcessEvents(body.StreamName, false)
		if err != nil {
			transport.SendError(w, processEventsOperation, err)
			return err
		}

		res := processEventsResponse{
			Events: events,
		}
		transport.SendJSON(w, processEventsOperation, res)
		return nil
	}
}
