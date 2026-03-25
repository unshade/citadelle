package handlers

import (
	"encoding/base64"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(group *fuego.Server) {
	usersGroup := fuego.Group(group, "/users", option.Tags("users"))
	fuego.Post(usersGroup, "/", h.CreateUser)
	fuego.Get(usersGroup, "/{uuid}", h.GetUser)
}

func (h *UserHandler) GetUser(c fuego.ContextNoBody) (*ApiResponse[models.User], error) {
	id, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[models.User](err)
	}
	user, err := h.userService.GetUser(c.Context(), id)
	if err != nil {
		return NewErrorResponse[models.User](err)
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

func (h *UserHandler) CreateUser(c fuego.ContextWithBody[CreateUserRequest]) (*ApiResponse[CreateUserResponse], error) {
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[CreateUserResponse](err)
	}

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

	id, err := h.userService.CreateUser(c.Context(), services.CreateUserInput{
		Salt:               salt,
		EncryptedMasterKey: encryptedMasterKey,
		EncryptedChallenge: encryptedChallenge,
		ClearChallenge:     body.ClearChallenge,
	})
	if err != nil {
		return NewErrorResponse[CreateUserResponse](err)
	}
	return NewApiResponse(&CreateUserResponse{Uuid: id.String()}, "user created")
}
