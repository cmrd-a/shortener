package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

//go:generate easyjson file.go

//easyjson:json
type StoredURL struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
}

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
	for _, str := range strings.Split(string(data), "\n") {
		s := StoredURL{}
		s.UnmarshalJSON([]byte(str))
		r.cache.Add(s.ID, s.OriginalURL)
	}
	return r, nil
}

func (r FileRepository) Get(short string) (string, error) {
	url, err := r.cache.Get(short)
	return url, err
}

func (r FileRepository) Add(short, original string) error {
	err := r.cache.Add(short, original)
	if err != nil {
		return err
	}
	file, _ := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	s := StoredURL{short, original}
	data, err := json.Marshal(&s)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = file.Write(data)
	return err

}

func (r FileRepository) AddBatch(ctx context.Context, b map[string]string) error {
	r.cache.AddBatch(ctx, b)
	file, _ := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	var result []byte
	for short, original := range b {
		data, err := json.Marshal(StoredURL{short, original})
		if err != nil {
			return err
		}
		data = append(data, '\n')
		result = append(result, data...)
	}

	_, err := file.Write(result)
	if err != nil {
		return err
	}
	return nil
}

func (r FileRepository) Ping(ctx context.Context) error {
	return nil
}
