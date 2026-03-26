package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/middleware"
)

const testSecret = "test-secret-key"

func makeToken(t *testing.T, userID string, secret string, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(userID))
}

func TestAuthenticate(t *testing.T) {
	m := middleware.NewJWTAuthMiddleware(testSecret)
	handler := m.Authenticate(http.HandlerFunc(okHandler))

	t.Run("passes request with valid token", func(t *testing.T) {
		userID := uuid.New().String()
		token := makeToken(t, userID, testSecret, time.Hour)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, userID, rec.Body.String())
	})

	t.Run("rejects request with no Authorization header", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects malformed Authorization header", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "NotBearer token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects token signed with wrong secret", func(t *testing.T) {
		token := makeToken(t, uuid.New().String(), "wrong-secret", time.Hour)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		token := makeToken(t, uuid.New().String(), testSecret, -time.Hour)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestGetUserID(t *testing.T) {
	t.Run("returns empty string when key not in context", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		assert.Equal(t, "", middleware.GetUserID(req.Context()))
	})
}
