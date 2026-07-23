package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector holds application KPI metrics for Prometheus/Grafana.
type Collector struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	ReviewsCreated      prometheus.Counter
	AnalyzeTotal        *prometheus.CounterVec
	InferenceDuration   prometheus.Histogram
	DecisionsTotal      *prometheus.CounterVec
	AutoScoreQuality    prometheus.Histogram
}

var active *Collector

// Register creates and registers KPI metrics. Safe to call once at startup.
func Register() *Collector {
	if active != nil {
		return active
	}

	c := &Collector{
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "app_review_http_requests_total",
			Help: "Total HTTP requests handled by the API.",
		}, []string{"method", "route", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "app_review_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		ReviewsCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "app_review_reviews_created_total",
			Help: "Total reviews created.",
		}),
		AnalyzeTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "app_review_llm_analyze_total",
			Help: "LLM analyze attempts by outcome.",
		}, []string{"status"}),
		InferenceDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "app_review_llm_inference_duration_seconds",
			Help:    "End-to-end LLM inference latency for analyze.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60, 120},
		}),
		DecisionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "app_review_decisions_total",
			Help: "Classification decisions by category and sentiment.",
		}, []string{"category", "sentiment"}),
		AutoScoreQuality: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "app_review_auto_score_quality",
			Help:    "Automatic decision quality scores (1-5).",
			Buckets: []float64{1, 2, 3, 4, 5},
		}),
	}

	prometheus.MustRegister(
		c.HTTPRequestsTotal,
		c.HTTPRequestDuration,
		c.ReviewsCreated,
		c.AnalyzeTotal,
		c.InferenceDuration,
		c.DecisionsTotal,
		c.AutoScoreQuality,
	)

	active = c
	return c
}

func C() *Collector {
	return active
}

func RecordHTTP(method, route string, status int, duration time.Duration) {
	if active == nil {
		return
	}
	statusLabel := prometheus.Labels{
		"method": method,
		"route":  route,
		"status": formatStatus(status),
	}
	active.HTTPRequestsTotal.With(statusLabel).Inc()
	active.HTTPRequestDuration.With(prometheus.Labels{
		"method": method,
		"route":  route,
	}).Observe(duration.Seconds())
}

func RecordReviewCreated() {
	if active != nil {
		active.ReviewsCreated.Inc()
	}
}

func RecordAnalyzeSuccess(latencyMs int) {
	if active == nil {
		return
	}
	active.AnalyzeTotal.WithLabelValues("success").Inc()
	if latencyMs > 0 {
		active.InferenceDuration.Observe(float64(latencyMs) / 1000)
	}
}

func RecordAnalyzeError() {
	if active != nil {
		active.AnalyzeTotal.WithLabelValues("error").Inc()
	}
}

func RecordDecision(category, sentiment string) {
	if active != nil {
		active.DecisionsTotal.WithLabelValues(category, sentiment).Inc()
	}
}

func RecordAutoScore(quality int) {
	if active != nil {
		active.AutoScoreQuality.Observe(float64(quality))
	}
}

func formatStatus(code int) string {
	return strconv.Itoa(code)
}
