package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/transport"
)

type deleteStreamRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

func (router Router) deleteStream(w io.Writer, r request) error {
	var body deleteStreamRequest
	err := parseReq(r, body)
	if err != nil {
		transport.SendJSON(w, deleteStreamOperation, err)
		return err
	}

	err = router.bus.DeleteStream(body.StreamName)
	if err != nil {
		transport.SendJSON(w, deleteStreamOperation, err)
		return err
	}
	transport.SendJSON(w, deleteStreamOperation, nil)
	return nil
}
