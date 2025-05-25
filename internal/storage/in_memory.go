package storage

import (
	"context"
	"errors"
)

type InMemoryRepository struct {
	store map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{store: make(map[string]string)}
}

func (a InMemoryRepository) Get(short string) (string, error) {
	original, ok := a.store[short]
	if !ok {
		return "", errors.New("url not found")
	}
	return original, nil
}

func (a InMemoryRepository) CheckOriginalExist(original string) (string, bool) {
	for key, value := range a.store {
		if value == original {
			return key, true
		}
	}
	return "", false
}

func (a InMemoryRepository) Add(short, original string) error {
	if oldShort, ok := a.CheckOriginalExist(original); ok {
		return NewOriginalExistError(oldShort)
	}
	a.store[short] = original
	return nil
}
func (a InMemoryRepository) AddBatch(ctx context.Context, b map[string]string) error {
	for short, original := range b {
		a.store[short] = original
	}
	return nil
}

func (a InMemoryRepository) Ping(ctx context.Context) error {
	return nil
}
