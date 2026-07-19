package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("secret-key-it-will-change")
 
func GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,                               
		"exp": time.Now().Add(24 * time.Hour).Unix(), 
		"iat": time.Now().Unix(),                    
	}

	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil 
	})
	if err != nil || !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	
	userID, _ := claims["sub"].(string)
	return userID, nil
}