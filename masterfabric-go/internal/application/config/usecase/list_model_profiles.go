package usecase

import (
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
)

type ListModelProfilesUseCase struct {
	profiles []model.ModelProfile
}

func NewListModelProfilesUseCase(profiles []model.ModelProfile) *ListModelProfilesUseCase {
	if len(profiles) == 0 {
		profiles = model.DefaultModelProfiles()
	}
	return &ListModelProfilesUseCase{profiles: profiles}
}

func (uc *ListModelProfilesUseCase) Execute() []model.ModelProfile {
	out := make([]model.ModelProfile, len(uc.profiles))
	copy(out, uc.profiles)
	return out
}

func (uc *ListModelProfilesUseCase) Find(id string) (model.ModelProfile, bool) {
	return model.FindProfile(uc.profiles, id)
}
