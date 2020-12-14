package controllers

import (
	"encoding/json"
	"io"

	"github.com/chill-and-code/event-bus/transport"
)

type eventWriter interface {
	WriteEvent(streamName string, event json.RawMessage) error
}

type writeEventRequest struct {
	StreamName string          `json:"stream_name" type:"string"`
	Event      json.RawMessage `json:"event" type:"[]byte"`
}

func (router Router) writeEvent(bus eventWriter) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body writeEventRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, writeEventOperation, err)
			return err
		}

		err = bus.WriteEvent(body.StreamName, body.Event)
		if err != nil {
			transport.SendError(w, writeEventOperation, err)
			return err
		}

		transport.SendJSON(w, writeEventOperation, nil)
		return nil
	}
}
