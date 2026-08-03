package invite

import (
	"net/http"

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

func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		response.LegacyError(w, http.StatusBadRequest, "token required")
		return
	}
	preview, err := h.org.PreviewInvite(r.Context(), token)
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, preview)
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	token := chi.URLParam(r, "token")
	if token == "" {
		response.LegacyError(w, http.StatusBadRequest, "token required")
		return
	}
	result, err := h.org.AcceptInvite(r.Context(), token, userID)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "invite invalid or expired" {
			status = http.StatusGone
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, result)
}
