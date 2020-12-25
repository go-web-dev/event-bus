//+build integration

package integration

import (
	"fmt"
	"net"
	"time"

	"github.com/go-web-dev/event-bus/models"
)

func (s *appSuite) Test_DeleteStream_Success() {
	conn := s.newConn()

	s.write(conn, "delete_stream", `{"stream_name": "s1-name"}`)

	s.assertDeleteStreamRes(conn, s1)
}

func (s *appSuite) Test_DeleteStream_StreamNotFoundError() {
	someStream := models.Stream{
		ID:   "stream-id",
		Name: "stream-name",
		CreatedAt: testTime,
	}
	s.Require().NoError(s.db.DropAll())
	s.dbSet(someStream.Key(), someStream.Value())
	conn := s.newConn()
	expectedDBStream := `{
		"id":"stream-id",
		"name":"stream-name",
		"created_at":"2020-12-15T05:28:31.490416Z"
	}`

	s.write(conn, "delete_stream", `{"stream_name": "not-found-stream-name"}`)

	var res response
	s.read(conn, &res)
	s.Equal("delete_stream", res.Operation)
	s.False(res.Status)
	s.Empty(res.Body)
	s.Empty(res.Context)
	s.Equal("stream 'not-found-stream-name' not found", res.Reason)

	dbStream, err := s.dbGet(someStream.Key())
	s.JSONEq(expectedDBStream, dbStream)
	s.Nil(err)
}

func (s *appSuite) Test_DeleteStream_Concurrent_Success() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	s.wg.Add(1)
	go func() {
		s.write(conn1, "delete_stream", `{"stream_name": "s1-name"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.write(conn2, "delete_stream", `{"stream_name": "s2-name"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.write(conn3, "delete_stream", `{"stream_name": "s3-name"}`)
		s.wg.Done()
	}()
	s.wg.Wait()

	s.assertDeleteStreamRes(conn1, s1)
	s.assertDeleteStreamRes(conn2, s2)
	s.assertDeleteStreamRes(conn3, s3)
}

func (s *appSuite) Test_DeleteStream_Concurrent_Error() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	s.wg.Add(1)
	go func() {
		s.write(conn1, "delete_stream", `{"stream_name": "s1-name"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.write(conn2, "delete_stream", `{"stream_name": "s1-name"}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.write(conn3, "delete_stream", `{"stream_name": "s1-name"}`)
		s.wg.Done()
	}()
	s.wg.Wait()

	s.assertDeleteStreamRes(conn1, s1)
	s.assertDeleteStreamErr(conn2, s1)
	s.assertDeleteStreamErr(conn3, s1)
}

func (s *appSuite) assertDeleteStreamRes(conn net.Conn, stream models.Stream) {
	var res response
	s.read(conn, &res)
	s.Equal("delete_stream", res.Operation)
	s.True(res.Status)
	s.Empty(res.Body)

	dbStream, err := s.dbGet(stream.Key())
	s.Empty(dbStream)
	s.EqualError(err, "Key not found")
}

func (s *appSuite) assertDeleteStreamErr(conn net.Conn, stream models.Stream) {
	var res response
	s.read(conn, &res)
	s.Equal("delete_stream", res.Operation)
	s.False(res.Status)
	s.Equal(fmt.Sprintf("stream '%s' not found", stream.Name), res.Reason)

	dbStream, err := s.dbGet(stream.Key())
	s.Equal("", dbStream)
	s.EqualError(err, "Key not found")
}
