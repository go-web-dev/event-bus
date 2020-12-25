//+build integration

package integration

import (
	"net"

	"github.com/go-web-dev/event-bus/models"
)

type getStreamInfoResponseBody struct {
	Stream models.Stream `json:"stream"`
}

func (s *appSuite) Test_GetStreamInfo_Success() {
	conn := s.newConn()

	s.write(conn, "get_stream_info", `{"stream_name": "s1-name"}`)

	s.assertGetStreamInfoRes(conn)
}

func (s *appSuite) Test_GetStreamInfo_StreamNotFoundError() {
	conn := s.newConn()

	s.write(conn, "get_stream_info", `{"stream_name": "not-found-stream-name"}`)

	var res response
	s.read(conn, &res)
	s.Equal("get_stream_info", res.Operation)
	s.False(res.Status)
	s.Empty(res.Body)
	s.Equal("stream 'not-found-stream-name' not found", res.Reason)
}

func (s *appSuite) Test_GetStreamInfo_Concurrent() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	go func() {
		s.write(conn1, "get_stream_info", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn2, "get_stream_info", `{"stream_name": "s1-name"}`)
	}()
	go func() {
		s.write(conn3, "get_stream_info", `{"stream_name": "s1-name"}`)
	}()

	s.assertGetStreamInfoRes(conn1)
	s.assertGetStreamInfoRes(conn2)
	s.assertGetStreamInfoRes(conn3)
}

func (s *appSuite) assertGetStreamInfoRes(conn net.Conn) {
	expectedResBody := getStreamInfoResponseBody{
		Stream: s1,
	}
	var res response
	s.read(conn, &res)
	s.Equal("get_stream_info", res.Operation)
	s.True(res.Status)
	s.Empty(res.Reason)
	var body getStreamInfoResponseBody
	s.JSONUnmarshal(res.Body, &body)
	s.Equal(expectedResBody, body)
}
