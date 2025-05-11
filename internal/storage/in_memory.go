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

func (a InMemoryRepository) Get(key string) (string, error) {
	value, ok := a.store[key]
	if !ok {
		return "", errors.New("url not found")
	}
	return value, nil
}

func (a InMemoryRepository) Add(key, value string) error {
	a.store[key] = value
	return nil
}
