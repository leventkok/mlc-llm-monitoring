package dto

import "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"

type CreateReviewRequest struct {
	AppName string `json:"app_name"`
	Store   string `json:"store"`
	Rating  int    `json:"rating"`
	Text    string `json:"text"`
}

type SaveDecisionRequest struct {
	ReviewID  string `json:"review_id"`
	Category  string `json:"category"`
	Sentiment string `json:"sentiment"`
	RawOutput string `json:"raw_output"`
	LatencyMs int    `json:"latency_ms"`
}

type CreateScoreRequest struct {
	DecisionID      string `json:"decision_id"`
	Quality         int    `json:"quality"`
	CorrectCategory string `json:"correct_category"`
}

type ReviewResponse = model.Review
type DecisionResponse = model.Decision
type ScoreResponse = model.Score
type MetricsResponse = model.Metrics
