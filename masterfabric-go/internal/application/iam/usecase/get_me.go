package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/dto"
	orgUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/org/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	domainErr "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/errors"
)

type GetMeUseCase struct {
	users repository.UserRepository
	org   *orgUC.Service
}

func NewGetMeUseCase(users repository.UserRepository, org *orgUC.Service) *GetMeUseCase {
	return &GetMeUseCase{users: users, org: org}
}

func (uc *GetMeUseCase) Execute(ctx context.Context, userID string) (dto.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return dto.UserResponse{}, errors.New("unauthorized")
	}

	user, err := uc.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainErr.ErrNotFound) {
			return dto.UserResponse{}, errors.New("user not found")
		}
		var de *domainErr.DomainError
		if errors.As(err, &de) && errors.Is(de.Kind, domainErr.ErrNotFound) {
			return dto.UserResponse{}, errors.New("user not found")
		}
		return dto.UserResponse{}, errors.New("user not found")
	}

	return toUserResponse(user, uc.org, ctx), nil
}
