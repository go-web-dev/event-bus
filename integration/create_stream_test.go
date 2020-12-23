//+build integration

package integration

import (
	"net"
	"sync"
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
	s.Equal("create_stream", res.Operation)
	s.False(res.Status)
	s.Equal("stream: 'existing-stream-name' already exists", res.Reason)
	s.Nil(res.Context)
}

func (s *appSuite) Test_CreateStream_Concurrent_Success() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.write(conn1, "create_stream", `{"stream_name": "some-stream-1"}`)
		wg.Done()
	}()
	wg.Wait()
	wg.Add(1)
	go func() {
		s.write(conn2, "create_stream", `{"stream_name": "some-stream-2"}`)
		wg.Done()
	}()
	wg.Wait()
	wg.Add(1)
	go func() {
		s.write(conn3, "create_stream", `{"stream_name": "some-stream-3"}`)
		wg.Done()
	}()
	wg.Wait()

	s.assertCreateStreamRes(conn1, "some-stream-1")
	s.assertCreateStreamRes(conn2, "some-stream-2")
	s.assertCreateStreamRes(conn3, "some-stream-3")
}

func (s *appSuite) Test_CreateStream_Concurrent_Error() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.write(conn1, "create_stream", `{"stream_name": "some-stream"}`)
		wg.Done()
	}()
	wg.Wait()
	wg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.write(conn2, "create_stream", `{"stream_name": "some-stream"}`)
		wg.Done()
	}()
	wg.Wait()
	wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.write(conn3, "create_stream", `{"stream_name": "some-stream"}`)
		wg.Done()
	}()
	wg.Wait()

	s.assertCreateStreamRes(conn1, "some-stream")
	s.assertCreateStreamErr(conn2)
	s.assertCreateStreamErr(conn3)
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

	s.JSONEq(s.JSONMarshal(body.Stream), s.dbGet(body.Stream.Key()))
}

func (s *appSuite) assertCreateStreamErr(conn net.Conn) {
	var res response
	s.read(conn, &res)
	s.False(res.Status)
	s.Equal("create_stream", res.Operation)
	s.Equal("stream: 'some-stream' already exists", res.Reason)
}
