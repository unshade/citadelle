package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/repositories"
)

type UserService interface {
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (uuid.UUID, error)
}

type CreateUserInput struct {
	Salt               []byte
	MasterKeyNonce     []byte
	EncryptedMasterKey []byte
	ChallengeNonce     []byte
	EncryptedChallenge []byte
	ClearChallenge     string
}

type userService struct {
	users repositories.UsersRepo
}

func NewUserService(users repositories.UsersRepo) UserService {
	return &userService{users: users}
}

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.users.GetById(ctx, id)
}

func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (uuid.UUID, error) {
	id := uuid.New()
	user := models.User{
		Id:                 id,
		Salt:               input.Salt,
		MasterKeyNonce:     input.MasterKeyNonce,
		EncryptedMasterKey: input.EncryptedMasterKey,
		ChallengeNonce:     input.ChallengeNonce,
		EncryptedChallenge: input.EncryptedChallenge,
		ClearChallenge:     input.ClearChallenge,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
