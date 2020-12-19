package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testTimeStr = "2020-12-15T05:28:31.490416Z"
)

var (
	testTime, _ = time.Parse(time.RFC3339, testTimeStr)
	testEvt     = Event{
		ID:        "event-id",
		StreamID:  "stream-id",
		Status:    0,
		CreatedAt: testTime,
		Body:      []byte(`{"key": "value"}`),
	}
)

func Test_Event_Key(t *testing.T) {
	expected := []byte("event:stream-id:0:2020-12-15T05:28:31Z:event-id")

	bs := testEvt.Key(0)

	assert.Equal(t, expected, bs)
}

func Test_Event_Value(t *testing.T) {
	expected := `{
		"id":"event-id",
		"stream_id":"stream-id",
		"status":0,
		"created_at":"2020-12-15T05:28:31.490416Z",
		"body":{"key":"value"}
	}`

	bs := testEvt.Value()

	assert.JSONEq(t, expected, string(bs))
}

func Test_Event_ExpiresAt(t *testing.T) {
	expiresAt := testEvt.ExpiresAt()

	assert.Equal(t, uint64(testTime.Add(720*time.Hour).Unix()), expiresAt)
}
