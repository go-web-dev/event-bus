package testutils

import (
	"go.uber.org/zap"
)

// NewLogger creates a new development logger.
// To be used only for testing purposes
func NewLogger() (*zap.Logger, error) {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := loggerCfg.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
