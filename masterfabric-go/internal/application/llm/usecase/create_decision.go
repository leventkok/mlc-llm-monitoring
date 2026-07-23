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

type CreateDecisionUseCase struct {
	reviews repository.ReviewRepository
}

func NewCreateDecisionUseCase(reviews repository.ReviewRepository) *CreateDecisionUseCase {
	return &CreateDecisionUseCase{reviews: reviews}
}

func (uc *CreateDecisionUseCase) Execute(ctx context.Context, userID string, req dto.SaveDecisionRequest) (model.Decision, error) {
	if req.ReviewID == "" || req.Category == "" || req.Sentiment == "" {
		return model.Decision{}, errors.New("review_id, category and sentiment are required")
	}
	if err := validate.Category(req.Category); err != nil {
		return model.Decision{}, err
	}
	if err := validate.Sentiment(req.Sentiment); err != nil {
		return model.Decision{}, err
	}
	if err := validate.MaxLen("raw_output", req.RawOutput, validate.MaxRawOutput); err != nil {
		return model.Decision{}, err
	}

	decision := model.Decision{
		ID:        uuid.NewString(),
		ReviewID:  req.ReviewID,
		Category:  req.Category,
		Sentiment: req.Sentiment,
		RawOutput: req.RawOutput,
		LatencyMs: req.LatencyMs,
	}
	if err := uc.reviews.CreateDecision(ctx, decision, userID); err != nil {
		if errors.Is(err, pgLlm.ErrNotFound) {
			return model.Decision{}, errors.New("review not found")
		}
		if errors.Is(err, pgLlm.ErrDuplicateDecision) {
			return model.Decision{}, errors.New("decision already exists for this review")
		}
		return model.Decision{}, errors.New("could not save decision")
	}

	if err := persistAutoScore(ctx, uc.reviews, userID, decision); err != nil {
		return model.Decision{}, errors.New("could not save auto score")
	}

	return decision, nil
}
