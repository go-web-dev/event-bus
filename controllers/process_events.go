package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"
)

type eventProcessor interface {
	ProcessEvents(streamName string, retry bool) ([]services.Event, error)
}

type processEventsRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type processEventsResponse struct {
	Events []services.Event `json:"events"`
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
