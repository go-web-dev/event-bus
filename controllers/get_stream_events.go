package controllers

import (
	"io"

	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

type streamEventsGetter interface {
	GetStreamEvents(streamName string) ([]models.Event, error)
}

type getStreamEventsRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type getStreamEventsResponse struct {
	Events []models.Event `json:"events"`
}

func (router Router) getStreamEvents(bus streamEventsGetter) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body getStreamEventsRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, getStreamEventsOperation, err)
			return err
		}

		events, err := bus.GetStreamEvents(body.StreamName)
		if err != nil {
			transport.SendError(w, getStreamEventsOperation, err)
			return err
		}

		res := getStreamEventsResponse{
			Events: events,
		}
		transport.SendJSON(w, getStreamEventsOperation, res)
		return nil
	}
}
