package service

import (
	"fmt"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Repository interface {
	Get(key string) (value string, err error)
	Add(key, value string) error
}
type URLService struct {
	baseURL    string
	repository Repository
}

func NewURLService(baseURL string, repo Repository) *URLService {
	return &URLService{baseURL, repo}
}

func (s *URLService) createID() string {

	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (s *URLService) Shorten(originalURL string) (string, error) {
	id := s.createID()
	err := s.repository.Add(id, originalURL)
	if err != nil {
		return "", err
	}
	shortURL := fmt.Sprintf("%s/%s", s.baseURL, id)
	return shortURL, nil
}

func (s *URLService) GetOriginal(id string) (string, error) {
	original, err := s.repository.Get(id)
	return original, err
}
