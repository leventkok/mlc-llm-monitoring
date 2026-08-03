package org

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

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.org.ListMembersForUser(r.Context(), userID)
	if err != nil {
		response.LegacyError(w, http.StatusForbidden, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, list)
}

func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.org.ListInvitesForMember(r.Context(), userID)
	if err != nil {
		response.LegacyError(w, http.StatusForbidden, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req orgDTO.CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Role == "" {
		req.Role = "company_member"
	}
	inv, err := h.org.CreateInviteForMember(r.Context(), userID, req)
	if err != nil {
		response.LegacyError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, inv)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	targetUserID := chi.URLParam(r, "userId")
	if targetUserID == "" {
		response.LegacyError(w, http.StatusBadRequest, "user id required")
		return
	}
	if err := h.org.RemoveMemberForAdmin(r.Context(), userID, targetUserID); err != nil {
		status := http.StatusBadRequest
		switch err.Error() {
		case "organization membership required", "company admin role required":
			status = http.StatusForbidden
		case "member not found":
			status = http.StatusNotFound
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, orgDTO.MessageResponse{Message: "member removed and account deleted"})
}
