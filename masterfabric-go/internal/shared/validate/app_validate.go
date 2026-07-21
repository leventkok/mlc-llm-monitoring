package validate

import (
	"errors"
	"strings"
)

var (
	ErrInvalidCategory  = errors.New("invalid category")
	ErrInvalidSentiment = errors.New("invalid sentiment")
	ErrInvalidRating    = errors.New("rating must be between 1 and 5")
	ErrInvalidStore     = errors.New("invalid store")
	ErrFieldTooLong     = errors.New("field exceeds maximum length")
)

var (
	allowedCategories = map[string]struct{}{
		"bug": {}, "feature": {}, "praise": {}, "spam": {}, "other": {},
	}
	allowedSentiments = map[string]struct{}{
		"positive": {}, "negative": {}, "neutral": {},
	}
	allowedStores = map[string]struct{}{
		"play": {}, "appstore": {},
	}
)

const (
	MaxReviewText    = 10_000
	MaxAppName       = 200
	MaxRawOutput     = 64 << 10
	MaxUsername      = 64
	MaxEmail         = 254
	DefaultListLimit = 100
	MaxListLimit     = 500
)

func Category(v string) error {
	if _, ok := allowedCategories[v]; !ok {
		return ErrInvalidCategory
	}
	return nil
}

func Sentiment(v string) error {
	if _, ok := allowedSentiments[v]; !ok {
		return ErrInvalidSentiment
	}
	return nil
}

func Rating(v int) error {
	if v < 1 || v > 5 {
		return ErrInvalidRating
	}
	return nil
}

func Store(v string) error {
	if v == "" {
		return nil
	}
	if _, ok := allowedStores[v]; !ok {
		return ErrInvalidStore
	}
	return nil
}

func MaxLen(field, value string, max int) error {
	if len(value) > max {
		return errors.New(field + ": " + ErrFieldTooLong.Error())
	}
	return nil
}

func Email(v string) error {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" || len(v) > MaxEmail || !strings.Contains(v, "@") {
		return errors.New("invalid email address")
	}
	return nil
}

func Username(v string) error {
	v = strings.TrimSpace(v)
	if v == "" || len(v) > MaxUsername {
		return errors.New("invalid username")
	}
	return nil
}

func ListLimit(limit int) int {
	if limit <= 0 {
		return DefaultListLimit
	}
	if limit > MaxListLimit {
		return MaxListLimit
	}
	return limit
}

func ListOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
