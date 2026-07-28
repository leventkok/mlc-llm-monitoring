package config

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
)

type SwitchRepo struct {
	pool *pgxpool.Pool
}

func NewSwitchRepo(pool *pgxpool.Pool) *SwitchRepo {
	return &SwitchRepo{pool: pool}
}

func scanSwitch(row pgx.Row) (model.ModelSwitchRequest, error) {
	var req model.ModelSwitchRequest
	var requestedBy *string
	var completedAt *time.Time
	err := row.Scan(
		&req.ID, &req.ProfileID, &req.RequestModel, &req.EngineModel, &req.LocalAdapter,
		&req.Status, &req.ErrorMessage, &requestedBy,
		&req.CreatedAt, &req.UpdatedAt, &completedAt,
	)
	if requestedBy != nil {
		req.RequestedBy = *requestedBy
	}
	req.CompletedAt = completedAt
	return req, err
}

func (r *SwitchRepo) Create(ctx context.Context, req model.ModelSwitchRequest) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO model_switch_requests
		 (id, profile_id, request_model, engine_model, local_adapter, status, requested_by)
		 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')::uuid)`,
		req.ID, req.ProfileID, req.RequestModel, req.EngineModel, req.LocalAdapter,
		req.Status, req.RequestedBy,
	)
	return err
}

func (r *SwitchRepo) GetLatest(ctx context.Context) (*model.ModelSwitchRequest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, profile_id, request_model, engine_model, local_adapter,
		        status, COALESCE(error_message, ''), requested_by::text,
		        created_at, updated_at, completed_at
		 FROM model_switch_requests
		 ORDER BY created_at DESC LIMIT 1`,
	)
	req, err := scanSwitch(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *SwitchRepo) ClaimNextPending(ctx context.Context) (*model.ModelSwitchRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx,
		`SELECT id, profile_id, request_model, engine_model, local_adapter,
		        status, COALESCE(error_message, ''), requested_by::text,
		        created_at, updated_at, completed_at
		 FROM model_switch_requests
		 WHERE status = $1
		 ORDER BY created_at ASC
		 LIMIT 1
		 FOR UPDATE SKIP LOCKED`,
		model.SwitchStatusPending,
	)
	req, err := scanSwitch(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`UPDATE model_switch_requests SET status = $1, updated_at = now() WHERE id = $2`,
		model.SwitchStatusRunning, req.ID,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	req.Status = model.SwitchStatusRunning
	return &req, nil
}

func (r *SwitchRepo) UpdateStatus(ctx context.Context, id, status, errMsg string) error {
	var completedAt any
	if status == model.SwitchStatusCompleted || status == model.SwitchStatusFailed {
		completedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE model_switch_requests
		 SET status = $2, error_message = NULLIF($3, ''), updated_at = now(), completed_at = $4
		 WHERE id = $1`,
		id, status, errMsg, completedAt,
	)
	return err
}
