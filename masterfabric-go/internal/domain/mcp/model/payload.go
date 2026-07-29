package model

// Payload is the WebMCP → Go backend request shape (FINAL BOSS gist).
type Payload struct {
	Action   string         `json:"action"`
	ReviewID string         `json:"review_id,omitempty"`
	Query    string         `json:"query,omitempty"`
	Spec     map[string]any `json:"spec,omitempty"`
	Offset   int            `json:"offset,omitempty"`
	Limit    int            `json:"limit,omitempty"`
}

// RichResult is the enriched MCP response returned to the frontend.
type RichResult struct {
	Action     string `json:"action,omitempty"`
	ReviewID   string `json:"review_id,omitempty"`
	Category   string `json:"category,omitempty"`
	Sentiment  string `json:"sentiment,omitempty"`
	RawOutput  string `json:"raw_output,omitempty"`
	LatencyMs  int    `json:"latency_ms,omitempty"`
	DecisionID string `json:"decision_id,omitempty"`
	RepoName   string `json:"repo_name,omitempty"`
	Query      string `json:"query,omitempty"`
	Answer     string `json:"answer,omitempty"`
	Mode          string  `json:"mode,omitempty"`
	DatasetID     string  `json:"dataset_id,omitempty"`
	Processed     int     `json:"processed,omitempty"`
	AccuracyPct   float64 `json:"accuracy_pct,omitempty"`
	Items         []BatchItem `json:"items,omitempty"`
}

// BatchItem is one row from dataset_batch_analyze MCP action.
type BatchItem struct {
	ReviewID          string `json:"review_id"`
	Text              string `json:"text,omitempty"`
	ExpectedCategory  string `json:"expected_category,omitempty"`
	ExpectedSentiment string `json:"expected_sentiment,omitempty"`
	Category          string `json:"category,omitempty"`
	Sentiment         string `json:"sentiment,omitempty"`
	Match             bool   `json:"match,omitempty"`
	LatencyMs         int    `json:"latency_ms,omitempty"`
}
