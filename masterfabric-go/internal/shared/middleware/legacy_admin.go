package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/admin"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type legacyAdminKey struct{}

// LegacyRequireAdmin restricts routes to users listed in ADMIN_USERNAMES.
func LegacyRequireAdmin(users repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := LegacyUserID(r.Context())
			if !ok {
				response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			id, err := uuid.Parse(userID)
			if err != nil {
				response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			user, err := users.GetByID(r.Context(), id)
			if err != nil || user == nil {
				response.LegacyError(w, http.StatusForbidden, "admin access required")
				return
			}

			if !admin.IsPlatformAdmin(user) {
				response.LegacyError(w, http.StatusForbidden, "admin access required")
				return
			}

			ctx := context.WithValue(r.Context(), legacyAdminKey{}, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
