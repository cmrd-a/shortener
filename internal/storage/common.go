package storage

import (
	"context"

	"github.com/cmrd-a/shortener/internal/config"
)

type Repository interface {
	Get(string) (string, error)
	Add(string, string) error
	AddBatch(context.Context, map[string]string) error
	Ping(context.Context) error
}

func MakeRepository(cfg *config.Config) (Repository, error) {
	if cfg.DatabaseDSN != "" {
		return NewPgRepository(cfg.DatabaseDSN)
	}
	inMemoryRepo := NewInMemoryRepository()
	if cfg.FileStoragePath != "" {
		return NewFileRepository(cfg.FileStoragePath, inMemoryRepo)
	}
	return inMemoryRepo, nil
}
