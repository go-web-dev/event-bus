package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"
)

type processMessagesRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type messageProcessor struct {
	w io.Writer
}

func (p messageProcessor) Process(msg services.Message) error {
	transport.SendJSON(p.w, processMessagesOperation, msg)
	return nil
}

func (router Router) processMessages(w io.Writer, r request) error {
	var body processMessagesRequest
	err := parseReq(r, body)
	if err != nil {
		transport.SendError(w, processMessagesOperation, err)
		return err
	}

	processor := messageProcessor{w: w}
	err = router.bus.ProcessMessages(body.StreamName, processor)
	if err != nil {
		transport.SendError(w, processMessagesOperation, err)
		return err
	}
	return nil
}
