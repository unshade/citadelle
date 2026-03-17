package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"gorm.io/gorm"
)

type ServerNodesRepo interface {
	Create(ctx context.Context, node models.ServerNode) error
	Delete(ctx context.Context, uuid uuid.UUID) error
	GetChildrens(ctx context.Context, uuid uuid.UUID) ([]*models.ServerNode, error)
	FindBySha256Path(ctx context.Context, b64Sha256Path string) ([]*models.ServerNode, error)
}

type ServerNodes struct {
	db *gorm.DB
}

func NewServerNodesRepo(db *gorm.DB) ServerNodesRepo {
	return &ServerNodes{db: db}
}

func (r *ServerNodes) Create(ctx context.Context, node models.ServerNode) error {
	return gorm.G[models.ServerNode](r.db).Create(ctx, &node)
}

func (r *ServerNodes) Delete(ctx context.Context, uuid uuid.UUID) error {
	linesAffected, err := gorm.G[models.ServerNode](r.db).Where("uuid = ?", uuid).Delete(ctx)
	if err != nil {
		return err
	}

	if linesAffected != 1 {
		return errors.New("critical, multiple rows affected. Better load a backup bro")
	}

	return nil
}

func (r *ServerNodes) GetChildrens(ctx context.Context, uuid uuid.UUID) ([]*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("parent_id = ?", uuid).Find(ctx)
}

func (r *ServerNodes) FindBySha256Path(ctx context.Context, b64Sha256Path string) ([]*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("b64_sha256_path = ?", b64Sha256Path).Find(ctx)
}
