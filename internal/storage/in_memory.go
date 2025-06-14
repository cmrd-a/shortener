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

func (r InMemoryRepository) Get(short string) (string, error) {
	storedURL, ok := r.store[short]
	if !ok {
		return "", errors.New("url not found")
	}
	return storedURL.OriginalURL, nil
}

func (r InMemoryRepository) checkOriginalExist(original string) (string, bool) {
	for key, value := range r.store {
		if value.OriginalURL == original {
			return key, true
		}
	}
	return "", false
}

func (r InMemoryRepository) Add(short, original string, userID int64) error {
	if oldShort, ok := r.checkOriginalExist(original); ok {
		return NewOriginalExistError(oldShort)
	}
	r.store[short] = StoredURL{ShortID: short, OriginalURL: original, UserID: userID}
	return nil
}

func (r InMemoryRepository) AddBatch(ctx context.Context, userID int64, b map[string]string) error {
	for short, original := range b {
		r.store[short] = StoredURL{ShortID: short, OriginalURL: original, UserID: userID}
	}
	return nil
}

func (r InMemoryRepository) Ping(ctx context.Context) error {
	return nil
}

func (r InMemoryRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	var urls []StoredURL
	for _, value := range r.store {
		if value.UserID == userID {
			urls = append(urls, value)
		}
	}
	return urls, nil
}

func (r InMemoryRepository) MarkDeletedUserURLs(ctx context.Context, userID int64, shortIDs ...string) {
	for _, shortID := range shortIDs {
		if r.store[shortID].UserID == userID {
			v := r.store[shortID]
			v.IsDeleted = true
			r.store[shortID] = v
		}
	}
}
