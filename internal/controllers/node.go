package controllers

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
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

func (fc *NodeController) Register(group *fuego.Server) {
	filesGroup := fuego.Group(group, "/nodes", option.Tags("nodes"))

	fuego.Post(filesGroup, "/", fc.CreateNode)
	fuego.Post(filesGroup, "/{uuid}", fc.SaveNode)
	fuego.Get(filesGroup, "/{uuid}", fc.IndexDirectory)
	fuego.Delete(filesGroup, "/{uuid}", fc.DeleteNode)
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

	parsedParentUuid, err := uuid.Parse(*body.ParentUuid)
	if err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}

	serverNode := models.ServerNode{
		Id:               uuid.New(),
		EncryptedKey:     encryptedKey,
		Nonce:            nonce,
		Version:          body.Version,
		EncryptedName:    encryptedName,
		B64EncryptedPath: body.B64EncryptedPath,
		IsDirectory:      body.IsDirectory,
		ParentId:         parsedParentUuid,
	}

	if err := fc.Database.ServerNodes.Create(c.Context(), serverNode); err != nil {
		return NewErrorResponse[CreateNodeResponse](err)
	}

	return NewApiResponse(&CreateNodeResponse{
		Uuid: serverNode.Id.String(),
	}, "file node created")
}

type SaveFileResponse struct{}

func (fc *NodeController) SaveNode(c fuego.ContextNoBody) (*ApiResponse[SaveFileResponse], error) {
	uuid := c.PathParam("uuid")

	storagePath := filepath.Join(dataRoot, uuid)
	filePath := filepath.Join(storagePath, "content")

	if err := os.MkdirAll(storagePath, 0777); err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	encryptedFile, header, err := c.Request().FormFile("encryptedFile")
	if err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	rawEncryptedFile := make([]byte, header.Size)
	_, err = encryptedFile.Read(rawEncryptedFile)
	if err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	err = os.WriteFile(filePath, rawEncryptedFile, 0777)
	if err != nil {
		return NewErrorResponse[SaveFileResponse](err)
	}

	return NewApiResponse(&SaveFileResponse{}, "file saved")
}

type IndexFilesResponse struct {
	Nodes []*models.ServerNode `json:"nodes"`
}

func (fc *NodeController) IndexDirectory(c fuego.ContextNoBody) (*ApiResponse[IndexFilesResponse], error) {
	uuid, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[IndexFilesResponse](err)
	}

	nodes, err := fc.Database.ServerNodes.GetChildrens(c.Context(), uuid)
	if err != nil {
		return NewErrorResponse[IndexFilesResponse](err)
	}

	return NewApiResponse(&IndexFilesResponse{Nodes: nodes}, "files retrieved")
}

type DeleteNodeResponse struct{}

func (fc *NodeController) DeleteNode(c fuego.ContextNoBody) (*ApiResponse[DeleteNodeResponse], error) {
	nodeUuid, err := uuid.Parse(c.PathParam("uuid"))
	if err != nil {
		return NewErrorResponse[DeleteNodeResponse](err)
	}

	uuids, err := fc.Database.ServerNodes.DeleteRecursive(c.Context(), nodeUuid)
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
