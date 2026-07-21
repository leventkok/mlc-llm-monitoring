package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/leventkok/mlc-llm-monitoring/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tokenString := extractToken(r)
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"token required"}`))
			return
		}

		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid or expired token"}`))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	if c, err := r.Cookie(auth.SessionCookieName()); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	sameSite, secure := sessionCookieOptions()
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   86400,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	sameSite, secure := sessionCookieOptions()
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}

func sessionCookieOptions() (http.SameSite, bool) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("GO_ENV")), "production") {
		return http.SameSiteNoneMode, true
	}
	return http.SameSiteLaxMode, false
}

func isSecureCookie() bool {
	_, secure := sessionCookieOptions()
	return secure
}
