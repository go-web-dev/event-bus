package config

import (
	"github.com/spf13/viper"

	"github.com/go-web-dev/event-bus/logging"
)

// Configuration fields
const (
	auth = "auth"
)

// ClientCredentials represents the credentials every client has to make a request with
type ClientCredentials struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// ClientAuth represents all allowed client to make requests to Event Bus
type ClientAuth map[string]ClientCredentials

// NewManager creates a new configuration manager
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

// Manager represents the configuration manager
type Manager struct {
	viper      *viper.Viper
	clientAuth ClientAuth
}

func (m *Manager) loadClientAuth() error {
	return m.viper.UnmarshalKey(auth, &m.clientAuth)
}

// GetAuth gets all allowed clients authentication details from config file
func (m *Manager) GetAuth() ClientAuth {
	return m.clientAuth
}
