// Package testutil provides shared mock implementations for use in tests.
// Each mock implements an interface by delegating to a function field,
// so each test can configure exactly the behaviour it needs.
package testutil

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
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
