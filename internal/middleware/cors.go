package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var parsedAllowedOrigins map[string]struct{}

func InitCORS() {
	allowed := os.Getenv("ALLOWED_ORIGINS")
	if allowed == "" {
		if os.Getenv("GO_ENV") == "production" {
			log.Fatal("ALLOWED_ORIGINS must be set in production")
		}
		parsedAllowedOrigins = map[string]struct{}{"*": {}}
		return
	}

	parsedAllowedOrigins = make(map[string]struct{})
	for _, o := range strings.Split(allowed, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			parsedAllowedOrigins[o] = struct{}{}
		}
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		_, allowAll := parsedAllowedOrigins["*"]

		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := parsedAllowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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

// Chain composes middleware in order (first runs first on request).
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

var corsOnce sync.Once

func EnsureCORS() {
	corsOnce.Do(InitCORS)
}
