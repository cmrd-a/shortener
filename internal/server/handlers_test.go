package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func TestAddLinkHandler(t *testing.T) {
	s := CreateNewServer()
	s.Prepare()
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
		{name: "happy path", params: params{method: http.MethodPost, url: "/", body: "https://ya.ru"}, want: want{stausCode: http.StatusCreated}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes := strings.NewReader(tt.params.body)
			req := httptest.NewRequest(tt.params.method, tt.params.url, bodyBytes)
			res := executeRequest(req, s)
			resBody := res.Body.String()

			assert.Equal(t, tt.want.stausCode, res.Code)
			assert.GreaterOrEqual(t, len(resBody), tt.want.minBodyLength)
		})
	}
}

func TestGetLinkHandler(t *testing.T) {
	s := CreateNewServer()
	s.Prepare()

	originalLink := "https://ya.ru"
	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalLink))
	res1 := executeRequest(req1, s)
	linkID := res1.Body.String()

	req := httptest.NewRequest(http.MethodGet, linkID, nil)
	res := executeRequest(req, s)
	header := res.Header().Get("location")

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, originalLink, header)

}

func TestShortenHandler(t *testing.T) {
	s := CreateNewServer()
	s.Prepare()

	tests := []struct {
		name      string
		reqBody   string
		resStatus int
		resLen    int
	}{
		{name: "happy path", reqBody: "{\"url\": \"https://mail.ru\"}", resStatus: http.StatusCreated, resLen: 10},
		{name: "empty body", reqBody: "", resStatus: http.StatusBadRequest},
		{name: "empty url", reqBody: "{\"url\": \"\"}", resStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBodyBytes := strings.NewReader(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", reqBodyBytes)
			req.Header.Set("Content-Type", "application/json")

			res := executeRequest(req, s)

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
