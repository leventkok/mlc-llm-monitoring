package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	configModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/repository"
)

type SwitchModelUseCase struct {
	profiles *ListModelProfilesUseCase
	config   repository.ConfigRepository
	switches repository.ModelSwitchRepository
}

func NewSwitchModelUseCase(
	profiles *ListModelProfilesUseCase,
	config repository.ConfigRepository,
	switches repository.ModelSwitchRepository,
) *SwitchModelUseCase {
	return &SwitchModelUseCase{profiles: profiles, config: config, switches: switches}
}

func (uc *SwitchModelUseCase) Execute(ctx context.Context, profileID, requestedBy string) (configModel.ModelSwitchRequest, error) {
	profile, ok := uc.profiles.Find(profileID)
	if !ok {
		return configModel.ModelSwitchRequest{}, errors.New("unknown model profile")
	}

	req := configModel.ModelSwitchRequest{
		ID:           uuid.NewString(),
		ProfileID:    profile.ID,
		RequestModel: profile.RequestModel,
		EngineModel:  profile.EngineModel,
		LocalAdapter: profile.LocalAdapter,
		Status:       configModel.SwitchStatusPending,
		RequestedBy:  requestedBy,
	}

	if err := uc.switches.Create(ctx, req); err != nil {
		return configModel.ModelSwitchRequest{}, err
	}

	// API-side hot-swap: next analyze requests use the new model id immediately.
	uc.config.SwitchLLM(profile.RequestModel, profile.LocalAdapter)

	return req, nil
}
