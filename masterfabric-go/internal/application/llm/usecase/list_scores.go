package usecase

import (
	"context"
	"errors"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
)

type ListScoresUseCase struct {
	reviews repository.ReviewRepository
}

func NewListScoresUseCase(reviews repository.ReviewRepository) *ListScoresUseCase {
	return &ListScoresUseCase{reviews: reviews}
}

func (uc *ListScoresUseCase) Execute(ctx context.Context, userID string, limit, offset int) ([]model.Score, error) {
	scores, err := uc.reviews.ListScores(ctx, userID, limit, offset)
	if err != nil {
		return nil, errors.New("could not list scores")
	}
	if scores == nil {
		scores = []model.Score{}
	}
	return scores, nil
}
