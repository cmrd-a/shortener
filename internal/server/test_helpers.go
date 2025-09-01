package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// executeRequest is a helper function to execute HTTP requests in tests
func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

// TestSetup represents the configuration for setting up test URLs and authentication
type TestSetup struct {
	URLs        []string
	SkipAuth    bool
	ResStatus   int
	ExpectEmpty bool
}

// setupTestURLs creates test URLs and returns the authentication cookie from the first request
func setupTestURLs(s *Server, urls []string) *http.Cookie {
	if len(urls) == 0 {
		return nil
	}

	var authCookie *http.Cookie
	for _, url := range urls {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))
		req.Body.Close()
		res := executeRequest(req, s)

		if authCookie == nil {
			// Get the auth cookie from first request
			result := res.Result()
			for _, cookie := range result.Cookies() {
				if cookie.Name == "Authorization" {
					authCookie = cookie
					break
				}
			}
			result.Body.Close()
		}
	}

	return authCookie
}

// createUserURLsRequest creates a request for getting user URLs with optional authentication
func createUserURLsRequest(setup TestSetup, authCookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	// Add auth cookie if not skipping auth and we have one
	if !setup.SkipAuth && authCookie != nil {
		req.AddCookie(authCookie)
	}

	return req
}

// assertUserURLsResponse validates the response from the user URLs endpoint
func assertUserURLsResponse(t *testing.T, res *httptest.ResponseRecorder, setup TestSetup) {
	assert.Equal(t, setup.ResStatus, res.Code)

	if setup.ResStatus == http.StatusOK {
		assert.Greater(t, len(res.Body.String()), 0)
		h := res.Header().Get("Content-Type")
		assert.Equal(t, "application/json", h)
	}
}

// runUserURLsTest executes a complete user URLs test with the given setup
func runUserURLsTest(t *testing.T, s *Server, setup TestSetup) {
	// Setup URLs if needed
	authCookie := setupTestURLs(s, setup.URLs)

	// Create request for getting user URLs
	req := createUserURLsRequest(setup, authCookie)

	// Execute request
	res := executeRequest(req, s)

	// Assert response
	assertUserURLsResponse(t, res, setup)
}
