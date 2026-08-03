package repository

import (
	"context"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
)

// ReviewRepository persists reviews, decisions, scores, and metrics.
type ReviewRepository interface {
	CreateReview(ctx context.Context, review model.Review) error
	GetReviewForScope(ctx context.Context, id string, scope model.ReviewScope) (model.Review, error)
	ListReviews(ctx context.Context, scope model.ReviewScope, limit, offset int) ([]model.Review, error)
	CreateDecision(ctx context.Context, decision model.Decision, scope model.ReviewScope) error
	ListDecisions(ctx context.Context, scope model.ReviewScope, limit, offset int) ([]model.Decision, error)
	CreateScore(ctx context.Context, score model.Score, scope model.ReviewScope) error
	ListScores(ctx context.Context, scope model.ReviewScope, limit, offset int) ([]model.Score, error)
	GetMetrics(ctx context.Context, scope model.ReviewScope) (model.Metrics, error)
}
