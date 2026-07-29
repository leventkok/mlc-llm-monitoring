package usecase

import (
	"context"
	"errors"
	"strings"

	datasetModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/dataset/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/hfdataset"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/mlc"
)

type DatasetReader interface {
	ListRows(ctx context.Context, offset, limit int) (datasetModel.ReviewPage, error)
	ExportCSV(ctx context.Context, offset, limit int) ([]byte, error)
	DatasetID() string
}

type ListDatasetReviewsUseCase struct {
	reader DatasetReader
}

func NewListDatasetReviewsUseCase(reader DatasetReader) *ListDatasetReviewsUseCase {
	return &ListDatasetReviewsUseCase{reader: reader}
}

func (uc *ListDatasetReviewsUseCase) Execute(ctx context.Context, offset, limit int) (datasetModel.ReviewPage, error) {
	return uc.reader.ListRows(ctx, offset, limit)
}

type ExportDatasetCSVUseCase struct {
	reader DatasetReader
}

func NewExportDatasetCSVUseCase(reader DatasetReader) *ExportDatasetCSVUseCase {
	return &ExportDatasetCSVUseCase{reader: reader}
}

func (uc *ExportDatasetCSVUseCase) Execute(ctx context.Context, offset, limit int) ([]byte, error) {
	return uc.reader.ExportCSV(ctx, offset, limit)
}

type BatchAnalyzeDatasetUseCase struct {
	reader DatasetReader
	mlc    *mlc.Client
}

func NewBatchAnalyzeDatasetUseCase(reader DatasetReader, mlcClient *mlc.Client) *BatchAnalyzeDatasetUseCase {
	return &BatchAnalyzeDatasetUseCase{reader: reader, mlc: mlcClient}
}

func (uc *BatchAnalyzeDatasetUseCase) Execute(ctx context.Context, offset, limit int) (datasetModel.BatchAnalyzeResult, error) {
	page, err := uc.reader.ListRows(ctx, offset, limit)
	if err != nil {
		return datasetModel.BatchAnalyzeResult{}, err
	}
	if uc.mlc == nil {
		return datasetModel.BatchAnalyzeResult{}, errMLCNotConfigured
	}

	items := make([]datasetModel.BatchAnalyzeItem, 0, len(page.Rows))
	matches := 0
	labeled := 0

	for _, row := range page.Rows {
		item := datasetModel.BatchAnalyzeItem{
			ReviewID:          row.ReviewID,
			Text:              row.Text,
			ExpectedCategory:  row.Category,
			ExpectedSentiment: row.Sentiment,
		}
		if strings.TrimSpace(row.Text) == "" {
			continue
		}
		cls, err := uc.mlc.ClassifyReview(ctx, row.Text)
		if err != nil {
			return datasetModel.BatchAnalyzeResult{}, err
		}
		item.Category = cls.Category
		item.Sentiment = cls.Sentiment
		item.LatencyMs = cls.LatencyMs
		if row.Category != "" && row.Sentiment != "" {
			labeled++
			if row.Category == cls.Category && row.Sentiment == cls.Sentiment {
				item.Match = true
				matches++
			}
		}
		items = append(items, item)
	}

	accuracy := 0.0
	if labeled > 0 {
		accuracy = float64(matches) / float64(labeled) * 100
	}

	return datasetModel.BatchAnalyzeResult{
		DatasetID: page.DatasetID,
		Processed: len(items),
		Accuracy:  accuracy,
		Items:     items,
	}, nil
}

var errMLCNotConfigured = errors.New("mlc inference is not configured")

// Ensure interfaces.
var _ DatasetReader = (*hfdataset.Client)(nil)
