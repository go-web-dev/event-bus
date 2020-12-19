package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/testutils"
)

var (
	errTest = errors.New("some test error")
)

type dbSuite struct {
	testutils.Suite
	db          db
	loggerEntry zapcore.Entry
}

func (s *dbSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), &s.loggerEntry)
	s.db = db{DB: testutils.NewBadger(s.T())}
}

func (s *dbSuite) TearDownTest() {
	s.Require().NoError(s.db.DropAll())
}

func (s *dbSuite) Test_db_txn_Success() {
	txn1 := func(txn *badger.Txn) error {
		s.Require().NoError(txn.Set([]byte("key1"), []byte("value1")))
		return nil
	}
	txn2 := func(txn *badger.Txn) error {
		s.Require().NoError(txn.Set([]byte("key2"), []byte("value2")))
		return nil
	}
	txn3 := func(txn *badger.Txn) error {
		s.Require().NoError(txn.Set([]byte("key3"), []byte("value3")))
		return nil
	}

	err := s.db.txn(true, txn1, txn2, txn3)

	s.Require().NoError(err)
	k1, v1, _ := s.get("key1")
	s.Equal("key1", k1)
	s.Equal("value1", v1)
	k2, v2, _ := s.get("key2")
	s.Equal("key2", k2)
	s.Equal("value2", v2)
	k3, v3, _ := s.get("key3")
	s.Equal("key3", k3)
	s.Equal("value3", v3)
}

func (s *dbSuite) Test_db_txn_Error() {
	txn1 := func(txn *badger.Txn) error {
		s.Require().NoError(txn.Set([]byte("key1"), []byte("value1")))
		return nil
	}
	txn2 := func(txn *badger.Txn) error {
		s.Require().NoError(txn.Set([]byte("key2"), []byte("value2")))
		return errTest
	}
	txn3 := func(txn *badger.Txn) error {
		s.Require().NoError(txn.Set([]byte("key3"), []byte("value3")))
		return nil
	}

	err := s.db.txn(true, txn1, txn2, txn3)

	s.Equal(errTest, err)
	_, _, found := s.get("key1")
	s.False(found)
	_, _, found = s.get("key2")
	s.False(found)
	_, _, found = s.get("key3")
	s.False(found)
}

func (s *dbSuite) Test_db_set_Success() {
	txn := s.db.NewTransaction(true)
	err := s.db.set([]byte("key"), []byte("value"), 0)(txn)
	s.Require().NoError(err)

	err = txn.Commit()

	s.Require().NoError(err)
	k1, v1, found := s.get("key")
	s.Require().True(found)
	s.Equal("key", k1)
	s.Equal("value", v1)
}

func (s *dbSuite) Test_db_set_WithExpiration() {
	txn := s.db.NewTransaction(true)
	expiresAt := uint64(time.Now().Add(100 * time.Millisecond).Unix())

	err := s.db.set([]byte("key"), []byte("value"), expiresAt)(txn)
	s.Require().NoError(err)
	err = txn.Commit()

	s.Require().NoError(err)
	time.Sleep(125 * time.Millisecond)
	_, _, found := s.get("key")
	s.False(found)
}

func (s *dbSuite) Test_db_delete_Success() {
	setTxn := s.db.NewTransaction(true)
	key1, key2 := []byte("key1"), []byte("key2")
	val1, val2 := []byte("value1"), []byte("value2")
	s.Require().NoError(setTxn.Set(key1, val1))
	s.Require().NoError(setTxn.Set(key2, val2))
	s.Require().NoError(setTxn.Commit())
	deleteTxn := s.db.NewTransaction(true)

	err := s.db.delete(key1, key2)(deleteTxn)
	s.Require().NoError(err)
	err = deleteTxn.Commit()

	s.Require().NoError(err)
	_, _, found := s.get("key1")
	s.False(found)
	_, _, found = s.get("key2")
	s.False(found)
}

func (s *dbSuite) Test_db_delete_Error() {
	setTxn := s.db.NewTransaction(true)
	key, val := []byte("key"), []byte("value")
	s.Require().NoError(setTxn.Set(key, val))
	s.Require().NoError(setTxn.Commit())
	deleteTxn := s.db.NewTransaction(false)

	err := s.db.delete(key)(deleteTxn)
	s.EqualError(err, "No sets or deletes are allowed in a read-only transaction")
	err = deleteTxn.Commit()

	s.Require().NoError(err)
	_, _, found := s.get("key")
	s.True(found)
}

