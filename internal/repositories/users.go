package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"gorm.io/gorm"
)

type UsersRepo interface {
	GetById(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, user models.User) error
}

type Users struct {
	db *gorm.DB
}

func NewUsersRepo(db *gorm.DB) UsersRepo {
	return &Users{db: db}
}

func (r *Users) GetById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return gorm.G[*models.User](r.db).Where("id = ?", id).First(ctx)
}

func (r *Users) Create(ctx context.Context, user models.User) error {
	return gorm.G[models.User](r.db).Create(ctx, &user)
}
