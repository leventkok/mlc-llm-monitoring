package dto

type StoreApp struct {
	Store     string `json:"store"`
	AppID     string `json:"app_id"`
	AppName   string `json:"app_name"`
	Developer string `json:"developer"`
	IconURL   string `json:"icon_url,omitempty"`
}

type SearchAppsRequest struct {
	Query   string `json:"query"`
	Country string `json:"country,omitempty"`
	Lang    string `json:"lang,omitempty"`
}

type SearchAppsResponse struct {
	Play     []StoreApp `json:"play"`
	AppStore []StoreApp `json:"appstore"`
}

type CreateAuditRequest struct {
	ClientName          string `json:"client_name"`
	AppDisplayName      string `json:"app_display_name"`
	PlayAppID           string `json:"play_app_id"`
	AppStoreAppID       string `json:"appstore_app_id"`
	Country             string `json:"country,omitempty"`
	Lang                string `json:"lang,omitempty"`
	PlayReviewLimit     int    `json:"play_review_limit,omitempty"`
	AppStoreReviewLimit int    `json:"appstore_review_limit,omitempty"`
	Mode                string `json:"mode"`
}

type AuditResponse struct {
	ID              string `json:"id"`
	ClientName      string `json:"client_name"`
	AppDisplayName  string `json:"app_display_name"`
	PlayAppID           string `json:"play_app_id,omitempty"`
	AppStoreAppID       string `json:"appstore_app_id,omitempty"`
	Country             string `json:"country,omitempty"`
	Lang                string `json:"lang,omitempty"`
	PlayReviewLimit     int    `json:"play_review_limit,omitempty"`
	AppStoreReviewLimit int    `json:"appstore_review_limit,omitempty"`
	Mode                string `json:"mode"`
	Status          string `json:"status"`
	Step            string `json:"step"`
	PlayFetched     int    `json:"play_fetched"`
	AppStoreFetched int    `json:"appstore_fetched"`
	TotalReviews    int    `json:"total_reviews"`
	AnalyzedCount   int    `json:"analyzed_count"`
	Truncated       bool   `json:"truncated"`
	ErrorMessage    string `json:"error_message,omitempty"`
	StartedAt       string `json:"started_at,omitempty"`
	CompletedAt     string `json:"completed_at,omitempty"`
	CreatedAt       string `json:"created_at"`
}

type ReportResponse struct {
	Audit    AuditResponse `json:"audit"`
	Insights *InsightsDTO  `json:"insights,omitempty"`
}

type InsightsDTO struct {
	ExecutiveSummary string                 `json:"executive_summary"`
	Statistics       map[string]any         `json:"statistics"`
	RootCauses       []map[string]any       `json:"root_causes"`
	ActionPlan       []map[string]any       `json:"action_plan"`
	ReportMeta       map[string]any         `json:"report_meta,omitempty"`
	GeneratedAt      string                 `json:"generated_at"`
}
