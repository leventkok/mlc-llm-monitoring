package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/leventkok/mlc-llm-monitoring/internal/models"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
)

type ConfigHandler struct {
	store *storage.ConfigStore
}

func NewConfigHandler(store *storage.ConfigStore) *ConfigHandler {
	return &ConfigHandler{store: store}
}

func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.Get())
}

func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req models.Config
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.AppName == "" || req.Model == "" || req.Version == "" {
		writeError(w, http.StatusBadRequest, "app_name, model and version are required")
		return
	}

	updated := h.store.Update(req)
	writeJSON(w, http.StatusOK, updated)
}
