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

func NewFileController(database repositories.Database) *FileController {
	return &FileController{Database: database}
}

func (fc *FileController) Register(group *fuego.Server) {
	filesGroup := fuego.Group(group, "/files", option.Tags("files"))

	fuego.Post(filesGroup, "/", fc.CreateFile)
	fuego.Post(filesGroup, "/{id}", fc.SaveFile)
	fuego.Get(filesGroup, "/index/{b64Sha256Path}", fc.IndexFiles)
	// fuego.Delete(filesGroup, "/{path...}", fc.DeleteFile)
	// fuego.Put(filesGroup, "/{path...}", fc.UpdateFile)
}

type CreateFileRequest struct {
	B64EncryptedEncryptionKey string `json:"b64EncryptedEncryptionKey"`
	B64EncryptionNonce        string `json:"b64EncryptionNonce"`
	B64EncryptedName          string `json:"b64EncryptedName"`
	B64Sha256Path             string `json:"b64Sha256Path"`
	Version                   uint64 `json:"version"`
}

type CreateFileResponse struct {
	Uuid string `json:"uuid"`
}

func (fc *FileController) CreateFile(c fuego.ContextWithBody[CreateFileRequest]) (*ApiResponse[CreateFileResponse], error) {
	body, err := c.Body()
	if err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(body.B64EncryptedEncryptionKey)
	if err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	nonce, err := base64.StdEncoding.DecodeString(body.B64EncryptionNonce)
	if err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	encryptedName, err := base64.StdEncoding.DecodeString(body.B64EncryptedName)
	if err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	serverNode := models.ServerNode{
		Id:            uuid.New(),
		EncryptedKey:  encryptedKey,
		Nonce:         nonce,
		Version:       body.Version,
		EncryptedName: encryptedName,
		B64Sha256Path: body.B64Sha256Path,
	}

	if err := fc.Database.ServerNodes.Create(c.Context(), serverNode); err != nil {
		return NewErrorResponse[CreateFileResponse](err)
	}

	return NewApiResponse(&CreateFileResponse{
		Uuid: serverNode.Id.String(),
	}, "file node created")
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

type FileNode struct {
	Id               string `json:"id"`
	B64EncryptedName string `json:"b64EncryptedName"`
}

type IndexFilesResponse struct {
	Files []FileNode `json:"files"`
}

func (fc *FileController) IndexFiles(c fuego.ContextWithBody[fuego.ContextNoBody]) (*ApiResponse[IndexFilesResponse], error) {
	b64Sha256Path := c.PathParam("b64Sha256Path")

	nodes, err := fc.Database.ServerNodes.FindBySha256Path(c.Context(), b64Sha256Path)
	if err != nil {
		return NewErrorResponse[IndexFilesResponse](err)
	}

	files := make([]FileNode, len(nodes))
	for i, node := range nodes {
		files[i] = FileNode{
			Id:               node.Id.String(),
			B64EncryptedName: base64.StdEncoding.EncodeToString(node.EncryptedName),
		}
	}

	return NewApiResponse(&IndexFilesResponse{Files: files}, "files retrieved")
}
