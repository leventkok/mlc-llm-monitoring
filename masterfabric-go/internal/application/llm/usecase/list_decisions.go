package usecase

import (
	"context"
	"errors"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
)

type ListDecisionsUseCase struct {
	reviews repository.ReviewRepository
}

func NewListDecisionsUseCase(reviews repository.ReviewRepository) *ListDecisionsUseCase {
	return &ListDecisionsUseCase{reviews: reviews}
}

func (uc *ListDecisionsUseCase) Execute(ctx context.Context, userID string, limit, offset int) ([]model.Decision, error) {
	decisions, err := uc.reviews.ListDecisions(ctx, userID, limit, offset)
	if err != nil {
		return nil, errors.New("could not list decisions")
	}
	if decisions == nil {
		decisions = []model.Decision{}
	}
	return decisions, nil
}
