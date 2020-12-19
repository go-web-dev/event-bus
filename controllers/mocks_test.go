package controllers

import (
	"encoding/json"

	"github.com/stretchr/testify/mock"

	"github.com/go-web-dev/event-bus/config"
	"github.com/go-web-dev/event-bus/models"
)

type busMock struct {
	mock.Mock
}

func (m *busMock) CreateStream(streamName string) (models.Stream, error) {
	args := m.Called(streamName)
	return args.Get(0).(models.Stream), args.Error(1)
}

func (m *busMock) DeleteStream(streamName string) error {
	args := m.Called(streamName)
	return args.Error(0)
}

func (m *busMock) GetStreamInfo(streamName string) (models.Stream, error) {
	args := m.Called(streamName)
	return args.Get(0).(models.Stream), args.Error(1)
}

func (m *busMock) GetStreamEvents(streamName string) ([]models.Event, error) {
	args := m.Called(streamName)
	return args.Get(0).([]models.Event), args.Error(1)
}

func (m *busMock) WriteEvent(streamName string, event json.RawMessage) error {
	args := m.Called(streamName, event)
	return args.Error(0)
}

func (m *busMock) MarkEvent(streamName string, status uint8) error {
	args := m.Called(streamName, status)
	return args.Error(0)
}

func (m *busMock) ProcessEvents(streamName string, retry bool) ([]models.Event, error) {
	args := m.Called(streamName, retry)
	return args.Get(0).([]models.Event), args.Error(1)
}

type cfgMock struct {
	mock.Mock
}

func (m *cfgMock) GetLoggerLevel() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *cfgMock) GetLoggerOutput() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *cfgMock) GetAuth() config.ClientAuth {
	args := m.Called()
	return args.Get(0).(config.ClientAuth)
}
