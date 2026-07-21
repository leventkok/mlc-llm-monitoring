package response

import (
	"encoding/json"
	"net/http"
)

// LegacyJSON writes legacy API JSON responses: {"error":"..."} on failure paths.
type legacyError struct {
	Error string `json:"error"`
}

func LegacyJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func LegacyError(w http.ResponseWriter, status int, message string) {
	LegacyJSON(w, status, legacyError{Error: message})
}
