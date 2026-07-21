package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/internal/auth"
	"github.com/leventkok/mlc-llm-monitoring/internal/middleware"
	"github.com/leventkok/mlc-llm-monitoring/internal/models"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
	"github.com/leventkok/mlc-llm-monitoring/internal/validate"
)

type AuthHandler struct {
	store storage.UserStore
}

func NewAuthHandler(store storage.UserStore) *AuthHandler {
	return &AuthHandler{store: store}
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type messageResponse struct {
	Message string `json:"message"`
}

type validateResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST")
		return
	}

	var req registerRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	if err := validate.Email(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Username(req.Username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "password could not be processed")
		return
	}

	user := models.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	ctx := r.Context()
	if err := h.store.Create(ctx, user); err != nil {
		if errors.Is(err, storage.ErrEmailTaken) || errors.Is(err, storage.ErrUsernameTaken) {
			writeError(w, http.StatusConflict, "registration failed; check email and username")
			return
		}
		log.Printf("register error: %v", err)
		writeError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{
		ID: user.ID, Email: user.Email, Username: user.Username,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST")
		return
	}

	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	ctx := r.Context()

	user, err := h.store.FindByEmail(ctx, req.Email)
	if err != nil {
		_ = auth.CheckPassword(string(auth.DummyHash), req.Password)
		writeError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token could not be generated")
		return
	}

	middleware.SetSessionCookie(w, token)
	writeJSON(w, http.StatusOK, messageResponse{Message: "signed in"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.store.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID: user.ID, Email: user.Email, Username: user.Username,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, messageResponse{Message: "signed out"})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token could not be generated")
		return
	}

	middleware.SetSessionCookie(w, token)
	writeJSON(w, http.StatusOK, messageResponse{Message: "token refreshed"})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req changePasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	user, err := h.store.FindByID(ctx, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.OldPassword); err != nil {
		writeError(w, http.StatusUnauthorized, "could not change password")
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "password could not be processed")
		return
	}

	user.PasswordHash = string(newHash)
	if err := h.store.Update(ctx, user); err != nil {
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}

	middleware.ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, messageResponse{Message: "password changed; sign in again"})
}

type updateMeRequest struct {
	Username string `json:"username"`
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateMeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if err := validate.Username(req.Username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	user, err := h.store.FindByID(ctx, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if existing, err := h.store.FindByUsername(ctx, req.Username); err == nil && existing.ID != userID {
		writeError(w, http.StatusConflict, "this username is already taken")
		return
	}

	user.Username = req.Username
	if err := h.store.Update(ctx, user); err != nil {
		writeError(w, http.StatusInternalServerError, "could not be updated")
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID: user.ID, Email: user.Email, Username: user.Username,
	})
}

func (h *AuthHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.store.Delete(r.Context(), userID); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete account")
		return
	}

	middleware.ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, messageResponse{Message: "account deleted"})
}

func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, validateResponse{Valid: true, UserID: userID})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return err
	}
	return nil
}
