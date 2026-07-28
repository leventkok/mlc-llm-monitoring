package model

import "time"

const (
	SwitchStatusPending   = "pending"
	SwitchStatusRunning   = "running"
	SwitchStatusCompleted = "completed"
	SwitchStatusFailed    = "failed"
)

// ModelSwitchRequest tracks an adapter/model swap queued for the local mlc-agent.
type ModelSwitchRequest struct {
	ID            string     `json:"id"`
	ProfileID     string     `json:"profile_id"`
	RequestModel  string     `json:"request_model"`
	EngineModel   string     `json:"engine_model"`
	LocalAdapter  string     `json:"local_adapter"`
	Status        string     `json:"status"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	RequestedBy   string     `json:"requested_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}
