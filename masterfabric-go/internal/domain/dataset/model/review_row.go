package model

// ReviewRow is a remote Hugging Face dataset row (InferReview app review schema).
type ReviewRow struct {
	ReviewID   string `json:"review_id"`
	Store      string `json:"store"`
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
	Rating     int    `json:"rating"`
	Text       string `json:"text"`
	Language   string `json:"language"`
	Category   string `json:"category,omitempty"`
	Sentiment  string `json:"sentiment,omitempty"`
}

type ReviewPage struct {
	DatasetID string      `json:"dataset_id"`
	Offset    int         `json:"offset"`
	Limit     int         `json:"limit"`
	Rows      []ReviewRow `json:"rows"`
}

type BatchAnalyzeItem struct {
	ReviewID          string `json:"review_id"`
	Text              string `json:"text"`
	ExpectedCategory  string `json:"expected_category,omitempty"`
	ExpectedSentiment string `json:"expected_sentiment,omitempty"`
	Category          string `json:"category"`
	Sentiment         string `json:"sentiment"`
	Match             bool   `json:"match"`
	LatencyMs         int    `json:"latency_ms"`
}

type BatchAnalyzeResult struct {
	DatasetID string             `json:"dataset_id"`
	Processed int                `json:"processed"`
	Accuracy  float64            `json:"accuracy_pct"`
	Items     []BatchAnalyzeItem `json:"items"`
}
