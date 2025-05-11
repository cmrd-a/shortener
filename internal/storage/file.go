package storage

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/cmrd-a/shortener/internal/config"
)

//go:generate easyjson file.go

//easyjson:json
type StoredURL struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
}

type FileRepository struct{}

func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

func (a FileRepository) Add(key, value string) error {
	file, _ := os.OpenFile(config.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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
func (a FileRepository) Get(key string) (string, error) {
	data, err := os.ReadFile(config.FileStoragePath)
	if err != nil {
		return "", err
	}
	for _, str := range strings.Split(string(data), "\n") {
		s := StoredURL{}
		s.UnmarshalJSON([]byte(str))
		if s.ID == key {
			return s.OriginalURL, nil
		}
	}
	return "", errors.New("url not found")
}
