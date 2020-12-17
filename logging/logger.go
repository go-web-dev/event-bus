package logging

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

// Init initializes application logger
func Init() error {
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	l, err := loggerCfg.Build()
	Logger = l
	if err != nil {
		return err
	}
	l.Sync()
	return nil
}
