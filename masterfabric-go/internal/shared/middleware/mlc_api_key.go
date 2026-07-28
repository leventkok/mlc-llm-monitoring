package middleware

import (
	"net/http"
	"strings"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

// RequireMLCAPIKey protects hybrid agent routes with the same secret as the MLC gateway.
func RequireMLCAPIKey(expectedKey string) func(http.Handler) http.Handler {
	expectedKey = strings.TrimSpace(expectedKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedKey == "" {
				response.LegacyError(w, http.StatusServiceUnavailable, "agent API not configured")
				return
			}
			got := strings.TrimSpace(r.Header.Get("X-MLC-API-Key"))
			if got == "" {
				got = strings.TrimSpace(r.Header.Get("Authorization"))
				got = strings.TrimPrefix(got, "Bearer ")
			}
			if got != expectedKey {
				response.LegacyError(w, http.StatusUnauthorized, "invalid API key")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
