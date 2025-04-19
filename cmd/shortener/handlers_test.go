package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func TestAddLinkHandler(t *testing.T) {
	s := CreateNewServer()
	s.MountHandlers()
	type want struct {
		stausCode     int
		response      string
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
		{name: "1 post", params: params{method: "POST", url: "/", body: "https://ya.ru"}, want: want{stausCode: http.StatusCreated}},
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
	s.MountHandlers()

	originalLink := "https://ya.ru"
	req1 := httptest.NewRequest("POST", "/", strings.NewReader(originalLink))
	res1 := executeRequest(req1, s)
	linkID := res1.Body.String()

	req := httptest.NewRequest("GET", linkID, nil)
	res := executeRequest(req, s)
	header := res.Header().Get("location")

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, originalLink, header)

}
