package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
	pgLlm "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/llm"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/validate"
)

type CreateScoreUseCase struct {
	reviews repository.ReviewRepository
}

func NewCreateScoreUseCase(reviews repository.ReviewRepository) *CreateScoreUseCase {
	return &CreateScoreUseCase{reviews: reviews}
}

func (uc *CreateScoreUseCase) Execute(ctx context.Context, userID string, req dto.CreateScoreRequest) (model.Score, error) {
	if req.Quality < 1 || req.Quality > 5 {
		return model.Score{}, errors.New("quality must be between 1 and 5")
	}
	if req.CorrectCategory != "" {
		if err := validate.Category(req.CorrectCategory); err != nil {
			return model.Score{}, err
		}
	}

	score := model.Score{
		ID:              uuid.NewString(),
		DecisionID:      req.DecisionID,
		Quality:         req.Quality,
		CorrectCategory: req.CorrectCategory,
		ScoredBy:        userID,
	}
	if err := uc.reviews.CreateScore(ctx, score, userID); err != nil {
		if errors.Is(err, pgLlm.ErrAlreadyScored) {
			return model.Score{}, errors.New("decision already scored")
		}
		if errors.Is(err, pgLlm.ErrNotFound) {
			return model.Score{}, errors.New("decision not found")
		}
		return model.Score{}, errors.New("could not save score")
	}
	return score, nil
}
