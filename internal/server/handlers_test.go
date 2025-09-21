package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cmrd-a/shortener/internal/config"
	"github.com/cmrd-a/shortener/internal/logger"
	"github.com/cmrd-a/shortener/internal/service"
	"github.com/cmrd-a/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg = config.NewConfig(false)
var zl, _ = logger.NewLogger(cfg.LogLevel)
var ctx = context.Background()
var repo, _ = storage.MakeRepository(ctx, cfg)
var generator = service.NewShortGenerator()
var server = NewServer(zl, service.NewURLService(generator, cfg.BaseURL, repo), "")

func TestAddLinkHandler(t *testing.T) {
	type want struct {
		stausCode     int
		minBodyLength int
	}
	type params struct {
		method string
		url    string
		body   string
	}
	tests := []struct {
		name   string
		params params
		want   want
	}{
		{name: "happy_path", params: params{method: http.MethodPost, url: "/", body: "https://zed.dev"}, want: want{stausCode: http.StatusCreated}},
		{name: "empty_body", params: params{method: http.MethodPost, url: "/", body: ""}, want: want{stausCode: http.StatusBadRequest}},
		{name: "invalid_method", params: params{method: http.MethodGet, url: "/", body: "https://example.com"}, want: want{stausCode: http.StatusMethodNotAllowed}},
		{name: "long_url", params: params{method: http.MethodPost, url: "/", body: "https://verylongexampleurl.com/with/many/path/segments/and/query/parameters?param1=value1&param2=value2"}, want: want{stausCode: http.StatusCreated}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes := strings.NewReader(tt.params.body)
			req := httptest.NewRequest(tt.params.method, tt.params.url, bodyBytes)
			res := executeRequest(req, server)
			resBody := res.Body.String()

			assert.Equal(t, tt.want.stausCode, res.Code)
			assert.GreaterOrEqual(t, len(resBody), tt.want.minBodyLength)
		})
	}
}

func TestGetLinkHandler(t *testing.T) {
	originalLink := "https://ya.ru"
	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalLink))
	res1 := executeRequest(req1, server)
	linkID := res1.Body.String()

	req := httptest.NewRequest(http.MethodGet, linkID, nil)
	res := executeRequest(req, server)
	header := res.Header().Get("location")

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, originalLink, header)

}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name      string
		reqBody   string
		resStatus int
		resLen    int
		compress  bool
	}{
		{name: "happy_path_without_compress", reqBody: "{\"url\": \"https://protobuf.dev\"}", resStatus: http.StatusCreated, resLen: 10, compress: false},
		{name: "happy_path_with_compress", reqBody: "{\"url\": \"https://microservices.io\"}", resStatus: http.StatusCreated, resLen: 10, compress: true},
		{name: "empty_body", reqBody: "", resStatus: http.StatusBadRequest},
		{name: "empty_url", reqBody: "{\"url\": \"\"}", resStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.compress == true {
				var buf bytes.Buffer
				zw := gzip.NewWriter(&buf)
				zw.Write([]byte(tt.reqBody))
				zw.Close()
				req = httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(buf.Bytes()))
				req.Header.Set("Accept-Encoding", "gzip")
				req.Header.Set("Content-Encoding", "gzip")
			} else {
				req = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.reqBody))
			}
			req.Header.Set("Content-Type", "application/json")
			res := executeRequest(req, server)

			assert.Equal(t, tt.resStatus, res.Code)
			if res.Code == http.StatusCreated {
				resString := res.Body.String()
				resLen := len(resString)
				assert.GreaterOrEqual(t, resLen, tt.resLen)

				h := res.Header().Get("Content-Type")
				assert.Equal(t, "application/json", h)
			}

		})
	}
}

