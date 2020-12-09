package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/chill-and-code/event-bus/logging"
	"github.com/chill-and-code/event-bus/models"
	"github.com/chill-and-code/event-bus/services"
	"github.com/chill-and-code/event-bus/transport"

	"go.uber.org/zap"
)

const (
	createStreamOperation    = "create_stream"
	deleteStreamOperation    = "delete_stream"
	writeMessageOperation    = "write_message"
	processMessagesOperation = "process_messages"
	exitOperation            = "exit"
	decodeOperation          = "decode_request"
)

type request struct {
	Operation string          `json:"operation"`
	Body      json.RawMessage `json:"body"`
}

type operator func(io.Writer, request) error

type EventBus interface {
	CreateStream(streamName string) (services.Stream, error)
	DeleteStream(streamName string) error
	WriteMessage(streamName string, message json.RawMessage) error
	ProcessMessages(streamName string, processor services.MessageProcessor) error
}

type Router struct {
	bus        EventBus
	operations map[string]operator
}

func NewRouter(b EventBus) Router {
	router := Router{
		bus: b,
	}
	router.operations = map[string]operator{
		createStreamOperation:    router.createStream,
		deleteStreamOperation:    router.deleteStream,
		writeMessageOperation:    router.writeMessage,
		processMessagesOperation: router.processMessages,
	}
	return router
}

func (router Router) Switch(w io.Writer, r io.Reader) (bool, error) {
	var req request
	err := transport.Decode(r, &req)
	if err != nil {
		transport.SendError(w, decodeOperation, models.InvalidJSONError{})
		return false, fmt.Errorf("could not decode request :%s", err)
	}

	ops := make([]string, 0)
	for op := range router.operations {
		ops = append(ops, op)
	}
	notFoundErr := fmt.Errorf(
		"operation must be one of: '%s'",
		strings.Join(ops, "', '"),
	)

	if req.Operation == "" {
		transport.SendError(w, decodeOperation, notFoundErr)
		return false, notFoundErr
	}

	if req.Operation == exitOperation {
		return true, nil
	}

	operation, ok := router.operations[req.Operation]
	if !ok {
		transport.SendError(w, decodeOperation, notFoundErr)
		return false, notFoundErr
	}
	return false, operation(w, req)
}

func parseReq(r request, body interface{}) error {
	decodedFields := transport.DecodeFields(body)
	if r.Body == nil {
		e := models.OperationRequestError{
			Body: decodedFields,
		}
		return e
	}
	err := transport.Decode(bytes.NewReader(r.Body), &body)
	if err != nil {
		logging.Logger.Debug("could not decode request body", zap.Error(err))
		return models.InvalidJSONError{}
	}
	return nil
}
