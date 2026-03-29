package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
	"github.com/unshade/citadelle/internal/testutil"
)

func TestGetUser_ReturnsUser_WhenFound(t *testing.T) {
	userID := uuid.New()
	expected := &models.User{Id: userID, ClearChallenge: "challenge"}

	repo := &testutil.MockUsersRepo{
		GetByIdFunc: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			assert.Equal(t, userID, id)
			return expected, nil
		},
	}

	svc := services.NewUserService(repo)
	got, err := svc.GetUser(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetUser_ReturnsError_WhenNotFound(t *testing.T) {
	repo := &testutil.MockUsersRepo{
		GetByIdFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, errors.New("record not found")
		},
	}

	svc := services.NewUserService(repo)
	_, err := svc.GetUser(context.Background(), uuid.New())

	require.Error(t, err)
}

func TestCreateUser_ReturnsNewUUID_OnSuccess(t *testing.T) {
	var createdUser models.User

	repo := &testutil.MockUsersRepo{
		CreateFunc: func(_ context.Context, u models.User) error {
			createdUser = u
			return nil
		},
	}

	input := services.CreateUserInput{
		Salt:               []byte("salt"),
		MasterKeyNonce:     []byte("mk-nonce"),
		EncryptedMasterKey: []byte("key"),
		ChallengeNonce:     []byte("ch-nonce"),
		EncryptedChallenge: []byte("challenge"),
		ClearChallenge:     "clear",
	}

	svc := services.NewUserService(repo)
	id, err := svc.CreateUser(context.Background(), input)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)

	// The repo received a user whose fields match the input
	assert.Equal(t, id, createdUser.Id)
	assert.Equal(t, input.Salt, createdUser.Salt)
	assert.Equal(t, input.MasterKeyNonce, createdUser.MasterKeyNonce)
	assert.Equal(t, input.EncryptedMasterKey, createdUser.EncryptedMasterKey)
	assert.Equal(t, input.ChallengeNonce, createdUser.ChallengeNonce)
	assert.Equal(t, input.EncryptedChallenge, createdUser.EncryptedChallenge)
	assert.Equal(t, input.ClearChallenge, createdUser.ClearChallenge)
}

func TestCreateUser_ReturnsError_WhenRepoFails(t *testing.T) {
	repo := &testutil.MockUsersRepo{
		CreateFunc: func(_ context.Context, _ models.User) error {
			return errors.New("db connection lost")
		},
	}

	svc := services.NewUserService(repo)
	id, err := svc.CreateUser(context.Background(), services.CreateUserInput{})

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}
