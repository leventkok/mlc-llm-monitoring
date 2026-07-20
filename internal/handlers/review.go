package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/leventkok/mlc-llm-monitoring/internal/middleware"
	"github.com/leventkok/mlc-llm-monitoring/internal/models"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
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
	w.Header().Set("Content-Type", "application/json")

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

	review := models.Review{
		ID:      uuid.NewString(),
		UserID:  userID,
		AppName: req.AppName,
		Store:   req.Store,
		Rating:  req.Rating,
		Text:    req.Text,
	}
	if err := h.store.CreateReview(review); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save review")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	review, err := h.store.GetReviewForUser(id, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}

	json.NewEncoder(w).Encode(review)
}

func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	reviews, err := h.store.ListReviews(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list reviews")
		return
	}
	if reviews == nil {
		reviews = []models.Review{}
	}
	json.NewEncoder(w).Encode(reviews)
}

func (h *ReviewHandler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	decisions, err := h.store.ListDecisions(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list decisions")
		return
	}
	if decisions == nil {
		decisions = []models.Decision{}
	}
	json.NewEncoder(w).Encode(decisions)
}

type createScoreRequest struct {
	DecisionID      string `json:"decision_id"`
	Quality         int    `json:"quality"`
	CorrectCategory string `json:"correct_category"`
}

func (h *ReviewHandler) CreateScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	belongs, err := h.store.DecisionBelongsToUser(req.DecisionID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not verify decision")
		return
	}
	if !belongs {
		writeError(w, http.StatusNotFound, "decision not found")
		return
	}

	score := models.Score{
		ID:              uuid.NewString(),
		DecisionID:      req.DecisionID,
		Quality:         req.Quality,
		CorrectCategory: req.CorrectCategory,
		ScoredBy:        userID,
	}
	if err := h.store.CreateScore(score); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save score")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(score)
}

func (h *ReviewHandler) ListScores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	scores, err := h.store.ListScores(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list scores")
		return
	}
	if scores == nil {
		scores = []models.Score{}
	}
	json.NewEncoder(w).Encode(scores)
}

func (h *ReviewHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	metrics, err := h.store.GetMetrics(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not compute metrics")
		return
	}
	json.NewEncoder(w).Encode(metrics)
}

type saveDecisionRequest struct {
	ReviewID  string `json:"review_id"`
	Category  string `json:"category"`
	Sentiment string `json:"sentiment"`
	RawOutput string `json:"raw_output"`
	LatencyMs int    `json:"latency_ms"`
}

func (h *ReviewHandler) SaveDecision(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	if _, err := h.store.GetReviewForUser(req.ReviewID, userID); err != nil {
		writeError(w, http.StatusNotFound, "review not found")
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
	if err := h.store.CreateDecision(decision); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save decision")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(decision)
}
