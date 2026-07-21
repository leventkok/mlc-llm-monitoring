package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/leventkok/mlc-llm-monitoring/internal/models"
)

func (s *PostgresStore) CreateReview(ctx context.Context, r models.Review) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO reviews (id, user_id, app_name, store, rating, text)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		r.ID, r.UserID, r.AppName, r.Store, r.Rating, r.Text,
	)
	return err
}

func (s *PostgresStore) ListReviews(ctx context.Context, userID string, limit, offset int) ([]models.Review, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, app_name, store, rating, text, created_at
		 FROM reviews WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]models.Review, 0, min(limit, 64))
	for rows.Next() {
		var r models.Review
		if err := rows.Scan(&r.ID, &r.UserID, &r.AppName, &r.Store, &r.Rating, &r.Text, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

func (s *PostgresStore) GetReviewForUser(ctx context.Context, id, userID string) (models.Review, error) {
	var r models.Review
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, app_name, store, rating, text, created_at
		 FROM reviews WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&r.ID, &r.UserID, &r.AppName, &r.Store, &r.Rating, &r.Text, &r.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Review{}, ErrNotFound
		}
		return models.Review{}, err
	}
	return r, nil
}

func (s *PostgresStore) CreateDecision(ctx context.Context, d models.Decision, userID string) error {
	tag, err := s.pool.Exec(ctx,
		`INSERT INTO decisions (id, review_id, category, sentiment, raw_output, latency_ms)
		 SELECT $1, $2, $3, $4, $5, $6
		 FROM reviews WHERE id = $2 AND user_id = $7`,
		d.ID, d.ReviewID, d.Category, d.Sentiment, d.RawOutput, d.LatencyMs, userID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateDecision
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListDecisions(ctx context.Context, userID string, limit, offset int) ([]models.Decision, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT d.id, d.review_id, d.category, d.sentiment, d.raw_output, d.latency_ms, d.created_at
		 FROM decisions d
		 JOIN reviews r ON r.id = d.review_id
		 WHERE r.user_id = $1
		 ORDER BY d.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decisions := make([]models.Decision, 0, min(limit, 64))
	for rows.Next() {
		var d models.Decision
		if err := rows.Scan(&d.ID, &d.ReviewID, &d.Category, &d.Sentiment, &d.RawOutput, &d.LatencyMs, &d.CreatedAt); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	return decisions, rows.Err()
}

func (s *PostgresStore) CreateScore(ctx context.Context, sc models.Score, userID string) error {
	var correct pgtype.Text
	if sc.CorrectCategory != "" {
		correct = pgtype.Text{String: sc.CorrectCategory, Valid: true}
	}

	tag, err := s.pool.Exec(ctx,
		`INSERT INTO scores (id, decision_id, quality, correct_category, scored_by)
		 SELECT $1, $2, $3, $4, $5
		 FROM decisions d
		 JOIN reviews r ON r.id = d.review_id
		 WHERE d.id = $2 AND r.user_id = $6`,
		sc.ID, sc.DecisionID, sc.Quality, correct, sc.ScoredBy, userID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyScored
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListScores(ctx context.Context, userID string, limit, offset int) ([]models.Score, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT sc.id, sc.decision_id, sc.quality, COALESCE(sc.correct_category, ''), COALESCE(sc.scored_by::text, ''), sc.created_at
		 FROM scores sc
		 JOIN decisions d ON d.id = sc.decision_id
		 JOIN reviews r ON r.id = d.review_id
		 WHERE r.user_id = $1
		 ORDER BY sc.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make([]models.Score, 0, min(limit, 64))
	for rows.Next() {
		var sc models.Score
		if err := rows.Scan(&sc.ID, &sc.DecisionID, &sc.Quality, &sc.CorrectCategory, &sc.ScoredBy, &sc.CreatedAt); err != nil {
			return nil, err
		}
		scores = append(scores, sc)
	}
	return scores, rows.Err()
}

type Metrics struct {
	TotalReviews    int            `json:"total_reviews"`
	TotalDecisions  int            `json:"total_decisions"`
	TotalScores     int            `json:"total_scores"`
	CategoryCounts  map[string]int `json:"category_counts"`
	SentimentCounts map[string]int `json:"sentiment_counts"`
	AvgQuality      float64        `json:"avg_quality"`
	AvgLatencyMs    float64        `json:"avg_latency_ms"`
	AccuracyPct     float64        `json:"accuracy_pct"`
}

const metricsQuery = `
SELECT
  (SELECT count(*)::int FROM reviews WHERE user_id = $1),
  (SELECT count(*)::int FROM decisions d JOIN reviews r ON r.id = d.review_id WHERE r.user_id = $1),
  (SELECT count(*)::int FROM scores sc JOIN decisions d ON d.id = sc.decision_id JOIN reviews r ON r.id = d.review_id WHERE r.user_id = $1),
  (SELECT coalesce(avg(sc.quality), 0) FROM scores sc JOIN decisions d ON d.id = sc.decision_id JOIN reviews r ON r.id = d.review_id WHERE r.user_id = $1),
  (SELECT coalesce(avg(d.latency_ms), 0) FROM decisions d JOIN reviews r ON r.id = d.review_id WHERE r.user_id = $1),
  (SELECT count(*)::int FROM scores sc JOIN decisions d ON d.id = sc.decision_id JOIN reviews r ON r.id = d.review_id
     WHERE r.user_id = $1 AND sc.correct_category IS NOT NULL AND sc.correct_category <> ''),
  (SELECT count(*)::int FROM scores sc JOIN decisions d ON d.id = sc.decision_id JOIN reviews r ON r.id = d.review_id
     WHERE r.user_id = $1 AND sc.correct_category IS NOT NULL AND sc.correct_category <> '' AND sc.correct_category = d.category),
  (SELECT coalesce(json_object_agg(category, cnt), '{}')::text FROM (
      SELECT d.category, count(*)::int AS cnt FROM decisions d JOIN reviews r ON r.id = d.review_id WHERE r.user_id = $1 GROUP BY d.category
   ) s),
  (SELECT coalesce(json_object_agg(sentiment, cnt), '{}')::text FROM (
      SELECT d.sentiment, count(*)::int AS cnt FROM decisions d JOIN reviews r ON r.id = d.review_id WHERE r.user_id = $1 GROUP BY d.sentiment
   ) s)
`

func (s *PostgresStore) GetMetrics(ctx context.Context, userID string) (Metrics, error) {
	var m Metrics
	var catJSON, sentJSON string
	var totalGraded, correct int

	err := s.pool.QueryRow(ctx, metricsQuery, userID).Scan(
		&m.TotalReviews,
		&m.TotalDecisions,
		&m.TotalScores,
		&m.AvgQuality,
		&m.AvgLatencyMs,
		&totalGraded,
		&correct,
		&catJSON,
		&sentJSON,
	)
	if err != nil {
		return m, err
	}

	m.CategoryCounts = map[string]int{}
	m.SentimentCounts = map[string]int{}
	_ = json.Unmarshal([]byte(catJSON), &m.CategoryCounts)
	_ = json.Unmarshal([]byte(sentJSON), &m.SentimentCounts)

	if totalGraded > 0 {
		m.AccuracyPct = float64(correct) / float64(totalGraded) * 100
	}
	return m, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
