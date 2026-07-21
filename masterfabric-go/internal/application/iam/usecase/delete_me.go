package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	domainErr "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/errors"
)

type DeleteMeUseCase struct {
	users repository.UserRepository
}

func NewDeleteMeUseCase(users repository.UserRepository) *DeleteMeUseCase {
	return &DeleteMeUseCase{users: users}
}

func (uc *DeleteMeUseCase) Execute(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("unauthorized")
	}

	if err := uc.users.Delete(ctx, id); err != nil {
		if domainErr.HTTPStatusCode(err) == 404 {
			return errors.New("user not found")
		}
		return errors.New("could not delete account")
	}
	return nil
}
