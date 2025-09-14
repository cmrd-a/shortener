package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnet middleware checks if the client IP from X-Real-IP header
// is within the trusted subnet
func TrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if trustedSubnet == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, trustedNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(clientIP)
			if ip == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !trustedNet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
