package config

import (
	"github.com/spf13/viper"

	"github.com/chill-and-code/event-bus/logging"
)

const (
	auth = "auth"
)

type ClientCredentials struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type ClientAuth map[string]ClientCredentials

func NewManager(filename string) (*Manager, error) {
	logger := logging.Logger
	m := viper.New()
	m.SetConfigFile(filename)
	err := m.ReadInConfig()
	if err != nil {
		logger.Error("could not read config file")
		return nil, err
	}
	manager := &Manager{
		viper: m,
	}
	err = manager.loadClientAuth()
	if err != nil {
		logger.Error("could not load client auth")
		return nil, err
	}
	return manager, nil
}

type Manager struct {
	viper      *viper.Viper
	clientAuth ClientAuth
}

func (m *Manager) loadClientAuth() error {
	return m.viper.UnmarshalKey(auth, &m.clientAuth)
}

func (m *Manager) GetAuth() ClientAuth {
	return m.clientAuth
}
