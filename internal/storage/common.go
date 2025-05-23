package storage

import "github.com/cmrd-a/shortener/internal/config"

type Repository interface {
	Get(short string) (original string, err error)
	Add(short, original string) error
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
