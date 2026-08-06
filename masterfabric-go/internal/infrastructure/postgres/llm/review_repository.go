package llm

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

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
	var orgID *string
	if review.OrgID != "" {
		orgID = &review.OrgID
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reviews (id, user_id, org_id, app_name, store, rating, text)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		review.ID, review.UserID, orgID, review.AppName, review.Store, review.Rating, review.Text,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "could not save review", err)
	}
	return nil
}

func (r *ReviewRepo) ListReviews(ctx context.Context, scope model.ReviewScope, limit, offset int) ([]model.Review, error) {
	query, args := scopeListQuery(
		`SELECT id, user_id, COALESCE(org_id::text, ''), app_name, store, rating, text, created_at FROM reviews`,
		scope, limit, offset,
	)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "could not list reviews", err)
	}
	defer rows.Close()
	return scanReviews(rows)
}

func (r *ReviewRepo) GetReviewForScope(ctx context.Context, id string, scope model.ReviewScope) (model.Review, error) {
	query, args := scopeOneQuery(
		`SELECT id, user_id, COALESCE(org_id::text, ''), app_name, store, rating, text, created_at FROM reviews WHERE id = $1`,
		id, scope,
	)
	var rv model.Review
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&rv.ID, &rv.UserID, &rv.OrgID, &rv.AppName, &rv.Store, &rv.Rating, &rv.Text, &rv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Review{}, ErrNotFound
		}
		return model.Review{}, domainErr.New(domainErr.ErrInternal, "could not get review", err)
	}
	return rv, nil
}

