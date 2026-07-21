package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
)

type ChangePasswordUseCase struct {
	users repository.UserRepository
}

func NewChangePasswordUseCase(users repository.UserRepository) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{users: users}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("unauthorized")
	}
	if err := infraAuth.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	user, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	if err := infraAuth.CheckPassword(user.PasswordHash, req.OldPassword); err != nil {
		return errors.New("could not change password")
	}

	newHash, err := infraAuth.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("password could not be processed")
	}

	user.PasswordHash = string(newHash)
	if err := uc.users.Update(ctx, user); err != nil {
		return errors.New("could not change password")
	}
	return nil
}
