package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/pagination"
	"gorm.io/gorm"
)

type ServerNodesRepo interface {
	Create(ctx context.Context, node models.ServerNode) error
	GetByIdAndUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) (*models.ServerNode, error)
	DeleteByIdAndUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) error
	DeleteRecursiveByUserId(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID) ([]uuid.UUID, error)
	GetChildrensByUserId(ctx context.Context, parentUUID uuid.UUID, userId uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error)
	GetRootNodesByUserId(ctx context.Context, userId uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error)
	SetFavourite(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID, isFavourite bool) error
	GetFavouritesByUserId(ctx context.Context, userId uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error)
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

// getDescendantsByUserId is an internal helper that fetches ALL descendants without
// pagination — it is used exclusively by DeleteRecursiveByUserId which must process
// every child node regardless of count.
func (r *ServerNodes) getDescendantsByUserId(ctx context.Context, parentUUID uuid.UUID, userId uuid.UUID) ([]*models.ServerNode, error) {
	var allDescendants []*models.ServerNode
	toProcess := []uuid.UUID{parentUUID}

	for len(toProcess) > 0 {
		currentParent := toProcess[0]
		toProcess = toProcess[1:]

		var children []*models.ServerNode
		if err := r.db.WithContext(ctx).
			Where("parent_id = ? AND proprietary_id = ?", currentParent, userId).
			Find(&children).Error; err != nil {
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

func (r *ServerNodes) GetChildrensByUserId(ctx context.Context, parentUUID uuid.UUID, userId uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error) {
	p = p.Normalize()

	var total int64
	if err := r.db.WithContext(ctx).Model(&models.ServerNode{}).
		Where("parent_id = ? AND proprietary_id = ?", parentUUID, userId).
		Count(&total).Error; err != nil {
		return nil, pagination.Result{}, err
	}

	var nodes []*models.ServerNode
	if err := r.db.WithContext(ctx).
		Where("parent_id = ? AND proprietary_id = ?", parentUUID, userId).
		Offset(p.Offset()).Limit(p.Limit()).
		Find(&nodes).Error; err != nil {
		return nil, pagination.Result{}, err
	}

	return nodes, pagination.Result{Page: p.Page, PerPage: p.PerPage, Total: uint64(total)}, nil
}

func (r *ServerNodes) GetRootNodesByUserId(ctx context.Context, userId uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error) {
	p = p.Normalize()

	var total int64
	if err := r.db.WithContext(ctx).Model(&models.ServerNode{}).
		Where("parent_id IS NULL AND proprietary_id = ?", userId).
		Count(&total).Error; err != nil {
		return nil, pagination.Result{}, err
	}

	var nodes []*models.ServerNode
	if err := r.db.WithContext(ctx).
		Where("parent_id IS NULL AND proprietary_id = ?", userId).
		Offset(p.Offset()).Limit(p.Limit()).
		Find(&nodes).Error; err != nil {
		return nil, pagination.Result{}, err
	}

	return nodes, pagination.Result{Page: p.Page, PerPage: p.PerPage, Total: uint64(total)}, nil
}

func (r *ServerNodes) SetFavourite(ctx context.Context, nodeUUID uuid.UUID, userId uuid.UUID, isFavourite bool) error {
	rowsAffected, err := gorm.G[models.ServerNode](r.db).
		Select("IsFavourite").
		Where("id = ? AND proprietary_id = ?", nodeUUID, userId).
		Updates(ctx, models.ServerNode{IsFavourite: isFavourite})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("node not found or not accessible")
	}
	return nil
}

func (r *ServerNodes) GetFavouritesByUserId(ctx context.Context, userId uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error) {
	p = p.Normalize()

	var total int64
	if err := r.db.WithContext(ctx).Model(&models.ServerNode{}).
		Where("is_favourite = true AND proprietary_id = ?", userId).
		Count(&total).Error; err != nil {
		return nil, pagination.Result{}, err
	}

	var nodes []*models.ServerNode
	if err := r.db.WithContext(ctx).
		Where("is_favourite = true AND proprietary_id = ?", userId).
		Offset(p.Offset()).Limit(p.Limit()).
		Find(&nodes).Error; err != nil {
		return nil, pagination.Result{}, err
	}

	return nodes, pagination.Result{Page: p.Page, PerPage: p.PerPage, Total: uint64(total)}, nil
}
