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
type ShortGenerator struct {
}

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

func (s *URLService) Shorten(originalURL string) (string, error) {
	short := s.generator.Generate()
	err := s.repository.Add(short, originalURL)
	var myErr *storage.OriginalExistError
	if errors.As(err, &myErr) {
		return "", NewOriginalExistError(fmt.Sprintf("%s/%s", s.baseURL, myErr.Short))
	}
	if err != nil {
		return "", err
	}
	shortURL := fmt.Sprintf("%s/%s", s.baseURL, short)
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
