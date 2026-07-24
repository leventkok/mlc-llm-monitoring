package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var (
	inferenceTotal          uint64
	inferenceLatencyMsTotal uint64
)

func main() {
	addr := envOr("LISTEN_ADDR", ":8080")
	upstreamRaw := envOr("MLC_UPSTREAM", "http://mlc-engine:8081")
	expectedKey := strings.TrimSpace(os.Getenv("MLC_API_KEY"))

	upstream, err := url.Parse(upstreamRaw)
	if err != nil {
		log.Fatalf("invalid MLC_UPSTREAM %q: %v", upstreamRaw, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error %s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if err := probeUpstream(upstream); err != nil {
			http.Error(w, `{"status":"starting"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/metrics", handleMetrics)
	mux.Handle("/v1/", withAPIKey(expectedKey, trackInference(proxy)))

	log.Printf("mlc-proxy listening on %s -> %s", addr, upstreamRaw)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func probeUpstream(upstream *url.URL) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(upstream.String() + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstream health status %d", resp.StatusCode)
	}
	return nil
}

func withAPIKey(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expected != "" && r.Header.Get("X-MLC-API-Key") != expected {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func trackInference(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/chat/completions") {
			start := time.Now()
			next.ServeHTTP(w, r)
			atomic.AddUint64(&inferenceTotal, 1)
			atomic.AddUint64(&inferenceLatencyMsTotal, uint64(time.Since(start).Milliseconds()))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	total := atomic.LoadUint64(&inferenceTotal)
	latSum := atomic.LoadUint64(&inferenceLatencyMsTotal)
	_, _ = fmt.Fprintf(w, "# HELP mlc_inference_requests_total Total MLC inference requests.\n")
	_, _ = fmt.Fprintf(w, "# TYPE mlc_inference_requests_total counter\n")
	_, _ = fmt.Fprintf(w, "mlc_inference_requests_total %d\n", total)
	_, _ = fmt.Fprintf(w, "# HELP mlc_inference_latency_ms_sum Sum of inference latency in milliseconds.\n")
	_, _ = fmt.Fprintf(w, "# TYPE mlc_inference_latency_ms_sum counter\n")
	_, _ = fmt.Fprintf(w, "mlc_inference_latency_ms_sum %d\n", latSum)
}
