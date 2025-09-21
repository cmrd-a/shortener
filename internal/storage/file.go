package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// FileRepository implements the Repository interface using file-based persistence with in-memory caching.
type FileRepository struct {
	path  string
	cache *InMemoryRepository
}

// NewFileRepository creates a new FileRepository instance that persists data to a file while using an in-memory cache for fast access.
func NewFileRepository(path string, cache *InMemoryRepository) (*FileRepository, error) {
	r := &FileRepository{path, cache}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return r, nil
	}
	data, err := os.ReadFile(r.path)
	if err != nil {
		return nil, err
	}
	for str := range strings.SplitSeq(string(data), "\n") {
		s := StoredURL{}
		err := s.UnmarshalJSON([]byte(str))
		if err != nil {
			return nil, err
		}
		err = r.cache.Add(context.TODO(), s.ShortID, s.OriginalURL, s.UserID)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Get retrieves the original URL for a given short URL identifier from the cache.
func (r FileRepository) Get(ctx context.Context, short string) (string, error) {
	return r.cache.Get(ctx, short)
}

// Add stores a new URL mapping both in cache and persists it to the file.
func (r FileRepository) Add(ctx context.Context, short, original string, userID int64) error {
	return r.AddBatch(ctx, userID, StoredURL{ShortID: short, OriginalURL: original, UserID: userID})
}

// AddBatch stores multiple URL mappings both in cache and appends them to the file.
func (r FileRepository) AddBatch(ctx context.Context, userID int64, batch ...StoredURL) error {
	err := r.cache.AddBatch(ctx, userID, batch...)
	if err != nil {
		return err
	}
	file, _ := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	var result []byte
	for _, url := range batch {
		data, err := json.Marshal(url)
		if err != nil {
			return err
		}
		data = append(data, '\n')
		result = append(result, data...)
	}

	_, err = file.Write(result)
	if err != nil {
		return err
	}
	return nil
}

// Ping checks the health of the repository (always returns nil for file storage).
func (r FileRepository) Ping(ctx context.Context) error {
	return nil
}

// GetUserURLs retrieves all URLs created by a specific user from the cache.
func (r FileRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	return r.cache.GetUserURLs(ctx, userID)
}

// GetStats retrieves statistics about the stored URLs and users from the cache.
func (r FileRepository) GetStats(ctx context.Context) (Stats, error) {
	return r.cache.GetStats(ctx)
}

// MarkDeletedUserURLs marks the specified URLs as deleted in cache and rewrites the entire file.
func (r FileRepository) MarkDeletedUserURLs(ctx context.Context, urls ...URLForDelete) {
	r.cache.MarkDeletedUserURLs(ctx, urls...)

	file, _ := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()
	all := r.cache.GetAll()
	var result []byte
	for _, url := range all {
		data, err := json.Marshal(url)
		if err != nil {
			fmt.Printf("error while marshalling %v", err)
		}
		data = append(data, '\n')
		result = append(result, data...)
	}

	_, err := file.Write(result)
	if err != nil {
		fmt.Printf("error while writing file %v", err)
	}

}

// Close closes the FileRepository by ensuring all data is flushed to disk.
// This method saves all current data to the file during graceful shutdown.
func (r *FileRepository) Close() error {
	// Ensure all data is written to file before closing
	file, err := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	all := r.cache.GetAll()
	var result []byte
	for _, url := range all {
		data, err := json.Marshal(url)
		if err != nil {
			return fmt.Errorf("error marshalling URL data: %v", err)
		}
		result = append(result, data...)
		result = append(result, '\n')
	}
	_, err = file.Write(result)
	return err
}
