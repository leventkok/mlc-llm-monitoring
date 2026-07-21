package model

import "time"

// Review is a user-submitted app store review.
type Review struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	AppName   string    `json:"app_name"`
	Store     string    `json:"store"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// Decision is an LLM classification result for a review.
type Decision struct {
	ID        string    `json:"id"`
	ReviewID  string    `json:"review_id"`
	Category  string    `json:"category"`
	Sentiment string    `json:"sentiment"`
	RawOutput string    `json:"raw_output"`
	LatencyMs int       `json:"latency_ms"`
	CreatedAt time.Time `json:"created_at"`
}

// Score is a human quality rating for a decision.
type Score struct {
	ID              string    `json:"id"`
	DecisionID      string    `json:"decision_id"`
	Quality         int       `json:"quality"`
	CorrectCategory string    `json:"correct_category,omitempty"`
	ScoredBy        string    `json:"scored_by"`
	CreatedAt       time.Time `json:"created_at"`
}

// Metrics aggregates review monitoring statistics for a user.
type Metrics struct {
	TotalReviews    int            `json:"total_reviews"`
	TotalDecisions  int            `json:"total_decisions"`
	TotalScores     int            `json:"total_scores"`
	CategoryCounts  map[string]int `json:"category_counts"`
	SentimentCounts map[string]int `json:"sentiment_counts"`
	AvgQuality      float64        `json:"avg_quality"`
	AvgLatencyMs    float64        `json:"avg_latency_ms"`
	AccuracyPct     float64        `json:"accuracy_pct"`
}
