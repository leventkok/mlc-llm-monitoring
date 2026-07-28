package usecase

import (
	"context"

	configModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/repository"
)

type AgentModelSwitchUseCase struct {
	switches repository.ModelSwitchRepository
}

func NewAgentModelSwitchUseCase(switches repository.ModelSwitchRepository) *AgentModelSwitchUseCase {
	return &AgentModelSwitchUseCase{switches: switches}
}

func (uc *AgentModelSwitchUseCase) ClaimNext(ctx context.Context) (*configModel.ModelSwitchRequest, error) {
	return uc.switches.ClaimNextPending(ctx)
}

func (uc *AgentModelSwitchUseCase) Report(ctx context.Context, id, status, errMsg string) error {
	switch status {
	case configModel.SwitchStatusRunning, configModel.SwitchStatusCompleted, configModel.SwitchStatusFailed:
	default:
		status = configModel.SwitchStatusFailed
		if errMsg == "" {
			errMsg = "invalid status reported by agent"
		}
	}
	return uc.switches.UpdateStatus(ctx, id, status, errMsg)
}
