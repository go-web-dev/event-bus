package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/transport"
)

type streamDeleter interface {
	DeleteStream(streamName string) error
}

type deleteStreamRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

func (router Router) deleteStream(bus streamDeleter) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body deleteStreamRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, deleteStreamOperation, err)
			return err
		}

		err = bus.DeleteStream(body.StreamName)
		if err != nil {
			transport.SendError(w, deleteStreamOperation, err)
			return err
		}

		transport.SendJSON(w, deleteStreamOperation, nil)
		return nil
	}
}
