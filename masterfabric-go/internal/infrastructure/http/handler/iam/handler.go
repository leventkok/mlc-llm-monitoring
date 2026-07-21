package iam

import (
	"encoding/json"
	"net/http"

	iamUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/dto"
	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/middleware"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

// Handler serves legacy /auth/* routes with original JSON shapes.
type Handler struct {
	registerUC        *iamUC.RegisterUseCase
	loginUC           *iamUC.LoginUseCase
	getMeUC           *iamUC.GetMeUseCase
	updateMeUC        *iamUC.UpdateMeUseCase
	deleteMeUC        *iamUC.DeleteMeUseCase
	refreshUC         *iamUC.RefreshUseCase
	changePasswordUC  *iamUC.ChangePasswordUseCase
}

func NewHandler(
	registerUC *iamUC.RegisterUseCase,
	loginUC *iamUC.LoginUseCase,
	getMeUC *iamUC.GetMeUseCase,
	updateMeUC *iamUC.UpdateMeUseCase,
	deleteMeUC *iamUC.DeleteMeUseCase,
	refreshUC *iamUC.RefreshUseCase,
	changePasswordUC *iamUC.ChangePasswordUseCase,
) *Handler {
	return &Handler{
		registerUC:       registerUC,
		loginUC:          loginUC,
		getMeUC:          getMeUC,
		updateMeUC:       updateMeUC,
		deleteMeUC:       deleteMeUC,
		refreshUC:        refreshUC,
		changePasswordUC: changePasswordUC,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	user, err := h.registerUC.Execute(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "registration failed; check email and username" {
			status = http.StatusConflict
		} else if err.Error() != "registration failed" {
			status = http.StatusBadRequest
		}
		response.LegacyError(w, status, err.Error())
		return
	}

	response.LegacyJSON(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	token, err := h.loginUC.Execute(r.Context(), req)
	if err != nil {
		response.LegacyError(w, http.StatusUnauthorized, err.Error())
		return
	}

	infraAuth.SetSessionCookie(w, token)
	response.LegacyJSON(w, http.StatusOK, dto.MessageResponse{Message: "signed in"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.getMeUC.Execute(r.Context(), userID)
	if err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UpdateMeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	user, err := h.updateMeUC.Execute(r.Context(), userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "this username is already taken" {
			status = http.StatusConflict
		} else if err.Error() == "user not found" {
			status = http.StatusNotFound
		} else {
			status = http.StatusBadRequest
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, user)
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.deleteMeUC.Execute(r.Context(), userID); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		response.LegacyError(w, status, err.Error())
		return
	}

	infraAuth.ClearSessionCookie(w)
	response.LegacyJSON(w, http.StatusOK, dto.MessageResponse{Message: "account deleted"})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	infraAuth.ClearSessionCookie(w)
	response.LegacyJSON(w, http.StatusOK, dto.MessageResponse{Message: "signed out"})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	token, err := h.refreshUC.Execute(r.Context(), userID)
	if err != nil {
		response.LegacyError(w, http.StatusInternalServerError, err.Error())
		return
	}

	infraAuth.SetSessionCookie(w, token)
	response.LegacyJSON(w, http.StatusOK, dto.MessageResponse{Message: "token refreshed"})
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	response.LegacyJSON(w, http.StatusOK, dto.ValidateResponse{Valid: true, UserID: userID})
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.ChangePasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if err := h.changePasswordUC.Execute(r.Context(), userID, req); err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "user not found":
			status = http.StatusNotFound
		case "could not change password":
			status = http.StatusUnauthorized
		default:
			status = http.StatusBadRequest
		}
		response.LegacyError(w, status, err.Error())
		return
	}

	infraAuth.ClearSessionCookie(w)
	response.LegacyJSON(w, http.StatusOK, dto.MessageResponse{Message: "password changed; sign in again"})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return err
	}
	return nil
}
