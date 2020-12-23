package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/testutils"
)

const (
	testTimeStr = "2020-12-15T05:28:31.490416Z"
)

var (
	testTime, _ = time.Parse(time.RFC3339, testTimeStr)
)

type busSuite struct {
	testutils.Suite
	db          db
	bus         *Bus
	loggerEntry zapcore.Entry
}

func (s *busSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), &s.loggerEntry)
	s.db = db{DB: testutils.NewBadger(s.T())}
}

func (s *busSuite) SetupTest() {
	s.bus = &Bus{
		db:      s.db,
		Streams: map[string]models.Stream{},
	}
}

func (s *busSuite) TearDownTest() {
	s.Require().NoError(s.db.DropAll())
}

func (s *busSuite) Test_Bus_New() {
	expected := &Bus{
		db:      db{s.db},
		Streams: map[string]models.Stream{},
	}

	bus := NewBus(s.db)

	s.Require().NotNil(bus)
	s.Equal(expected, bus)
}

func (s *busSuite) Test_Bus_Init_Success() {
	stream1 := models.Stream{
		ID:        "stream-id1",
		Name:      "stream-name1",
		CreatedAt: testTime,
	}
	stream2 := models.Stream{
		ID:        "stream-id2",
		Name:      "stream-name2",
		CreatedAt: testTime,
	}
	expected := map[string]models.Stream{
		stream1.Name: stream1,
		stream2.Name: stream2,
	}
	s.setStreams(stream1, stream2)

	err := s.bus.Init()

	s.Require().NoError(err)
	s.Equal(expected, s.bus.Streams)
}

func (s *busSuite) Test_Bus_Init_Error() {
	stream := models.Stream{
		ID:        "stream-id",
		Name:      "stream-name",
		CreatedAt: testTime,
	}
	expected := map[string]models.Stream{
		stream.Name: stream,
	}
	s.setStreams(stream)
	s.setRaw([]byte("stream:zzz-stream-id"), []byte(`{`))

	err := s.bus.Init()

	s.EqualError(err, "unexpected end of JSON input")
	s.Equal(expected, s.bus.Streams)
}

func (s *busSuite) Test_Bus_CreateStream_Success() {
	streamName := "hello-stream"

	stream, err := s.bus.CreateStream(streamName)

	s.Require().NoError(err)
	s.Equal(streamName, stream.Name)
}

func (s *busSuite) Test_Bus_CreateStream_AlreadyExistsError() {
	streamName := "hello-stream"
	s.bus.Streams[streamName] = models.Stream{}

	stream, err := s.bus.CreateStream(streamName)

	s.EqualError(err, "stream: 'hello-stream' already exists")
	s.Empty(stream)
}

func (s *busSuite) Test_Bus_DeleteStream_Success() {
	streamName := "stream-name"
	streamID := "stream-id"
	stream := models.Stream{
		ID:        streamID,
		Name:      streamName,
		CreatedAt: testTime,
	}
	evt1 := models.Event{
		ID:       "evt1-id",
		StreamID: streamID,
	}
	evt2 := models.Event{
		ID:       "evt2-id",
		StreamID: streamID,
	}
	s.setStreams(stream)
	s.setEvents(evt1, evt2)
	s.bus.Streams[streamName] = stream

	err := s.bus.DeleteStream(streamName)

	s.Require().NoError(err)
	s.Empty(s.bus.Streams)
	s.Empty(s.fetchEvents())
}

func (s *busSuite) Test_Bus_DeleteStream_StreamNotFoundError() {
	err := s.bus.DeleteStream("stream-name")

	s.EqualError(err, "stream 'stream-name' not found")
}

func (s *busSuite) Test_Bus_GetStreamInfo_Success() {
	streamName := "stream-name"
	expected := models.Stream{
		ID:        "stream-id",
		Name:      streamName,
		CreatedAt: testTime,
	}
	s.setStreams(expected)
	s.bus.Streams[streamName] = expected

	stream, err := s.bus.GetStreamInfo(streamName)

	s.Require().NoError(err)
	s.Equal(expected, stream)
}

func (s *busSuite) Test_Bus_GetStreamInfo_StreamNotFoundError() {
	stream, err := s.bus.GetStreamInfo("stream-name")

	s.EqualError(err, "stream 'stream-name' not found")
	s.Empty(stream)
}

func (s *busSuite) Test_Bus_GetStreamEvents_Success() {
	streamName := "stream-name"
	streamID := "stream-id"
	anotherStreamID := "another-stream-id"
	stream1 := models.Stream{
		ID:        streamID,
		Name:      streamName,
		CreatedAt: testTime,
	}
	stream2 := models.Stream{
		ID:        anotherStreamID,
		Name:      "another-stream",
		CreatedAt: testTime,
	}
	s1Evt1 := models.Event{
		ID:       "s1-evt1-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   0,
	}
	s1Evt2 := models.Event{
		ID:       "s1-evt2-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   1,
	}
	s1Evt3 := models.Event{
		ID:       "s1-evt3-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   2,
	}
	s2Evt1 := models.Event{
		ID:       "s2-evt1-id",
		StreamID: anotherStreamID,
		Body:     []byte("{}"),
		Status:   0,
	}
	expectedEvents := []models.Event{s1Evt1, s1Evt2, s1Evt3}
	s.setStreams(stream1, stream2)
	s.setEvents(s1Evt1, s1Evt2, s1Evt3, s2Evt1)
	s.bus.Streams[streamName] = stream1

	events, err := s.bus.GetStreamEvents(streamName)

	s.Require().NoError(err)
	s.Equal(expectedEvents, events)
}

