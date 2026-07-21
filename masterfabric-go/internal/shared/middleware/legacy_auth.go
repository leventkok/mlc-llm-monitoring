package middleware

import (
	"context"
	"net/http"
	"strings"

	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type legacyContextKey string

const LegacyUserIDKey legacyContextKey = "legacyUserID"

// LegacyAuth validates Bearer JWT or HttpOnly session cookie (legacy API).
func LegacyAuth(jwt *infraAuth.AppJWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractLegacyToken(r)
			if token == "" {
				response.LegacyError(w, http.StatusUnauthorized, "token required")
				return
			}

			userID, err := jwt.ValidateToken(token)
			if err != nil {
				response.LegacyError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), LegacyUserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LegacyUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(LegacyUserIDKey).(string)
	return id, ok && id != ""
}

func extractLegacyToken(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	if c, err := r.Cookie(infraAuth.SessionCookieName()); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}
