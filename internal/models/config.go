package models

type Config struct {
	AppName string `json:"app_name"`
	Model   string `json:"model"`
	Version string `json:"version"`
}