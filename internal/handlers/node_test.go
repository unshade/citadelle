package handlers_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/handlers"
	"github.com/unshade/citadelle/internal/middleware"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
	"github.com/unshade/citadelle/internal/testutil"
)

// injectUserID sets the user ID in the context the same way the JWT middleware does.
func injectUserID(ctx *fuego.MockContext[any, any], userID uuid.UUID) {
	ctx.CommonCtx = context.WithValue(ctx.CommonCtx, middleware.UserIDContextKey, userID.String())
}

func injectUserIDBody[B any](ctx *fuego.MockContext[B, any], userID uuid.UUID) {
	ctx.CommonCtx = context.WithValue(ctx.CommonCtx, middleware.UserIDContextKey, userID.String())
}

func TestCreateNode(t *testing.T) {
	userID := uuid.New()
	nodeID := uuid.New()

	keyNonce := []byte("key-nonce")
	encKey := []byte("enc-key")
	contentNonce := []byte("content-nonce")
	nameNonce := []byte("name-nonce")
	encName := []byte("enc-name")

	validBody := handlers.CreateNodeRequest{
		B64KeyNonce:               base64.StdEncoding.EncodeToString(keyNonce),
		B64EncryptedEncryptionKey: base64.StdEncoding.EncodeToString(encKey),
		B64ContentNonce:           base64.StdEncoding.EncodeToString(contentNonce),
		B64NameNonce:              base64.StdEncoding.EncodeToString(nameNonce),
		B64EncryptedName:          base64.StdEncoding.EncodeToString(encName),
		B64PathNonce:              "pathnonce==",
		B64EncryptedPath:          "encpath==",
		IsDirectory:               false,
		Version:                   1,
	}

	t.Run("creates node and returns UUID", func(t *testing.T) {
		mock := &testutil.MockNodeService{
			CreateNodeFunc: func(_ context.Context, uid uuid.UUID, input services.CreateNodeInput) (uuid.UUID, error) {
				assert.Equal(t, userID, uid)
				assert.Equal(t, keyNonce, input.KeyNonce)
				assert.Equal(t, encKey, input.EncryptedKey)
				assert.Equal(t, contentNonce, input.ContentNonce)
				assert.Equal(t, nameNonce, input.NameNonce)
				assert.Equal(t, encName, input.EncryptedName)
				return nodeID, nil
			},
		}
		h := handlers.NewNodeHandler(mock)

		ctx := fuego.NewMockContext[handlers.CreateNodeRequest, any](validBody, nil)
		injectUserIDBody(ctx, userID)

		resp, err := h.CreateNode(ctx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, nodeID.String(), resp.Data.Uuid)
	})

	t.Run("returns error when no user ID in context", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContext[handlers.CreateNodeRequest, any](validBody, nil)
		// no user ID injected

		_, err := h.CreateNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error for invalid base64 key", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		body := handlers.CreateNodeRequest{
			B64KeyNonce:               "!!!not-base64!!!",
			B64EncryptedEncryptionKey: base64.StdEncoding.EncodeToString(encKey),
			B64ContentNonce:           base64.StdEncoding.EncodeToString(contentNonce),
			B64NameNonce:              base64.StdEncoding.EncodeToString(nameNonce),
			B64EncryptedName:          base64.StdEncoding.EncodeToString(encName),
		}
		ctx := fuego.NewMockContext[handlers.CreateNodeRequest, any](body, nil)
		injectUserIDBody(ctx, userID)

		_, err := h.CreateNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockNodeService{
			CreateNodeFunc: func(_ context.Context, _ uuid.UUID, _ services.CreateNodeInput) (uuid.UUID, error) {
				return uuid.Nil, errors.New("db error")
			},
		}
		h := handlers.NewNodeHandler(mock)

		ctx := fuego.NewMockContext[handlers.CreateNodeRequest, any](validBody, nil)
		injectUserIDBody(ctx, userID)

		_, err := h.CreateNode(ctx)
		require.Error(t, err)
	})
}

