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

type FileController struct {
	Database repositories.Database
}

func NewFileController() *FileController {
	return &FileController{}
}

func (fc *FileController) Register(group *fuego.Server) {
	filesGroup := fuego.Group(group, "/files", option.Tags("files"))

	fuego.Post(filesGroup, "/", fc.CreateFile)
	fuego.Post(filesGroup, "/{id}", fc.SaveFile)
	// fuego.Get(filesGroup, "/{path...}", fc.IndexFiles)
	// fuego.Delete(filesGroup, "/{path...}", fc.DeleteFile)
	// fuego.Put(filesGroup, "/{path...}", fc.UpdateFile)
}


type CreateFileRequest struct {
	Id                        uuid.UUID `json:"uuid"`
	EncryptedB64EncryptionKey string    `json:"encryptedB64EncryptionKey"`
	B64EncryptionNonce        string    `json:"B64EncryptionNonce"`
	Version                   uint64    `json:"version"`
}

type CreateFileResponse struct{}

func (fc *FileController) CreateFile(c fuego.ContextWithBody[CreateFileRequest]) (*ApiResponse[CreateFileResponse], error) {
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(body.EncryptedB64EncryptionKey)
	if err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	nonce, err := base64.StdEncoding.DecodeString(body.B64EncryptionNonce)
	if err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	serverNode := models.ServerNode{
		Id:           body.Id,
		EncryptedKey: encryptedKey,
		Nonce:        nonce,
		Version:      body.Version,
	}

	fc.Database.ServerNodes.Create(serverNode)

	return NewApiResponse(&CreateFileResponse{}, "file node created")
}

type SaveFileResponse struct{}

func (fc *FileController) SaveFile(c fuego.ContextNoBody) (*ApiResponse[SaveFileResponse], error) {
	id := c.PathParam("id")

	storagePath := filepath.Join(dataRoot, id)
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
