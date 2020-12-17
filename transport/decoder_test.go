package transport

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-web-dev/event-bus/models"
)

type person struct {
	Name string `json:"name" type:"string"`
	Age  int    `json:"age,omitempty" type:"int"`
}

func TestDecode_Success(t *testing.T) {
	expected := person{
		Name: "steve",
	}
	var p person

	err := Decode(bytes.NewReader([]byte(`{"name": "steve"}`)), &p)

	assert.NoError(t, err)
	assert.Equal(t, expected, p)
}

func TestDecode_Error(t *testing.T) {
	var p person

	err := Decode(bytes.NewReader([]byte(`}`)), &p)

	assert.EqualError(t, err, "invalid character '}' looking for beginning of value")
	assert.Empty(t, p)
}

func TestDecodeFields(t *testing.T) {
	expected := []models.RequiredField{
		{
			Name:     "name",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "age",
			Type:     "int",
			Required: false,
		},
	}

	requiredFields := DecodeFields(person{})

	assert.Equal(t, expected, requiredFields)
}
