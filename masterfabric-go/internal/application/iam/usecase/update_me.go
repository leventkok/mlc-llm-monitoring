package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	pgIam "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/iam"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/validate"
)

type UpdateMeUseCase struct {
	users repository.UserRepository
}

func NewUpdateMeUseCase(users repository.UserRepository) *UpdateMeUseCase {
	return &UpdateMeUseCase{users: users}
}

func (uc *UpdateMeUseCase) Execute(ctx context.Context, userID string, req dto.UpdateMeRequest) (dto.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return dto.UserResponse{}, errors.New("unauthorized")
	}
	if err := validate.Username(req.Username); err != nil {
		return dto.UserResponse{}, err
	}

	user, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return dto.UserResponse{}, errors.New("user not found")
	}

	if existing, err := uc.users.GetByUsername(ctx, req.Username); err == nil && existing.ID != id {
		return dto.UserResponse{}, errors.New("this username is already taken")
	}

	user.Username = req.Username
	if err := uc.users.Update(ctx, user); err != nil {
		if errors.Is(err, pgIam.ErrUsernameTaken) {
			return dto.UserResponse{}, errors.New("this username is already taken")
		}
		return dto.UserResponse{}, errors.New("could not be updated")
	}

	return dto.UserResponse{ID: user.ID.String(), Email: user.Email, Username: user.Username}, nil
}
