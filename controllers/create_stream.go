package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/transport"
)

type createStreamRequest struct {
	StreamName string `json:"stream_name"`
}

func (router Router) createStream(w io.Writer, r request) error {
	var body createStreamRequest
	err := parseReq(r, body)
	if err != nil {
		transport.SendJSON(w, createStreamOperation, err)
		return err
	}

	s, err := router.bus.Create(body.StreamName)
	if err != nil {
		transport.SendJSON(w, createStreamOperation, err)
		return err
	}
	transport.SendJSON(w, createStreamOperation, s)
	return nil
}
