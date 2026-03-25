package services

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/repositories"
)

type AuthService interface {
	GetChallenge(ctx context.Context, userID uuid.UUID) (string, error)
	VerifyChallenge(ctx context.Context, userID uuid.UUID, clearChallenge string) (string, error)
}

type authService struct {
	users     repositories.UsersRepo
	jwtSecret []byte
}

func NewAuthService(users repositories.UsersRepo, jwtSecret string) AuthService {
	return &authService{users: users, jwtSecret: []byte(jwtSecret)}
}

func (s *authService) GetChallenge(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.users.GetById(ctx, userID)
	if err != nil {
		return "", errors.New("user not found")
	}
	return base64.StdEncoding.EncodeToString(user.EncryptedChallenge), nil
}

func (s *authService) VerifyChallenge(ctx context.Context, userID uuid.UUID, clearChallenge string) (string, error) {
	user, err := s.users.GetById(ctx, userID)
	if err != nil {
		return "", errors.New("user not found")
	}
	if clearChallenge != user.ClearChallenge {
		return "", errors.New("invalid challenge response")
	}
	return s.generateToken(userID)
}

func (s *authService) generateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	})
	return token.SignedString(s.jwtSecret)
}
