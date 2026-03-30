package services_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
	"github.com/unshade/citadelle/internal/testutil"
)

// --- CreateNode ---

func TestCreateNode_ReturnsNewUUID_OnSuccess(t *testing.T) {
	var savedNode models.ServerNode
	userID := uuid.New()

	nodes := &testutil.MockNodesRepo{
		CreateFunc: func(_ context.Context, n models.ServerNode) error {
			savedNode = n
			return nil
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	id, err := svc.CreateNode(context.Background(), userID, services.CreateNodeInput{
		KeyNonce:      []byte("key-nonce"),
		EncryptedKey:  []byte("key"),
		ContentNonce:  []byte("content-nonce"),
		NameNonce:     []byte("name-nonce"),
		EncryptedName: []byte("name"),
		IsDirectory:   false,
		Version:       1,
	})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	assert.Equal(t, id, savedNode.Id)
	assert.Equal(t, userID, savedNode.ProprietaryId)
}

func TestCreateNode_ReturnsError_WhenRepoFails(t *testing.T) {
	nodes := &testutil.MockNodesRepo{
		CreateFunc: func(_ context.Context, _ models.ServerNode) error {
			return errors.New("db error")
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	_, err := svc.CreateNode(context.Background(), uuid.New(), services.CreateNodeInput{})

	require.Error(t, err)
}

// --- SaveNode ---

func TestSaveNode_ReturnsError_WhenFileTooLarge(t *testing.T) {
	svc := services.NewNodeService(&testutil.MockNodesRepo{}, &testutil.MockFileStorage{})

	const tooBig = 101 * 1024 * 1024 // 101 MB — over the 100 MB limit
	err := svc.SaveNode(context.Background(), uuid.New(), uuid.New(), bytes.NewReader(nil), tooBig)

	require.Error(t, err)
	assert.Equal(t, "file too large", err.Error())
}

func TestSaveNode_ReturnsError_WhenUserDoesNotOwnNode(t *testing.T) {
	nodes := &testutil.MockNodesRepo{
		GetByIdAndUserIdFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.ServerNode, error) {
			return nil, errors.New("node not found or not accessible")
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	err := svc.SaveNode(context.Background(), uuid.New(), uuid.New(), bytes.NewReader([]byte("data")), 4)

	require.Error(t, err)
	assert.Equal(t, "node not found or not accessible", err.Error())
}

func TestSaveNode_WritesFileToStorage_WhenValid(t *testing.T) {
	nodeID := uuid.New()
	fileContent := []byte("encrypted content")
	var writtenID string
	var writtenData []byte

	nodes := &testutil.MockNodesRepo{
		GetByIdAndUserIdFunc: func(_ context.Context, id uuid.UUID, _ uuid.UUID) (*models.ServerNode, error) {
			return &models.ServerNode{Id: id}, nil
		},
	}
	store := &testutil.MockFileStorage{
		WriteFunc: func(id string, r io.Reader) error {
			writtenID = id
			writtenData, _ = io.ReadAll(r)
			return nil
		},
	}

	svc := services.NewNodeService(nodes, store)
	err := svc.SaveNode(context.Background(), uuid.New(), nodeID, bytes.NewReader(fileContent), int64(len(fileContent)))

	require.NoError(t, err)
	assert.Equal(t, nodeID.String(), writtenID)
	assert.Equal(t, fileContent, writtenData)
}

// --- DownloadNode ---

func TestDownloadNode_ReturnsError_WhenNodeIsDirectory(t *testing.T) {
	nodeID := uuid.New()

	nodes := &testutil.MockNodesRepo{
		GetByIdAndUserIdFunc: func(_ context.Context, id uuid.UUID, _ uuid.UUID) (*models.ServerNode, error) {
			return &models.ServerNode{Id: id, IsDirectory: true}, nil
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	_, _, err := svc.DownloadNode(context.Background(), uuid.New(), nodeID)

	require.Error(t, err)
	assert.Equal(t, "cannot download a directory", err.Error())
}

func TestDownloadNode_ReturnsNodeAndFileData_WhenValid(t *testing.T) {
	nodeID := uuid.New()
	fileData := []byte("encrypted file bytes")
	node := &models.ServerNode{
		Id:            nodeID,
		IsDirectory:   false,
		EncryptedKey:  []byte("key"),
		EncryptedName: []byte("name"),
	}

	nodes := &testutil.MockNodesRepo{
		GetByIdAndUserIdFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.ServerNode, error) {
			return node, nil
		},
	}
	store := &testutil.MockFileStorage{
		ReadFunc: func(id string) ([]byte, error) {
			assert.Equal(t, nodeID.String(), id)
			return fileData, nil
		},
	}

	svc := services.NewNodeService(nodes, store)
	gotNode, gotData, err := svc.DownloadNode(context.Background(), uuid.New(), nodeID)

	require.NoError(t, err)
	assert.Equal(t, node, gotNode)
	assert.Equal(t, fileData, gotData)
}

func TestDownloadNode_ReturnsError_WhenUserDoesNotOwnNode(t *testing.T) {
	nodes := &testutil.MockNodesRepo{
		GetByIdAndUserIdFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.ServerNode, error) {
			return nil, errors.New("node not found or not accessible")
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	_, _, err := svc.DownloadNode(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
}

// --- IndexDirectory ---

func TestIndexDirectory_ReturnsRootNodes_WhenUUIDIsRoot(t *testing.T) {
	userID := uuid.New()
	rootNodes := []*models.ServerNode{{Id: uuid.New()}, {Id: uuid.New()}}

	nodes := &testutil.MockNodesRepo{
		GetRootNodesByUserIdFunc: func(_ context.Context, uid uuid.UUID) ([]*models.ServerNode, error) {
			assert.Equal(t, userID, uid)
			return rootNodes, nil
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})

	for _, rootParam := range []string{"root", "", "00000000-0000-0000-0000-000000000000"} {
		got, err := svc.IndexDirectory(context.Background(), userID, rootParam)
		require.NoError(t, err, "param: %q", rootParam)
		assert.Equal(t, rootNodes, got, "param: %q", rootParam)
	}
}

func TestIndexDirectory_ReturnsChildren_WhenParentUUIDGiven(t *testing.T) {
	userID := uuid.New()
	parentID := uuid.New()
	children := []*models.ServerNode{{Id: uuid.New()}}

	nodes := &testutil.MockNodesRepo{
		GetChildrensByUserIdFunc: func(_ context.Context, pid uuid.UUID, uid uuid.UUID) ([]*models.ServerNode, error) {
			assert.Equal(t, parentID, pid)
			assert.Equal(t, userID, uid)
			return children, nil
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	got, err := svc.IndexDirectory(context.Background(), userID, parentID.String())

	require.NoError(t, err)
	assert.Equal(t, children, got)
}

// --- DeleteNode ---

func TestDeleteNode_DeletesDBRecordsAndFiles_OnSuccess(t *testing.T) {
	nodeID := uuid.New()
	child1 := uuid.New()
	child2 := uuid.New()
	deletedIDs := []uuid.UUID{nodeID, child1, child2}

	var storageDeleted []string

	nodes := &testutil.MockNodesRepo{
		DeleteRecursiveByUserIdFunc: func(_ context.Context, id uuid.UUID, _ uuid.UUID) ([]uuid.UUID, error) {
			assert.Equal(t, nodeID, id)
			return deletedIDs, nil
		},
	}
	store := &testutil.MockFileStorage{
		DeleteFunc: func(id string) error {
			storageDeleted = append(storageDeleted, id)
			return nil
		},
	}

	svc := services.NewNodeService(nodes, store)
	err := svc.DeleteNode(context.Background(), uuid.New(), nodeID)

	require.NoError(t, err)

	// Every UUID returned by the repo must have its files deleted
	expected := []string{nodeID.String(), child1.String(), child2.String()}
	assert.ElementsMatch(t, expected, storageDeleted)
}

// --- SetFavourite ---

func TestSetFavourite_UpdatesNode_OnSuccess(t *testing.T) {
	nodeID := uuid.New()
	userID := uuid.New()

	nodes := &testutil.MockNodesRepo{
		SetFavouriteFunc: func(_ context.Context, nid uuid.UUID, uid uuid.UUID, fav bool) error {
			assert.Equal(t, nodeID, nid)
			assert.Equal(t, userID, uid)
			assert.True(t, fav)
			return nil
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	err := svc.SetFavourite(context.Background(), userID, nodeID, true)

	require.NoError(t, err)
}

func TestSetFavourite_ReturnsError_WhenRepoFails(t *testing.T) {
	nodes := &testutil.MockNodesRepo{
		SetFavouriteFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ bool) error {
			return errors.New("node not found or not accessible")
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	err := svc.SetFavourite(context.Background(), uuid.New(), uuid.New(), true)

	require.Error(t, err)
}

// --- GetFavourites ---

func TestGetFavourites_ReturnsFavouriteNodes_OnSuccess(t *testing.T) {
	userID := uuid.New()
	favNodes := []*models.ServerNode{
		{Id: uuid.New(), IsFavourite: true},
		{Id: uuid.New(), IsFavourite: true},
	}

	nodes := &testutil.MockNodesRepo{
		GetFavouritesByUserIdFunc: func(_ context.Context, uid uuid.UUID) ([]*models.ServerNode, error) {
			assert.Equal(t, userID, uid)
			return favNodes, nil
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	got, err := svc.GetFavourites(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, favNodes, got)
}

func TestGetFavourites_ReturnsError_WhenRepoFails(t *testing.T) {
	nodes := &testutil.MockNodesRepo{
		GetFavouritesByUserIdFunc: func(_ context.Context, _ uuid.UUID) ([]*models.ServerNode, error) {
			return nil, errors.New("db error")
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	_, err := svc.GetFavourites(context.Background(), uuid.New())

	require.Error(t, err)
}

func TestDeleteNode_ReturnsError_WhenNodeNotFound(t *testing.T) {
	nodes := &testutil.MockNodesRepo{
		DeleteRecursiveByUserIdFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]uuid.UUID, error) {
			return nil, errors.New("node not found or not accessible")
		},
	}

	svc := services.NewNodeService(nodes, &testutil.MockFileStorage{})
	err := svc.DeleteNode(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
}
