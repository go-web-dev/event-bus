package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"
)

type retrier struct {
	w io.Writer
}

func (p retrier) Process(evt services.Event) error {
	transport.SendJSON(p.w, retryEventsOperation, evt)
	return nil
}

type retryEventsRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

func (router Router) retryEvents(bus eventProcessor) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body retryEventsRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, retryEventsOperation, err)
			return err
		}

		p := retrier{w: w}
		err = bus.ProcessEvents(body.StreamName, p, true)
		if err != nil {
			transport.SendError(w, retryEventsOperation, err)
			return err
		}
		return nil
	}
}
