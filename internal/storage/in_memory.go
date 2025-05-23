package storage

import (
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

func (a InMemoryRepository) Add(short, original string) error {
	a.store[short] = original
	return nil
}
