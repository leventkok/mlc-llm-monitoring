package auth

import (
	"context"
	"errors"
	"runtime"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/semaphore"
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 12 characters")
	bcryptSem           = semaphore.NewWeighted(int64(max(4, runtime.NumCPU())))
)

func ValidatePassword(pw string) error {
	if len(pw) < 12 {
		return ErrPasswordTooShort
	}
	if len(pw) > 72 {
		return errors.New("password must be at most 72 characters")
	}
	return nil
}

func HashPassword(pw string) ([]byte, error) {
	if err := bcryptSem.Acquire(context.Background(), 1); err != nil {
		return nil, err
	}
	defer bcryptSem.Release(1)
	return bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
}

func CheckPassword(hash, pw string) error {
	if err := bcryptSem.Acquire(context.Background(), 1); err != nil {
		return err
	}
	defer bcryptSem.Release(1)
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

// DummyHash is used on failed login to reduce email enumeration via timing.
var DummyHash = func() []byte {
	h, err := bcrypt.GenerateFromPassword([]byte("timing-padding-secret"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return h
}()
