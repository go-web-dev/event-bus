package config

import (
	"errors"

	"github.com/spf13/viper"
)

// Configuration fields
const (
	auth         = "auth"
	loggerLevel  = "logger.level"
	loggerOutput = "logger.output"
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
	m := viper.New()
	m.SetConfigFile(filename)
	err := m.ReadInConfig()
	if err != nil {
		return nil, err
	}
	manager := &Manager{
		viper: m,
	}
	err = manager.loadClientAuth()
	if err != nil {
		return nil, err
	}
	manager.setDefaults()

	return manager, nil
}

// Manager represents the configuration manager
type Manager struct {
	viper      *viper.Viper
	clientAuth ClientAuth
}

func (m *Manager) loadClientAuth() error {
	err := m.viper.UnmarshalKey(auth, &m.clientAuth)
	if err != nil {
		return err
	}
	if len(m.clientAuth) == 0 {
		return errors.New("'auth' field is required")
	}
	return nil
}

func (m *Manager) setDefaults() {
	m.viper.SetDefault(loggerLevel, "debug")
	m.viper.SetDefault(loggerOutput, "stdout")
}

// GetLoggerOutput gets logger output from config file
func (m *Manager) GetLoggerOutput() []string {
	return m.viper.GetStringSlice(loggerOutput)
}

// GetLoggerLevel gets logger atomic level level
func (m *Manager) GetLoggerLevel() string {
	return m.viper.GetString(loggerLevel)
}

// GetAuth gets all allowed clients authentication details from config file
func (m *Manager) GetAuth() ClientAuth {
	return m.clientAuth
}
