package audit

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	auditModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/audit/model"
)

var ErrNotFound = errors.New("audit not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, a auditModel.Audit) (auditModel.Audit, error) {
	id := uuid.New()
	var orgUUID *uuid.UUID
	if a.OrgID != "" {
		parsed, err := uuid.Parse(a.OrgID)
		if err == nil {
			orgUUID = &parsed
		}
	}
	userUUID, err := uuid.Parse(a.UserID)
	if err != nil {
		return auditModel.Audit{}, err
	}
	var created time.Time
	err = r.db.QueryRow(ctx,
		`INSERT INTO app_audits (
			id, user_id, org_id, client_name, app_display_name,
			play_app_id, appstore_app_id, country, lang,
			play_review_limit, appstore_review_limit,
			mode, status, step, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,now()) RETURNING created_at`,
		id, userUUID, orgUUID, a.ClientName, a.AppDisplayName,
		nullStr(a.PlayAppID), nullStr(a.AppStoreAppID), a.Country, a.Lang,
		a.PlayReviewLimit, a.AppStoreReviewLimit,
		a.Mode, auditModel.StatusQueued, auditModel.StepCrawl,
	).Scan(&created)
	if err != nil {
		return auditModel.Audit{}, err
	}
	a.ID = id.String()
	a.Status = auditModel.StatusQueued
	a.Step = auditModel.StepCrawl
	a.CreatedAt = created.UTC()
	return a, nil
}

func (r *Repository) Get(ctx context.Context, id, userID, orgID string) (auditModel.Audit, error) {
	auditUUID, err := uuid.Parse(id)
	if err != nil {
		return auditModel.Audit{}, ErrNotFound
	}
	var a auditModel.Audit
	var uid uuid.UUID
	var orgUUID *uuid.UUID
	var playID, appstoreID, errMsg *string
	var started, completed *time.Time
	err = r.db.QueryRow(ctx,
		`SELECT id, user_id, org_id, client_name, app_display_name, play_app_id, appstore_app_id,
		        country, lang, play_review_limit, appstore_review_limit,
		        mode, status, step, play_fetched, appstore_fetched, total_reviews, analyzed_count,
		        truncated, error_message, started_at, completed_at, created_at
		 FROM app_audits WHERE id = $1`, auditUUID,
	).Scan(
		&auditUUID, &uid, &orgUUID, &a.ClientName, &a.AppDisplayName, &playID, &appstoreID,
		&a.Country, &a.Lang, &a.PlayReviewLimit, &a.AppStoreReviewLimit,
		&a.Mode, &a.Status, &a.Step, &a.PlayFetched, &a.AppStoreFetched, &a.TotalReviews,
		&a.AnalyzedCount, &a.Truncated, &errMsg, &started, &completed, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auditModel.Audit{}, ErrNotFound
		}
		return auditModel.Audit{}, err
	}
	a.ID = auditUUID.String()
	a.UserID = uid.String()
	if orgUUID != nil {
		a.OrgID = orgUUID.String()
	}
	if playID != nil {
		a.PlayAppID = *playID
	}
	if appstoreID != nil {
		a.AppStoreAppID = *appstoreID
	}
	if errMsg != nil {
		a.ErrorMessage = *errMsg
	}
	a.StartedAt = started
	a.CompletedAt = completed
	if orgID != "" && a.OrgID != orgID {
		return auditModel.Audit{}, ErrNotFound
	}
	if orgID == "" && a.UserID != userID {
		return auditModel.Audit{}, ErrNotFound
	}
	return a, nil
}

