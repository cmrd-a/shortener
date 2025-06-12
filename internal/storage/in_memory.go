package storage

import (
	"context"
	"errors"
)

type InMemoryRepository struct {
	store map[string]StoredURL
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{store: make(map[string]StoredURL)}
}

func (a InMemoryRepository) Get(short string) (string, error) {
	storedURL, ok := a.store[short]
	if !ok {
		return "", errors.New("url not found")
	}
	return storedURL.OriginalURL, nil
}

func (a InMemoryRepository) CheckOriginalExist(original string) (string, bool) {
	for key, value := range a.store {
		if value.OriginalURL == original {
			return key, true
		}
	}
	return "", false
}

func (a InMemoryRepository) Add(short, original string) error {
	if oldShort, ok := a.CheckOriginalExist(original); ok {
		return NewOriginalExistError(oldShort)
	}
	a.store[short] = StoredURL{ShortID: short, OriginalURL: original, UserID: 0}
	return nil
}

func (a InMemoryRepository) AddBatch(ctx context.Context, b map[string]string) error {
	for short, original := range b {
		a.store[short] = StoredURL{ShortID: short, OriginalURL: original, UserID: 0}
	}
	return nil
}

func (a InMemoryRepository) Ping(ctx context.Context) error {
	return nil
}

func (a InMemoryRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	var urls []StoredURL
	for _, value := range a.store {
		if value.UserID == userID {
			urls = append(urls, value)
		}
	}
	return urls, nil
}
