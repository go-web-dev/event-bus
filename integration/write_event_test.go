//+build integration

package integration

import (
	"net"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/go-web-dev/event-bus/models"
)

func (s *appSuite) Test_WriteEvent_Success() {
	conn := s.newConn()

	s.write(conn, "write_event", `{"stream_name": "s3-name", "event": {"e1-field": "e1-value"}}`)

	s.assertWriteEventRes(conn)
	events := s.sortedEvents(s.dbFetch("event:s3-id:0"))
	s.Len(events, 1)
	s.assertDBEvent(events[0], "s3-id", `{"e1-field": "e1-value"}`)
}

func (s *appSuite) Test_WriteEvent_StreamNotFoundError() {
	conn := s.newConn()

	s.write(conn, "write_event", `{"stream_name": "nonexistent-stream-name", "event": {}}`)

	var res response
	s.read(conn, &res)
	s.Equal("write_event", res.Operation)
	s.False(res.Status)
	s.Equal("stream 'nonexistent-stream-name' not found", res.Reason)
	s.Empty(res.Body)
	values := s.dbFetch("event:nonexistent-stream-id:0")
	s.Len(values, 0)
}

func (s *appSuite) Test_WriteEvent_Concurrent() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	conn3 := s.newConn()

	s.wg.Add(1)
	go func() {
		s.write(conn1, "write_event", `{"stream_name": "s3-name", "event": {"e1-field": "e1-value"}}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(50*time.Millisecond)
		s.write(conn2, "write_event", `{"stream_name": "s3-name", "event": {"e2-field": "e2-value"}}`)
		s.wg.Done()
	}()
	s.wg.Wait()
	s.wg.Add(1)
	go func() {
		time.Sleep(100*time.Millisecond)
		s.write(conn3, "write_event", `{"stream_name": "s3-name", "event": {"e3-field": "e3-value"}}`)
		s.wg.Done()
	}()
	s.wg.Wait()

	s.assertWriteEventRes(conn1)
	s.assertWriteEventRes(conn2)
	s.assertWriteEventRes(conn3)
	events := s.sortedEvents(s.dbFetch("event:s3-id:0"))
	s.Len(events, 3)
	s.assertDBEvent(events[0], "s3-id", `{"e1-field": "e1-value"}`)
	s.assertDBEvent(events[1], "s3-id", `{"e2-field": "e2-value"}`)
	s.assertDBEvent(events[2], "s3-id", `{"e3-field": "e3-value"}`)
}

func (s *appSuite) assertWriteEventRes(conn net.Conn) {
	var res response
	s.read(conn, &res)
	s.Equal("write_event", res.Operation)
	s.True(res.Status)
	s.Empty(res.Reason)
	s.Empty(res.Body)
}

func (s *appSuite) assertDBEvent(evt models.Event, streamID, evtBody string) {
	s.Equal(streamID, evt.StreamID)
	s.JSONEq(evtBody, string(evt.Body))
	s.Equal(uint8(0), evt.Status)
	s.InDelta(time.Now().UTC().Unix(), evt.CreatedAt.Unix(), 1)
	_, err := uuid.Parse(evt.ID)
	s.Require().NoError(err)
}

func (s *appSuite) sortedEvents(values [][]byte) []models.Event {
	events := make([]models.Event, 0)
	for _, val := range values {
		var evt models.Event
		s.JSONUnmarshal(val, &evt)
		events = append(events, evt)
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].CreatedAt.UnixNano() < events[j].CreatedAt.UnixNano()
	})
	return events
}