func (r *Repository) List(ctx context.Context, userID, orgID string, limit int) ([]auditModel.Audit, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows pgx.Rows
	var err error
	if orgID != "" {
		orgUUID, parseErr := uuid.Parse(orgID)
		if parseErr != nil {
			return nil, parseErr
		}
		rows, err = r.db.Query(ctx,
			`SELECT id, user_id, org_id, client_name, app_display_name, play_app_id, appstore_app_id,
			        country, lang, play_review_limit, appstore_review_limit,
			        mode, status, step, play_fetched, appstore_fetched, total_reviews, analyzed_count,
			        truncated, error_message, started_at, completed_at, created_at
			 FROM app_audits WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2`, orgUUID, limit)
	} else {
		userUUID, parseErr := uuid.Parse(userID)
		if parseErr != nil {
			return nil, parseErr
		}
		rows, err = r.db.Query(ctx,
			`SELECT id, user_id, org_id, client_name, app_display_name, play_app_id, appstore_app_id,
			        country, lang, play_review_limit, appstore_review_limit,
			        mode, status, step, play_fetched, appstore_fetched, total_reviews, analyzed_count,
			        truncated, error_message, started_at, completed_at, created_at
			 FROM app_audits WHERE user_id = $1 AND org_id IS NULL ORDER BY created_at DESC LIMIT $2`, userUUID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAudits(rows)
}

func (r *Repository) UpdateProgress(ctx context.Context, id string, fields map[string]any) error {
	auditUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	status, _ := fields["status"].(string)
	step, _ := fields["step"].(string)
	playFetched, _ := fields["play_fetched"].(int)
	appstoreFetched, _ := fields["appstore_fetched"].(int)
	totalReviews, _ := fields["total_reviews"].(int)
	analyzedCount, _ := fields["analyzed_count"].(int)
	truncated, _ := fields["truncated"].(bool)
	errMsg, _ := fields["error_message"].(string)
	_, err = r.db.Exec(ctx,
		`UPDATE app_audits SET
		 status = COALESCE(NULLIF($2,''), status),
		 step = COALESCE(NULLIF($3,''), step),
		 play_fetched = CASE WHEN $4 >= 0 THEN $4 ELSE play_fetched END,
		 appstore_fetched = CASE WHEN $5 >= 0 THEN $5 ELSE appstore_fetched END,
		 total_reviews = CASE WHEN $6 >= 0 THEN $6 ELSE total_reviews END,
		 analyzed_count = CASE WHEN $7 >= 0 THEN $7 ELSE analyzed_count END,
		 truncated = $8,
		 error_message = NULLIF($9,''),
		 started_at = COALESCE(started_at, CASE WHEN $2 = 'running' THEN now() ELSE started_at END),
		 completed_at = CASE WHEN $2 IN ('completed','failed') THEN now() ELSE completed_at END
		 WHERE id = $1`,
		auditUUID, status, step, playFetched, appstoreFetched, totalReviews, analyzedCount, truncated, errMsg,
	)
	return err
}

func (r *Repository) InsertReviews(ctx context.Context, auditID string, rows []auditModel.Review) (int, error) {
	auditUUID, err := uuid.Parse(auditID)
	if err != nil {
		return 0, err
	}
	inserted := 0
	for _, row := range rows {
		id := uuid.New()
		var reviewedAt *time.Time
		if row.ReviewedAt != nil {
			reviewedAt = row.ReviewedAt
		}
		tag, err := r.db.Exec(ctx,
			`INSERT INTO audit_reviews (id, audit_id, store, store_review_id, app_name, rating, text, reviewed_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			 ON CONFLICT (audit_id, store, store_review_id) DO NOTHING`,
			id, auditUUID, row.Store, row.StoreReviewID, row.AppName, row.Rating, row.Text, reviewedAt,
		)
		if err != nil {
			return inserted, err
		}
		inserted += int(tag.RowsAffected())
	}
	return inserted, nil
}

func (r *Repository) ListUnclassified(ctx context.Context, auditID string, limit int) ([]auditModel.Review, error) {
	auditUUID, err := uuid.Parse(auditID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 500
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, audit_id, store, store_review_id, app_name, rating, text, reviewed_at, category, sentiment, raw_output
		 FROM audit_reviews WHERE audit_id = $1 AND category IS NULL ORDER BY reviewed_at DESC NULLS LAST LIMIT $2`,
		auditUUID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []auditModel.Review
	for rows.Next() {
		var rv auditModel.Review
		var id, aid uuid.UUID
		var reviewedAt *time.Time
		var cat, sent, raw *string
		if err := rows.Scan(&id, &aid, &rv.Store, &rv.StoreReviewID, &rv.AppName, &rv.Rating, &rv.Text,
			&reviewedAt, &cat, &sent, &raw); err != nil {
			return nil, err
		}
		rv.ID = id.String()
		rv.AuditID = aid.String()
		rv.ReviewedAt = reviewedAt
		if cat != nil {
			rv.Category = *cat
		}
		if sent != nil {
			rv.Sentiment = *sent
		}
		if raw != nil {
			rv.RawOutput = *raw
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateReviewClassification(ctx context.Context, id, category, sentiment, raw string) error {
	reviewUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx,
		`UPDATE audit_reviews SET category = $2, sentiment = $3, raw_output = $4 WHERE id = $1`,
		reviewUUID, category, sentiment, raw,
	)
	return err
}

func (r *Repository) ComputeStatistics(ctx context.Context, auditID string) (auditModel.Statistics, error) {
	auditUUID, err := uuid.Parse(auditID)
	if err != nil {
		return auditModel.Statistics{}, err
	}
	stats := auditModel.Statistics{
		CategoryCounts:  map[string]int{},
		SentimentCounts: map[string]int{},
		RatingCounts:    map[string]int{},
	}
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*),
		        COUNT(*) FILTER (WHERE store = 'play'),
		        COUNT(*) FILTER (WHERE store = 'appstore'),
		        COALESCE(AVG(rating), 0),
		        COALESCE(AVG(rating) FILTER (WHERE store = 'play'), 0),
		        COALESCE(AVG(rating) FILTER (WHERE store = 'appstore'), 0),
		        COALESCE(100.0 * COUNT(*) FILTER (WHERE rating <= 2) / NULLIF(COUNT(*), 0), 0)
		 FROM audit_reviews WHERE audit_id = $1`, auditUUID,
	).Scan(&stats.TotalReviews, &stats.PlayCount, &stats.AppStoreCount,
		&stats.AvgRating, &stats.AvgRatingPlay, &stats.AvgRatingAppStore, &stats.LowStarPct)
	if err != nil {
		return stats, err
	}

	catRows, err := r.db.Query(ctx,
		`SELECT COALESCE(category,'unknown'), COUNT(*) FROM audit_reviews WHERE audit_id = $1 GROUP BY 1`, auditUUID)
	if err != nil {
		return stats, err
	}
	for catRows.Next() {
		var k string
		var v int
		_ = catRows.Scan(&k, &v)
		stats.CategoryCounts[k] = v
	}
	catRows.Close()

	sentRows, err := r.db.Query(ctx,
		`SELECT COALESCE(sentiment,'unknown'), COUNT(*) FROM audit_reviews WHERE audit_id = $1 GROUP BY 1`, auditUUID)
	if err != nil {
		return stats, err
	}
	for sentRows.Next() {
		var k string
		var v int
		_ = sentRows.Scan(&k, &v)
		stats.SentimentCounts[k] = v
	}
	sentRows.Close()

	ratingRows, err := r.db.Query(ctx,
		`SELECT rating::text, COUNT(*) FROM audit_reviews WHERE audit_id = $1 GROUP BY 1`, auditUUID)
	if err != nil {
		return stats, err
	}
	for ratingRows.Next() {
		var k string
		var v int
		_ = ratingRows.Scan(&k, &v)
		stats.RatingCounts[k] = v
	}
	ratingRows.Close()
	return stats, nil
}

func (r *Repository) SampleNegativeReviews(ctx context.Context, auditID string, limit int) ([]auditModel.Review, error) {
	auditUUID, err := uuid.Parse(auditID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 80
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, audit_id, store, store_review_id, app_name, rating, text, reviewed_at, category, sentiment, raw_output
		 FROM audit_reviews WHERE audit_id = $1 AND rating <= 3
		 ORDER BY rating ASC, reviewed_at DESC NULLS LAST LIMIT $2`, auditUUID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []auditModel.Review
	for rows.Next() {
		var rv auditModel.Review
		var id, aid uuid.UUID
		var reviewedAt *time.Time
		var cat, sent, raw *string
		if err := rows.Scan(&id, &aid, &rv.Store, &rv.StoreReviewID, &rv.AppName, &rv.Rating, &rv.Text,
			&reviewedAt, &cat, &sent, &raw); err != nil {
			return nil, err
		}
		rv.ID = id.String()
		rv.AuditID = aid.String()
		rv.ReviewedAt = reviewedAt
		if cat != nil {
			rv.Category = *cat
		}
		if sent != nil {
			rv.Sentiment = *sent
		}
		if raw != nil {
			rv.RawOutput = *raw
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

func (r *Repository) SaveInsights(ctx context.Context, insights auditModel.Insights) error {
	auditUUID, err := uuid.Parse(insights.AuditID)
	if err != nil {
		return err
	}
	statsJSON, _ := json.Marshal(insights.Statistics)
	rootJSON, _ := json.Marshal(insights.RootCauses)
	planJSON, _ := json.Marshal(insights.ActionPlan)
	_, err = r.db.Exec(ctx,
		`INSERT INTO audit_insights (audit_id, executive_summary, statistics, root_causes, action_plan, generated_at)
		 VALUES ($1,$2,$3,$4,$5,now())
		 ON CONFLICT (audit_id) DO UPDATE SET
		   executive_summary = EXCLUDED.executive_summary,
		   statistics = EXCLUDED.statistics,
		   root_causes = EXCLUDED.root_causes,
		   action_plan = EXCLUDED.action_plan,
		   generated_at = now()`,
		auditUUID, insights.ExecutiveSummary, statsJSON, rootJSON, planJSON,
	)
	return err
}

func (r *Repository) GetInsights(ctx context.Context, auditID string) (*auditModel.Insights, error) {
	auditUUID, err := uuid.Parse(auditID)
	if err != nil {
		return nil, err
	}
	var ins auditModel.Insights
	var statsJSON, rootJSON, planJSON []byte
	err = r.db.QueryRow(ctx,
		`SELECT audit_id, executive_summary, statistics, root_causes, action_plan, generated_at
		 FROM audit_insights WHERE audit_id = $1`, auditUUID,
	).Scan(&auditUUID, &ins.ExecutiveSummary, &statsJSON, &rootJSON, &planJSON, &ins.GeneratedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ins.AuditID = auditUUID.String()
	_ = json.Unmarshal(statsJSON, &ins.Statistics)
	_ = json.Unmarshal(rootJSON, &ins.RootCauses)
	_ = json.Unmarshal(planJSON, &ins.ActionPlan)
	return &ins, nil
}

func (r *Repository) Delete(ctx context.Context, id, userID, orgID string) error {
	auditUUID, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	a, err := r.Get(ctx, id, userID, orgID)
	if err != nil {
		return err
	}
	if a.ID == "" {
		return ErrNotFound
	}
	tag, err := r.db.Exec(ctx, `DELETE FROM app_audits WHERE id = $1`, auditUUID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanAudits(rows pgx.Rows) ([]auditModel.Audit, error) {
	var out []auditModel.Audit
	for rows.Next() {
		var a auditModel.Audit
		var id, uid uuid.UUID
		var orgUUID *uuid.UUID
		var playID, appstoreID, errMsg *string
		var started, completed *time.Time
		if err := rows.Scan(
			&id, &uid, &orgUUID, &a.ClientName, &a.AppDisplayName, &playID, &appstoreID,
			&a.Country, &a.Lang, &a.PlayReviewLimit, &a.AppStoreReviewLimit,
			&a.Mode, &a.Status, &a.Step, &a.PlayFetched, &a.AppStoreFetched, &a.TotalReviews,
			&a.AnalyzedCount, &a.Truncated, &errMsg, &started, &completed, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		a.ID = id.String()
		a.UserID = uid.String()
		if orgUUID != nil {
			a.OrgID = orgUUID.String()
		}
		if playID != nil {
			a.PlayAppID = *playID
		}
		if appstoreID != nil {
			a.AppStoreAppID = *appstoreID
		}
		if errMsg != nil {
			a.ErrorMessage = *errMsg
		}
		a.StartedAt = started
		a.CompletedAt = completed
		out = append(out, a)
	}
	return out, rows.Err()
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
