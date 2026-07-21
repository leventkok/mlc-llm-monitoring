package usecase

import (
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/repository"
)

type GetConfigUseCase struct {
	config repository.ConfigRepository
}

func NewGetConfigUseCase(config repository.ConfigRepository) *GetConfigUseCase {
	return &GetConfigUseCase{config: config}
}

func (uc *GetConfigUseCase) Execute() model.Config {
	return uc.config.Get()
}
