package handlers

import (
	"encoding/base64"
	"errors"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/middleware"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/pagination"
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
	fuego.Get(nodesGroup, "/favourites", h.GetFavourites, authOption) // exact route before /{uuid}
	fuego.Get(nodesGroup, "/{uuid}/download", h.DownloadNode, authOption)
	fuego.Get(nodesGroup, "/{uuid}", h.IndexDirectory, authOption)
	fuego.Delete(nodesGroup, "/{uuid}", h.DeleteNode, authOption)
	fuego.Put(nodesGroup, "/{uuid}/favourite", h.SetFavourite, authOption)
}

// --- CreateNode ---

// CreateNodeRequest carries all encrypted node metadata.
// Each sealed value is split into nonce + ciphertext — never concatenated.
// ContentNonce is an empty string for directory nodes.
type CreateNodeRequest struct {
	B64ContentNonce  string  `json:"b64ContentNonce"`
	B64NameNonce     string  `json:"b64NameNonce"`
	B64EncryptedName string  `json:"b64EncryptedName"`
	B64PathNonce     string  `json:"b64PathNonce"`
	B64EncryptedPath string  `json:"b64EncryptedPath"`
	IsDirectory      bool    `json:"isDirectory"`
	ParentUuid       *string `json:"parentUuid"`
	Version          uint64  `json:"version"`
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

	contentNonce, err := base64.StdEncoding.DecodeString(body.B64ContentNonce)
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}
	nameNonce, err := base64.StdEncoding.DecodeString(body.B64NameNonce)
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
		ContentNonce:     contentNonce,
		NameNonce:        nameNonce,
		EncryptedName:    encryptedName,
		B64PathNonce:     body.B64PathNonce,
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

func (h *NodeHandler) IndexDirectory(c fuego.ContextNoBody) (*PaginatedApiResponse[*models.ServerNode], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return nil, errors.New("unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	p := pagination.Params{
		Page:    uint64(c.QueryParamInt("page")),
		PerPage: uint64(c.QueryParamInt("perPage")),
	}.Normalize()

	nodes, result, err := h.nodeService.IndexDirectory(c.Context(), userID, c.PathParam("uuid"), p)
	if err != nil {
		return nil, err
	}
	return NewPaginatedApiResponse(nodes, result, "files retrieved"), nil
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

// --- SetFavourite ---

type SetFavouriteRequest struct {
	IsFavourite bool `json:"isFavourite"`
}

type SetFavouriteResponse struct{}

func (h *NodeHandler) SetFavourite(c fuego.ContextWithBody[SetFavouriteRequest]) (*ApiResponse[SetFavouriteResponse], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[SetFavouriteResponse](errors.New("unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[SetFavouriteResponse](errors.New("invalid user ID"))
	}
	nodeID, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[SetFavouriteResponse](err)
	}
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[SetFavouriteResponse](err)
	}
	if err := h.nodeService.SetFavourite(c.Context(), userID, nodeID, body.IsFavourite); err != nil {
		return NewErrorResponse[SetFavouriteResponse](err)
	}
	return NewApiResponse(&SetFavouriteResponse{}, "favourite updated")
}

// --- GetFavourites ---

func (h *NodeHandler) GetFavourites(c fuego.ContextNoBody) (*PaginatedApiResponse[*models.ServerNode], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return nil, errors.New("unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	p := pagination.Params{
		Page:    uint64(c.QueryParamInt("page")),
		PerPage: uint64(c.QueryParamInt("perPage")),
	}.Normalize()

	nodes, result, err := h.nodeService.GetFavourites(c.Context(), userID, p)
	if err != nil {
		return nil, err
	}
	return NewPaginatedApiResponse(nodes, result, "favourites retrieved"), nil
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
	w.Header().Set("X-Content-Nonce", base64.StdEncoding.EncodeToString(node.ContentNonce))
	w.Header().Set("X-Name-Nonce", base64.StdEncoding.EncodeToString(node.NameNonce))
	w.Header().Set("X-Encrypted-Name", base64.StdEncoding.EncodeToString(node.EncryptedName))

	_, err = w.Write(data)
	return nil, err
}
