package storage

import (
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

func (r FileRepository) Add(key, value string) error {
	r.cache.Add(key, value)
	file, _ := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	s := StoredURL{key, value}
	data, err := json.Marshal(&s)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = file.Write(data)
	return err

}
func (r FileRepository) Get(key string) (string, error) {
	url, err := r.cache.Get(key)
	return url, err
}
