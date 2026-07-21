package usecase

import (
	"errors"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/repository"
)

type UpdateConfigUseCase struct {
	config repository.ConfigRepository
}

func NewUpdateConfigUseCase(config repository.ConfigRepository) *UpdateConfigUseCase {
	return &UpdateConfigUseCase{config: config}
}

func (uc *UpdateConfigUseCase) Execute(cfg model.Config) (model.Config, error) {
	if cfg.AppName == "" || cfg.Model == "" || cfg.Version == "" {
		return model.Config{}, errors.New("app_name, model and version are required")
	}
	return uc.config.Update(cfg), nil
}
