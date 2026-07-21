package usecase

import (
	"context"
	"errors"

	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
)

type RefreshUseCase struct {
	jwt *infraAuth.AppJWTService
}

func NewRefreshUseCase(jwt *infraAuth.AppJWTService) *RefreshUseCase {
	return &RefreshUseCase{jwt: jwt}
}

func (uc *RefreshUseCase) Execute(_ context.Context, userID string) (string, error) {
	token, err := uc.jwt.GenerateToken(userID)
	if err != nil {
		return "", errors.New("token could not be generated")
	}
	return token, nil
}
