package service

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	generator       Generator
	baseURL         string
	repository      storage.Repository
	delUserURLsChan chan storage.URLForDelete
}

func NewURLService(generator Generator, baseURL string, repo storage.Repository) *URLService {
	s := URLService{
		generator:       generator,
		baseURL:         baseURL,
		repository:      repo,
		delUserURLsChan: make(chan storage.URLForDelete, 1024),
	}
	go s.deleteUserURLsJob()
	return &s
}

func (s *URLService) addBaseURL(shortID string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, shortID)
}

func (s *URLService) Shorten(ctx context.Context, originalURL string, userID int64) (string, error) {
	shortID := s.generator.Generate()
	err := s.repository.Add(ctx, shortID, originalURL, userID)
	var myErr *storage.ErrOriginalExist
	if errors.As(err, &myErr) {
		return "", NewOriginalExistError(s.addBaseURL(myErr.Short))
	}
	if err != nil {
		return "", err
	}
	shortURL := s.addBaseURL(shortID)
	return shortURL, nil
}

func (s *URLService) ShortenBatch(ctx context.Context, userID int64, corOriginals map[string]string) (map[string]string, error) {
	shorts := make(map[string]string, len(corOriginals))
	shortsOriginals := make(map[string]string, len(corOriginals))
	for corrID, original := range corOriginals {
		short := s.generator.Generate()
		shorts[corrID] = short
		shortsOriginals[short] = original
	}

	err := s.repository.AddBatch(ctx, userID, shortsOriginals)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(corOriginals))
	for corrID := range corOriginals {
		result[corrID] = s.addBaseURL(shorts[corrID])
	}
	return result, nil
}

func (s *URLService) GetOriginal(ctx context.Context, id string) (string, error) {
	original, err := s.repository.Get(ctx, id)
	return original, err
}

func (s *URLService) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}

func (s *URLService) GetUserURLs(ctx context.Context, id int64) ([]SvcURL, error) {
	storedURLs, err := s.repository.GetUserURLs(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(storedURLs) == 0 {
		return nil, nil
	}
	svcURLs := make([]SvcURL, len(storedURLs))
	for i, stored := range storedURLs {
		svcURLs[i] = SvcURL{OriginalURL: stored.OriginalURL, UserID: stored.UserID, ShortURL: s.addBaseURL(stored.ShortID)}
	}
	return svcURLs, nil
}

func (s *URLService) DeleteUserURLs(ctx context.Context, userID int64, shortIDs ...string) {
	for _, shortID := range shortIDs {
		s.delUserURLsChan <- storage.URLForDelete{UserID: userID, ShortID: shortID}
	}
}

func (s *URLService) deleteUserURLsJob() {
	ticker := time.NewTicker(5 * time.Second)

	var deletions []storage.URLForDelete
	for {
		select {
		case deletion := <-s.delUserURLsChan:
			deletions = append(deletions, deletion)
		case <-ticker.C:
			if len(deletions) == 0 {
				continue
			}
			s.repository.MarkDeletedUserURLs(context.Background(), deletions...)
			deletions = nil
		}
	}
}
