package config

import (
	"encoding/json"
	"net/http"

	configUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/config/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type Handler struct {
	getUC    *configUC.GetConfigUseCase
	updateUC *configUC.UpdateConfigUseCase
}

func NewHandler(getUC *configUC.GetConfigUseCase, updateUC *configUC.UpdateConfigUseCase) *Handler {
	return &Handler{getUC: getUC, updateUC: updateUC}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	response.LegacyJSON(w, http.StatusOK, h.getUC.Execute())
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.Config
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	updated, err := h.updateUC.Execute(req)
	if err != nil {
		response.LegacyError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, updated)
}
