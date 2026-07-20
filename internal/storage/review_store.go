package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/leventkok/mlc-llm-monitoring/internal/models"
)

func (s *PostgresStore) CreateReview(r models.Review) error {
	_, err := s.pool.Exec(
		context.Background(),
		`INSERT INTO reviews (id, user_id, app_name, store, rating, text)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		r.ID, r.UserID, r.AppName, r.Store, r.Rating, r.Text,
	)
	return err
}

func (s *PostgresStore) ListReviews(userID string) ([]models.Review, error) {
	rows, err := s.pool.Query(
		context.Background(),
		`SELECT id, user_id, app_name, store, rating, text, created_at
		 FROM reviews WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		if err := rows.Scan(&r.ID, &r.UserID, &r.AppName, &r.Store, &r.Rating, &r.Text, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

func (s *PostgresStore) GetReviewForUser(id, userID string) (models.Review, error) {
	var r models.Review
	err := s.pool.QueryRow(
		context.Background(),
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

func (s *PostgresStore) CreateDecision(d models.Decision) error {
	_, err := s.pool.Exec(
		context.Background(),
		`INSERT INTO decisions (id, review_id, category, sentiment, raw_output, latency_ms)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		d.ID, d.ReviewID, d.Category, d.Sentiment, d.RawOutput, d.LatencyMs,
	)
	return err
}

func (s *PostgresStore) ListDecisions(userID string) ([]models.Decision, error) {
	rows, err := s.pool.Query(
		context.Background(),
		`SELECT d.id, d.review_id, d.category, d.sentiment, d.raw_output, d.latency_ms, d.created_at
		 FROM decisions d
		 JOIN reviews r ON r.id = d.review_id
		 WHERE r.user_id = $1
		 ORDER BY d.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []models.Decision
	for rows.Next() {
		var d models.Decision
		if err := rows.Scan(&d.ID, &d.ReviewID, &d.Category, &d.Sentiment, &d.RawOutput, &d.LatencyMs, &d.CreatedAt); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	return decisions, rows.Err()
}

func (s *PostgresStore) CreateScore(sc models.Score) error {
	_, err := s.pool.Exec(
		context.Background(),
		`INSERT INTO scores (id, decision_id, quality, correct_category, scored_by)
		 VALUES ($1, $2, $3, $4, $5)`,
		sc.ID, sc.DecisionID, sc.Quality, nullIfEmpty(sc.CorrectCategory), sc.ScoredBy,
	)
	return err
}

func (s *PostgresStore) ListScores(userID string) ([]models.Score, error) {
	rows, err := s.pool.Query(
		context.Background(),
		`SELECT sc.id, sc.decision_id, sc.quality, COALESCE(sc.correct_category, ''), COALESCE(sc.scored_by::text, ''), sc.created_at
		 FROM scores sc
		 JOIN decisions d ON d.id = sc.decision_id
		 JOIN reviews r ON r.id = d.review_id
		 WHERE r.user_id = $1
		 ORDER BY sc.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []models.Score
	for rows.Next() {
		var sc models.Score
		if err := rows.Scan(&sc.ID, &sc.DecisionID, &sc.Quality, &sc.CorrectCategory, &sc.ScoredBy, &sc.CreatedAt); err != nil {
			return nil, err
		}
		scores = append(scores, sc)
	}
	return scores, rows.Err()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
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

func (s *PostgresStore) GetMetrics(userID string) (Metrics, error) {
	ctx := context.Background()
	var m Metrics
	m.CategoryCounts = map[string]int{}
	m.SentimentCounts = map[string]int{}

	s.pool.QueryRow(ctx, `SELECT count(*) FROM reviews WHERE user_id = $1`, userID).Scan(&m.TotalReviews)
	s.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM decisions d
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1`, userID).Scan(&m.TotalDecisions)
	s.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM scores sc
		JOIN decisions d ON d.id = sc.decision_id
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1`, userID).Scan(&m.TotalScores)

	catRows, err := s.pool.Query(ctx, `
		SELECT d.category, count(*)
		FROM decisions d
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1
		GROUP BY d.category`, userID)
	if err != nil {
		return m, err
	}
	for catRows.Next() {
		var cat string
		var n int
		catRows.Scan(&cat, &n)
		m.CategoryCounts[cat] = n
	}
	catRows.Close()

	sentRows, err := s.pool.Query(ctx, `
		SELECT d.sentiment, count(*)
		FROM decisions d
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1
		GROUP BY d.sentiment`, userID)
	if err != nil {
		return m, err
	}
	for sentRows.Next() {
		var sent string
		var n int
		sentRows.Scan(&sent, &n)
		m.SentimentCounts[sent] = n
	}
	sentRows.Close()

	s.pool.QueryRow(ctx, `
		SELECT COALESCE(avg(sc.quality), 0)
		FROM scores sc
		JOIN decisions d ON d.id = sc.decision_id
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1`, userID).Scan(&m.AvgQuality)
	s.pool.QueryRow(ctx, `
		SELECT COALESCE(avg(d.latency_ms), 0)
		FROM decisions d
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1`, userID).Scan(&m.AvgLatencyMs)

	var total, correct int
	s.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM scores sc
		JOIN decisions d ON d.id = sc.decision_id
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1
		  AND sc.correct_category IS NOT NULL AND sc.correct_category <> ''`, userID).Scan(&total)
	s.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM scores sc
		JOIN decisions d ON d.id = sc.decision_id
		JOIN reviews r ON r.id = d.review_id
		WHERE r.user_id = $1
		  AND sc.correct_category IS NOT NULL
		  AND sc.correct_category <> ''
		  AND sc.correct_category = d.category`, userID).Scan(&correct)
	if total > 0 {
		m.AccuracyPct = float64(correct) / float64(total) * 100
	}

	return m, nil
}

func (s *PostgresStore) DecisionBelongsToUser(decisionID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(
		context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM decisions d
			JOIN reviews r ON r.id = d.review_id
			WHERE d.id = $1 AND r.user_id = $2
		)`,
		decisionID, userID,
	).Scan(&exists)
	return exists, err
}
