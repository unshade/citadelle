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
	GetByIdAndUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) (*models.ServerNode, error)
	DeleteByIdAndUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) error
	DeleteRecursiveByUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) ([]uuid.UUID, error)
	GetChildrensByUserId(ctx context.Context, parentUUID uuid.UUID, userId uuid.UUID) ([]*models.ServerNode, error)
	GetRootNodesByUserId(ctx context.Context, userId uuid.UUID) ([]*models.ServerNode, error)
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

func (r *ServerNodes) GetByIdAndUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) (*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("id = ? AND proprietary_id = ?", nodeUUID, userId).First(ctx)
}

func (r *ServerNodes) DeleteByIdAndUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND proprietary_id = ?", nodeUUID, userId).Delete(&models.ServerNode{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("node not found or not accessible")
	}
	return nil
}

func (r *ServerNodes) DeleteRecursiveByUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) ([]uuid.UUID, error) {
	// First verify the user owns this node
	_, err := r.GetByIdAndUserId(ctx, nodeUUID, userId)
	if err != nil {
		return nil, errors.New("node not found or not accessible")
	}

	descendants, err := r.getDescendantsByUserId(ctx, nodeUUID, userId)
	if err != nil {
		return nil, err
	}

	uuids := make([]uuid.UUID, 0, len(descendants)+1)
	uuids = append(uuids, nodeUUID)
	for _, node := range descendants {
		uuids = append(uuids, node.Id)
	}

	err = r.db.WithContext(ctx).Where("id IN ? AND proprietary_id = ?", uuids, userId).Delete(&models.ServerNode{}).Error
	if err != nil {
		return nil, err
	}

	return uuids, nil
}

func (r *ServerNodes) getDescendantsByUserId(ctx context.Context, parentUUID uuid.UUID, userId uuid.UUID) ([]*models.ServerNode, error) {
	var allDescendants []*models.ServerNode
	var toProcess = []uuid.UUID{parentUUID}

	for len(toProcess) > 0 {
		currentParent := toProcess[0]
		toProcess = toProcess[1:]

		children, err := r.GetChildrensByUserId(ctx, currentParent, userId)
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

func (r *ServerNodes) GetChildrensByUserId(ctx context.Context, parentUUID uuid.UUID, userId uuid.UUID) ([]*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("parent_id = ? AND proprietary_id = ?", parentUUID, userId).Find(ctx)
}

func (r *ServerNodes) GetRootNodesByUserId(ctx context.Context, userId uuid.UUID) ([]*models.ServerNode, error) {
	return gorm.G[*models.ServerNode](r.db).Where("parent_id IS NULL AND proprietary_id = ?", userId).Find(ctx)
}
