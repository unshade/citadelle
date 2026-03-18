package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
)

type UsersRepo interface {
	GetById(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, user models.User) error
}
