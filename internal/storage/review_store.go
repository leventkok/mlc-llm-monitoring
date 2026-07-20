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
		`INSERT INTO reviews (id, app_name, store, rating, text)
		 VALUES ($1, $2, $3, $4, $5)`,
		r.ID, r.AppName, r.Store, r.Rating, r.Text,
	)
	return err
}

func (s *PostgresStore) ListReviews() ([]models.Review, error) {
	rows, err := s.pool.Query(
		context.Background(),
		`SELECT id, app_name, store, rating, text, created_at
		 FROM reviews ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		if err := rows.Scan(&r.ID, &r.AppName, &r.Store, &r.Rating, &r.Text, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

func (s *PostgresStore) GetReview(id string) (models.Review, error) {
	var r models.Review
	err := s.pool.QueryRow(
		context.Background(),
		`SELECT id, app_name, store, rating, text, created_at FROM reviews WHERE id = $1`,
		id,
	).Scan(&r.ID, &r.AppName, &r.Store, &r.Rating, &r.Text, &r.CreatedAt)
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

func (s *PostgresStore) ListDecisions() ([]models.Decision, error) {
	rows, err := s.pool.Query(
		context.Background(),
		`SELECT id, review_id, category, sentiment, raw_output, latency_ms, created_at
		 FROM decisions ORDER BY created_at DESC`,
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

func (s *PostgresStore) ListScores() ([]models.Score, error) {
	rows, err := s.pool.Query(
		context.Background(),
		`SELECT id, decision_id, quality, COALESCE(correct_category, ''), COALESCE(scored_by::text, ''), created_at
		 FROM scores ORDER BY created_at DESC`,
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