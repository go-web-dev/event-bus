//+build integration

package integration

import (
	"fmt"
	"net"
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

	s.assertCreateStreamRes(conn, "test-stream")
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

	s.write(conn, "create_stream", `{"stream_name": "s1-name"}`)

	s.assertCreateStreamErr(conn, "s1-name")
}

func (s *appSuite) Test_CreateStream_Concurrent_Success() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	s.wg.Add(1)
	go func() {
		s.write(conn1, "create_stream", `{"stream_name": "some-stream-1"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		s.write(conn2, "create_stream", `{"stream_name": "some-stream-2"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		s.write(conn3, "create_stream", `{"stream_name": "some-stream-3"}`)
		s.wg.Done()
	}()
	s.wg.Wait()

	s.assertCreateStreamRes(conn1, "some-stream-1")
	s.assertCreateStreamRes(conn2, "some-stream-2")
	s.assertCreateStreamRes(conn3, "some-stream-3")
}

func (s *appSuite) Test_CreateStream_Concurrent_Error() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	s.wg.Add(1)
	go func() {
		s.write(conn1, "create_stream", `{"stream_name": "some-stream"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.write(conn2, "create_stream", `{"stream_name": "some-stream"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.write(conn3, "create_stream", `{"stream_name": "some-stream"}`)
		s.wg.Done()
	}()
	s.wg.Wait()

	s.assertCreateStreamRes(conn1, "some-stream")
	s.assertCreateStreamErr(conn2, "some-stream")
	s.assertCreateStreamErr(conn3, "some-stream")
}

func (s *appSuite) assertCreateStreamRes(conn net.Conn, streamName string) {
	var res response
	var body createStreamResponseBody
	s.read(conn, &res)
	s.True(res.Status)
	s.Equal("create_stream", res.Operation)

	s.JSONUnmarshal(res.Body, &body)
	s.Equal(streamName, body.Stream.Name)
	s.InDelta(time.Now().UTC().Unix(), body.Stream.CreatedAt.Unix(), 1)
	_, err := uuid.Parse(body.Stream.ID)
	s.Nil(err)

	dbStream, err := s.dbGet(body.Stream.Key())
	s.Require().NoError(err)
	s.JSONEq(s.JSONMarshal(body.Stream), dbStream)
}

func (s *appSuite) assertCreateStreamErr(conn net.Conn, streamName string) {
	var res response
	s.read(conn, &res)
	s.False(res.Status)
	s.Equal("create_stream", res.Operation)
	s.Equal(fmt.Sprintf("stream: '%s' already exists", streamName), res.Reason)
}
