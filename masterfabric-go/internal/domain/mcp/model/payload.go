package model

// Payload is the WebMCP → Go backend request shape (FINAL BOSS gist).
type Payload struct {
	Action   string         `json:"action"`
	ReviewID string         `json:"review_id,omitempty"`
	Query    string         `json:"query,omitempty"`
	Spec     map[string]any `json:"spec,omitempty"`
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
	Mode       string `json:"mode,omitempty"`
}
