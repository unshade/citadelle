package controllers

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/middleware"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/repositories"
)

const dataRoot = "./data"

type NodeController struct {
	Database repositories.Database
}

func NewNodeController(database repositories.Database) *NodeController {
	return &NodeController{Database: database}
}

func (fc *NodeController) Register(group *fuego.Server, authMiddleware *middleware.JWTAuthMiddleware) {
	// Create auth option to protect routes
	authOption := option.Middleware(authMiddleware.Authenticate)

	filesGroup := fuego.Group(group, "/nodes", option.Tags("nodes"))

	fuego.Post(filesGroup, "/", fc.CreateNode, authOption)
	fuego.Post(filesGroup, "/{uuid}", fc.SaveNode, authOption)
	fuego.Get(filesGroup, "/{uuid}/download", fc.DownloadNode, authOption)
	fuego.Get(filesGroup, "/{uuid}", fc.IndexDirectory, authOption)
	fuego.Delete(filesGroup, "/{uuid}", fc.DeleteNode, authOption)
}

type CreateFileRequest struct {
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

func (fc *NodeController) CreateNode(c fuego.ContextWithBody[CreateFileRequest]) (*ApiResponse[CreateNodeResponse], error) {
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}

	// Get authenticated user ID from context
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[CreateNodeResponse](errors.New("unauthorized"))
	}

	proprietaryId, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](errors.New("invalid user ID"))
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

	// Handle parent UUID - if "root" or empty, it's a root level item (nil parent)
	var parentIdPtr *uuid.UUID
	if body.ParentUuid != nil && *body.ParentUuid != "" && *body.ParentUuid != "root" {
		parsedParentUuid, err := uuid.Parse(*body.ParentUuid)
		if err != nil {
			return NewErrorResponse[CreateNodeResponse](err)
		}
		parentIdPtr = &parsedParentUuid
	}

	serverNode := models.ServerNode{
		Id:               uuid.New(),
		EncryptedKey:     encryptedKey,
		Nonce:            nonce,
		Version:          body.Version,
		EncryptedName:    encryptedName,
		B64EncryptedPath: body.B64EncryptedPath,
		IsDirectory:      body.IsDirectory,
		ParentId:         parentIdPtr,
		ProprietaryId:    proprietaryId,
	}

	if err := fc.Database.ServerNodes.Create(c.Context(), serverNode); err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}

	return NewApiResponse(&CreateNodeResponse{
		Uuid: serverNode.Id.String(),
	}, "file node created")
}

type SaveFileResponse struct{}

const maxUploadSize = 100 << 20 // 100 MB

func (fc *NodeController) SaveNode(c fuego.ContextNoBody) (*ApiResponse[SaveFileResponse], error) {
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[SaveFileResponse](errors.New("unauthorized"))
	}

	proprietaryId, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[SaveFileResponse](errors.New("invalid user ID"))
	}

	nodeUuid, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	_, err = fc.Database.ServerNodes.GetByIdAndUserId(c.Context(), nodeUuid, proprietaryId)
	if err != nil {
		return NewErrorResponse[SaveFileResponse](errors.New("node not found or not accessible"))
	}

	storagePath := filepath.Join(dataRoot, nodeUuid.String())
	filePath := filepath.Join(storagePath, "content")

	if err := os.MkdirAll(storagePath, 0700); err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	encryptedFile, header, err := c.Request().FormFile("encryptedFile")
	if err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	if header.Size > maxUploadSize {
		return NewErrorResponse[SaveFileResponse](errors.New("file too large"))
	}

	rawEncryptedFile := make([]byte, header.Size)
	_, err = encryptedFile.Read(rawEncryptedFile)
	if err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	err = os.WriteFile(filePath, rawEncryptedFile, 0600)
	if err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	return NewApiResponse(&SaveFileResponse{}, "file saved")
}

type IndexFilesResponse struct {
	Nodes []*models.ServerNode `json:"nodes"`
}

func (fc *NodeController) IndexDirectory(c fuego.ContextNoBody) (*ApiResponse[IndexFilesResponse], error) {
	// Get authenticated user ID from context
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[IndexFilesResponse](errors.New("unauthorized"))
	}

	proprietaryId, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[IndexFilesResponse](errors.New("invalid user ID"))
	}

	// Handle "root" or empty UUID for root directory
	var nodes []*models.ServerNode
	uuidParam := c.PathParam("uuid")

	if uuidParam == "root" || uuidParam == "" || uuidParam == "00000000-0000-0000-0000-000000000000" {
		nodes, err = fc.Database.ServerNodes.GetRootNodesByUserId(c.Context(), proprietaryId)
	} else {
		parentUUID, parseErr := uuid.Parse(uuidParam)
		if parseErr != nil {
			return NewErrorResponse[IndexFilesResponse](parseErr)
		}
		nodes, err = fc.Database.ServerNodes.GetChildrensByUserId(c.Context(), parentUUID, proprietaryId)
	}

	if err != nil {
		return NewErrorResponse[IndexFilesResponse](err)
	}

	return NewApiResponse(&IndexFilesResponse{Nodes: nodes}, "files retrieved")
}

type DeleteNodeResponse struct{}

func (fc *NodeController) DeleteNode(c fuego.ContextNoBody) (*ApiResponse[DeleteNodeResponse], error) {
	// Get authenticated user ID from context
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return NewErrorResponse[DeleteNodeResponse](errors.New("unauthorized"))
	}

	proprietaryId, err := uuid.Parse(userIDStr)
	if err != nil {
		return NewErrorResponse[DeleteNodeResponse](errors.New("invalid user ID"))
	}

	nodeUuid, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[DeleteNodeResponse](err)
	}

	uuids, err := fc.Database.ServerNodes.DeleteRecursiveByUserId(c.Context(), nodeUuid, proprietaryId)
	if err != nil {
		return NewErrorResponse[DeleteNodeResponse](err)
	}

	for _, id := range uuids {
		storagePath := filepath.Join(dataRoot, id.String())
		if err := os.RemoveAll(storagePath); err != nil {
			return NewErrorResponse[DeleteNodeResponse](err)
		}
	}

	return NewApiResponse(&DeleteNodeResponse{}, "node deleted")
}

func (fc *NodeController) DownloadNode(c fuego.ContextNoBody) (any, error) {
	// Get authenticated user ID from context
	userIDStr := middleware.GetUserID(c.Context())
	if userIDStr == "" {
		return nil, errors.New("unauthorized")
	}

	proprietaryId, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	nodeUuid, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return nil, err
	}

	node, err := fc.Database.ServerNodes.GetByIdAndUserId(c.Context(), nodeUuid, proprietaryId)
	if err != nil {
		return nil, err
	}

	if node.IsDirectory {
		return nil, errors.New("cannot download a directory")
	}

	storagePath := filepath.Join(dataRoot, nodeUuid.String())
	filePath := filepath.Join(storagePath, "content")

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	w := c.Response()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("X-Encrypted-Key", base64.StdEncoding.EncodeToString(node.EncryptedKey))
	w.Header().Set("X-Encryption-Nonce", base64.StdEncoding.EncodeToString(node.Nonce))
	w.Header().Set("X-Encrypted-Name", base64.StdEncoding.EncodeToString(node.EncryptedName))

	_, err = w.Write(fileData)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
