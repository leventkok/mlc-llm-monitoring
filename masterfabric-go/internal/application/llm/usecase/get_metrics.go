package usecase

import (
	"context"
	"errors"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
)

type GetMetricsUseCase struct {
	reviews repository.ReviewRepository
}

func NewGetMetricsUseCase(reviews repository.ReviewRepository) *GetMetricsUseCase {
	return &GetMetricsUseCase{reviews: reviews}
}

func (uc *GetMetricsUseCase) Execute(ctx context.Context, userID string) (model.Metrics, error) {
	metrics, err := uc.reviews.GetMetrics(ctx, userID)
	if err != nil {
		return model.Metrics{}, errors.New("could not compute metrics")
	}
	return metrics, nil
}
