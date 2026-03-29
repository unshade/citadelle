package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/handlers"
	"github.com/unshade/citadelle/internal/services"
	"github.com/unshade/citadelle/internal/testutil"
)

func TestGetChallenge(t *testing.T) {
	userID := uuid.New()

	t.Run("returns challenge for valid user ID", func(t *testing.T) {
		mock := &testutil.MockAuthService{
			GetChallengeFunc: func(_ context.Context, id uuid.UUID) (services.ChallengeData, error) {
				assert.Equal(t, userID, id)
				return services.ChallengeData{Nonce: "nonce==", Ciphertext: "ciphertext=="}, nil
			},
		}
		h := handlers.NewAuthHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = userID.String()

		resp, err := h.GetChallenge(ctx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "nonce==", resp.Data.B64ChallengeNonce)
		assert.Equal(t, "ciphertext==", resp.Data.B64EncryptedChallenge)
	})

	t.Run("returns error for invalid UUID", func(t *testing.T) {
		h := handlers.NewAuthHandler(&testutil.MockAuthService{})

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = "not-a-uuid"

		_, err := h.GetChallenge(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockAuthService{
			GetChallengeFunc: func(_ context.Context, _ uuid.UUID) (services.ChallengeData, error) {
				return services.ChallengeData{}, errors.New("user not found")
			},
		}
		h := handlers.NewAuthHandler(mock)

		ctx := fuego.NewMockContextNoBody()
		ctx.PathParams["uuid"] = userID.String()

		_, err := h.GetChallenge(ctx)
		require.Error(t, err)
	})
}

func TestVerifyChallenge(t *testing.T) {
	userID := uuid.New()

	t.Run("returns token for valid credentials", func(t *testing.T) {
		mock := &testutil.MockAuthService{
			VerifyChallengeFunc: func(_ context.Context, id uuid.UUID, challenge string) (string, error) {
				assert.Equal(t, userID, id)
				assert.Equal(t, "mysecret", challenge)
				return "jwt.token.here", nil
			},
		}
		h := handlers.NewAuthHandler(mock)

		body := handlers.VerifyRequest{
			UserUUID:       userID.String(),
			ClearChallenge: "mysecret",
		}
		ctx := fuego.NewMockContext[handlers.VerifyRequest, any](body, nil)

		resp, err := h.VerifyChallenge(ctx)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "jwt.token.here", resp.Data.Token)
	})

	t.Run("returns error for invalid user UUID", func(t *testing.T) {
		h := handlers.NewAuthHandler(&testutil.MockAuthService{})

		body := handlers.VerifyRequest{
			UserUUID:       "not-a-uuid",
			ClearChallenge: "mysecret",
		}
		ctx := fuego.NewMockContext[handlers.VerifyRequest, any](body, nil)

		_, err := h.VerifyChallenge(ctx)
		require.Error(t, err)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		mock := &testutil.MockAuthService{
			VerifyChallengeFunc: func(_ context.Context, _ uuid.UUID, _ string) (string, error) {
				return "", errors.New("wrong challenge")
			},
		}
		h := handlers.NewAuthHandler(mock)

		body := handlers.VerifyRequest{
			UserUUID:       userID.String(),
			ClearChallenge: "wrongsecret",
		}
		ctx := fuego.NewMockContext[handlers.VerifyRequest, any](body, nil)

		_, err := h.VerifyChallenge(ctx)
		require.Error(t, err)
	})
}
