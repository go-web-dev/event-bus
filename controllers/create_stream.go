package controllers

import (
	"fmt"
	"io"

	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"
)

type streamCreator interface {
	CreateStream(streamName string) (services.Stream, error)
}

type createStreamRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

func (router Router) createStream(bus streamCreator) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body createStreamRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendJSON(w, createStreamOperation, err)
			return err
		}

		fmt.Println("stream name", body.StreamName)
		s, err := bus.CreateStream(body.StreamName)
		if err != nil {
			transport.SendJSON(w, createStreamOperation, err)
			return err
		}
		transport.SendJSON(w, createStreamOperation, s)
		return nil
	}
}
