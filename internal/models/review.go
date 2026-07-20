package models

import "time"

type Review struct {
	ID        string    `json:"id"`
	AppName   string    `json:"app_name"`
	Store     string    `json:"store"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Decision struct {
	ID        string    `json:"id"`
	ReviewID  string    `json:"review_id"`
	Category  string    `json:"category"`  
	Sentiment string    `json:"sentiment"` 
	RawOutput string    `json:"raw_output"`
	LatencyMs int       `json:"latency_ms"`
	CreatedAt time.Time `json:"created_at"`
}

type Score struct {
	ID              string    `json:"id"`
	DecisionID      string    `json:"decision_id"`
	Quality         int       `json:"quality"`                     
	CorrectCategory string    `json:"correct_category,omitempty"`  
	ScoredBy        string    `json:"scored_by"`
	CreatedAt       time.Time `json:"created_at"`
}