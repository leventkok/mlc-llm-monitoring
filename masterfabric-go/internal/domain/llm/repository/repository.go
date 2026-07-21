package repository

import (
	"context"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
)

// ReviewRepository persists reviews, decisions, scores, and metrics.
type ReviewRepository interface {
	CreateReview(ctx context.Context, review model.Review) error
	GetReviewForUser(ctx context.Context, id, userID string) (model.Review, error)
	ListReviews(ctx context.Context, userID string, limit, offset int) ([]model.Review, error)
	CreateDecision(ctx context.Context, decision model.Decision, userID string) error
	ListDecisions(ctx context.Context, userID string, limit, offset int) ([]model.Decision, error)
	CreateScore(ctx context.Context, score model.Score, userID string) error
	ListScores(ctx context.Context, userID string, limit, offset int) ([]model.Score, error)
	GetMetrics(ctx context.Context, userID string) (model.Metrics, error)
}
