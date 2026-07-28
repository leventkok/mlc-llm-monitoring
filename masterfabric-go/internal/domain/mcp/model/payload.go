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
	ReviewID  string `json:"review_id,omitempty"`
	Category  string `json:"category"`
	Sentiment string `json:"sentiment"`
	RawOutput string `json:"raw_output"`
	LatencyMs int    `json:"latency_ms"`
	DecisionID string `json:"decision_id,omitempty"`
}
