package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

const sessionCookieName = "session"

func SessionCookieName() string {
	return sessionCookieName
}

func InitJWT() {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		if os.Getenv("GO_ENV") == "production" {
			log.Fatal("JWT_SECRET must be set in production")
		}
		s = "dev-only-secret-change-in-production-local"
		log.Println("WARNING: JWT_SECRET not set, using dev-only fallback")
	}
	if os.Getenv("GO_ENV") == "production" && len(s) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters in production")
	}
	jwtSecret = []byte(s)
}

func GenerateToken(userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil || !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.UserID == "" {
		return "", jwt.ErrTokenInvalidClaims
	}
	return claims.UserID, nil
}
