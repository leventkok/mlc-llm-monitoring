package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlowLegacyRequest(t *testing.T) {
	cases := []struct {
		method string
		path   string
		slow   bool
	}{
		{http.MethodPost, "/mcp", true},
		{http.MethodPost, "/reviews/abc/analyze", true},
		{http.MethodGet, "/reviews", false},
		{http.MethodPost, "/reviews", false},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if got := slowLegacyRequest(req); got != tc.slow {
			t.Fatalf("%s %s: slow=%v want %v", tc.method, tc.path, got, tc.slow)
		}
	}
}

func TestRequestTimeoutLegacy_SlowRouteGetsLongerDeadline(t *testing.T) {
	var ctxDeadline time.Time
	handler := RequestTimeoutLegacy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Fatal("expected deadline")
		}
		ctxDeadline = deadline
	}))

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	remaining := time.Until(ctxDeadline)
	if remaining < 20*time.Second || remaining > 28*time.Second {
		t.Fatalf("expected ~28s deadline, remaining=%s", remaining)
	}
}