func TestShortenBatchHandler(t *testing.T) {
	tests := []struct {
		name      string
		reqBody   string
		resStatus int
		resLen    int
	}{
		{
			name:      "happy_path_batch",
			reqBody:   `[{"correlation_id": "1", "original_url": "https://golang.org"}, {"correlation_id": "2", "original_url": "https://github.com"}]`,
			resStatus: http.StatusCreated,
			resLen:    10,
		},
		{
			name:      "empty_body",
			reqBody:   "",
			resStatus: http.StatusOK,
		},
		{
			name:      "invalid_json",
			reqBody:   `{"invalid": json}`,
			resStatus: http.StatusBadRequest,
		},
		{
			name:      "empty_array",
			reqBody:   `[]`,
			resStatus: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.reqBody))
			req.Header.Set("Content-Type", "application/json")
			res := executeRequest(req, server)

			assert.Equal(t, tt.resStatus, res.Code)
			if res.Code == http.StatusCreated {
				resString := res.Body.String()
				resLen := len(resString)
				assert.GreaterOrEqual(t, resLen, tt.resLen)

				h := res.Header().Get("Content-Type")
				assert.Equal(t, "application/json", h)
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name      string
		resStatus int
	}{
		{
			name:      "ping_success",
			resStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			res := executeRequest(req, server)

			assert.Equal(t, tt.resStatus, res.Code)
		})
	}
}

func TestGetLinkHandlerErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		linkID    string
		resStatus int
		setupLink bool
		setupURL  string
	}{
		{
			name:      "non_existent_link",
			linkID:    "/nonexistent",
			resStatus: http.StatusBadRequest,
			setupLink: false,
		},
		{
			name:      "empty_link_id",
			linkID:    "/",
			resStatus: http.StatusMethodNotAllowed,
			setupLink: false,
		},
		{
			name:      "valid_redirect",
			linkID:    "",
			resStatus: http.StatusTemporaryRedirect,
			setupLink: true,
			setupURL:  "https://example.com/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var linkID string
			if tt.setupLink {
				// Create a link first
				req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.setupURL))
				res1 := executeRequest(req1, server)
				linkID = res1.Body.String()
			} else {
				linkID = tt.linkID
			}

			req := httptest.NewRequest(http.MethodGet, linkID, nil)
			res := executeRequest(req, server)

			assert.Equal(t, tt.resStatus, res.Code)

			if tt.resStatus == http.StatusTemporaryRedirect {
				header := res.Header().Get("location")
				assert.Equal(t, tt.setupURL, header)
			}
		})
	}
}

func TestGetUserURLsHandler(t *testing.T) {
	tests := []TestSetup{
		{
			URLs:        []string{"https://example.com/1", "https://example.com/2"},
			ResStatus:   http.StatusOK,
			ExpectEmpty: false,
			SkipAuth:    false,
		},
		{
			URLs:        []string{},
			ResStatus:   http.StatusNoContent,
			ExpectEmpty: true,
			SkipAuth:    true,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			runUserURLsTest(t, server, tt)
		})
	}
}

func TestStatsHandler(t *testing.T) {
	type want struct {
		statusCode int
		response   string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "successful stats request",
			want: want{
				statusCode: http.StatusOK,
				response:   `{"urls":0,"users":0}`,
			},
		},
		{
			name: "forbidden without X-Real-IP",
			want: want{
				statusCode: http.StatusForbidden,
				response:   "Forbidden\n",
			},
		},
		{
			name: "forbidden with untrusted IP",
			want: want{
				statusCode: http.StatusForbidden,
				response:   "Forbidden\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated repository for each test
			isolatedRepo := storage.NewInMemoryRepository()
			isolatedGenerator := service.NewShortGenerator()
			isolatedService := service.NewURLService(isolatedGenerator, cfg.BaseURL, isolatedRepo)

			request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)

			var testServer *Server
			switch tt.name {
			case "successful stats request":
				request.Header.Set("X-Real-IP", "192.168.1.1")
				testServer = NewServer(zl, isolatedService, "192.168.1.0/24")
			case "forbidden without X-Real-IP":
				// Don't set X-Real-IP header
				testServer = NewServer(zl, isolatedService, "192.168.1.0/24")
			case "forbidden with untrusted IP":
				request.Header.Set("X-Real-IP", "10.0.0.1")
				testServer = NewServer(zl, isolatedService, "192.168.1.0/24")
			}

			w := httptest.NewRecorder()
			testServer.Router.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if res.Header.Get("Content-Type") == "application/json" {
				assert.JSONEq(t, tt.want.response, string(resBody))
			} else {
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}
