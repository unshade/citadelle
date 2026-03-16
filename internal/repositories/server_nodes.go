package repositories

import (
	"context"

	"github.com/unshade/citadelle/internal/models"
	"gorm.io/gorm"
)

type ServerNodesRepo interface {
	Create(ctx context.Context, node models.ServerNode) error
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

func (r *ServerNodes) FindBySha256Path(ctx context.Context, b64Sha256Path string) ([]*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("b64_sha256_path = ?", b64Sha256Path).Find(ctx)
}
