package controllers

import (
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/transport"
)

func (s *controllersSuite) Test_GetStreamInfo_Success() {
	expectedStream := models.Stream{
		ID:        "some-stream-id",
		Name:      "some-stream-name",
		CreatedAt: testTime,
	}
	expectedRes := transport.Response{
		Operation: "get_stream_info",
		Status:    true,
		Body: JSON{
			"stream": JSON{
				"id":         expectedStream.ID,
				"name":       expectedStream.Name,
				"created_at": testTimeStr,
			},
		},
	}
	s.write("get_stream_info", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("GetStreamInfo", "some-stream-name").
		Return(expectedStream, nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}

func (s *controllersSuite) Test_GetStreamInfo_NilReqError() {
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
	s.testNilRequest("get_stream_info", ctx, fields)
}

func (s *controllersSuite) Test_GetStreamInfo_ParseReqError() {
	s.testParseRequest("get_stream_info", `{"stream_name": 1}`)
}

func (s *controllersSuite) Test_GetStreamInfo_ServiceError() {
	expectedRes := transport.Response{
		Operation: "get_stream_info",
		Status:    false,
		Reason:    errTest.Error(),
	}
	s.write("get_stream_info", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("GetStreamInfo", "some-stream-name").
		Return(models.Stream{}, errTest).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(errTest, err)
	s.False(exited)
	s.Equal(expectedRes, s.read())
}
