package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/chill-and-code/event-bus/logging"
	"github.com/chill-and-code/event-bus/testutils"
)

const (
	testCfgFile = "config_test.yaml"
)

type configSuite struct {
	suite.Suite
	manager *Manager
}

func (s *configSuite) SetupSuite() {
	logger, err := testutils.NewLogger()
	s.Require().NoError(err)
	logging.Logger = logger
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

func (s *configSuite) Test_NewManager_LoadAuthError() {
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

func TestConfig(t *testing.T) {
	suite.Run(t, new(configSuite))
}
