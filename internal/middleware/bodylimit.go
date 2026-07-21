package middleware

import (
	"net/http"
)

const defaultMaxBodyBytes = 1 << 20 // 1 MiB

func MaxBodyBytes(max int64) func(http.Handler) http.Handler {
	if max <= 0 {
		max = defaultMaxBodyBytes
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil && r.Method != http.MethodGet && r.Method != http.MethodHead {
				r.Body = http.MaxBytesReader(w, r.Body, max)
			}
			next.ServeHTTP(w, r)
		})
	}
}
