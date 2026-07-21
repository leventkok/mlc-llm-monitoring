package usecase

import (
	"context"
	"errors"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
)

type ListReviewsUseCase struct {
	reviews repository.ReviewRepository
}

func NewListReviewsUseCase(reviews repository.ReviewRepository) *ListReviewsUseCase {
	return &ListReviewsUseCase{reviews: reviews}
}

func (uc *ListReviewsUseCase) Execute(ctx context.Context, userID string, limit, offset int) ([]model.Review, error) {
	reviews, err := uc.reviews.ListReviews(ctx, userID, limit, offset)
	if err != nil {
		return nil, errors.New("could not list reviews")
	}
	if reviews == nil {
		reviews = []model.Review{}
	}
	return reviews, nil
}
