package storage

import (
	"context"
	"encoding/json"
	"errors"
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
		err = r.cache.Add(s.ShortID, s.OriginalURL, s.UserID)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r FileRepository) Get(short string) (string, error) {
	return r.cache.Get(short)
}

func (r FileRepository) Add(short, original string, userID int64) error {
	err := r.cache.Add(short, original, userID)
	if err != nil {
		return err
	}
	file, _ := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	s := StoredURL{ShortID: short, OriginalURL: original, UserID: userID}
	data, err := json.Marshal(&s)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = file.Write(data)
	return err

}

func (r FileRepository) AddBatch(ctx context.Context, userID int64, b map[string]string) error {
	err := r.cache.AddBatch(ctx, userID, b)
	if err != nil {
		return err
	}
	file, _ := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	var result []byte
	for short, original := range b {
		data, err := json.Marshal(StoredURL{ShortID: short, OriginalURL: original, UserID: userID})
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

func (r FileRepository) MarkDeletedUserURLs(ctx context.Context, userID int64, shortIDs ...string) {
	r.cache.MarkDeletedUserURLs(ctx, userID, shortIDs...)
	//TODO
}
