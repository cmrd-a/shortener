package storage

import (
	"context"

	"github.com/cmrd-a/shortener/internal/config"
)

type Repository interface {
	Get(context.Context, string) (string, error)
	Add(context.Context, string, string, int64) error
	AddBatch(context.Context, int64, map[string]string) error
	Ping(context.Context) error
	GetUserURLs(context.Context, int64) ([]StoredURL, error)
	MarkDeletedUserURLs(context.Context, ...URLForDelete)
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
