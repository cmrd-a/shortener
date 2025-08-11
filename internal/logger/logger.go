package logger

import (
	"go.uber.org/zap"
)

// NewLogger creates a new zap logger with the specified log level.
func NewLogger(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return zl, nil
}
