package handlers_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/handlers"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
	"github.com/unshade/citadelle/internal/testutil"
)

func TestGetUser(t *testing.T) {
	userID := uuid.New()

	t.Run("returns user for valid ID", func(t *testing.T) {
		expected := &models.User{Id: userID}
		mock := &testutil.MockUserService{
			GetUserFunc: func(_ context.Context, id uuid.UUID) (*models.User, error) {
				assert.Equal(t, userID, id)
				return expected, nil
			},
		}
		h := handlers.NewUserHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = userID.String()

		resp, err := h.GetUser(ctx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, userID, resp.Data.Id)
	})

	t.Run("returns error for invalid UUID", func(t *testing.T) {
		h := handlers.NewUserHandler(&testutil.MockUserService{})

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = "not-a-uuid"

		_, err := h.GetUser(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockUserService{
			GetUserFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
				return nil, errors.New("not found")
			},
		}
		h := handlers.NewUserHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = userID.String()

		_, err := h.GetUser(ctx)
		require.Error(t, err)
	})
}

func TestCreateUser(t *testing.T) {
	salt := []byte("salt-bytes")
	encKey := []byte("enc-master-key")
	encChallenge := []byte("enc-challenge")

	validBody := handlers.CreateUserRequest{
		B64Salt:               base64.StdEncoding.EncodeToString(salt),
		B64EncryptedMasterKey: base64.StdEncoding.EncodeToString(encKey),
		B64EncryptedChallenge: base64.StdEncoding.EncodeToString(encChallenge),
		ClearChallenge:        "mysecret",
	}

	t.Run("creates user and returns UUID", func(t *testing.T) {
		newID := uuid.New()
		mock := &testutil.MockUserService{
			CreateUserFunc: func(_ context.Context, input services.CreateUserInput) (uuid.UUID, error) {
				assert.Equal(t, salt, input.Salt)
				assert.Equal(t, encKey, input.EncryptedMasterKey)
				assert.Equal(t, encChallenge, input.EncryptedChallenge)
				assert.Equal(t, "mysecret", input.ClearChallenge)
				return newID, nil
			},
		}
		h := handlers.NewUserHandler(mock)

		ctx := fuego.NewMockContext[handlers.CreateUserRequest, any](validBody, nil)

		resp, err := h.CreateUser(ctx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, newID.String(), resp.Data.Uuid)
	})

	t.Run("returns error for invalid base64 salt", func(t *testing.T) {
		h := handlers.NewUserHandler(&testutil.MockUserService{})

		body := handlers.CreateUserRequest{
			B64Salt:               "!!!not-base64!!!",
			B64EncryptedMasterKey: base64.StdEncoding.EncodeToString(encKey),
			B64EncryptedChallenge: base64.StdEncoding.EncodeToString(encChallenge),
		}
		ctx := fuego.NewMockContext[handlers.CreateUserRequest, any](body, nil)

		_, err := h.CreateUser(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockUserService{
			CreateUserFunc: func(_ context.Context, _ services.CreateUserInput) (uuid.UUID, error) {
				return uuid.Nil, errors.New("db error")
			},
		}
		h := handlers.NewUserHandler(mock)

		ctx := fuego.NewMockContext[handlers.CreateUserRequest, any](validBody, nil)

		_, err := h.CreateUser(ctx)
		require.Error(t, err)
	})
}
