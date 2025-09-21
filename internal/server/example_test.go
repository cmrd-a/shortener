// Package server provides example tests demonstrating how to use the ShortenHandler
// and other HTTP handlers for the URL shortening service.
package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/cmrd-a/shortener/internal/service"
	"github.com/cmrd-a/shortener/internal/storage"
	"go.uber.org/zap"
)

type MockService struct{}

func (m *MockService) Shorten(ctx context.Context, original string, userID int64) (short string, err error) {
	if original == "" {
		return "", fmt.Errorf("empty URL")
	}
	return "http://localhost:8080/abc123", nil
}

func (m *MockService) ShortenBatch(ctx context.Context, userID int64, corrOrig map[string]string) (corrShort map[string]string, err error) {
	corrShort = make(map[string]string)
	for corr, orig := range corrOrig {
		if orig == "" {
			return nil, fmt.Errorf("empty URL for correlation %s", corr)
		}
		corrShort[corr] = "http://localhost:8080/batch" + corr
	}
	return corrShort, nil
}

func (m *MockService) GetOriginal(ctx context.Context, short string) (original string, err error) {
	return "https://example.com", nil
}

func (m *MockService) Ping(ctx context.Context) (err error) {
	return nil
}

func (m *MockService) GetUserURLs(ctx context.Context, userID int64) (urls []service.SvcURL, err error) {
	return []service.SvcURL{}, nil
}

func (m *MockService) DeleteUserURLs(ctx context.Context, userID int64, shortIDs ...string) {
	// Mock implementation - do nothing
}

func (m *MockService) GetStats(context.Context) (storage.Stats, error) {
	return storage.Stats{
		URLs:  0,
		Users: 0,
	}, nil
}

func setupTestServer() *Server {
	// Set JWT secret for auth middleware
	os.Setenv("JWT_SECRET", "test-secret-key")

	logger := zap.NewNop() // Use no-op logger for tests
	mockService := &MockService{}
	return NewServer(logger, mockService, "")
}

// ExampleShortenHandler demonstrates the basic usage of the ShortenHandler
// for shortening a URL via JSON API endpoint.
//
// The handler expects a JSON request body with a "url" field and returns
// a JSON response with a "result" field containing the shortened URL.
func ExampleShortenHandler() {
	server := setupTestServer()
	defer os.Unsetenv("JWT_SECRET")

	// Create a test request with JSON body
	requestBody := `{"url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Execute the request
	server.Router.ServeHTTP(recorder, req)

	// Print results
	fmt.Printf("Status: %d\n", recorder.Code)
	fmt.Printf("Content-Type: %s\n", recorder.Header().Get("Content-Type"))

	// The response body will contain a JSON with the shortened URL
	body := recorder.Body.String()
	if strings.Contains(body, `"result"`) && recorder.Code == http.StatusCreated {
		fmt.Println("Response contains shortened URL")
	}

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Response contains shortened URL
}

// ExampleShortenHandler_emptyURL demonstrates how the ShortenHandler
// handles validation errors when an empty URL is provided.
//
// The handler should return a 400 Bad Request status when the URL field
// is empty or missing, preventing invalid URLs from being processed.
func ExampleShortenHandler_emptyURL() {
	server := setupTestServer()
	defer os.Unsetenv("JWT_SECRET")

	// Create a test request with empty URL
	requestBody := `{"url": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Execute the request
	server.Router.ServeHTTP(recorder, req)

	// Print results
	fmt.Printf("Status: %d\n", recorder.Code)

	if recorder.Code == http.StatusBadRequest {
		fmt.Println("Empty URL rejected")
	}

	// Output:
	// Status: 400
	// Empty URL rejected
}
