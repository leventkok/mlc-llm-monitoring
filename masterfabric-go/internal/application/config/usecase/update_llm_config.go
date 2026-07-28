package usecase

import (
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/repository"
)

type UpdateLLMConfigUseCase struct {
	config repository.ConfigRepository
}

func NewUpdateLLMConfigUseCase(config repository.ConfigRepository) *UpdateLLMConfigUseCase {
	return &UpdateLLMConfigUseCase{config: config}
}

func (uc *UpdateLLMConfigUseCase) Execute(cfg model.LLMConfig) (model.LLMConfig, error) {
	return uc.config.UpdateLLM(cfg)
}