func (s *busSuite) Test_Bus_GetStreamEvents_EmptyResult() {
	streamName := "stream-name"
	s.bus.Streams[streamName] = models.Stream{}

	stream, err := s.bus.GetStreamEvents(streamName)

	s.Nil(err)
	s.Empty(stream)
}

func (s *busSuite) Test_Bus_GetStreamEvents_StreamNotFoundError() {
	stream, err := s.bus.GetStreamEvents("stream-name")

	s.EqualError(err, "stream 'stream-name' not found")
	s.Empty(stream)
}

func (s *busSuite) Test_Bus_GetStreamEvents_UnmarshalJSONError() {
	streamName := "stream-name"
	streamID := "stream-id"
	evt := models.Event{
		ID:       "evt1-id",
		StreamID: streamID,
	}
	s.setRaw(evt.Key(0), []byte("}"))
	s.bus.Streams[streamName] = models.Stream{ID: streamID}

	stream, err := s.bus.GetStreamEvents(streamName)

	s.Nil(err)
	s.Empty(stream)
	s.Equal("could not unmarshal message", s.loggerEntry.Message)
}

func (s *busSuite) Test_Bus_WriteEvent_Success() {
	streamName := "stream-name"
	streamID := "stream-id"
	evtBody := `{"k": "v"}`
	s.bus.Streams[streamName] = models.Stream{ID: streamID}

	err := s.bus.WriteEvent(streamName, json.RawMessage(evtBody))

	s.Require().NoError(err)
	events := s.fetchEvents()
	s.Require().Len(events, 1)
	s.Equal(streamID, events[0].StreamID)
	s.JSONEq(evtBody, string(events[0].Body))
}

func (s *busSuite) Test_Bus_WriteEvent_StreamNotFoundError() {
	err := s.bus.WriteEvent("stream-name", []byte(`{"k", "v"}`))

	s.EqualError(err, "stream 'stream-name' not found")
	s.Empty(s.fetchEvents())
}

func (s *busSuite) Test_Bus_MarkEvent_Success() {
	streamName := "stream-name"
	streamID := "stream-id"
	evtID := "evt1-id"
	evt := models.Event{
		ID:       evtID,
		StreamID: streamID,
		Body:     []byte("{}"),
	}
	s.setEvents(evt)
	s.bus.Streams[streamName] = models.Stream{ID: streamID}

	err := s.bus.MarkEvent(evtID, 1)

	s.Require().NoError(err)
	events := s.fetchEvents()
	s.Require().Len(events, 1)
	s.Equal(streamID, events[0].StreamID)
	s.Equal(uint8(1), events[0].Status)
}

func (s *busSuite) Test_Bus_MarkEvent_EventNotFound() {
	err := s.bus.MarkEvent("evt-id", 1)

	s.EqualError(err, "event 'evt-id' not found")
}

func (s *busSuite) Test_Bus_ProcessEvents_Success() {
	streamName := "stream-name"
	streamID := "stream-id"
	anotherStreamID := "another-stream-id"
	stream1 := models.Stream{
		ID:        streamID,
		Name:      streamName,
		CreatedAt: testTime,
	}
	stream2 := models.Stream{
		ID:        anotherStreamID,
		Name:      "another-stream-name",
		CreatedAt: testTime,
	}
	s1Evt1 := models.Event{
		ID:       "s1-evt1-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   0,
	}
	s1Evt2 := models.Event{
		ID:       "s1-evt2-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   1,
	}
	s1Evt3 := models.Event{
		ID:       "s1-evt3-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   2,
	}
	s2Evt1 := models.Event{
		ID:       "s2-evt1-id",
		StreamID: anotherStreamID,
		Body:     []byte("{}"),
		Status:   0,
	}
	s.setStreams(stream1, stream2)
	s.setEvents(s1Evt1, s1Evt2, s1Evt3, s2Evt1)
	s.bus.Streams[stream1.Name] = stream1
	s.bus.Streams[stream2.Name] = stream2

	events, err := s.bus.ProcessEvents(streamName, false)

	s.Require().NoError(err)
	s.Equal([]models.Event{s1Evt1}, events)

	events, err = s.bus.ProcessEvents(streamName, true)

	s.Require().NoError(err)
	s.Equal([]models.Event{s1Evt3}, events)
}

func (s *busSuite) Test_Bus_ProcessEvents_EmptyResult() {
	streamName := "stream-name"
	streamID := "stream-id"
	evt := models.Event{
		ID:       "evt-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   2,
	}
	s.setEvents(evt)
	s.bus.Streams[streamName] = models.Stream{ID: streamID}

	events, err := s.bus.ProcessEvents(streamName, false)

	s.Require().NoError(err)
	s.Empty(events)
}

func (s *busSuite) Test_Bus_ProcessEvents_StreamNotFound() {
	events, err := s.bus.ProcessEvents("non-existent-stream", false)

	s.EqualError(err, "stream 'non-existent-stream' not found")
	s.Empty(events)
}

func (s *busSuite) setRaw(key, value []byte) {
	s.Require().NoError(s.db.txn(true, s.db.set(key, value, 0)))
}

func (s *busSuite) setStreams(streams ...models.Stream) {
	for _, stream := range streams {
		s.setRaw(stream.Key(), stream.Value())
	}
}

func (s *busSuite) setEvents(events ...models.Event) {
	for _, evt := range events {
		s.setRaw(evt.Key(evt.Status), evt.Value())
	}
}

func (s *busSuite) fetchEvents() []models.Event {
	events := make([]models.Event, 0)
	var evtVar models.Event
	txn := s.db.fetch("event:", &evtVar, func(result fetchResult) {
		evt := result.item.(*models.Event)
		events = append(events, *evt)
	})
	s.Require().NoError(s.db.txn(false, txn))
	return events
}

func Test_BusSuite(t *testing.T) {
	suite.Run(t, new(busSuite))
}
