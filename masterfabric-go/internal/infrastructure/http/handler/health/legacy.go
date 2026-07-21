package health

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type legacyHealthResponse struct {
	Status string `json:"status"`
}

// LegacyHealth serves /health and /ready with legacy JSON shapes.
func LegacyHealth(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ready" {
			if pool == nil || pool.Ping(r.Context()) != nil {
				response.LegacyError(w, http.StatusServiceUnavailable, "database unavailable")
				return
			}
			response.LegacyJSON(w, http.StatusOK, legacyHealthResponse{Status: "ready"})
			return
		}
		response.LegacyJSON(w, http.StatusOK, legacyHealthResponse{Status: "ok"})
	}
}

// PingDB is used at startup.
func PingDB(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}
	return pool.Ping(ctx)
}
