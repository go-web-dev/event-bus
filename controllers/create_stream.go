package controllers

import (
	"io"

	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

type streamCreator interface {
	CreateStream(streamName string) (models.Stream, error)
}

type createStreamRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type createStreamResponse struct {
	Stream models.Stream `json:"stream"`
}

func (router Router) createStream(bus streamCreator) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body createStreamRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, createStreamOperation, err)
			return err
		}

		s, err := bus.CreateStream(body.StreamName)
		if err != nil {
			transport.SendError(w, createStreamOperation, err)
			return err
		}

		res := createStreamResponse{
			Stream: s,
		}
		transport.SendJSON(w, createStreamOperation, res)
		return nil
	}
}
