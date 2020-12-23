//+build integration

package integration

import (
	"time"

	"github.com/go-web-dev/event-bus/models"
	"github.com/google/uuid"
)

type createStreamResponseBody struct {
	Stream models.Stream `json:"stream"`
}

func (s *appSuite) Test_CreateStream_Success() {
	conn := s.newConn()

	s.write(conn, "create_stream", `{"stream_name": "test-stream"}`)

	var res response
	s.read(conn, &res)
	var body createStreamResponseBody
	s.JSONUnmarshal(res.Body, &body)
	s.Equal("create_stream", res.Operation)
	s.True(res.Status)
	s.Equal("test-stream", body.Stream.Name)
	s.InDelta(time.Now().UTC().Unix(), body.Stream.CreatedAt.Unix(), 1)
	_, err := uuid.Parse(body.Stream.ID)
	s.Nil(err)
	s.JSONEq(s.JSONMarshal(body.Stream), s.dbGet(body.Stream.Key()))
}

func (s *appSuite) Test_CreateStream_MissingBodyError() {
	expectedErrCtx := JSON{
		"body": []interface{}{
			JSON{
				"name":     "stream_name",
				"required": true,
				"type":     "string",
			},
		},
	}
	conn := s.newConn()

	s.write(conn, "create_stream", ``)

	var res response
	s.read(conn, &res)
	s.Equal("create_stream", res.Operation)
	s.False(res.Status)
	s.Equal("missing required fields", res.Reason)
	s.Equal(expectedErrCtx, res.Context)
}

func (s *appSuite) Test_CreateStream_StreamAlreadyExistsError() {
	conn := s.newConn()

	s.write(conn, "create_stream", `{"stream_name": "existing-stream-name"}`)

	var res response
	s.read(conn, &res)
	s.Equal("create_streams", res.Operation)
	s.False(res.Status)
	s.Equal("stream: 'existing-stream-name' already exists", res.Reason)
	s.Nil(res.Context)
}
