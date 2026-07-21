package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
	pgIam "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/iam"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/validate"
)

type RegisterUseCase struct {
	users repository.UserRepository
}

func NewRegisterUseCase(users repository.UserRepository) *RegisterUseCase {
	return &RegisterUseCase{users: users}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	if err := validate.Email(req.Email); err != nil {
		return dto.UserResponse{}, err
	}
	if err := validate.Username(req.Username); err != nil {
		return dto.UserResponse{}, err
	}
	if err := infraAuth.ValidatePassword(req.Password); err != nil {
		return dto.UserResponse{}, err
	}

	hash, err := infraAuth.HashPassword(req.Password)
	if err != nil {
		return dto.UserResponse{}, errors.New("password could not be processed")
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	if err := uc.users.Create(ctx, user); err != nil {
		if errors.Is(err, pgIam.ErrEmailTaken) || errors.Is(err, pgIam.ErrUsernameTaken) {
			return dto.UserResponse{}, errors.New("registration failed; check email and username")
		}
		return dto.UserResponse{}, errors.New("registration failed")
	}

	return dto.UserResponse{ID: user.ID.String(), Email: user.Email, Username: user.Username}, nil
}
