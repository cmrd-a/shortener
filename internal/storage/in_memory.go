package storage

import (
	"context"
	"errors"
	"sync"
)

// InMemoryRepository implements the Repository interface using in-memory maps for storage.
type InMemoryRepository struct {
	store     map[string]StoredURL
	userIndex map[int64][]string
	mu        *sync.Mutex
}

// NewInMemoryRepository creates a new InMemoryRepository instance with initialized storage maps.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		store:     make(map[string]StoredURL),
		userIndex: make(map[int64][]string),
		mu:        &sync.Mutex{},
	}
}

// Get retrieves the original URL for a given short URL identifier.
func (r InMemoryRepository) Get(ctx context.Context, short string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

// Add stores a new URL mapping in the repository.
func (r InMemoryRepository) Add(ctx context.Context, short, original string, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if oldShort, ok := r.checkOriginalExist(original); ok {
		return NewOriginalExistError(oldShort)
	}
	r.store[short] = StoredURL{ShortID: short, OriginalURL: original, UserID: userID}
	r.userIndex[userID] = append(r.userIndex[userID], short)
	return nil
}

// AddBatch stores multiple URL mappings in a single operation.
func (r InMemoryRepository) AddBatch(ctx context.Context, userID int64, batch ...StoredURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, url := range batch {
		r.store[url.ShortID] = StoredURL{ShortID: url.ShortID, OriginalURL: url.OriginalURL, UserID: userID}
	}
	return nil
}

// Ping checks the health of the repository (always returns nil for in-memory storage).
func (r InMemoryRepository) Ping(ctx context.Context) error {
	return nil
}

// GetUserURLs retrieves all non-deleted URLs created by a specific user.
func (r InMemoryRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

// MarkDeletedUserURLs marks the specified URLs as deleted for the given users.
func (r InMemoryRepository) MarkDeletedUserURLs(ctx context.Context, urls ...URLForDelete) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, url := range urls {
		if r.store[url.ShortID].UserID == url.UserID {
			v := r.store[url.ShortID]
			v.IsDeleted = true
			r.store[url.ShortID] = v
		}
	}
}

// GetAll returns all stored URLs (used primarily for testing and debugging).
func (r InMemoryRepository) GetAll() map[string]StoredURL {
	return r.store
}
