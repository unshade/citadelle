package controllers

import (
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/unshade/citadelle/internal/services"
)

// UserController handles all user-related HTTP routes
type UserController struct{}

// NewUserController creates a new UserController
func NewUserController() *UserController {
	return &UserController{}
}

// Register creates route group for user endpoints under the given group (e.g., /api)
func (uc *UserController) Register(group *fuego.Server) {
	// Create /users sub-group under the parent group
	usersGroup := fuego.Group(group, "/users", option.Tags("users"))

	// Register routes under /users
	fuego.Post(usersGroup, "/", uc.CreateUser)
	fuego.Post(usersGroup, "/login", uc.Login)
}

// CreateUserRequest represents the input for user creation
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// CreateUserResponse represents the output for user creation
type CreateUserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

// CreateUser handles user registration
func (uc *UserController) CreateUser(c fuego.ContextWithBody[CreateUserRequest]) (*CreateUserResponse, error) {
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

// LoginRequest represents the input for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the output for user login
type LoginResponse struct {
	Message string `json:"message"`
}

// Login handles user authentication
// Note: In a real implementation, you would return a JWT token or session cookie
func (uc *UserController) Login(c fuego.ContextWithBody[LoginRequest]) (*LoginResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}

	user, masterKey, err := services.Login(body.Email, body.Password)
	if err != nil {
		return nil, err
	}

	// In production, don't log the masterKey!
	_ = masterKey
	_ = user

	return &LoginResponse{
		Message: "Login successful",
	}, nil
}
