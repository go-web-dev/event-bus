package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"
)

type streamInfoGetter interface {
	GetStreamInfo(streamName string) (services.Stream, error)
}

type getStreamInfoRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

func (router Router) getStreamInfo(bus streamInfoGetter) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body getStreamInfoRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendJSON(w, getStreamInfoOperation, err)
			return err
		}

		s, err := bus.GetStreamInfo(body.StreamName)
		if err != nil {
			transport.SendJSON(w, getStreamInfoOperation, err)
			return err
		}
		transport.SendJSON(w, getStreamInfoOperation, s)
		return nil
	}
}