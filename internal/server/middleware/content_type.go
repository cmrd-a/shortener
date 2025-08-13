package middleware

import (
	"net/http"
	"slices"
	"strings"
)

// CheckContentType returns middleware that validates Content-Type header for API endpoints.
// It ensures that POST and DELETE requests to /api/ endpoints have application/json content type.
func CheckContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if slices.Contains([]string{http.MethodPost, http.MethodDelete}, req.Method) &&
			strings.Contains(req.RequestURI, "/api/") &&
			req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "only Content-Type:application/json is supported", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(res, req)
	})
}
