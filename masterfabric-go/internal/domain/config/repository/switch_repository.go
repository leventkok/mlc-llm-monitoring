package repository

import (
	"context"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
)

// ModelSwitchRepository persists hot-swap queue entries for the local mlc-agent.
type ModelSwitchRepository interface {
	Create(ctx context.Context, req model.ModelSwitchRequest) error
	GetLatest(ctx context.Context) (*model.ModelSwitchRequest, error)
	ClaimNextPending(ctx context.Context) (*model.ModelSwitchRequest, error)
	UpdateStatus(ctx context.Context, id, status, errMsg string) error
}
