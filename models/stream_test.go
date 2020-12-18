package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testStream = Stream{
		ID:        "stream-id",
		Name:      "stream-name",
		CreatedAt: testTime,
	}
)

func Test_Stream_Key(t *testing.T) {
	expected := []byte("stream:stream-id")

	bs := testStream.Key()

	assert.Equal(t, expected, bs)
}

func Test_Stream_Value(t *testing.T) {
	expected := `{
		"id":"stream-id",
		"name":"stream-name",
		"created_at":"2020-12-15T05:28:31.490416Z"
	}`

	bs := testStream.Value()

	assert.JSONEq(t, expected, string(bs))
}
