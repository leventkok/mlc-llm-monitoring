package llm

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	domainErr "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/errors"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyScored     = errors.New("decision already scored")
	ErrDuplicateDecision = errors.New("decision already exists for review")
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) CreateReview(ctx context.Context, review model.Review) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reviews (id, user_id, app_name, store, rating, text)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		review.ID, review.UserID, review.AppName, review.Store, review.Rating, review.Text,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "could not save review", err)
	}
	return nil
}

func (r *ReviewRepo) ListReviews(ctx context.Context, userID string, limit, offset int) ([]model.Review, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, app_name, store, rating, text, created_at
		 FROM reviews WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "could not list reviews", err)
	}
	defer rows.Close()

	reviews := make([]model.Review, 0, min(limit, 64))
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(&rv.ID, &rv.UserID, &rv.AppName, &rv.Store, &rv.Rating, &rv.Text, &rv.CreatedAt); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "could not scan review", err)
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}

func (r *ReviewRepo) GetReviewForUser(ctx context.Context, id, userID string) (model.Review, error) {
	var rv model.Review
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, app_name, store, rating, text, created_at
		 FROM reviews WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&rv.ID, &rv.UserID, &rv.AppName, &rv.Store, &rv.Rating, &rv.Text, &rv.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Review{}, ErrNotFound
		}
		return model.Review{}, domainErr.New(domainErr.ErrInternal, "could not get review", err)
	}
	return rv, nil
}

func (r *ReviewRepo) CreateDecision(ctx context.Context, d model.Decision, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO decisions (id, review_id, category, sentiment, raw_output, latency_ms)
		 SELECT $1, $2, $3, $4, $5, $6
		 FROM reviews WHERE id = $2 AND user_id = $7`,
		d.ID, d.ReviewID, d.Category, d.Sentiment, d.RawOutput, d.LatencyMs, userID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateDecision
		}
		return domainErr.New(domainErr.ErrInternal, "could not save decision", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ReviewRepo) ListDecisions(ctx context.Context, userID string, limit, offset int) ([]model.Decision, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT d.id, d.review_id, d.category, d.sentiment, d.raw_output, d.latency_ms, d.created_at
		 FROM decisions d
		 JOIN reviews r ON r.id = d.review_id
		 WHERE r.user_id = $1
		 ORDER BY d.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "could not list decisions", err)
	}
	defer rows.Close()

	decisions := make([]model.Decision, 0, min(limit, 64))
	for rows.Next() {
		var d model.Decision
		if err := rows.Scan(&d.ID, &d.ReviewID, &d.Category, &d.Sentiment, &d.RawOutput, &d.LatencyMs, &d.CreatedAt); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "could not scan decision", err)
		}
		decisions = append(decisions, d)
	}
	return decisions, rows.Err()
}

func (r *ReviewRepo) CreateScore(ctx context.Context, sc model.Score, userID string) error {
	var correct pgtype.Text
	if sc.CorrectCategory != "" {
		correct = pgtype.Text{String: sc.CorrectCategory, Valid: true}
	}

	tag, err := r.pool.Exec(ctx,
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
		return domainErr.New(domainErr.ErrInternal, "could not save score", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ReviewRepo) ListScores(ctx context.Context, userID string, limit, offset int) ([]model.Score, error) {
	rows, err := r.pool.Query(ctx,
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
		return nil, domainErr.New(domainErr.ErrInternal, "could not list scores", err)
	}
	defer rows.Close()

	scores := make([]model.Score, 0, min(limit, 64))
	for rows.Next() {
		var sc model.Score
		if err := rows.Scan(&sc.ID, &sc.DecisionID, &sc.Quality, &sc.CorrectCategory, &sc.ScoredBy, &sc.CreatedAt); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "could not scan score", err)
		}
		scores = append(scores, sc)
	}
	return scores, rows.Err()
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

func (r *ReviewRepo) GetMetrics(ctx context.Context, userID string) (model.Metrics, error) {
	var m model.Metrics
	var catJSON, sentJSON string
	var totalGraded, correct int

	err := r.pool.QueryRow(ctx, metricsQuery, userID).Scan(
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
		return m, domainErr.New(domainErr.ErrInternal, "could not compute metrics", err)
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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
