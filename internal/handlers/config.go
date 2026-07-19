package handlers

import (
	"encoding/json"
	"net/http"
)

func Config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"app_name": "mlc-llm-monitoring",
		"model":    "gemma-3-1b-it",
		"version":  "0.1.0",
	})
}