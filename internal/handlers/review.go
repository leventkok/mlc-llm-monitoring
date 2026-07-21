package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/leventkok/mlc-llm-monitoring/internal/middleware"
	"github.com/leventkok/mlc-llm-monitoring/internal/models"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
	"github.com/leventkok/mlc-llm-monitoring/internal/validate"
)

type ReviewHandler struct {
	store *storage.PostgresStore
}

func NewReviewHandler(store *storage.PostgresStore) *ReviewHandler {
	return &ReviewHandler{store: store}
}

type createReviewRequest struct {
	AppName string `json:"app_name"`
	Store   string `json:"store"`
	Rating  int    `json:"rating"`
	Text    string `json:"text"`
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Text == "" || req.AppName == "" {
		writeError(w, http.StatusBadRequest, "app_name and text are required")
		return
	}
	if err := validate.MaxLen("app_name", req.AppName, validate.MaxAppName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.MaxLen("text", req.Text, validate.MaxReviewText); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Store(req.Store); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Rating(req.Rating); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	review := models.Review{
		ID:      uuid.NewString(),
		UserID:  userID,
		AppName: req.AppName,
		Store:   req.Store,
		Rating:  req.Rating,
		Text:    req.Text,
	}
	if err := h.store.CreateReview(r.Context(), review); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save review")
		return
	}

	writeJSON(w, http.StatusCreated, review)
}

func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "review id required")
		return
	}

	review, err := h.store.GetReviewForUser(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}

	writeJSON(w, http.StatusOK, review)
}

func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := listParams(r)
	reviews, err := h.store.ListReviews(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list reviews")
		return
	}
	if reviews == nil {
		reviews = []models.Review{}
	}
	writeJSON(w, http.StatusOK, reviews)
}

func (h *ReviewHandler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := listParams(r)
	decisions, err := h.store.ListDecisions(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list decisions")
		return
	}
	if decisions == nil {
		decisions = []models.Decision{}
	}
	writeJSON(w, http.StatusOK, decisions)
}

type createScoreRequest struct {
	DecisionID      string `json:"decision_id"`
	Quality         int    `json:"quality"`
	CorrectCategory string `json:"correct_category"`
}

func (h *ReviewHandler) CreateScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Quality < 1 || req.Quality > 5 {
		writeError(w, http.StatusBadRequest, "quality must be between 1 and 5")
		return
	}
	if req.CorrectCategory != "" {
		if err := validate.Category(req.CorrectCategory); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	score := models.Score{
		ID:              uuid.NewString(),
		DecisionID:      req.DecisionID,
		Quality:         req.Quality,
		CorrectCategory: req.CorrectCategory,
		ScoredBy:        userID,
	}
	if err := h.store.CreateScore(r.Context(), score, userID); err != nil {
		if errors.Is(err, storage.ErrAlreadyScored) {
			writeError(w, http.StatusConflict, "decision already scored")
			return
		}
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "decision not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not save score")
		return
	}

	writeJSON(w, http.StatusCreated, score)
}

func (h *ReviewHandler) ListScores(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := listParams(r)
	scores, err := h.store.ListScores(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list scores")
		return
	}
	if scores == nil {
		scores = []models.Score{}
	}
	writeJSON(w, http.StatusOK, scores)
}

func (h *ReviewHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	metrics, err := h.store.GetMetrics(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not compute metrics")
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

type saveDecisionRequest struct {
	ReviewID  string `json:"review_id"`
	Category  string `json:"category"`
	Sentiment string `json:"sentiment"`
	RawOutput string `json:"raw_output"`
	LatencyMs int    `json:"latency_ms"`
}

func (h *ReviewHandler) SaveDecision(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req saveDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.ReviewID == "" || req.Category == "" || req.Sentiment == "" {
		writeError(w, http.StatusBadRequest, "review_id, category and sentiment are required")
		return
	}
	if err := validate.Category(req.Category); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Sentiment(req.Sentiment); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.MaxLen("raw_output", req.RawOutput, validate.MaxRawOutput); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	decision := models.Decision{
		ID:        uuid.NewString(),
		ReviewID:  req.ReviewID,
		Category:  req.Category,
		Sentiment: req.Sentiment,
		RawOutput: req.RawOutput,
		LatencyMs: req.LatencyMs,
	}
	if err := h.store.CreateDecision(r.Context(), decision, userID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "review not found")
			return
		}
		if errors.Is(err, storage.ErrDuplicateDecision) {
			writeError(w, http.StatusConflict, "decision already exists for this review")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not save decision")
		return
	}

	writeJSON(w, http.StatusCreated, decision)
}

func listParams(r *http.Request) (limit, offset int) {
	q := r.URL.Query()
	limit = validate.ListLimit(parseInt(q.Get("limit"), validate.DefaultListLimit))
	offset = validate.ListOffset(parseInt(q.Get("offset"), 0))
	return limit, offset
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