func (r *ReviewRepo) CreateDecision(ctx context.Context, d model.Decision, scope model.ReviewScope) error {
	reviewFilter, reviewArgs := reviewExistsFilter(scope, 7)
	args := []any{d.ID, d.ReviewID, d.Category, d.Sentiment, d.RawOutput, d.LatencyMs}
	args = append(args, reviewArgs...)

	tag, err := r.pool.Exec(ctx,
		`INSERT INTO decisions (id, review_id, category, sentiment, raw_output, latency_ms)
		 SELECT $1, $2, $3, $4, $5, $6
		 FROM reviews WHERE id = $2 AND `+reviewFilter,
		args...,
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

func (r *ReviewRepo) ListDecisions(ctx context.Context, scope model.ReviewScope, limit, offset int) ([]model.Decision, error) {
	reviewFilter, filterArgs := reviewJoinFilter(scope, 3)
	args := append([]any{limit, offset}, filterArgs...)

	rows, err := r.pool.Query(ctx,
		`SELECT d.id, d.review_id, d.category, d.sentiment, d.raw_output, d.latency_ms, d.created_at
		 FROM decisions d
		 JOIN reviews r ON r.id = d.review_id
		 WHERE `+reviewFilter+`
		 ORDER BY d.created_at DESC
		 LIMIT $1 OFFSET $2`,
		args...,
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

func (r *ReviewRepo) CreateScore(ctx context.Context, sc model.Score, scope model.ReviewScope) error {
	var correct pgtype.Text
	if sc.CorrectCategory != "" {
		correct = pgtype.Text{String: sc.CorrectCategory, Valid: true}
	}
	reviewFilter, filterArgs := reviewJoinFilter(scope, 6)
	args := []any{sc.ID, sc.DecisionID, sc.Quality, correct, sc.ScoredBy}
	args = append(args, filterArgs...)

	tag, err := r.pool.Exec(ctx,
		`INSERT INTO scores (id, decision_id, quality, correct_category, scored_by)
		 SELECT $1, $2, $3, $4, $5
		 FROM decisions d
		 JOIN reviews r ON r.id = d.review_id
		 WHERE d.id = $2 AND `+reviewFilter,
		args...,
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

func (r *ReviewRepo) ListScores(ctx context.Context, scope model.ReviewScope, limit, offset int) ([]model.Score, error) {
	reviewFilter, filterArgs := reviewJoinFilter(scope, 3)
	args := append([]any{limit, offset}, filterArgs...)

	rows, err := r.pool.Query(ctx,
		`SELECT sc.id, sc.decision_id, sc.quality, COALESCE(sc.correct_category, ''), COALESCE(sc.scored_by::text, ''), sc.created_at
		 FROM scores sc
		 JOIN decisions d ON d.id = sc.decision_id
		 JOIN reviews r ON r.id = d.review_id
		 WHERE `+reviewFilter+`
		 ORDER BY sc.created_at DESC
		 LIMIT $1 OFFSET $2`,
		args...,
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

func (r *ReviewRepo) GetMetrics(ctx context.Context, scope model.ReviewScope) (model.Metrics, error) {
	reviewWhere, args := reviewJoinWhereArgs(scope)
	decisionJoin := `decisions d JOIN reviews r ON r.id = d.review_id`
	scoreJoin := `scores sc JOIN decisions d ON d.id = sc.decision_id JOIN reviews r ON r.id = d.review_id`

	query := `
SELECT
  (SELECT count(*)::int FROM reviews r WHERE ` + reviewWhere + `),
  (SELECT count(*)::int FROM ` + decisionJoin + ` WHERE ` + reviewWhere + `),
  (SELECT count(*)::int FROM ` + scoreJoin + ` WHERE ` + reviewWhere + `),
  (SELECT coalesce(avg(sc.quality), 0) FROM ` + scoreJoin + ` WHERE ` + reviewWhere + `),
  (SELECT coalesce(avg(d.latency_ms), 0) FROM ` + decisionJoin + ` WHERE ` + reviewWhere + `),
  (SELECT count(*)::int FROM ` + scoreJoin + ` WHERE ` + reviewWhere + ` AND sc.quality >= 4),
  (SELECT coalesce(json_object_agg(category, cnt), '{}')::text FROM (
      SELECT d.category, count(*)::int AS cnt FROM ` + decisionJoin + ` WHERE ` + reviewWhere + ` GROUP BY d.category
   ) s),
  (SELECT coalesce(json_object_agg(sentiment, cnt), '{}')::text FROM (
      SELECT d.sentiment, count(*)::int AS cnt FROM ` + decisionJoin + ` WHERE ` + reviewWhere + ` GROUP BY d.sentiment
   ) s)`

	expanded := make([]any, 0, len(args)*8)
	for i := 0; i < 8; i++ {
		expanded = append(expanded, args...)
	}

	var m model.Metrics
	var catJSON, sentJSON string
	var compliant int
	err := r.pool.QueryRow(ctx, query, expanded...).Scan(
		&m.TotalReviews,
		&m.TotalDecisions,
		&m.TotalScores,
		&m.AvgQuality,
		&m.AvgLatencyMs,
		&compliant,
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
	if m.TotalScores > 0 {
		m.AccuracyPct = float64(compliant) / float64(m.TotalScores) * 100
	}
	return m, nil
}

func reviewWhereArgs(scope model.ReviewScope) (string, []any) {
	if scope.IsOrg() {
		// Org workspace + member's legacy personal rows (org_id NULL before backfill).
		return `(org_id = $1 OR (org_id IS NULL AND user_id = $2))`, []any{scope.OrgID, scope.UserID}
	}
	return `user_id = $1 AND org_id IS NULL`, []any{scope.UserID}
}

func reviewJoinWhereArgs(scope model.ReviewScope) (string, []any) {
	if scope.IsOrg() {
		return `(r.org_id = $1 OR (r.org_id IS NULL AND r.user_id = $2))`, []any{scope.OrgID, scope.UserID}
	}
	return `r.user_id = $1 AND r.org_id IS NULL`, []any{scope.UserID}
}

func reviewExistsFilter(scope model.ReviewScope, startArg int) (string, []any) {
	if scope.IsOrg() {
		return `(org_id = $` + itoa(startArg) + ` OR (org_id IS NULL AND user_id = $` + itoa(startArg+1) + `))`, []any{scope.OrgID, scope.UserID}
	}
	return `user_id = $` + itoa(startArg) + ` AND org_id IS NULL`, []any{scope.UserID}
}

func reviewJoinFilter(scope model.ReviewScope, startArg int) (string, []any) {
	if scope.IsOrg() {
		return `(r.org_id = $` + itoa(startArg) + ` OR (r.org_id IS NULL AND r.user_id = $` + itoa(startArg+1) + `))`, []any{scope.OrgID, scope.UserID}
	}
	return `r.user_id = $` + itoa(startArg) + ` AND r.org_id IS NULL`, []any{scope.UserID}
}

func scopeListQuery(base string, scope model.ReviewScope, limit, offset int) (string, []any) {
	where, args := reviewWhereArgs(scope)
	args = append(args, limit, offset)
	n := len(args)
	return base + ` WHERE ` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(n-1) + ` OFFSET $` + itoa(n), args
}

func scopeOneQuery(base string, id string, scope model.ReviewScope) (string, []any) {
	if scope.IsOrg() {
		return base + ` AND (org_id = $2 OR (org_id IS NULL AND user_id = $3))`, []any{id, scope.OrgID, scope.UserID}
	}
	return base + ` AND user_id = $2 AND org_id IS NULL`, []any{id, scope.UserID}
}

func scanReviews(rows pgx.Rows) ([]model.Review, error) {
	reviews := make([]model.Review, 0, 32)
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(&rv.ID, &rv.UserID, &rv.OrgID, &rv.AppName, &rv.Store, &rv.Rating, &rv.Text, &rv.CreatedAt); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "could not scan review", err)
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}

func itoa(n int) string {
	return strconv.Itoa(n)
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
