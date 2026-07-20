package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/leventkok/mlc-llm-monitoring/internal/llm"
	"github.com/leventkok/mlc-llm-monitoring/internal/middleware"
	"github.com/leventkok/mlc-llm-monitoring/internal/models"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
)


type ReviewHandler struct {
	store    *storage.PostgresStore
	analyzer llm.Analyzer
}

func NewReviewHandler(store *storage.PostgresStore, analyzer llm.Analyzer) *ReviewHandler {
	return &ReviewHandler{store: store, analyzer: analyzer}
}


type createReviewRequest struct {
	AppName string `json:"app_name"`
	Store   string `json:"store"`
	Rating  int    `json:"rating"`
	Text    string `json:"text"`
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reviews, err := h.store.ListReviews()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list reviews")
		return
	}
	json.NewEncoder(w).Encode(reviews)
}



type analyzeRequest struct {
	ReviewID string `json:"review_id"`
}


func (h *ReviewHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	review, err := h.store.GetReview(req.ReviewID)
	if err != nil {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}

	result, err := h.analyzer.Analyze(review.Text)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "analysis failed")
		return
	}

	decision := models.Decision{
		ID:        uuid.NewString(),
		ReviewID:  review.ID,
		Category:  result.Category,
		Sentiment: result.Sentiment,
		RawOutput: result.RawOutput,
		LatencyMs: result.LatencyMs,
	}
	if err := h.store.CreateDecision(decision); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save decision")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(decision)
}

func (h *ReviewHandler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decisions, err := h.store.ListDecisions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list decisions")
		return
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

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var req createScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Quality < 1 || req.Quality > 5 {
		writeError(w, http.StatusBadRequest, "quality must be between 1 and 5")
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
	scores, err := h.store.ListScores()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list scores")
		return
	}
	json.NewEncoder(w).Encode(scores)
}