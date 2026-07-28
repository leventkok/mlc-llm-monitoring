package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	configUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/config/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/analyzelog"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/middleware"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type Handler struct {
	getLLMUC       *configUC.GetLLMConfigUseCase
	updateLLMUC    *configUC.UpdateLLMConfigUseCase
	listProfilesUC *configUC.ListModelProfilesUseCase
	switchModelUC  *configUC.SwitchModelUseCase
	switchStatusUC *configUC.GetModelSwitchStatusUseCase
}

func NewHandler(
	getLLM *configUC.GetLLMConfigUseCase,
	updateLLM *configUC.UpdateLLMConfigUseCase,
	listProfiles *configUC.ListModelProfilesUseCase,
	switchModel *configUC.SwitchModelUseCase,
	switchStatus *configUC.GetModelSwitchStatusUseCase,
) *Handler {
	return &Handler{
		getLLMUC:       getLLM,
		updateLLMUC:    updateLLM,
		listProfilesUC: listProfiles,
		switchModelUC:  switchModel,
		switchStatusUC: switchStatus,
	}
}

func (h *Handler) GetLLMConfig(w http.ResponseWriter, r *http.Request) {
	response.LegacyJSON(w, http.StatusOK, h.getLLMUC.Execute())
}

func (h *Handler) UpdateLLMConfig(w http.ResponseWriter, r *http.Request) {
	var req model.LLMConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	updated, err := h.updateLLMUC.Execute(req)
	if err != nil {
		response.LegacyError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, updated)
}

func (h *Handler) AnalyzeLogs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	response.LegacyJSON(w, http.StatusOK, analyzelog.Recent(limit))
}

func (h *Handler) ListModelProfiles(w http.ResponseWriter, r *http.Request) {
	response.LegacyJSON(w, http.StatusOK, h.listProfilesUC.Execute())
}

type switchModelBody struct {
	ProfileID string `json:"profile_id"`
}

func (h *Handler) SwitchModel(w http.ResponseWriter, r *http.Request) {
	var body switchModelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.ProfileID == "" {
		response.LegacyError(w, http.StatusBadRequest, "profile_id is required")
		return
	}
	userID, _ := middleware.LegacyUserID(r.Context())
	req, err := h.switchModelUC.Execute(r.Context(), body.ProfileID, userID)
	if err != nil {
		response.LegacyError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusAccepted, req)
}

func (h *Handler) ModelSwitchStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.switchStatusUC.Execute(r.Context())
	if err != nil {
		response.LegacyError(w, http.StatusInternalServerError, "could not load switch status")
		return
	}
	response.LegacyJSON(w, http.StatusOK, status)
}
