//+build integration

package integration

import (
	"net"

	"github.com/go-web-dev/event-bus/models"
)

func (s *appSuite) Test_RetryEvents_Success() {
	conn := s.newConn()

	s.write(conn, "retry_events", `{"stream_name": "s1-name"}`)

	s.assertRetryEventsRes(conn)
}

func (s *appSuite) Test_RetryEvents_StreamNotFoundError() {
	conn := s.newConn()

	s.write(conn, "retry_events", `{"stream_name": "nonexistent-stream-name"}`)

	var res response
	s.read(conn, &res)
	s.Equal("retry_events", res.Operation)
	s.False(res.Status)
	s.Equal("stream 'nonexistent-stream-name' not found", res.Reason)
	s.Empty(res.Body)
	s.Empty(res.Context)
}

func (s *appSuite) Test_RetryEvents_Concurrent() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	go func() {
		s.write(conn1, "retry_events", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn2, "retry_events", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn3, "retry_events", `{"stream_name": "s1-name"}`)
	}()

	s.assertRetryEventsRes(conn1)
	s.assertRetryEventsRes(conn2)
	s.assertRetryEventsRes(conn3)
}

func (s *appSuite) assertRetryEventsRes(conn net.Conn) {
	expectedBody := []models.Event{
		s1evt4,
	}
	var res response
	s.read(conn, &res)
	s.Equal("retry_events", res.Operation)
	s.True(res.Status)
	s.Empty(res.Reason)
	var body getStreamEventsResponseBody
	s.JSONUnmarshal(res.Body, &body)
	s.Equal(expectedBody, body.Events)
}
