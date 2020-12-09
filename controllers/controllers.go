package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/chill-and-code/event-bus/models"
	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"
)

const (
	createStreamOperation = "create_stream"
	deleteStreamOperation = "delete_stream"
	exitOperation         = "exit"
	decodeOperation       = "decode_request"
)

type request struct {
	Operation string          `json:"operation"`
	Body      json.RawMessage `json:"body"`
}

type operator func(io.Writer, request) error

type Router struct {
	bus        services.Bus
	operations map[string]operator
}

func NewRouter(b services.Bus) Router {
	router := Router{
		bus: b,
	}
	router.operations = map[string]operator{
		createStreamOperation: router.createStream,
		deleteStreamOperation: router.deleteStream,
	}
	return router
}

func (router Router) Switch(w io.Writer, r io.Reader) (bool, error) {
	var req request
	err := transport.Decode(r, &req)
	if err != nil {
		transport.SendError(w, decodeOperation, err)
		return false, fmt.Errorf("could not decode request :%s", err)
	}

	if req.Operation == exitOperation {
		return true, nil
	}

	operation, ok := router.operations[req.Operation]
	if !ok {
		e := fmt.Errorf("operation '%s' not found", req.Operation)
		transport.SendError(w, decodeOperation, e)
		return false, e
	}
	return false, operation(w, req)
}

func parseReq(r request, body interface{}) error {
	if r.Body == nil {
		e := models.OperationRequestError{
			Fields: transport.DecodeFields(body),
		}
		return e
	}
	err := transport.Decode(bytes.NewReader(r.Body), &body)
	if err != nil {
		return err
	}
	return nil
}
