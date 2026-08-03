package orgadmin

import (
	"encoding/json"
	"net/http"

	orgDTO "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/org/dto"
	orgUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/org/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/middleware"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	org *orgUC.Service
}

func NewHandler(org *orgUC.Service) *Handler {
	return &Handler{org: org}
}

func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	list, err := h.org.ListOrganizations(r.Context())
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req orgDTO.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	org, err := h.org.CreateOrganization(r.Context(), req.Name)
	if err != nil {
		response.LegacyError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, org)
}

func (h *Handler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	if err := h.org.DeleteOrganization(r.Context(), orgID); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "organization not found" {
			status = http.StatusNotFound
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, orgDTO.MessageResponse{Message: "organization deleted"})
}

func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	list, err := h.org.ListInvites(r.Context(), orgID)
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.LegacyUserID(r.Context())
	orgID := chi.URLParam(r, "orgId")
	var req orgDTO.CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Role == "" {
		req.Role = "company_admin"
	}
	inv, err := h.org.CreateInvite(r.Context(), orgID, userID, req)
	if err != nil {
		response.LegacyError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, inv)
}
