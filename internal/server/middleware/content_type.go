package middleware

import (
	"net/http"
	"strings"
)

func CheckContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.RequestURI, "/api/") && req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "only Content-Type:application/json is supported", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(res, req)
	})
}
