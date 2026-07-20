package storage

import "errors"

var (
	ErrUserNotFound  = errors.New("User is not found")
	ErrUsernameTaken = errors.New("This username is already taken")
	ErrEmailTaken    = errors.New("This email is already taken")
	ErrNotFound      = errors.New("not found")
)
