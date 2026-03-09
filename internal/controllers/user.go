package controllers

import (
	"github.com/go-fuego/fuego"
	"github.com/unshade/citadelle/internal/services"
)

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type CreateUserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func CreateUser(c fuego.ContextWithBody[CreateUserRequest]) (*CreateUserResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}

	user, err := services.CreateUser(body.Email, body.Password)
	if err != nil {
		return nil, err
	}

	return &CreateUserResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Message string `json:"message"`
}

func Login(c fuego.ContextWithBody[LoginRequest]) (*LoginResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}

	user, masterKey, err := services.Login(body.Email, body.Password)
	if err != nil {
		return nil, err
	}

	_ = masterKey
	_ = user

	return &LoginResponse{
		Message: "Login successful",
	}, nil
}
