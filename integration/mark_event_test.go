//+build integration

package integration

import (
	"fmt"
	"net"
	"time"
)

func (s *appSuite) Test_MarkEvent_Success() {
	conn := s.newConn()

	s.write(conn, "mark_event", `{"event_id": "evt1-id", "status": 1}`)

	s.assertMarkEventRes(conn)
	s.assertDBMarkedEvent(1, true)
	s.assertDBMarkedEvent(0, false)
}

func (s *appSuite) Test_MarkEvent_EventNotFoundError() {
	conn := s.newConn()

	s.write(conn, "mark_event", `{"event_id": "nonexisting-evt-id", "status": 1}`)

	var res response
	s.read(conn, &res)
	s.Equal("mark_event", res.Operation)
	s.False(res.Status)
	s.Empty(res.Body)
	s.Equal("event 'nonexisting-evt-id' not found", res.Reason)
}

func (s *appSuite) Test_MarkEvent_InvalidStatusError() {
	conn := s.newConn()

	s.write(conn, "mark_event", `{"event_id": "evt1-id", "status": 5}`)

	s.assertMarkEventErr(conn)
	s.assertDBMarkedEvent(0, true)
	s.assertDBMarkedEvent(5, false)
}

func (s *appSuite) Test_MarkEvent_Concurrent_Success() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	s.wg.Add(1)
	go func() {
		s.write(conn1, "mark_event", `{"event_id": "evt1-id", "status": 2}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.write(conn2, "mark_event", `{"event_id": "evt1-id", "status": 0}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.write(conn3, "mark_event", `{"event_id": "evt1-id", "status": 1}`)
		s.wg.Done()
	}()
	s.wg.Wait()

	s.assertMarkEventRes(conn1)
	s.assertMarkEventRes(conn2)
	s.assertMarkEventRes(conn3)
	s.assertDBMarkedEvent(1, true)
	s.assertDBMarkedEvent(2, false)
	s.assertDBMarkedEvent(0, false)
}

func (s *appSuite) Test_MarkEvent_Concurrent_Error() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	s.wg.Add(1)
	go func() {
		s.write(conn1, "mark_event", `{"event_id": "evt1-id", "status": 2}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.write(conn2, "mark_event", `{"event_id": "evt1-id", "status": 5}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.write(conn3, "mark_event", `{"event_id": "evt1-id", "status": 1}`)
		s.wg.Done()
	}()
	s.wg.Wait()

	s.assertMarkEventRes(conn1)
	s.assertMarkEventErr(conn2)
	s.assertMarkEventRes(conn3)
	s.assertDBMarkedEvent(1, true)
	s.assertDBMarkedEvent(0, false)
	s.assertDBMarkedEvent(5, false)
}

func (s *appSuite) assertMarkEventRes(conn net.Conn) {
	var res response
	s.read(conn, &res)
	s.Equal("mark_event", res.Operation)
	s.True(res.Status)
	s.Empty(res.Body)
	s.Empty(res.Reason)
}

func (s *appSuite) assertMarkEventErr(conn net.Conn) {
	var res response
	s.read(conn, &res)
	s.Equal("mark_event", res.Operation)
	s.False(res.Status)
	s.Empty(res.Body)
	s.Equal("status must be one of: 0 - unprocessed', '1 - processed', '2 - retry", res.Reason)
}

func (s *appSuite) assertDBMarkedEvent(status uint8, found bool) {
	expectedEvt := fmt.Sprintf(`{
		"id":"evt1-id",
		"stream_id":"s1-id",
		"status":%d,
		"created_at":"2020-12-15T05:28:31.490416Z",
		"body":{"f1":"v1"}
	}`, status)
	if found {
		evt, err := s.dbGet(s1evt1.Key(status))
		s.Require().NoError(err)
		s.JSONEq(expectedEvt, evt)
		return
	}
	evt, err := s.dbGet(s1evt1.Key(status))
	s.EqualError(err, "Key not found")
	s.Empty(evt)
}
