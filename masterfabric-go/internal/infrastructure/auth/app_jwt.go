package auth

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const sessionCookieName = "session"

type legacyClaims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

// AppJWTService issues and validates legacy-compatible session tokens.
type AppJWTService struct {
	secret []byte
}

func NewAppJWTService(secret string) *AppJWTService {
	return &AppJWTService{secret: []byte(secret)}
}

func SessionCookieName() string {
	return sessionCookieName
}

func (s *AppJWTService) GenerateToken(userID string) (string, error) {
	now := time.Now()
	claims := legacyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *AppJWTService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &legacyClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil || !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(*legacyClaims)
	if !ok || claims.UserID == "" {
		return "", jwt.ErrTokenInvalidClaims
	}
	return claims.UserID, nil
}

func sessionCookieOptions() (http.SameSite, bool) {
	if os.Getenv("GO_ENV") == "production" {
		return http.SameSiteNoneMode, true
	}
	return http.SameSiteLaxMode, false
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	sameSite, secure := sessionCookieOptions()
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
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
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}
