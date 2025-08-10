package storage

import (
	"context"
	"errors"
)

type InMemoryRepository struct {
	store     map[string]StoredURL
	userIndex map[int64][]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		store:     make(map[string]StoredURL),
		userIndex: make(map[int64][]string),
	}
}

func (r InMemoryRepository) Get(ctx context.Context, short string) (string, error) {
	storedURL, ok := r.store[short]
	if !ok {
		return "", errors.New("url not found")
	}
	if storedURL.IsDeleted {
		return "", ErrURLIsDeleted
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

func (r InMemoryRepository) Add(ctx context.Context, short, original string, userID int64) error {
	if oldShort, ok := r.checkOriginalExist(original); ok {
		return NewOriginalExistError(oldShort)
	}
	r.store[short] = StoredURL{ShortID: short, OriginalURL: original, UserID: userID}
	r.userIndex[userID] = append(r.userIndex[userID], short)
	return nil
}

func (r InMemoryRepository) AddBatch(ctx context.Context, userID int64, batch ...StoredURL) error {
	for _, url := range batch {
		r.store[url.ShortID] = StoredURL{ShortID: url.ShortID, OriginalURL: url.OriginalURL, UserID: userID}
	}
	return nil
}

func (r InMemoryRepository) Ping(ctx context.Context) error {
	return nil
}

func (r InMemoryRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	shortIDs, exists := r.userIndex[userID]
	if !exists {
		return []StoredURL{}, nil
	}

	urls := make([]StoredURL, 0, len(shortIDs))
	for _, shortID := range shortIDs {
		if storedURL, ok := r.store[shortID]; ok && !storedURL.IsDeleted {
			urls = append(urls, storedURL)
		}
	}
	return urls, nil
}

func (r InMemoryRepository) MarkDeletedUserURLs(ctx context.Context, urls ...URLForDelete) {
	for _, url := range urls {
		if r.store[url.ShortID].UserID == url.UserID {
			v := r.store[url.ShortID]
			v.IsDeleted = true
			r.store[url.ShortID] = v
		}
	}
}

func (r InMemoryRepository) GetAll() map[string]StoredURL {
	return r.store
}
