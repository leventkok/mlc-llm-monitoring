package agent

import (
	"encoding/json"
	"net/http"

	configUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/config/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type Handler struct {
	agentUC *configUC.AgentModelSwitchUseCase
}

func NewHandler(agentUC *configUC.AgentModelSwitchUseCase) *Handler {
	return &Handler{agentUC: agentUC}
}

// ClaimNext returns the oldest pending switch for the local mlc-agent (X-MLC-API-Key).
func (h *Handler) ClaimNext(w http.ResponseWriter, r *http.Request) {
	req, err := h.agentUC.ClaimNext(r.Context())
	if err != nil {
		response.LegacyError(w, http.StatusInternalServerError, "could not claim switch request")
		return
	}
	if req == nil {
		response.LegacyJSON(w, http.StatusOK, map[string]any{"request": nil})
		return
	}
	response.LegacyJSON(w, http.StatusOK, map[string]any{"request": req})
}

type reportBody struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.LegacyError(w, http.StatusBadRequest, "id required")
		return
	}
	var body reportBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.agentUC.Report(r.Context(), id, body.Status, body.ErrorMessage); err != nil {
		response.LegacyError(w, http.StatusInternalServerError, "could not update switch status")
		return
	}
	response.LegacyJSON(w, http.StatusOK, map[string]string{"status": body.Status})
}
