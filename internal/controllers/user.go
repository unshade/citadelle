package controllers

import (
	"encoding/base64"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/repositories"
)

type UserController struct {
	Database repositories.Database
}

func NewUserController(database repositories.Database) *UserController {
	return &UserController{Database: database}
}

func (uc *UserController) Register(group *fuego.Server) {
	usersGroup := fuego.Group(group, "/users", option.Tags("users"))

	fuego.Post(usersGroup, "/", uc.CreateUser)
	fuego.Get(usersGroup, "/{uuid}", uc.GetUser)
}

func (uc *UserController) GetUser(c fuego.ContextNoBody) (*ApiResponse[models.User], error) {
	id, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return nil, err
	}
	user, err := uc.Database.Users.GetById(c.Context(), id)
	if err != nil {
		return nil, err
	}

	return NewApiResponse(user, "ok")
}

type CreateUserRequest struct {
	B64Salt               string `json:"b64Salt"`
	B64EncryptedMasterKey string `json:"b64EncryptedMasterKey"`
	B64EncryptedChallenge string `json:"b64EncryptedChallenge"`
	ClearChallenge        string `json:"clearChallenge"`
}

type CreateUserResponse struct {
	Uuid string `json:"uuid"`
}

func (uc *UserController) CreateUser(c fuego.ContextWithBody[CreateUserRequest]) (*ApiResponse[CreateUserResponse], error) {
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[CreateUserResponse](err)
	}

	generatedUuid := uuid.New()
	salt, err := base64.StdEncoding.DecodeString(body.B64Salt)
	if err != nil {
		return NewErrorResponse[CreateUserResponse](err)
	}

	encryptedMasterKey, err := base64.StdEncoding.DecodeString(body.B64EncryptedMasterKey)
	if err != nil {
		return NewErrorResponse[CreateUserResponse](err)
	}

	encryptedChallenge, err := base64.StdEncoding.DecodeString(body.B64EncryptedChallenge)
	if err != nil {
		return NewErrorResponse[CreateUserResponse](err)
	}

	user := models.User{
		Id:                 generatedUuid,
		Salt:               salt,
		EncryptedMasterKey: encryptedMasterKey,
		EncryptedChallenge: encryptedChallenge,
		ClearChallenge:     body.ClearChallenge,
	}

	if err := uc.Database.Users.Create(c.Context(), user); err != nil {
		return NewErrorResponse[CreateUserResponse](err)
	}

	return NewApiResponse(&CreateUserResponse{
		Uuid: user.Id.String(),
	}, "user created")
}
