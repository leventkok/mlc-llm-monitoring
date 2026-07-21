package storage

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameTaken  = errors.New("username already taken")
	ErrEmailTaken     = errors.New("email already taken")
	ErrNotFound       = errors.New("not found")
	ErrAlreadyScored  = errors.New("decision already scored")
	ErrDuplicateDecision = errors.New("decision already exists for review")
)
