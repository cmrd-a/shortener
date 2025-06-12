package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/cmrd-a/shortener/internal/storage"

	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Generator interface {
	Generate() string
}
type ShortGenerator struct{}

func NewShortGenerator() *ShortGenerator {
	return &ShortGenerator{}
}

func (g *ShortGenerator) Generate() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type URLService struct {
	generator  Generator
	baseURL    string
	repository storage.Repository
}

func NewURLService(generator Generator, baseURL string, repo storage.Repository) *URLService {
	return &URLService{generator, baseURL, repo}
}
func (s *URLService) IDtoURL(shortID string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, shortID)
}

func (s *URLService) Shorten(originalURL string) (string, error) {
	shortID := s.generator.Generate()
	err := s.repository.Add(shortID, originalURL)
	var myErr *storage.OriginalExistError
	if errors.As(err, &myErr) {
		return "", NewOriginalExistError(s.IDtoURL(myErr.Short))
	}
	if err != nil {
		return "", err
	}
	shortURL := s.IDtoURL(shortID)
	return shortURL, nil
}

func (s *URLService) ShortenBatch(ctx context.Context, corOriginals map[string]string) (map[string]string, error) {
	shorts := make(map[string]string, len(corOriginals))
	shortsOriginals := make(map[string]string, len(corOriginals))
	for corrID, original := range corOriginals {
		short := s.generator.Generate()
		shorts[corrID] = short
		shortsOriginals[short] = original
	}

	err := s.repository.AddBatch(ctx, shortsOriginals)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(corOriginals))
	for corrID := range corOriginals {
		result[corrID] = fmt.Sprintf("%s/%s", s.baseURL, shorts[corrID])
		result[corrID] = s.IDtoURL(shorts[corrID])
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

func (s *URLService) GetUserURLs(ctx context.Context, id int64) ([]storage.StoredURL, error) {
	// storedURLs, err := s.repository.GetUserURLs(ctx, id)
	// if err != nil {
	// 	return make([]storage.StoredURL)), err
	// }
	// return fmt.Sprintf("%s/%s", s.baseURL, shortURL), nil
	return nil, nil
}
