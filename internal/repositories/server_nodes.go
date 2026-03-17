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
	GetByID(ctx context.Context, uuid uuid.UUID) (*models.ServerNode, error)
	Delete(ctx context.Context, uuid uuid.UUID) error
	DeleteRecursive(ctx context.Context, uuid uuid.UUID) ([]uuid.UUID, error)
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

func (r *ServerNodes) GetByID(ctx context.Context, uuid uuid.UUID) (*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("id = ?", uuid).First(ctx)
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

func (r *ServerNodes) DeleteRecursive(ctx context.Context, nodeUUID uuid.UUID) ([]uuid.UUID, error) {
	descendants, err := r.getDescendants(ctx, nodeUUID)
	if err != nil {
		return nil, err
	}

	uuids := make([]uuid.UUID, 0, len(descendants)+1)
	uuids = append(uuids, nodeUUID)
	for _, node := range descendants {
		uuids = append(uuids, node.Id)
	}

	err = r.db.WithContext(ctx).Where("id IN ?", uuids).Delete(&models.ServerNode{}).Error
	if err != nil {
		return nil, err
	}

	return uuids, nil
}

func (r *ServerNodes) getDescendants(ctx context.Context, parentUUID uuid.UUID) ([]*models.ServerNode, error) {
	var allDescendants []*models.ServerNode
	var toProcess = []uuid.UUID{parentUUID}

	for len(toProcess) > 0 {
		currentParent := toProcess[0]
		toProcess = toProcess[1:]

		children, err := r.GetChildrens(ctx, currentParent)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			allDescendants = append(allDescendants, child)
			if child.IsDirectory {
				toProcess = append(toProcess, child.Id)
			}
		}
	}

	return allDescendants, nil
}

func (r *ServerNodes) GetChildrens(ctx context.Context, uuid uuid.UUID) ([]*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("parent_id = ?", uuid).Find(ctx)
}

func (r *ServerNodes) FindBySha256Path(ctx context.Context, b64Sha256Path string) ([]*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("b64_sha256_path = ?", b64Sha256Path).Find(ctx)
}
