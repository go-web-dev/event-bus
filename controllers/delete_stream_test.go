package controllers

import (
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

func (s *controllersSuite) Test_DeleteStream_Success() {
	expectedRes := transport.Response{
		Operation: "delete_stream",
		Status:    true,
	}
	s.write("delete_stream", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("DeleteStream", "some-stream-name").
		Return(nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_DeleteStream_NilReqError() {
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
	s.testNilRequest("delete_stream", ctx, fields)
}

func (s *controllersSuite) Test_DeleteStream_ParseReqError() {
	s.testParseRequest("delete_stream", `{"stream_name": 1}`)
}

func (s *controllersSuite) Test_DeleteStream_ServiceError() {
	expectedRes := transport.Response{
		Operation: "delete_stream",
		Status:    false,
		Reason:    errTest.Error(),
	}
	s.write("delete_stream", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("DeleteStream", "some-stream-name").
		Return(errTest).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(errTest, err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}
