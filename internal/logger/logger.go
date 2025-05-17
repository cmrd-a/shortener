package logger

import (
	"log"

	"go.uber.org/zap"
)

func NewLogger(level string) *zap.Logger {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		log.Fatal(err)
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	return zl
}
