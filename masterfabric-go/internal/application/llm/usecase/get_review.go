package usecase

import (
	"context"
	"errors"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
	pgLlm "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/llm"
)

type GetReviewUseCase struct {
	reviews repository.ReviewRepository
}

func NewGetReviewUseCase(reviews repository.ReviewRepository) *GetReviewUseCase {
	return &GetReviewUseCase{reviews: reviews}
}

func (uc *GetReviewUseCase) Execute(ctx context.Context, userID, reviewID string) (model.Review, error) {
	review, err := uc.reviews.GetReviewForUser(ctx, reviewID, userID)
	if err != nil {
		if errors.Is(err, pgLlm.ErrNotFound) {
			return model.Review{}, errors.New("review not found")
		}
		return model.Review{}, errors.New("review not found")
	}
	return review, nil
}
