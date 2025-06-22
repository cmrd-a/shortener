package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type FileRepository struct {
	path  string
	cache *InMemoryRepository
}

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

func (r FileRepository) Get(ctx context.Context, short string) (string, error) {
	return r.cache.Get(ctx, short)
}

func (r FileRepository) Add(ctx context.Context, short, original string, userID int64) error {
	return r.AddBatch(ctx, userID, StoredURL{ShortID: short, OriginalURL: original, UserID: userID})
}

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

func (r FileRepository) Ping(ctx context.Context) error {
	return nil
}

func (r FileRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	return r.cache.GetUserURLs(ctx, userID)
}

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
