package services

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/pagination"
	"github.com/unshade/citadelle/internal/repositories"
	"github.com/unshade/citadelle/internal/storage"
)

const maxUploadSize = 100 * 1024 * 1024 // 100 MB

type NodeService interface {
	CreateNode(ctx context.Context, userID uuid.UUID, input CreateNodeInput) (uuid.UUID, error)
	SaveNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, file io.Reader, size int64) error
	DownloadNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (*models.ServerNode, []byte, error)
	IndexDirectory(ctx context.Context, userID uuid.UUID, uuidParam string, p pagination.Params) ([]*models.ServerNode, pagination.Result, error)
	DeleteNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error
	SetFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, isFavourite bool) error
	GetFavourites(ctx context.Context, userID uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error)
}

type CreateNodeInput struct {
	ContentNonce     []byte
	NameNonce        []byte
	EncryptedName    []byte
	B64PathNonce     string
	B64EncryptedPath string
	IsDirectory      bool
	ParentId         *uuid.UUID
	Version          uint64
}

type nodeService struct {
	nodes   repositories.ServerNodesRepo
	storage storage.FileStorage
}

func NewNodeService(nodes repositories.ServerNodesRepo, storage storage.FileStorage) NodeService {
	return &nodeService{nodes: nodes, storage: storage}
}

func (s *nodeService) CreateNode(ctx context.Context, userID uuid.UUID, input CreateNodeInput) (uuid.UUID, error) {
	node := models.ServerNode{
		Id:               uuid.New(),
		ContentNonce:     input.ContentNonce,
		NameNonce:        input.NameNonce,
		EncryptedName:    input.EncryptedName,
		B64PathNonce:     input.B64PathNonce,
		B64EncryptedPath: input.B64EncryptedPath,
		Version:          input.Version,
		IsDirectory:      input.IsDirectory,
		ParentId:         input.ParentId,
		ProprietaryId:    userID,
	}
	if err := s.nodes.Create(ctx, node); err != nil {
		return uuid.Nil, err
	}
	return node.Id, nil
}

func (s *nodeService) SaveNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, file io.Reader, size int64) error {
	if size > maxUploadSize {
		return errors.New("file too large")
	}
	if _, err := s.nodes.GetByIdAndUserId(ctx, nodeID, userID); err != nil {
		return errors.New("node not found or not accessible")
	}
	return s.storage.Write(nodeID.String(), file)
}

func (s *nodeService) DownloadNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (*models.ServerNode, []byte, error) {
	node, err := s.nodes.GetByIdAndUserId(ctx, nodeID, userID)
	if err != nil {
		return nil, nil, err
	}
	if node.IsDirectory {
		return nil, nil, errors.New("cannot download a directory")
	}
	data, err := s.storage.Read(nodeID.String())
	if err != nil {
		return nil, nil, err
	}
	return node, data, nil
}

func (s *nodeService) IndexDirectory(ctx context.Context, userID uuid.UUID, uuidParam string, p pagination.Params) ([]*models.ServerNode, pagination.Result, error) {
	if uuidParam == "root" || uuidParam == "" || uuidParam == "00000000-0000-0000-0000-000000000000" {
		return s.nodes.GetRootNodesByUserId(ctx, userID, p)
	}
	parentUUID, err := uuid.Parse(uuidParam)
	if err != nil {
		return nil, pagination.Result{}, err
	}
	return s.nodes.GetChildrensByUserId(ctx, parentUUID, userID, p)
}

func (s *nodeService) SetFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, isFavourite bool) error {
	return s.nodes.SetFavourite(ctx, nodeID, userID, isFavourite)
}

func (s *nodeService) GetFavourites(ctx context.Context, userID uuid.UUID, p pagination.Params) ([]*models.ServerNode, pagination.Result, error) {
	return s.nodes.GetFavouritesByUserId(ctx, userID, p)
}

func (s *nodeService) DeleteNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error {
	uuids, err := s.nodes.DeleteRecursiveByUserId(ctx, nodeID, userID)
	if err != nil {
		return err
	}
	for _, id := range uuids {
		if err := s.storage.Delete(id.String()); err != nil {
			return err
		}
	}
	return nil
}
