package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents an application user (email + username auth).
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
