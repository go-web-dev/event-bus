package controllers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/go-web-dev/event-bus/config"
	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/testutils"
)

type controllersSuite struct {
	testutils.Suite
	bus    *busMock
	cfg    *cfgMock
	rw     *testutils.ReadWriter
	auth   config.ClientAuth
	router Router
}

func (s *controllersSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), nil)
	s.bus = new(busMock)
	s.cfg = new(cfgMock)
	s.router = NewRouter(s.bus, s.cfg)
}

func (s *controllersSuite) SetupTest() {
	s.rw = testutils.NewReadWriter()
	s.auth = config.ClientAuth{
		"root": config.ClientCredentials{
			ClientID:     "root_client_id",
			ClientSecret: "root_client_secret",
		},
	}
	s.bus.Test(s.T())
	s.cfg.Test(s.T())
}

func (s *controllersSuite) TearDownTest() {
	s.bus.AssertExpectations(s.T())
	s.cfg.AssertExpectations(s.T())
}

func (s *controllersSuite) Test_NewRouter() {
	router := NewRouter(s.bus, s.cfg)

	s.Equal(s.cfg, router.cfg)
	s.Len(router.operations, 10)
}

func (s *controllersSuite) Test_Switch_Success() {
	s.write("create_stream", `{"stream_name": "some-stream-name"}`)
	s.cfg.
		On("GetAuth").
		Return(s.auth).
		Once()
	s.bus.
		On("CreateStream", "some-stream-name").
		Return(models.Stream{}, nil).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
}

func (s *controllersSuite) Test_Switch_Health() {
	s.write("health", "")

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.False(exited)
}

func (s *controllersSuite) Test_Switch_Exit() {
	s.write("exit", "")

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Require().NoError(err)
	s.True(exited)
}

func (s *controllersSuite) Test_Switch_WrongCommandError() {
	s.write("abc", "")

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(models.OperationNotFoundError{}, err)
	s.False(exited)
}

func (s *controllersSuite) Test_Switch_AuthError() {
	s.write("create_stream", "")
	s.cfg.
		On("GetAuth").
		Return(config.ClientAuth{}).
		Once()

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(models.AuthError{}, err)
	s.False(exited)
}

func (s *controllersSuite) Test_Switch_DecodeError() {
	expectedErr := models.Error{
		Message: "invalid character ',' looking for beginning of object key string",
	}
	s.write("create_stream", "{")

	exited, err := s.router.Switch(s.rw, s.rw)

	s.Equal(expectedErr, err)
	s.False(exited)
}

func (s *controllersSuite) write(operation, body string) {
	auth := `{"client_id": "root_client_id", "client_secret": "root_client_secret"}`
	req := fmt.Sprintf(`{"operation": "%s"`, operation)
	if body != "" {
		req += fmt.Sprintf(`, "body": %s`, body)
	}
	if s.auth != nil {
		req += fmt.Sprintf(`, "auth": %s`, auth)
	}
	req += "}"

	_, err := s.rw.Write([]byte(req))
	s.Require().NoError(err)
}

func Test_ControllersSuite(t *testing.T) {
	suite.Run(t, new(controllersSuite))
}
