package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRootHandler(t *testing.T) {
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
		{name: "2 get", params: params{method: "GET", url: "/", body: "some red"}, want: want{stausCode: http.StatusTemporaryRedirect}},
	}
	for i, tt := range tests {
		bodyBytes := strings.NewReader(tt.params.body)
		req := httptest.NewRequest(tt.params.method, tt.params.url, bodyBytes)
		res := httptest.NewRecorder()
		RootHandler(res, req)

		resBody := res.Body.String()

		assert.Equal(t, tt.want.stausCode, res.Code)
		assert.GreaterOrEqual(t, len(resBody), tt.want.minBodyLength)
		if i == 0 {
			tests[1].params.url = strings.TrimPrefix(resBody, "http://example.com")
		}
	}
}
