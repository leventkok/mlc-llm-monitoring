package storage

import (
	"context"

	"github.com/leventkok/mlc-llm-monitoring/internal/models"
)

type UserStore interface {
	Create(ctx context.Context, user models.User) error
	FindByEmail(ctx context.Context, email string) (models.User, error)
	FindByUsername(ctx context.Context, username string) (models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
	Update(ctx context.Context, user models.User) error
	Delete(ctx context.Context, id string) error
}
