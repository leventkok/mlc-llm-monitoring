package usecase

import (
	"context"
	"errors"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/mlc"
	pgLlm "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/llm"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/metrics"
)

// ReviewClassifier runs inference for review text.
type ReviewClassifier interface {
	ClassifyReview(ctx context.Context, text string) (mlc.Classification, error)
}

type AnalyzeReviewUseCase struct {
	reviews    repository.ReviewRepository
	classifier ReviewClassifier
}

func NewAnalyzeReviewUseCase(reviews repository.ReviewRepository, classifier ReviewClassifier) *AnalyzeReviewUseCase {
	return &AnalyzeReviewUseCase{reviews: reviews, classifier: classifier}
}

func (uc *AnalyzeReviewUseCase) Execute(ctx context.Context, userID, reviewID string) (model.Decision, error) {
	if uc.classifier == nil {
		return model.Decision{}, errors.New("mlc inference is not configured")
	}

	review, err := uc.reviews.GetReviewForUser(ctx, reviewID, userID)
	if err != nil {
		if errors.Is(err, pgLlm.ErrNotFound) {
			return model.Decision{}, errors.New("review not found")
		}
		return model.Decision{}, errors.New("could not load review")
	}

	classification, err := uc.classifier.ClassifyReview(ctx, review.Text)
	if err != nil {
		metrics.RecordAnalyzeError()
		return model.Decision{}, errors.New("inference failed")
	}

	createUC := NewCreateDecisionUseCase(uc.reviews)
	decision, err := createUC.Execute(ctx, userID, dto.SaveDecisionRequest{
		ReviewID:  reviewID,
		Category:  classification.Category,
		Sentiment: classification.Sentiment,
		RawOutput: classification.RawOutput,
		LatencyMs: classification.LatencyMs,
	})
	if err != nil {
		return model.Decision{}, err
	}

	metrics.RecordAnalyzeSuccess(classification.LatencyMs)
	return decision, nil
}
