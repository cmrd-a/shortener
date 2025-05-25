package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/cmrd-a/shortener/internal/storage"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Service interface {
	Shorten(string) (string, error)
	ShortenBatch(context.Context, map[string]string) (map[string]string, error)
	GetOriginal(string) (string, error)
	Ping(context.Context) error
}
type URLService struct {
	baseURL    string
	repository storage.Repository
}

func NewURLService(baseURL string, repo storage.Repository) *URLService {
	return &URLService{baseURL, repo}
}

func (s *URLService) createShort() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (s *URLService) Shorten(originalURL string) (string, error) {
	id := s.createShort()
	err := s.repository.Add(id, originalURL)
	if err != nil {
		return "", err
	}
	shortURL := fmt.Sprintf("%s/%s", s.baseURL, id)
	return shortURL, nil
}

func (s *URLService) ShortenBatch(ctx context.Context, originals map[string]string) (map[string]string, error) {
	shorts := make(map[string]string, len(originals))
	shortsOriginals := make(map[string]string, len(originals))
	for corrID, original := range originals {
		short := s.createShort()
		shorts[corrID] = short
		shortsOriginals[short] = original
	}

	err := s.repository.AddBatch(ctx, shortsOriginals)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(originals))
	for corrID := range originals {
		result[corrID] = fmt.Sprintf("%s/%s", s.baseURL, shorts[corrID])
	}
	return result, nil
}

func (s *URLService) GetOriginal(id string) (string, error) {
	original, err := s.repository.Get(id)
	return original, err
}

func (s *URLService) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}
