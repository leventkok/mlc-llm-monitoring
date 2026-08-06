package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	llmUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/usecase"
)

// BackfillMissingAutoScores creates quality scores for decisions saved before auto-scoring existed.
func BackfillMissingAutoScores(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `
		SELECT d.id, d.category, d.sentiment, COALESCE(d.raw_output, ''), d.latency_ms, r.user_id::text
		FROM decisions d
		JOIN reviews r ON r.id = d.review_id
		LEFT JOIN scores s ON s.decision_id = d.id
		WHERE s.id IS NULL`)
	if err != nil {
		return fmt.Errorf("list decisions without scores: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			decisionID, category, sentiment, rawOutput, userID string
			latencyMs                                          int
		)
		if err := rows.Scan(&decisionID, &category, &sentiment, &rawOutput, &latencyMs, &userID); err != nil {
			return fmt.Errorf("scan decision for score backfill: %w", err)
		}
		quality := llmUC.ComputeAutoQuality(category, sentiment, rawOutput, latencyMs)
		if _, err := pool.Exec(ctx,
			`INSERT INTO scores (id, decision_id, quality, scored_by) VALUES ($1, $2, $3, $4)`,
			uuid.NewString(), decisionID, quality, userID,
		); err != nil {
			return fmt.Errorf("backfill score for decision %s: %w", decisionID, err)
		}
	}
	return rows.Err()
}
