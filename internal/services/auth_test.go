package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
	"github.com/unshade/citadelle/internal/testutil"
)

const testJWTSecret = "test-secret-min-32-characters-long!!"

func TestGetChallenge_ReturnsBase64Challenge_WhenUserExists(t *testing.T) {
	userID := uuid.New()
	encryptedChallenge := []byte("encrypted-challenge-bytes")

	repo := &testutil.MockUsersRepo{
		GetByIdFunc: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			assert.Equal(t, userID, id)
			return &models.User{
				Id:                 userID,
				EncryptedChallenge: encryptedChallenge,
			}, nil
		},
	}

	svc := services.NewAuthService(repo, testJWTSecret)
	b64, err := svc.GetChallenge(context.Background(), userID)

	require.NoError(t, err)
	assert.NotEmpty(t, b64)
}

func TestGetChallenge_ReturnsError_WhenUserNotFound(t *testing.T) {
	repo := &testutil.MockUsersRepo{
		GetByIdFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, errors.New("record not found")
		},
	}

	svc := services.NewAuthService(repo, testJWTSecret)
	_, err := svc.GetChallenge(context.Background(), uuid.New())

	require.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestVerifyChallenge_ReturnsSignedJWT_WhenChallengeMatches(t *testing.T) {
	userID := uuid.New()
	clearChallenge := "the-clear-challenge"

	repo := &testutil.MockUsersRepo{
		GetByIdFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{
				Id:             userID,
				ClearChallenge: clearChallenge,
			}, nil
		},
	}

	svc := services.NewAuthService(repo, testJWTSecret)
	tokenStr, err := svc.VerifyChallenge(context.Background(), userID, clearChallenge)

	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	// Parse the token and verify the user_id claim is correct
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(testJWTSecret), nil
	})
	require.NoError(t, err)
	claims := token.Claims.(jwt.MapClaims)
	assert.Equal(t, userID.String(), claims["user_id"])
}

func TestVerifyChallenge_ReturnsError_WhenChallengeIsWrong(t *testing.T) {
	userID := uuid.New()

	repo := &testutil.MockUsersRepo{
		GetByIdFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{Id: userID, ClearChallenge: "correct-challenge"}, nil
		},
	}

	svc := services.NewAuthService(repo, testJWTSecret)
	_, err := svc.VerifyChallenge(context.Background(), userID, "wrong-challenge")

	require.Error(t, err)
	assert.Equal(t, "invalid challenge response", err.Error())
}

func TestVerifyChallenge_ReturnsError_WhenUserNotFound(t *testing.T) {
	repo := &testutil.MockUsersRepo{
		GetByIdFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, errors.New("record not found")
		},
	}

	svc := services.NewAuthService(repo, testJWTSecret)
	_, err := svc.VerifyChallenge(context.Background(), uuid.New(), "any-challenge")

	require.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}
