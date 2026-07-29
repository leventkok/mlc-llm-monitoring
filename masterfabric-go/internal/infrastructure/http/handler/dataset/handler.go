package dataset

import (
	"net/http"
	"strconv"

	datasetUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/dataset/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

type Handler struct {
	listUC   *datasetUC.ListDatasetReviewsUseCase
	exportUC *datasetUC.ExportDatasetCSVUseCase
	batchUC  *datasetUC.BatchAnalyzeDatasetUseCase
}

func NewHandler(
	list *datasetUC.ListDatasetReviewsUseCase,
	export *datasetUC.ExportDatasetCSVUseCase,
	batch *datasetUC.BatchAnalyzeDatasetUseCase,
) *Handler {
	return &Handler{listUC: list, exportUC: export, batchUC: batch}
}

func parsePaging(r *http.Request) (offset, limit int) {
	offset = 0
	limit = 100
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			offset = n
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 500 {
		limit = 500
	}
	return offset, limit
}

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	offset, limit := parsePaging(r)
	page, err := h.listUC.Execute(r.Context(), offset, limit)
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, page)
}

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	offset, limit := parsePaging(r)
	csv, err := h.exportUC.Execute(r.Context(), offset, limit)
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csv)
}

func (h *Handler) BatchAnalyze(w http.ResponseWriter, r *http.Request) {
	offset, limit := parsePaging(r)
	if limit > 50 {
		limit = 50
	}
	result, err := h.batchUC.Execute(r.Context(), offset, limit)
	if err != nil {
		response.LegacyError(w, http.StatusBadGateway, err.Error())
		return
	}
	response.LegacyJSON(w, http.StatusOK, result)
}
