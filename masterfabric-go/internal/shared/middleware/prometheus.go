package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/metrics"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Prometheus records HTTP request KPIs when metrics are registered.
func Prometheus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}
		metrics.RecordHTTP(r.Method, route, sw.status, time.Since(start))
	})
}
