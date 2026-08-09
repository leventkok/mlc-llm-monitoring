package model

import "time"

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"

	StepCrawl    = "crawl"
	StepClassify = "classify"
	StepInsights = "insights"
	StepDone     = "done"

	ModeQuick = "quick"
	ModeFull  = "full"
)

type Audit struct {
	ID              string
	UserID          string
	OrgID           string
	ClientName      string
	AppDisplayName  string
	PlayAppID       string
	AppStoreAppID   string
	Mode            string
	Status          string
	Step            string
	PlayFetched     int
	AppStoreFetched int
	TotalReviews    int
	AnalyzedCount   int
	Truncated       bool
	ErrorMessage    string
	StartedAt       *time.Time
	CompletedAt     *time.Time
	CreatedAt       time.Time
}

type Review struct {
	ID            string
	AuditID       string
	Store         string
	StoreReviewID string
	AppName       string
	Rating        int
	Text          string
	ReviewedAt    *time.Time
	Category      string
	Sentiment     string
	RawOutput     string
}

type Statistics struct {
	TotalReviews   int                `json:"total_reviews"`
	PlayCount      int                `json:"play_count"`
	AppStoreCount  int                `json:"appstore_count"`
	AvgRating      float64            `json:"avg_rating"`
	AvgRatingPlay  float64            `json:"avg_rating_play"`
	AvgRatingAppStore float64         `json:"avg_rating_appstore"`
	LowStarPct     float64            `json:"low_star_pct"`
	CategoryCounts map[string]int     `json:"category_counts"`
	SentimentCounts map[string]int    `json:"sentiment_counts"`
	RatingCounts   map[string]int     `json:"rating_counts"`
}

type RootCause struct {
	Theme         string   `json:"theme"`
	Description   string   `json:"description"`
	AffectedRating string  `json:"affected_rating"`
	SampleCount   int      `json:"sample_count"`
	Examples      []string `json:"examples"`
}

type ActionItem struct {
	Priority       string `json:"priority"`
	HorizonDays    string `json:"horizon_days"`
	Action         string `json:"action"`
	OwnerHint      string `json:"owner_hint"`
	ExpectedImpact string `json:"expected_impact"`
}

type Insights struct {
	AuditID           string
	ExecutiveSummary  string
	Statistics        Statistics
	RootCauses        []RootCause
	ActionPlan        []ActionItem
	GeneratedAt       time.Time
}
