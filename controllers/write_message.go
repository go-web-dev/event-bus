package controllers

import (
	"encoding/json"
	"io"

	"github.com/chill-and-code/event-bus/transport"
)

type writeMessageRequest struct {
	StreamName string          `json:"stream_name" type:"string"`
	Message    json.RawMessage `json:"message" type:"[]byte"`
}

func (router Router) writeMessage(w io.Writer, r request) error {
	var body writeMessageRequest
	err := parseReq(r, body)
	if err != nil {
		transport.SendJSON(w, writeMessageOperation, err)
		return err
	}

	err = router.bus.WriteMessage(body.StreamName, body.Message)
	if err != nil {
		transport.SendJSON(w, writeMessageOperation, err)
		return err
	}
	transport.SendJSON(w, writeMessageOperation, nil)
	return nil
}
