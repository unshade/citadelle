package handlers

import (
	"encoding/base64"
	"errors"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/middleware"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/services"
)

type NodeHandler struct {
	nodeService services.NodeService
}

func NewNodeHandler(nodeService services.NodeService) *NodeHandler {
	return &NodeHandler{nodeService: nodeService}
}

func (h *NodeHandler) Register(group *fuego.Server, authMiddleware *middleware.JWTAuthMiddleware) {
	authOption := option.Middleware(authMiddleware.Authenticate)
	nodesGroup := fuego.Group(group, "/nodes", option.Tags("nodes"))

	fuego.Post(nodesGroup, "/", h.CreateNode, authOption)
	fuego.Post(nodesGroup, "/{uuid}", h.SaveNode, authOption)
	fuego.Get(nodesGroup, "/{uuid}/download", h.DownloadNode, authOption)
	fuego.Get(nodesGroup, "/{uuid}", h.IndexDirectory, authOption)
	fuego.Delete(nodesGroup, "/{uuid}", h.DeleteNode, authOption)
}

// mustParseUserID extracts and parses the authenticated user ID from context.
func mustParseUserID(c interface{ Context() interface{ Value(any) any } }) (uuid.UUID, error) {
	return uuid.Nil, nil // unused placeholder; each handler calls middleware.GetUserID directly
}

// --- CreateNode ---

type CreateNodeRequest struct {
	B64EncryptedEncryptionKey string  `json:"b64EncryptedEncryptionKey"`
	B64EncryptionNonce        string  `json:"b64EncryptionNonce"`
	B64EncryptedName          string  `json:"b64EncryptedName"`
	B64EncryptedPath          string  `json:"b64EncryptedPath"`
	IsDirectory               bool    `json:"isDirectory"`
	ParentUuid                *string `json:"parentUuid"`
	Version                   uint64  `json:"version"`
}

type CreateNodeResponse struct {
	Uuid string `json:"uuid"`
}

func (h *NodeHandler) CreateNode(c fuego.ContextWithBody[CreateNodeRequest]) (*ApiResponse[CreateNodeResponse], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[CreateNodeResponse](errors.New("unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](errors.New("invalid user ID"))
	}

	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(body.B64EncryptedEncryptionKey)
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}
	nonce, err := base64.StdEncoding.DecodeString(body.B64EncryptionNonce)
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}
	encryptedName, err := base64.StdEncoding.DecodeString(body.B64EncryptedName)
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}

	var parentID *uuid.UUID
	if body.ParentUuid != nil && *body.ParentUuid != "" && *body.ParentUuid != "root" {
		parsed, err := uuid.Parse(*body.ParentUuid)
		if err != nil {
			return NewErrorResponse[CreateNodeResponse](err)
		}
		parentID = &parsed
	}

	id, err := h.nodeService.CreateNode(c.Context(), userID, services.CreateNodeInput{
		EncryptedKey:     encryptedKey,
		Nonce:            nonce,
		EncryptedName:    encryptedName,
		B64EncryptedPath: body.B64EncryptedPath,
		IsDirectory:      body.IsDirectory,
		ParentId:         parentID,
		Version:          body.Version,
	})
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}
	return NewApiResponse(&CreateNodeResponse{Uuid: id.String()}, "file node created")
}

// --- SaveNode ---

type SaveNodeResponse struct{}

func (h *NodeHandler) SaveNode(c fuego.ContextNoBody) (*ApiResponse[SaveNodeResponse], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[SaveNodeResponse](errors.New("unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[SaveNodeResponse](errors.New("invalid user ID"))
	}
	nodeID, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[SaveNodeResponse](err)
	}

	file, header, err := c.Request().FormFile("encryptedFile")
	if err != nil {
		return NewErrorResponse[SaveNodeResponse](err)
	}

	if err := h.nodeService.SaveNode(c.Context(), userID, nodeID, file, header.Size); err != nil {
		return NewErrorResponse[SaveNodeResponse](err)
	}
	return NewApiResponse(&SaveNodeResponse{}, "file saved")
}

// --- IndexDirectory ---

type IndexDirectoryResponse struct {
	Nodes []*models.ServerNode `json:"nodes"`
}

func (h *NodeHandler) IndexDirectory(c fuego.ContextNoBody) (*ApiResponse[IndexDirectoryResponse], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[IndexDirectoryResponse](errors.New("unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[IndexDirectoryResponse](errors.New("invalid user ID"))
	}

	nodes, err := h.nodeService.IndexDirectory(c.Context(), userID, c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[IndexDirectoryResponse](err)
	}
	return NewApiResponse(&IndexDirectoryResponse{Nodes: nodes}, "files retrieved")
}

// --- DeleteNode ---

type DeleteNodeResponse struct{}

func (h *NodeHandler) DeleteNode(c fuego.ContextNoBody) (*ApiResponse[DeleteNodeResponse], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[DeleteNodeResponse](errors.New("unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[DeleteNodeResponse](errors.New("invalid user ID"))
	}
	nodeID, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[DeleteNodeResponse](err)
	}

	if err := h.nodeService.DeleteNode(c.Context(), userID, nodeID); err != nil {
		return NewErrorResponse[DeleteNodeResponse](err)
	}
	return NewApiResponse(&DeleteNodeResponse{}, "node deleted")
}

// --- DownloadNode ---

func (h *NodeHandler) DownloadNode(c fuego.ContextNoBody) (any, error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return nil, errors.New("unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	nodeID, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return nil, err
	}

	node, data, err := h.nodeService.DownloadNode(c.Context(), userID, nodeID)
	if err != nil {
		return nil, err
	}

	w := c.Response()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("X-Encrypted-Key", base64.StdEncoding.EncodeToString(node.EncryptedKey))
	w.Header().Set("X-Encryption-Nonce", base64.StdEncoding.EncodeToString(node.Nonce))
	w.Header().Set("X-Encrypted-Name", base64.StdEncoding.EncodeToString(node.EncryptedName))

	_, err = w.Write(data)
	return nil, err
}
