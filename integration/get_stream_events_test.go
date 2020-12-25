//+build integration

package integration

import (
	"net"

	"github.com/go-web-dev/event-bus/models"
)

type getStreamEventsResponseBody struct {
	Events []models.Event `json:"events"`
}

func (s *appSuite) Test_GetStreamEvents_Success() {
	conn := s.newConn()

	s.write(conn, "get_stream_events", `{"stream_name": "s1-name"}`)

	s.assertGetStreamEventsRes(conn)
}

func (s *appSuite) Test_GetStreamEvents_StreamNotFoundError() {
	conn := s.newConn()

	s.write(conn, "get_stream_events", `{"stream_name": "nonexistent-stream-name"}`)

	var res response
	s.read(conn, &res)
	s.Equal("get_stream_events", res.Operation)
	s.False(res.Status)
	s.Equal("stream 'nonexistent-stream-name' not found", res.Reason)
	s.Empty(res.Body)
}

func (s *appSuite) Test_GetStreamEvents_Concurrent() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	go func() {
		s.write(conn1, "get_stream_events", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn2, "get_stream_events", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn3, "get_stream_events", `{"stream_name": "s1-name"}`)
	}()

	s.assertGetStreamEventsRes(conn1)
	s.assertGetStreamEventsRes(conn2)
	s.assertGetStreamEventsRes(conn3)
}

func (s *appSuite) assertGetStreamEventsRes(conn net.Conn) {
	expectedBody := []models.Event{
		s1evt1,
		s1evt2,
		s1evt3,
		s1evt4,
	}
	var res response
	s.read(conn, &res)
	s.Equal("get_stream_events", res.Operation)
	s.True(res.Status)
	s.Empty(res.Reason)
	var body getStreamEventsResponseBody
	s.JSONUnmarshal(res.Body, &body)
	s.Equal(expectedBody, body.Events)
}
