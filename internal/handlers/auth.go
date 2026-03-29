package handlers

import (
	"errors"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/services"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(group *fuego.Server) {
	authGroup := fuego.Group(group, "/auth", option.Tags("auth"))
	fuego.Get(authGroup, "/challenge/{uuid}", h.GetChallenge)
	fuego.Post(authGroup, "/verify", h.VerifyChallenge)
}

// ChallengeResponse carries the sealed challenge as separate nonce + ciphertext.
type ChallengeResponse struct {
	B64ChallengeNonce     string `json:"b64ChallengeNonce"`
	B64EncryptedChallenge string `json:"b64EncryptedChallenge"`
}

func (h *AuthHandler) GetChallenge(c fuego.ContextNoBody) (*ApiResponse[ChallengeResponse], error) {
	userID, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[ChallengeResponse](err)
	}
	data, err := h.authService.GetChallenge(c.Context(), userID)
	if err != nil {
		return NewErrorResponse[ChallengeResponse](err)
	}
	return NewApiResponse(&ChallengeResponse{
		B64ChallengeNonce:     data.Nonce,
		B64EncryptedChallenge: data.Ciphertext,
	}, "challenge retrieved")
}

type VerifyRequest struct {
	UserUUID       string `json:"userUuid"`
	ClearChallenge string `json:"clearChallenge"`
}

type VerifyResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) VerifyChallenge(c fuego.ContextWithBody[VerifyRequest]) (*ApiResponse[VerifyResponse], error) {
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[VerifyResponse](err)
	}
	userID, err := uuid.Parse(body.UserUUID)
	if err != nil {
		return NewErrorResponse[VerifyResponse](errors.New("invalid user ID"))
	}
	token, err := h.authService.VerifyChallenge(c.Context(), userID, body.ClearChallenge)
	if err != nil {
		return NewErrorResponse[VerifyResponse](err)
	}
	return NewApiResponse(&VerifyResponse{Token: token}, "authentication successful")
}
