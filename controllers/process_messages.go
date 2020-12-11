package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"
)

type eventProcessor interface {
	ProcessEvents(streamName string, processor services.EventProcessor) error
}

type processEventsRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type processor struct {
	w io.Writer
}

func (p processor) Process(evt services.Event) error {
	transport.SendJSON(p.w, processEventsOperation, evt)
	return nil
}

func (router Router) processEvents(bus eventProcessor) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body processEventsRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, processEventsOperation, err)
			return err
		}

		p := processor{w: w}
		err = bus.ProcessEvents(body.StreamName, p)
		if err != nil {
			transport.SendError(w, processEventsOperation, err)
			return err
		}
		return nil
	}
}
