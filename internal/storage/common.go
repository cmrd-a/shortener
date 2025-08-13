package storage

import (
	"context"

	"github.com/cmrd-a/shortener/internal/config"
)

// Repository defines the interface for URL storage operations.
type Repository interface {
	Get(context.Context, string) (string, error)
	Add(context.Context, string, string, int64) error
	AddBatch(context.Context, int64, ...StoredURL) error
	Ping(context.Context) error
	GetUserURLs(context.Context, int64) ([]StoredURL, error)
	MarkDeletedUserURLs(context.Context, ...URLForDelete)
}

// MakeRepository creates a Repository instance based on the provided configuration.
// It returns a PostgreSQL repository if DatabaseDSN is provided,
// a file-backed repository if FileStoragePath is provided,
// or an in-memory repository as the default fallback.
func MakeRepository(ctx context.Context, cfg *config.Config) (Repository, error) {
	if cfg.DatabaseDSN != "" {
		return NewPgRepository(ctx, cfg.DatabaseDSN)
	}
	inMemoryRepo := NewInMemoryRepository()
	if cfg.FileStoragePath != "" {
		return NewFileRepository(cfg.FileStoragePath, inMemoryRepo)
	}
	return inMemoryRepo, nil
}
