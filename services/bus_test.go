package services

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
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
	s.db = db{DB: s.newBadger()}
}

func (s *busSuite) SetupTest() {
	s.bus = &Bus{
		db:      s.db,
		streams: map[string]models.Stream{},
	}
}

func (s *busSuite) TearDownTest() {
	s.Require().NoError(s.db.DropAll())
}

func (s *busSuite) Test_Bus_New() {
	expected := &Bus{
		db:      db{s.db},
		streams: map[string]models.Stream{},
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
	txn1 := s.db.set(stream1.Key(), stream1.Value(), 0)
	txn2 := s.db.set(stream2.Key(), stream2.Value(), 0)
	err := s.db.txn(true, txn1, txn2)
	s.Require().NoError(err)

	err = s.bus.Init()

	s.Require().NoError(err)
	s.Equal(expected, s.bus.streams)
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
	txn1 := s.db.set(stream.Key(), stream.Value(), 0)
	txn2 := s.db.set([]byte("stream:zzz-stream-id"), []byte(`{`), 0)
	err := s.db.txn(true, txn1, txn2)
	s.Require().NoError(err)

	err = s.bus.Init()

	s.EqualError(err, "unexpected end of JSON input")
	s.Equal(expected, s.bus.streams)
}

func (s *busSuite) Test_Bus_CreateStream_Success() {
	streamName := "hello-stream"

	stream, err := s.bus.CreateStream(streamName)

	s.Require().NoError(err)
	s.Equal(streamName, stream.Name)
}

func (s *busSuite) Test_Bus_CreateStream_AlreadyExistsError() {
	streamName := "hello-stream"
	s.bus.streams[streamName] = models.Stream{}

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
	txn1 := s.db.set(stream.Key(), stream.Value(), 0)
	txn2 := s.db.set(evt1.Key(0), evt1.Value(), 0)
	txn3 := s.db.set(evt2.Key(0), evt2.Value(), 0)
	s.Require().NoError(s.db.txn(true, txn1, txn2, txn3))
	s.bus.streams[streamName] = stream

	err := s.bus.DeleteStream(streamName)

	s.Require().NoError(err)
	s.Empty(s.bus.streams)
	fn := func(result fetchResult) {
		s.Fail(fmt.Sprintf("did not expect to see result: '%s'", result.key))
	}
	var streamVar models.Stream
	fetchStreamTxn := s.db.fetch("stream:", &streamVar, fn)
	var evtVar models.Event
	fetchEvtTxn := s.db.fetch("event:", &evtVar, fn)
	s.Require().NoError(s.db.txn(false, fetchStreamTxn, fetchEvtTxn))
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
	txn := s.db.set(expected.Key(), expected.Value(), 0)
	s.Require().NoError(s.db.txn(true, txn))
	s.bus.streams[streamName] = expected

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
	stream := models.Stream{
		ID:        streamID,
		Name:      streamName,
		CreatedAt: testTime,
	}
	evt1 := models.Event{
		ID:       "evt1-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   0,
	}
	evt2 := models.Event{
		ID:       "evt2-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   1,
	}
	evt3 := models.Event{
		ID:       "evt3-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   2,
	}
	expectedEvents := []models.Event{evt1, evt2, evt3}
	txn1 := s.db.set(stream.Key(), stream.Value(), 0)
	txn2 := s.db.set(evt1.Key(0), evt1.Value(), 0)
	txn3 := s.db.set(evt2.Key(1), evt2.Value(), 0)
	txn4 := s.db.set(evt3.Key(2), evt3.Value(), 0)
	s.Require().NoError(s.db.txn(true, txn1, txn2, txn3, txn4))
	s.bus.streams[streamName] = stream

	events, err := s.bus.GetStreamEvents(streamName)

	s.Require().NoError(err)
	s.Equal(expectedEvents, events)
}

func (s *busSuite) Test_Bus_GetStreamEvents_EmptyResult() {
	streamName := "stream-name"
	s.bus.streams[streamName] = models.Stream{}

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
	txn := s.db.set(evt.Key(0), []byte("}"), 0)
	s.Require().NoError(s.db.txn(true, txn))
	s.bus.streams[streamName] = models.Stream{ID: streamID}

	stream, err := s.bus.GetStreamEvents(streamName)

	s.Nil(err)
	s.Empty(stream)
	s.Equal("could not unmarshal message", s.loggerEntry.Message)
}

func (s busSuite) newBadger() *badger.DB {
	dbOptions := badger.DefaultOptions("")
	dbOptions.Logger = nil
	dbOptions = dbOptions.WithInMemory(true)
	badgerDB, err := badger.Open(dbOptions)
	s.Require().NoError(err)
	return badgerDB
}

func (s *busSuite) Test_Bus_WriteEvent_Success() {
	streamName := "stream-name"
	streamID := "stream-id"
	evtBody := `{"k": "v"}`
	s.bus.streams[streamName] = models.Stream{ID: streamID}

	err := s.bus.WriteEvent(streamName, json.RawMessage(evtBody))

	s.Require().NoError(err)
	var evtVar models.Event
	txn := s.db.fetch("event:", &evtVar, func(result fetchResult) {
		evt := result.item.(*models.Event)
		s.Equal(streamID, evt.StreamID)
		s.JSONEq(evtBody, string(evt.Body))
	})
	s.Require().NoError(s.db.txn(false, txn))
}

func (s *busSuite) Test_Bus_WriteEvent_StreamNotFoundError() {
	err := s.bus.WriteEvent("stream-name", []byte(`{"k", "v"}`))

	s.EqualError(err, "stream 'stream-name' not found")
	var evtVar models.Event
	txn := s.db.fetch("event:", &evtVar, func(result fetchResult) {
		s.Fail(fmt.Sprintf("did not expect to see result: '%s'", result.key))
	})
	s.Require().NoError(s.db.txn(false, txn))
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
	txn := s.db.set(evt.Key(0), evt.Value(), 0)
	s.Require().NoError(s.db.txn(true, txn))
	s.bus.streams[streamName] = models.Stream{ID: streamID}

	err := s.bus.MarkEvent(evtID, 1)

	s.Require().NoError(err)
	var evtVar models.Event
	txn = s.db.fetch("event:", &evtVar, func(result fetchResult) {
		evt := result.item.(*models.Event)
		s.Equal(streamID, evt.StreamID)
		s.Equal(uint8(1), evt.Status)
	})
	s.Require().NoError(s.db.txn(false, txn))
}

func (s *busSuite) Test_Bus_MarkEvent_EventNotFound() {
	err := s.bus.MarkEvent("evt-id", 1)

	s.EqualError(err, "event 'evt-id' not found")
}

func (s *busSuite) Test_Bus_ProcessEvents_Success() {
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
		Body:     []byte("{}"),
		Status:   0,
	}
	evt2 := models.Event{
		ID:       "evt2-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   1,
	}
	evt3 := models.Event{
		ID:       "evt3-id",
		StreamID: streamID,
		Body:     []byte("{}"),
		Status:   2,
	}
	txn1 := s.db.set(stream.Key(), stream.Value(), 0)
	txn2 := s.db.set(evt1.Key(0), evt1.Value(), 0)
	txn3 := s.db.set(evt2.Key(1), evt2.Value(), 0)
	txn4 := s.db.set(evt3.Key(2), evt3.Value(), 0)
	s.Require().NoError(s.db.txn(true, txn1, txn2, txn3, txn4))
	s.bus.streams[streamName] = stream

	events, err := s.bus.ProcessEvents(streamName, false)

	s.Require().NoError(err)
	s.Equal([]models.Event{evt1}, events)

	events, err = s.bus.ProcessEvents(streamName, true)

	s.Require().NoError(err)
	s.Equal([]models.Event{evt3}, events)
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
	txn := s.db.set(evt.Key(2), evt.Value(), 0)
	s.Require().NoError(s.db.txn(true, txn))
	s.bus.streams[streamName] = models.Stream{ID: streamID}

	events, err := s.bus.ProcessEvents(streamName, false)

	s.Require().NoError(err)
	s.Empty(events)
}

func (s *busSuite) Test_Bus_ProcessEvents_StreamNotFound() {
	events, err := s.bus.ProcessEvents("non-existent-stream", false)

	s.EqualError(err, "stream 'non-existent-stream' not found")
	s.Empty(events)
}

func Test_BusSuite(t *testing.T) {
	suite.Run(t, new(busSuite))
}
