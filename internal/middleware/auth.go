package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/leventkok/mlc-llm-monitoring/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"


func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"token required"}`))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid token format"}`))
			return
		}
		tokenString := parts[1]

		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid or expired token"}`))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		
		next(w, r.WithContext(ctx))
	}
}