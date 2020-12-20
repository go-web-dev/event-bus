package server

import (
	"io"

	"github.com/stretchr/testify/mock"
)

type routerMock struct {
	mock.Mock
}

func (m *routerMock) Switch(w io.Writer, r io.Reader) (bool, error) {
	args := m.Called(w, r)
	err, ok := args.Error(1).(error)
	if !ok {
		return args.Get(0).(bool), nil
	}
	return args.Get(0).(bool), err
}

type dbMock struct {
	mock.Mock
}

func (m *dbMock) Close() error {
	return m.Called().Error(0)
}
