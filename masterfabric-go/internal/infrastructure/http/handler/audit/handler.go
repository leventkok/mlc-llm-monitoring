package audit

import (
	"encoding/json"
	"net/http"

	auditDTO "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/audit/dto"
	auditUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/audit/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/middleware"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *auditUC.Service
}

func NewHandler(svc *auditUC.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var req auditDTO.SearchAppsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	out, err := h.svc.SearchApps(r.Context(), req.Query, req.Country, req.Lang)
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, out)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req auditDTO.CreateAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	out, err := h.svc.CreateAudit(r.Context(), userID, req)
	if err != nil {
		response.LegacyError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, out)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	out, err := h.svc.ListAudits(r.Context(), userID)
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	out, err := h.svc.GetAudit(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "audit not found" {
			status = http.StatusNotFound
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, out)
}

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	out, err := h.svc.GetReport(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "audit not found" {
			status = http.StatusNotFound
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, out)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.svc.DeleteAudit(r.Context(), userID, chi.URLParam(r, "id")); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "audit not found" {
			status = http.StatusNotFound
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
