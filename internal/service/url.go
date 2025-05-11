package service

import (
	"fmt"
	"math/rand"

	"github.com/cmrd-a/shortener/internal/config"
)

type Repository interface {
	Get(key string) (value string, err error)
	Add(key, value string) error
}
type URLService struct {
	repository Repository
}

func NewURLService(repo Repository) *URLService {
	return &URLService{repo}
}

func (s *URLService) createID() string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (s *URLService) save(id, originalURL string) error {
	err := s.repository.Add(id, originalURL)
	return err
}

func (s *URLService) Shorten(originalURL string) (string, error) {
	id := s.createID()
	err := s.save(id, originalURL)
	shortURL := fmt.Sprintf("%s/%s", config.BaseURL, id)
	return shortURL, err
}

func (s *URLService) GetOriginal(id string) (string, error) {
	original, err := s.repository.Get(id)
	return original, err
}
