package transport

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/testutils"
)

type encoderSuite struct {
	testutils.Suite
	loggerEntry zapcore.Entry
}

func (s *encoderSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), &s.loggerEntry)
}

func (s *encoderSuite) TestSendJSON_Success() {
	rw := testutils.NewReadWriter()
	expected := `{"operation":"great_op","status":true, "body": {"name": "steve"}}`

	SendJSON(rw, "great_op", person{Name: "steve"})

	s.JSONEq(expected, s.ReadAll(rw))
	s.Empty(s.ReadAll(rw))
}

func (s *encoderSuite) TestSendJSON_Success_NoBody() {
	rw := testutils.NewReadWriter()
	expected := `{"operation":"great_op","status":true}`

	SendJSON(rw, "great_op", nil)

	s.JSONEq(expected, s.ReadAll(rw))
	s.Empty(s.ReadAll(rw))
}

func (s *encoderSuite) TestSendJSON_Error() {
	rw := testutils.NewReadWriter()

	SendJSON(rw, "great_op", make(chan int))

	s.Equal(s.loggerEntry.Message, "could not encode json response")
	s.Empty(s.ReadAll(rw))
}

func (s *encoderSuite) TestSendError() {
	rw := testutils.NewReadWriter()
	expected := `{"operation":"great_op","status":false, "reason": "some weird reason"}`

	SendError(rw, "great_op", errors.New("some weird reason"))

	s.JSONEq(expected, s.ReadAll(rw))
	s.Empty(s.ReadAll(rw))
}

func (s *encoderSuite) TestSendError_WithContext() {
	rw := testutils.NewReadWriter()
	err := models.OperationRequestError{
		Body: []models.RequiredField{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "field2",
				Type:     "int",
				Required: false,
			},
		},
	}
	expected := `{
		"operation":"great_op",
		"status":false,
		"reason": "missing required fields",
		"context": {
			"body": [
				{"name": "field1", "type": "string", "required": true},
				{"name": "field2", "type": "int", "required": false}
			]
		}
	}`

	SendError(rw, "great_op", err)

	s.JSONEq(expected, s.ReadAll(rw))
	s.Empty(s.ReadAll(rw))
}

func Test_EncoderSuite(t *testing.T) {
	suite.Run(t, new(encoderSuite))
}
