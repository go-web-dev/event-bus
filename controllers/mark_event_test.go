package controllers

import (
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

func (s *controllersSuite) Test_MarkEvent_Success() {
	expectedRes := transport.Response{
		Operation: "mark_event",
		Status:    true,
	}
	s.write("mark_event", `{"event_id": "some-event-id", "status": 2}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("MarkEvent", "some-event-id", uint8(2)).
		Return(nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_MarkEvent_InvalidStatusError() {
	expectedErr := models.InvalidEventStatusError{}
	expectedRes := transport.Response{
		Operation: "mark_event",
		Status:    false,
		Reason:    expectedErr.Error(),
	}
	s.write("mark_event", `{"event_id": "some-event-id", "status": 3}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(expectedErr, err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_MarkEvent_NilReqError() {
	ctx := JSON{
		"body": []interface{}{
			JSON{
				"name":     "event_id",
				"type":     "string",
				"required": true,
			},
			JSON{
				"name":     "status",
				"type":     "uint8",
				"required": true,
			},
		},
	}
	fields := []models.RequiredField{
		{
			Name:     "event_id",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "status",
			Type:     "uint8",
			Required: true,
		},
	}
	s.testNilRequest("mark_event", ctx, fields)
}

func (s *controllersSuite) Test_MarkEvent_ParseReqError() {
	s.testParseRequest("mark_event", `{"event_id": true}`)
}

func (s *controllersSuite) Test_MarkEvent_ServiceError() {
	expectedRes := transport.Response{
		Operation: "mark_event",
		Status:    false,
		Reason:    errTest.Error(),
	}
	s.write("mark_event", `{"event_id": "some-event-id", "status": 2}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("MarkEvent", "some-event-id", uint8(2)).
		Return(errTest).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(errTest, err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}
