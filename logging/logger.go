package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger represents the application logger
var Logger *zap.Logger

type Settings struct {
	Level string
	Output []string
}

// Init initializes application logger
func Init(settings Settings) error {
	loggerCfg := zap.NewProductionConfig()
	level := zapcore.InfoLevel
	err := level.Set(settings.Level)
	if err != nil {
		return err
	}
	loggerCfg.Level = zap.NewAtomicLevelAt(level)
	loggerCfg.OutputPaths = settings.Output

	l, err := loggerCfg.Build()
	if err != nil {
		return err
	}
	Logger = l
	l.Sync()
	return nil
}
