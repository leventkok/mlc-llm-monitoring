package model

// Config holds runtime LLM configuration exposed to the frontend.
type Config struct {
	AppName string `json:"app_name"`
	Model   string `json:"model"`
	Version string `json:"version"`
}
