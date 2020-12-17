package testutils

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

type Suite struct {
	suite.Suite
}

func (s *Suite) ReadAll(t *testing.T, r io.Reader) string {
	if rc, ok := r.(io.Closer); ok {
		defer func() { require.NoError(t, rc.Close()) }()
	}
	bs, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	return string(bs)
}

func NewReadWriter() *ReadWriter {
	return &ReadWriter{}
}

type ReadWriter struct {
	buff bytes.Buffer
}

func (rw *ReadWriter) Read(bs []byte) (int, error) {
	return rw.buff.Read(bs)
}

func (rw *ReadWriter) Write(bs []byte) (int, error) {
	return rw.buff.Write(bs)
}

// Logger creates a new test logger
func Logger(t *testing.T, err *error) *zap.Logger {
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		if err != nil && e.Level == zap.ErrorLevel {
			*err = errors.New(e.Message)
		}
		return nil
	})))
	return logger
}
