package usecase

import (
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/repository"
)

type GetLLMConfigUseCase struct {
	config repository.ConfigRepository
}

func NewGetLLMConfigUseCase(config repository.ConfigRepository) *GetLLMConfigUseCase {
	return &GetLLMConfigUseCase{config: config}
}

func (uc *GetLLMConfigUseCase) Execute() model.LLMConfig {
	return uc.config.GetLLM()
}
