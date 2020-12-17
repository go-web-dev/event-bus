package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Init_Success(t *testing.T) {
	settings := Settings{Output: []string{"stdout"}}
	err := Init(settings)

	assert.NoError(t, err)
	assert.NotNil(t, Logger)
	Logger = nil
}

func Test_Init_Error(t *testing.T) {
	settings := Settings{
		Level: "z_level",
	}
	err := Init(settings)

	assert.EqualError(t, err, `unrecognized level: "z_level"`)
	assert.Nil(t, Logger)
	Logger = nil
}

func Test_Init_BuildConfig(t *testing.T) {
	settings := Settings{Output: []string{"http://www.google.com"}}
	err := Init(settings)

	assert.EqualError(t, err, `couldn't open sink "http://www.google.com": no sink found for scheme "http"`)
	assert.Nil(t, Logger)
	Logger = nil
}
