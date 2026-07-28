package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	configUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/config/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/analyzelog"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type Handler struct {
	getLLMUC    *configUC.GetLLMConfigUseCase
	updateLLMUC *configUC.UpdateLLMConfigUseCase
}

func NewHandler(getLLM *configUC.GetLLMConfigUseCase, updateLLM *configUC.UpdateLLMConfigUseCase) *Handler {
	return &Handler{getLLMUC: getLLM, updateLLMUC: updateLLM}
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