func (s *dbSuite) Test_db_fetch_Success() {
	evt1 := models.Event{
		ID:       "evt1-id",
		StreamID: "stream-id",
		Body:     json.RawMessage("{}"),
	}
	evt2 := models.Event{
		ID:       "evt2-id",
		StreamID: "stream-id",
		Body:     json.RawMessage("{}"),
	}
	setTxn := s.db.NewTransaction(true)
	s.Require().NoError(setTxn.Set(evt1.Key(0), evt1.Value()))
	s.Require().NoError(setTxn.Set(evt2.Key(0), evt2.Value()))
	s.Require().NoError(setTxn.Commit())

	fetchTxn := s.db.NewTransaction(false)
	var evtVar models.Event
	events := make([]models.Event, 0)
	err := s.db.fetch("event:", &evtVar, func(result fetchResult) {
		evt := result.item.(*models.Event)
		events = append(events, *evt)
	})(fetchTxn)

	s.Require().NoError(err)
	s.Equal([]models.Event{evt1, evt2}, events)
}

func (s *dbSuite) Test_db_fetch_Error() {
	setTxn := s.db.NewTransaction(true)
	s.Require().NoError(setTxn.Set([]byte("event:broken"), []byte("}")))
	s.Require().NoError(setTxn.Commit())

	fetchTxn := s.db.NewTransaction(false)
	var evtVar models.Event
	events := make([]models.Event, 0)
	err := s.db.fetch("event:", &evtVar, func(result fetchResult) {
		evt := result.item.(*models.Event)
		events = append(events, *evt)
	})(fetchTxn)

	s.EqualError(err, "invalid character '}' looking for beginning of value")
	s.Empty(events)
}

func (s *dbSuite) Test_db_streamEvents_Success() {
	evt1 := models.Event{
		ID:       "evt1-id",
		StreamID: "stream-id",
		Body:     json.RawMessage("{}"),
	}
	evt2 := models.Event{
		ID:       "evt2-id",
		StreamID: "stream-id",
		Body:     json.RawMessage("{}"),
	}
	setTxn := s.db.NewTransaction(true)
	s.Require().NoError(setTxn.Set(evt1.Key(0), evt1.Value()))
	s.Require().NoError(setTxn.Set(evt2.Key(0), evt2.Value()))
	s.Require().NoError(setTxn.Commit())

	events, err := s.db.streamEvents("event", nil)

	s.Require().NoError(err)
	s.Equal([]models.Event{evt1, evt2}, events)
}

func (s *dbSuite) Test_db_streamEvents_ChooseFunc() {
	evt1 := models.Event{
		ID:       "evt1-id",
		StreamID: "stream-id",
		Body:     json.RawMessage("{}"),
	}
	evt2 := models.Event{
		ID:       "evt2-id",
		StreamID: "stream-id",
		Body:     json.RawMessage("{}"),
	}
	setTxn := s.db.NewTransaction(true)
	s.Require().NoError(setTxn.Set(evt1.Key(0), evt1.Value()))
	s.Require().NoError(setTxn.Set(evt2.Key(0), evt2.Value()))
	s.Require().NoError(setTxn.Commit())

	events, err := s.db.streamEvents("event", func(item *badger.Item) bool {
		return bytes.Compare(evt1.Key(0), item.Key()) == 0
	})

	s.Require().NoError(err)
	s.Equal([]models.Event{evt1}, events)
}

func (s *dbSuite) Test_db_streamEvents_Error() {
	setTxn := s.db.NewTransaction(true)
	s.Require().NoError(setTxn.Set([]byte("event:broken"), []byte("}")))
	s.Require().NoError(setTxn.Commit())

	events, err := s.db.streamEvents("event", nil)

	s.Require().NoError(err)
	s.Empty(events)
	s.Equal("could not unmarshal message", s.loggerEntry.Message)
}

func (s *dbSuite) get(key string) (string, string, bool) {
	txn := s.db.NewTransaction(false)
	item, err := txn.Get([]byte(key))
	if err != nil {
		return "", "", false
	}
	return string(item.Key()), s.value(item), true
}

func (s *dbSuite) value(item *badger.Item) string {
	var value string
	s.Require().NoError(item.Value(func(val []byte) error {
		value = string(val)
		return nil
	}))
	return value
}

func Test_dbSuite(t *testing.T) {
	suite.Run(t, new(dbSuite))
}
