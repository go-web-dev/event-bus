package controllers

import (
	"encoding/json"

	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

func (s *controllersSuite) Test_ProcessEvents_Success() {
	streamID := "some-stream-id"
	events := []models.Event{
		{
			ID:        "evt1",
			StreamID:  streamID,
			Status:    0,
			CreatedAt: testTime,
			Body:      json.RawMessage("{}"),
		},
		{
			ID:        "evt2",
			StreamID:  streamID,
			Status:    0,
			CreatedAt: testTime,
			Body:      json.RawMessage("{}"),
		},
	}
	expectedRes := transport.Response{
		Operation: "process_events",
		Status:    true,
		Body: JSON{
			"events": []interface{}{
				JSON{
					"id":         events[0].ID,
					"stream_id":  streamID,
					"status":     float64(events[0].Status),
					"created_at": testTimeStr,
					"body":       JSON{},
				},
				JSON{
					"id":         events[1].ID,
					"stream_id":  streamID,
					"status":     float64(events[1].Status),
					"created_at": testTimeStr,
					"body":       JSON{},
				},
			},
		},
	}
	s.write("process_events", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("ProcessEvents", "some-stream-name", false).
		Return(events, nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_ProcessEvents_EmptyResult() {
	expectedRes := transport.Response{
		Operation: "process_events",
		Status:    true,
		Body: JSON{
			"events": []interface{}{},
		},
	}
	s.write("process_events", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("ProcessEvents", "some-stream-name", false).
		Return([]models.Event{}, nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_ProcessEvents_NilReqError() {
	ctx := JSON{
		"body": []interface{}{
			JSON{
				"name":     "stream_name",
				"type":     "string",
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
	}
	s.testNilRequest("process_events", ctx, fields)
}

func (s *controllersSuite) Test_ProcessEvents_ParseReqError() {
	s.testParseRequest("process_events", `{"stream_name": 1}`)
}

func (s *controllersSuite) Test_ProcessEvents_ServiceError() {
	expectedRes := transport.Response{
		Operation: "process_events",
		Status:    false,
		Reason:    errTest.Error(),
	}
	s.write("process_events", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("ProcessEvents", "some-stream-name", false).
		Return([]models.Event{}, errTest).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(errTest, err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}
