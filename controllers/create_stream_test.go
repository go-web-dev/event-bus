package controllers

import (
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

func (s *controllersSuite) Test_CreateStream_Success() {
	expectedStream := models.Stream{
		ID:        "stream-id",
		Name:      "some-stream-name",
		CreatedAt: testTime,
	}
	expectedRes := transport.Response{
		Operation: "create_stream",
		Status:    true,
		Body: JSON{
			"stream": JSON{
				"id":         expectedStream.ID,
				"name":       expectedStream.Name,
				"created_at": testTimeStr,
			},
		},
	}
	s.write("create_stream", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("CreateStream", "some-stream-name").
		Return(expectedStream, nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_CreateStream_NilReqError() {
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
	s.testNilRequest("create_stream", ctx, fields)
}

func (s *controllersSuite) Test_CreateStream_ParseReqError() {
	s.testParseRequest("create_stream", `{"stream_name": 1}`)
}

func (s *controllersSuite) Test_CreateStream_ServiceError() {
	expectedRes := transport.Response{
		Operation: "create_stream",
		Status:    false,
		Reason:    errTest.Error(),
	}
	s.write("create_stream", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("CreateStream", "some-stream-name").
		Return(models.Stream{}, errTest).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(errTest, err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}
