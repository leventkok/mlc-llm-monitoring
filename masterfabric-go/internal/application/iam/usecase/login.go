package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/dto"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
)

type LoginUseCase struct {
	users repository.UserRepository
	jwt   *infraAuth.AppJWTService
}

func NewLoginUseCase(users repository.UserRepository, jwt *infraAuth.AppJWTService) *LoginUseCase {
	return &LoginUseCase{users: users, jwt: jwt}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (string, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := uc.users.GetByEmail(ctx, req.Email)
	if err != nil {
		_ = infraAuth.CheckPassword(string(infraAuth.DummyHash), req.Password)
		return "", errors.New("incorrect email or password")
	}

	if err := infraAuth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return "", errors.New("incorrect email or password")
	}

	token, err := uc.jwt.GenerateToken(user.ID.String())
	if err != nil {
		return "", errors.New("token could not be generated")
	}
	return token, nil
}
