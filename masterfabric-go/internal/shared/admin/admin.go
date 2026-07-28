package admin

import (
	"os"
	"strings"
)

const defaultAdminUsernames = "admin"

// IsAdmin reports whether username is listed in ADMIN_USERNAMES (comma-separated).
func IsAdmin(username string) bool {
	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" {
		return false
	}
	for _, allowed := range AdminUsernames() {
		if username == allowed {
			return true
		}
	}
	return false
}

// AdminUsernames returns lowercase admin usernames from env.
func AdminUsernames() []string {
	raw := os.Getenv("ADMIN_USERNAMES")
	if strings.TrimSpace(raw) == "" {
		raw = defaultAdminUsernames
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
