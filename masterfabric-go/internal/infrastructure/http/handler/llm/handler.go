package llm

import (
	"encoding/json"
	"net/http"
	"strconv"

	llmUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/middleware"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/validate"
)

type Handler struct {
	createReviewUC   *llmUC.CreateReviewUseCase
	getReviewUC      *llmUC.GetReviewUseCase
	listReviewsUC    *llmUC.ListReviewsUseCase
	analyzeReviewUC  *llmUC.AnalyzeReviewUseCase
	createDecisionUC *llmUC.CreateDecisionUseCase
	listDecisionsUC  *llmUC.ListDecisionsUseCase
	createScoreUC    *llmUC.CreateScoreUseCase
	listScoresUC     *llmUC.ListScoresUseCase
	getMetricsUC     *llmUC.GetMetricsUseCase
}

func NewHandler(
	createReviewUC *llmUC.CreateReviewUseCase,
	getReviewUC *llmUC.GetReviewUseCase,
	listReviewsUC *llmUC.ListReviewsUseCase,
	analyzeReviewUC *llmUC.AnalyzeReviewUseCase,
	createDecisionUC *llmUC.CreateDecisionUseCase,
	listDecisionsUC *llmUC.ListDecisionsUseCase,
	createScoreUC *llmUC.CreateScoreUseCase,
	listScoresUC *llmUC.ListScoresUseCase,
	getMetricsUC *llmUC.GetMetricsUseCase,
) *Handler {
	return &Handler{
		createReviewUC:   createReviewUC,
		getReviewUC:      getReviewUC,
		listReviewsUC:    listReviewsUC,
		analyzeReviewUC:  analyzeReviewUC,
		createDecisionUC: createDecisionUC,
		listDecisionsUC:  listDecisionsUC,
		createScoreUC:    createScoreUC,
		listScoresUC:     listScoresUC,
		getMetricsUC:     getMetricsUC,
	}
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	review, err := h.createReviewUC.Execute(r.Context(), userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() != "could not save review" {
			status = http.StatusBadRequest
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, review)
}

func (h *Handler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		response.LegacyError(w, http.StatusBadRequest, "review id required")
		return
	}

	review, err := h.getReviewUC.Execute(r.Context(), userID, id)
	if err != nil {
		response.LegacyError(w, http.StatusNotFound, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, review)
}

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := listParams(r)
	reviews, err := h.listReviewsUC.Execute(r.Context(), userID, limit, offset)
	if err != nil {
		response.LegacyError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, reviews)
}

func (h *Handler) AnalyzeReview(w http.ResponseWriter, r *http.Request) {
	if h.analyzeReviewUC == nil {
		response.LegacyError(w, http.StatusServiceUnavailable, "server-side inference is not configured")
		return
	}

	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		response.LegacyError(w, http.StatusBadRequest, "review id required")
		return
	}

	decision, err := h.analyzeReviewUC.Execute(r.Context(), userID, id)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "review not found":
			status = http.StatusNotFound
		case "decision already exists for this review":
			status = http.StatusConflict
		case "mlc inference is not configured", "inference failed":
			status = http.StatusServiceUnavailable
		default:
			status = http.StatusBadRequest
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, decision)
}

func (h *Handler) SaveDecision(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.SaveDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	decision, err := h.createDecisionUC.Execute(r.Context(), userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "review not found":
			status = http.StatusNotFound
		case "decision already exists for this review":
			status = http.StatusConflict
		default:
			status = http.StatusBadRequest
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, decision)
}

func (h *Handler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := listParams(r)
	decisions, err := h.listDecisionsUC.Execute(r.Context(), userID, limit, offset)
	if err != nil {
		response.LegacyError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, decisions)
}

func (h *Handler) CreateScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	score, err := h.createScoreUC.Execute(r.Context(), userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "decision already scored":
			status = http.StatusConflict
		case "decision not found":
			status = http.StatusNotFound
		default:
			status = http.StatusBadRequest
		}
		response.LegacyError(w, status, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusCreated, score)
}

func (h *Handler) ListScores(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := listParams(r)
	scores, err := h.listScoresUC.Execute(r.Context(), userID, limit, offset)
	if err != nil {
		response.LegacyError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, scores)
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	metrics, err := h.getMetricsUC.Execute(r.Context(), userID)
	if err != nil {
		response.LegacyError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, metrics)
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
