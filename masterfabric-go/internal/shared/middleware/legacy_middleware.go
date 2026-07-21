package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type visitor struct {
	count    int
	windowAt time.Time
}

// LegacyRateLimit applies a per-IP fixed window limit to specific paths.
func LegacyRateLimit(maxPerWindow int, window time.Duration, paths ...string) func(http.Handler) http.Handler {
	pathSet := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		pathSet[p] = struct{}{}
	}

	var (
		mu       sync.Mutex
		visitors = map[string]*visitor{}
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := pathSet[r.URL.Path]; !ok {
				next.ServeHTTP(w, r)
				return
			}

			ip := clientIP(r)
			now := time.Now()

			mu.Lock()
			v, ok := visitors[ip]
			if !ok || now.Sub(v.windowAt) >= window {
				v = &visitor{count: 0, windowAt: now}
				visitors[ip] = v
			}
			v.count++
			allowed := v.count <= maxPerWindow
			mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", "60")
				response.LegacyError(w, http.StatusTooManyRequests, "too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// SecurityHeaders adds baseline response headers on every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// RequestTimeoutLegacy applies a 10s timeout to requests.
func RequestTimeoutLegacy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
