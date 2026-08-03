package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	AccountKindIndividual = "individual"
	AccountKindCompany    = "company"

	PlatformRoleUser  = "user"
	PlatformRoleAdmin = "platform_admin"
)

// User represents an application user (email + username auth).
type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	PasswordHash  string    `json:"-"`
	AccountKind   string    `json:"account_kind"`
	PlatformRole  string    `json:"platform_role"`
	CreatedAt     time.Time `json:"created_at"`
}
