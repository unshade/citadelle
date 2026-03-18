package controllers

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/repositories"
)

type AuthController struct {
	Database  repositories.Database
	JWTSecret []byte
}

func NewAuthController(database repositories.Database, jwtSecret string) *AuthController {
	return &AuthController{
		Database:  database,
		JWTSecret: []byte(jwtSecret),
	}
}

func (ac *AuthController) Register(group *fuego.Server) {
	authGroup := fuego.Group(group, "/auth", option.Tags("auth"))

	fuego.Get(authGroup, "/challenge/{uuid}", ac.GetChallenge)
	fuego.Post(authGroup, "/verify", ac.VerifyChallenge)
}

// ChallengeResponse returns the encrypted challenge for the user to decrypt
type ChallengeResponse struct {
	B64EncryptedChallenge string `json:"b64EncryptedChallenge"`
}

// GetChallenge returns the encrypted challenge for a user
// The client must decrypt this using their master key and send back the clear text
func (ac *AuthController) GetChallenge(c fuego.ContextNoBody) (*ApiResponse[ChallengeResponse], error) {
	userID, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[ChallengeResponse](err)
	}

	user, err := ac.Database.Users.GetById(c.Context(), userID)
	if err != nil {
		return NewErrorResponse[ChallengeResponse](errors.New("user not found"))
	}

	return NewApiResponse(&ChallengeResponse{
		B64EncryptedChallenge: base64.StdEncoding.EncodeToString(user.EncryptedChallenge),
	}, "challenge retrieved")
}

// VerifyRequest contains the decrypted challenge response
type VerifyRequest struct {
	UserUUID       string `json:"userUuid"`
	ClearChallenge string `json:"clearChallenge"`
}

// VerifyResponse contains the JWT token if verification succeeds
type VerifyResponse struct {
	Token string `json:"token"`
}

// VerifyChallenge checks if the decrypted challenge matches
// If successful, returns a JWT token
func (ac *AuthController) VerifyChallenge(c fuego.ContextWithBody[VerifyRequest]) (*ApiResponse[VerifyResponse], error) {
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[VerifyResponse](err)
	}

	userID, err := uuid.Parse(body.UserUUID)
	if err != nil {
		return NewErrorResponse[VerifyResponse](errors.New("invalid user ID"))
	}

	user, err := ac.Database.Users.GetById(c.Context(), userID)
	if err != nil {
		return NewErrorResponse[VerifyResponse](errors.New("user not found"))
	}

	// Verify the challenge
	if body.ClearChallenge != user.ClearChallenge {
		return NewErrorResponse[VerifyResponse](errors.New("invalid challenge response"))
	}

	// Generate JWT token
	token, err := ac.generateToken(userID)
	if err != nil {
		return NewErrorResponse[VerifyResponse](err)
	}

	return NewApiResponse(&VerifyResponse{
		Token: token,
	}, "authentication successful")
}

// generateToken creates a JWT token for the authenticated user
func (ac *AuthController) generateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiry
		"iat":     time.Now().Unix(),
	})

	return token.SignedString(ac.JWTSecret)
}
