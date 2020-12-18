package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/testutils"
)

const (
	testCfgFile = "config_test.yaml"
)

type configSuite struct {
	testutils.Suite
	manager *Manager
}

func (s *configSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), nil)
	m, err := NewManager(testCfgFile)
	s.Require().NoError(err)
	s.Require().NotNil(m)
	s.manager = m
}

func (s *configSuite) Test_NewManager_Success() {
	m, err := NewManager(testCfgFile)

	s.Require().NoError(err)
	s.Require().NotNil(m)
}

func (s *configSuite) Test_NewManager_ReadConfigError() {
	m, err := NewManager("does_not_exist")

	s.EqualError(err, `Unsupported Config Type ""`)
	s.Nil(m)
}

func (s *configSuite) Test_NewManager_LoadAuth_UnmarshalError() {
	tmpFileName := "tmp_cfg.yaml"
	file, err := os.Create(tmpFileName)
	s.Require().NoError(err)
	_, err = file.Write([]byte("auth:\n  some_client:\n  client_id: 2\n"))
	s.Require().NoError(err)

	m, err := NewManager(tmpFileName)

	s.EqualError(err, "1 error(s) decoding:\n\n* '[client_id]' expected a map, got 'int'")
	s.Nil(m)
	s.Require().NoError(os.Remove(tmpFileName))
}

func (s *configSuite) Test_NewManager_LoadAuth_EmptyError() {
	tmpFileName := "tmp_cfg.yaml"
	file, err := os.Create(tmpFileName)
	s.Require().NoError(err)
	_, err = file.Write([]byte("auth:\n"))
	s.Require().NoError(err)

	m, err := NewManager(tmpFileName)

	s.EqualError(err, "'auth' field is required")
	s.Nil(m)
	s.Require().NoError(os.Remove(tmpFileName))
}

func (s *configSuite) Test_Manager_setDefaults() {
	m := &Manager{viper: viper.New()}
	m.setDefaults()
	testCases := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{name: "GetLoggerLevel", actual: m.GetLoggerLevel(), expected: "debug"},
		{name: "GetLoggerOutput", actual: m.GetLoggerOutput(), expected: []string{"stdout"}},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			s.Equal(testCase.expected, testCase.actual)
		})
	}
}

func (s *configSuite) Test_GetLoggerLevel() {
	level := s.manager.GetLoggerLevel()

	s.Equal("debug", level)
}

func (s *configSuite) Test_GetLoggerOutput() {
	out := s.manager.GetLoggerOutput()

	s.Equal([]string{"stdout", "app.log"}, out)
}

func (s *configSuite) Test_GetAuth() {
	expected := ClientAuth{
		"client1": ClientCredentials{
			ClientID:     "client1_id",
			ClientSecret: "client1_secret",
		},
		"client2": ClientCredentials{
			ClientID:     "client2_id",
			ClientSecret: "client2_secret",
		},
	}

	clientAuth := s.manager.GetAuth()

	s.Equal(expected, clientAuth)
}

func Test_ConfigSuite(t *testing.T) {
	suite.Run(t, new(configSuite))
}
