//+build integration

package integration

import (
	"net"

	"github.com/go-web-dev/event-bus/models"
)

func (s *appSuite) Test_ProcessEvents_Success() {
	conn := s.newConn()

	s.write(conn, "process_events", `{"stream_name": "s1-name"}`)

	s.assertProcessEventsRes(conn)
}

func (s *appSuite) Test_ProcessEvents_StreamNotFoundError() {
	conn := s.newConn()

	s.write(conn, "process_events", `{"stream_name": "nonexistent-stream-name"}`)

	var res response
	s.read(conn, &res)
	s.Equal("process_events", res.Operation)
	s.False(res.Status)
	s.Equal("stream 'nonexistent-stream-name' not found", res.Reason)
	s.Empty(res.Body)
	s.Empty(res.Context)
}

func (s *appSuite) Test_ProcessEvents_Concurrent() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	go func() {
		s.write(conn1, "process_events", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn2, "process_events", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn3, "process_events", `{"stream_name": "s1-name"}`)
	}()

	s.assertProcessEventsRes(conn1)
	s.assertProcessEventsRes(conn2)
	s.assertProcessEventsRes(conn3)
}

func (s *appSuite) assertProcessEventsRes(conn net.Conn) {
	expectedBody := []models.Event{
		s1evt1,
		s1evt2,
	}
	var res response
	s.read(conn, &res)
	s.Equal("process_events", res.Operation)
	s.True(res.Status)
	s.Empty(res.Reason)
	var body getStreamEventsResponseBody
	s.JSONUnmarshal(res.Body, &body)
	s.Equal(expectedBody, body.Events)
}