func TestSaveNode(t *testing.T) {
	userID := uuid.New()
	nodeID := uuid.New()

	// multipartRequest builds a request with a file field named "encryptedFile".
	multipartRequest := func(t *testing.T, content []byte) *httptest.ResponseRecorder {
		t.Helper()
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, err := w.CreateFormFile("encryptedFile", "data.bin")
		require.NoError(t, err)
		_, err = io.Copy(fw, bytes.NewReader(content))
		require.NoError(t, err)
		require.NoError(t, w.Close())
		_ = content
		return httptest.NewRecorder()
	}

	t.Run("returns error when no user ID in context", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = nodeID.String()
		_ = multipartRequest(t, []byte("filedata"))

		_, err := h.SaveNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error for invalid node UUID", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = "not-a-uuid"

		_, err := h.SaveNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when no file in request", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = nodeID.String()
		// empty request — FormFile returns an error (no multipart file present)
		ctx.SetRequest(httptest.NewRequestWithContext(t.Context(), "POST", "/nodes/"+nodeID.String(), nil))

		_, err := h.SaveNode(ctx)
		require.Error(t, err)
	})
}

func TestIndexDirectory(t *testing.T) {
	userID := uuid.New()
	nodeID := uuid.New()

	nodes := []*models.ServerNode{
		{Id: nodeID, IsDirectory: false},
	}

	t.Run("returns nodes for valid directory", func(t *testing.T) {
		mock := &testutil.MockNodeService{
			IndexDirectoryFunc: func(_ context.Context, uid uuid.UUID, param string) ([]*models.ServerNode, error) {
				assert.Equal(t, userID, uid)
				assert.Equal(t, nodeID.String(), param)
				return nodes, nil
			},
		}
		h := handlers.NewNodeHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = nodeID.String()

		resp, err := h.IndexDirectory(ctx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data.Nodes, 1)
	})

	t.Run("returns error when no user ID in context", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = nodeID.String()

		_, err := h.IndexDirectory(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockNodeService{
			IndexDirectoryFunc: func(_ context.Context, _ uuid.UUID, _ string) ([]*models.ServerNode, error) {
				return nil, errors.New("not found")
			},
		}
		h := handlers.NewNodeHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = nodeID.String()

		_, err := h.IndexDirectory(ctx)
		require.Error(t, err)
	})
}

func TestDeleteNode(t *testing.T) {
	userID := uuid.New()
	nodeID := uuid.New()

	t.Run("deletes node successfully", func(t *testing.T) {
		mock := &testutil.MockNodeService{
			DeleteNodeFunc: func(_ context.Context, uid uuid.UUID, nid uuid.UUID) error {
				assert.Equal(t, userID, uid)
				assert.Equal(t, nodeID, nid)
				return nil
			},
		}
		h := handlers.NewNodeHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = nodeID.String()

		resp, err := h.DeleteNode(ctx)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("returns error when no user ID in context", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = nodeID.String()

		_, err := h.DeleteNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error for invalid node UUID", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = "not-a-uuid"

		_, err := h.DeleteNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockNodeService{
			DeleteNodeFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
				return errors.New("delete failed")
			},
		}
		h := handlers.NewNodeHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = nodeID.String()

		_, err := h.DeleteNode(ctx)
		require.Error(t, err)
	})
}

func TestDownloadNode(t *testing.T) {
	userID := uuid.New()
	nodeID := uuid.New()

	t.Run("returns error when no user ID in context", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = nodeID.String()

		_, err := h.DownloadNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error for invalid node UUID", func(t *testing.T) {
		h := handlers.NewNodeHandler(&testutil.MockNodeService{})

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = "not-a-uuid"

		_, err := h.DownloadNode(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockNodeService{
			DownloadNodeFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.ServerNode, []byte, error) {
				return nil, nil, errors.New("storage error")
			},
		}
		h := handlers.NewNodeHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		injectUserID(ctx, userID)
		ctx.PathParams["uuid"] = nodeID.String()

		_, err := h.DownloadNode(ctx)
		require.Error(t, err)
	})
}
