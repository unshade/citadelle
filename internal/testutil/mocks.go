// Package testutil provides shared mock implementations for use in tests.
// Each mock implements an interface by delegating to a function field,
// so each test can configure exactly the behaviour it needs.
package testutil

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
)

// --- UsersRepo mock ---

type MockUsersRepo struct {
	GetByIdFunc func(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateFunc  func(ctx context.Context, user models.User) error
}

func (m *MockUsersRepo) GetById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.GetByIdFunc(ctx, id)
}

func (m *MockUsersRepo) Create(ctx context.Context, user models.User) error {
	return m.CreateFunc(ctx, user)
}

// --- ServerNodesRepo mock ---

type MockNodesRepo struct {
	CreateFunc                  func(ctx context.Context, node models.ServerNode) error
	GetByIdAndUserIdFunc        func(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID) (*models.ServerNode, error)
	DeleteByIdAndUserIdFunc     func(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID) error
	DeleteRecursiveByUserIdFunc func(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID) ([]uuid.UUID, error)
	GetChildrensByUserIdFunc    func(ctx context.Context, parentID uuid.UUID, userID uuid.UUID) ([]*models.ServerNode, error)
	GetRootNodesByUserIdFunc    func(ctx context.Context, userID uuid.UUID) ([]*models.ServerNode, error)
	SetFavouriteFunc            func(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID, isFavourite bool) error
	GetFavouritesByUserIdFunc   func(ctx context.Context, userID uuid.UUID) ([]*models.ServerNode, error)
}

func (m *MockNodesRepo) Create(ctx context.Context, node models.ServerNode) error {
	return m.CreateFunc(ctx, node)
}

func (m *MockNodesRepo) GetByIdAndUserId(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID) (*models.ServerNode, error) {
	return m.GetByIdAndUserIdFunc(ctx, nodeID, userID)
}

func (m *MockNodesRepo) DeleteByIdAndUserId(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID) error {
	return m.DeleteByIdAndUserIdFunc(ctx, nodeID, userID)
}

func (m *MockNodesRepo) DeleteRecursiveByUserId(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID) ([]uuid.UUID, error) {
	return m.DeleteRecursiveByUserIdFunc(ctx, nodeID, userID)
}

func (m *MockNodesRepo) GetChildrensByUserId(ctx context.Context, parentID uuid.UUID, userID uuid.UUID) ([]*models.ServerNode, error) {
	return m.GetChildrensByUserIdFunc(ctx, parentID, userID)
}

func (m *MockNodesRepo) GetRootNodesByUserId(ctx context.Context, userID uuid.UUID) ([]*models.ServerNode, error) {
	return m.GetRootNodesByUserIdFunc(ctx, userID)
}

func (m *MockNodesRepo) SetFavourite(ctx context.Context, nodeID uuid.UUID, userID uuid.UUID, isFavourite bool) error {
	return m.SetFavouriteFunc(ctx, nodeID, userID, isFavourite)
}

func (m *MockNodesRepo) GetFavouritesByUserId(ctx context.Context, userID uuid.UUID) ([]*models.ServerNode, error) {
	return m.GetFavouritesByUserIdFunc(ctx, userID)
}

// --- AuthService mock ---

type MockAuthService struct {
	GetChallengeFunc    func(ctx context.Context, userID uuid.UUID) (services.ChallengeData, error)
	VerifyChallengeFunc func(ctx context.Context, userID uuid.UUID, clearChallenge string) (string, error)
}

func (m *MockAuthService) GetChallenge(ctx context.Context, userID uuid.UUID) (services.ChallengeData, error) {
	return m.GetChallengeFunc(ctx, userID)
}

func (m *MockAuthService) VerifyChallenge(ctx context.Context, userID uuid.UUID, clearChallenge string) (string, error) {
	return m.VerifyChallengeFunc(ctx, userID, clearChallenge)
}

// --- UserService mock ---

type MockUserService struct {
	GetUserFunc    func(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUserFunc func(ctx context.Context, input services.CreateUserInput) (uuid.UUID, error)
}

func (m *MockUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.GetUserFunc(ctx, id)
}

func (m *MockUserService) CreateUser(ctx context.Context, input services.CreateUserInput) (uuid.UUID, error) {
	return m.CreateUserFunc(ctx, input)
}

// --- NodeService mock ---

type MockNodeService struct {
	CreateNodeFunc     func(ctx context.Context, userID uuid.UUID, input services.CreateNodeInput) (uuid.UUID, error)
	SaveNodeFunc       func(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, file io.Reader, size int64) error
	DownloadNodeFunc   func(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (*models.ServerNode, []byte, error)
	IndexDirectoryFunc func(ctx context.Context, userID uuid.UUID, uuidParam string) ([]*models.ServerNode, error)
	DeleteNodeFunc     func(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error
	SetFavouriteFunc   func(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, isFavourite bool) error
	GetFavouritesFunc  func(ctx context.Context, userID uuid.UUID) ([]*models.ServerNode, error)
}

func (m *MockNodeService) CreateNode(ctx context.Context, userID uuid.UUID, input services.CreateNodeInput) (uuid.UUID, error) {
	return m.CreateNodeFunc(ctx, userID, input)
}

func (m *MockNodeService) SaveNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, file io.Reader, size int64) error {
	return m.SaveNodeFunc(ctx, userID, nodeID, file, size)
}

func (m *MockNodeService) DeleteNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) error {
	return m.DeleteNodeFunc(ctx, userID, nodeID)
}

func (m *MockNodeService) IndexDirectory(ctx context.Context, userID uuid.UUID, uuidParam string) ([]*models.ServerNode, error) {
	return m.IndexDirectoryFunc(ctx, userID, uuidParam)
}

func (m *MockNodeService) DownloadNode(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID) (*models.ServerNode, []byte, error) {
	return m.DownloadNodeFunc(ctx, userID, nodeID)
}

func (m *MockNodeService) SetFavourite(ctx context.Context, userID uuid.UUID, nodeID uuid.UUID, isFavourite bool) error {
	return m.SetFavouriteFunc(ctx, userID, nodeID, isFavourite)
}

func (m *MockNodeService) GetFavourites(ctx context.Context, userID uuid.UUID) ([]*models.ServerNode, error) {
	return m.GetFavouritesFunc(ctx, userID)
}

// --- FileStorage mock ---

type MockFileStorage struct {
	WriteFunc  func(id string, r io.Reader) error
	ReadFunc   func(id string) ([]byte, error)
	DeleteFunc func(id string) error
}

func (m *MockFileStorage) Write(id string, r io.Reader) error {
	return m.WriteFunc(id, r)
}

func (m *MockFileStorage) Read(id string) ([]byte, error) {
	return m.ReadFunc(id)
}

func (m *MockFileStorage) Delete(id string) error {
	return m.DeleteFunc(id)
}
