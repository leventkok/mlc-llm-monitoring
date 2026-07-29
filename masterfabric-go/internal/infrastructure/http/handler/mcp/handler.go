package mcp

import (
	"encoding/json"
	"net/http"
	"strings"

	datasetUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/dataset/usecase"
	mcpUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/mcp/usecase"
	llmUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/usecase"
	mcpModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/mcp/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/middleware"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type Handler struct {
	analyzeReviewUC *llmUC.AnalyzeReviewUseCase
	deepWiki        mcpUC.DeepWikiQuerier
	batchDatasetUC  *datasetUC.BatchAnalyzeDatasetUseCase
}

func NewHandler(
	analyzeReviewUC *llmUC.AnalyzeReviewUseCase,
	deepWiki mcpUC.DeepWikiQuerier,
	batchDatasetUC *datasetUC.BatchAnalyzeDatasetUseCase,
) *Handler {
	return &Handler{
		analyzeReviewUC: analyzeReviewUC,
		deepWiki:        deepWiki,
		batchDatasetUC:  batchDatasetUC,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.LegacyUserID(r.Context())
	if !ok {
		response.LegacyError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req mcpModel.Payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.LegacyError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	result, err := mcpUC.HandleMCPRequest(r.Context(), userID, req, h.analyzeReviewUC, h.deepWiki, h.batchDatasetUC)
	if err != nil {
		status := http.StatusBadRequest
		switch err.Error() {
		case "mlc inference is not configured", "inference failed":
			status = http.StatusServiceUnavailable
		case "deepwiki is not configured":
			status = http.StatusServiceUnavailable
		case "review not found":
			status = http.StatusNotFound
		case "decision already exists for this review":
			status = http.StatusConflict
		case "unsupported mcp action", "review_id is required", "spec.repo is required (owner/name)", "query is required for ask mode", "dataset batch analyze is not configured":
			status = http.StatusBadRequest
		default:
			if strings.Contains(err.Error(), "Error fetching wiki") || strings.Contains(err.Error(), "Repository not found") {
				status = http.StatusNotFound
			}
		}
		response.LegacyError(w, status, err.Error())
		return
	}

	response.LegacyJSON(w, http.StatusOK, result)
}
