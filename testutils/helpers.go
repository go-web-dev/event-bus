package testutils

import (
	"bytes"
	"github.com/dgraph-io/badger/v2"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

// Suite represents a testify suite with extra methods available
type Suite struct {
	suite.Suite
}

// ReadAll is an extra functionality in the testify suite for reading all from a reader
func (s *Suite) ReadAll(t *testing.T, r io.Reader) string {
	bs, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	return string(bs)
}

// NewReadWriter creates a new instance of testing ReadWriter
func NewReadWriter() *ReadWriter {
	return &ReadWriter{}
}

// ReadWriter represents a read writer for testing purposes
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
func Logger(t *testing.T, entry *zapcore.Entry) *zap.Logger {
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		if entry != nil && e.Level == zap.ErrorLevel {
			*entry = e
		}
		return nil
	}), zap.AddCaller()))
	return logger
}

func NewBadger(t *testing.T) *badger.DB {
	dbOptions := badger.DefaultOptions("")
	dbOptions.Logger = nil
	dbOptions = dbOptions.WithInMemory(true)
	badgerDB, err := badger.Open(dbOptions)
	require.NoError(t, err)
	return badgerDB
}
