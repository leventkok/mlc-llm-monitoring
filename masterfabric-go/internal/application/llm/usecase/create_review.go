package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/validate"
)

type CreateReviewUseCase struct {
	reviews repository.ReviewRepository
}

func NewCreateReviewUseCase(reviews repository.ReviewRepository) *CreateReviewUseCase {
	return &CreateReviewUseCase{reviews: reviews}
}

func (uc *CreateReviewUseCase) Execute(ctx context.Context, userID string, req dto.CreateReviewRequest) (model.Review, error) {
	if req.Text == "" || req.AppName == "" {
		return model.Review{}, errors.New("app_name and text are required")
	}
	if err := validate.MaxLen("app_name", req.AppName, validate.MaxAppName); err != nil {
		return model.Review{}, err
	}
	if err := validate.MaxLen("text", req.Text, validate.MaxReviewText); err != nil {
		return model.Review{}, err
	}
	if err := validate.Store(req.Store); err != nil {
		return model.Review{}, err
	}
	if err := validate.Rating(req.Rating); err != nil {
		return model.Review{}, err
	}

	review := model.Review{
		ID:      uuid.NewString(),
		UserID:  userID,
		AppName: req.AppName,
		Store:   req.Store,
		Rating:  req.Rating,
		Text:    req.Text,
	}
	if err := uc.reviews.CreateReview(ctx, review); err != nil {
		return model.Review{}, errors.New("could not save review")
	}
	return review, nil
}
