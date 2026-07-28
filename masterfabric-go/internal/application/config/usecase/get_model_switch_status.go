package usecase

import (
	"context"

	configModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/repository"
)

type GetModelSwitchStatusUseCase struct {
	switches repository.ModelSwitchRepository
}

func NewGetModelSwitchStatusUseCase(switches repository.ModelSwitchRepository) *GetModelSwitchStatusUseCase {
	return &GetModelSwitchStatusUseCase{switches: switches}
}

func (uc *GetModelSwitchStatusUseCase) Execute(ctx context.Context) (*configModel.ModelSwitchRequest, error) {
	return uc.switches.GetLatest(ctx)
}
