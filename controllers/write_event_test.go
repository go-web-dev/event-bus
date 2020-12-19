package controllers

import (
	"encoding/json"

	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

func (s *controllersSuite) Test_WriteEvent_Success() {
	expectedRes := transport.Response{
		Operation: "write_event",
		Status:    true,
	}
	evtBody := json.RawMessage(`{"k": "v"}`)
	s.write("write_event", `{"stream_name": "some-stream-name", "event": {"k": "v"}}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("WriteEvent", "some-stream-name", evtBody).
		Return(nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_WriteEvent_NilReqError() {
	ctx := JSON{
		"body": []interface{}{
			JSON{
				"name":     "stream_name",
				"type":     "string",
				"required": true,
			},
			JSON{
				"name":     "event",
				"type":     "[]byte",
				"required": true,
			},
		},
	}
	fields := []models.RequiredField{
		{
			Name:     "stream_name",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "event",
			Type:     "[]byte",
			Required: true,
		},
	}
	s.testNilRequest("write_event", ctx, fields)
}

func (s *controllersSuite) Test_WriteEvent_ParseReqError() {
	s.testParseRequest("write_event", `{"stream_name": 1}`)
}

func (s *controllersSuite) Test_WriteEvent_ServiceError() {
	expectedRes := transport.Response{
		Operation: "write_event",
		Status:    false,
		Reason:    errTest.Error(),
	}
	s.write("write_event", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("WriteEvent", "some-stream-name", json.RawMessage(nil)).
		Return(errTest).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(errTest, err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}
