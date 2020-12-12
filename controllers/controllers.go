package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
	"github.com/chill-and-code/event-bus/models"
	"github.com/chill-and-code/event-bus/transport"
)

const (
	healthOperation         = "health"
	createStreamOperation   = "create_stream"
	deleteStreamOperation   = "delete_stream"
	getStreamInfoOperation  = "get_stream_info"
	writeEventOperation     = "write_event"
	processEventsOperation  = "process_events"
	retryEventsOperation    = "retry_events"
	snapshotDBOperation     = "snapshot_db"
	exitOperation           = "exit"
	decodeOperation         = "decode_request"
)

type request struct {
	Operation string          `json:"operation"`
	Body      json.RawMessage `json:"body"`
}

type EventBus interface {
	streamCreator
	streamDeleter
	streamInfoGetter
	eventWriter
	eventProcessor
	dbSnapshotter
}

type operator func(io.Writer, request) error

type Router struct {
	operations map[string]operator
}

func NewRouter(b EventBus) Router {
	router := Router{}
	router.operations = map[string]operator{
		createStreamOperation:  router.createStream(b),
		deleteStreamOperation:  router.deleteStream(b),
		getStreamInfoOperation: router.getStreamInfo(b),
		writeEventOperation:    router.writeEvent(b),
		processEventsOperation: router.processEvents(b),
		snapshotDBOperation:    router.snapshotDB(b),
		healthOperation: func(w io.Writer, _ request) error {
			transport.SendJSON(w, healthOperation, nil)
			return nil
		},
		exitOperation: func(w io.Writer, _ request) error {
			transport.SendJSON(w, exitOperation, nil)
			return nil
		},
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
	operation, ok := router.operations[req.Operation]
	if !ok {
		transport.SendError(w, decodeOperation, notFoundErr)
		return false, notFoundErr
	}
	if req.Operation == exitOperation {
		return true, operation(w, req)
	}
	return false, operation(w, req)
}

func parseReq(r request, body interface{}) error {
	if r.Body == nil {
		decodedFields := transport.DecodeFields(reflect.Indirect(reflect.ValueOf(body)).Interface())
		e := models.OperationRequestError{
			Body: decodedFields,
		}
		return e
	}
	err := transport.Decode(bytes.NewReader(r.Body), body)
	if err != nil {
		logging.Logger.Debug("could not decode request body", zap.Error(err))
		return models.InvalidJSONError{}
	}
	return nil
}
