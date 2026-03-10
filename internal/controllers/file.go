package controllers

import (
	"encoding/base64"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/unshade/citadelle/internal/models"
)

const dataRoot = "./data"

type FileController struct{}

func NewFileController() *FileController {
	return &FileController{}
}

func (fc *FileController) Register(group *fuego.Server) {
	filesGroup := fuego.Group(group, "/files", option.Tags("files"))

	fuego.Post(filesGroup, "/", fc.CreateFile)
	// fuego.Get(filesGroup, "/{path...}", fc.IndexFiles)
	// fuego.Delete(filesGroup, "/{path...}", fc.DeleteFile)
	// fuego.Put(filesGroup, "/{path...}", fc.UpdateFile)
}

type PathRequest struct {
	Path string `json:"path" validate:"required"`
}

type FileInfoResponse struct {
	Path    string      `json:"path"`
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    fs.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
}

func fileInfoToResponse(path string, info fs.FileInfo) FileInfoResponse {
	return FileInfoResponse{
		Path:    path,
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}
}

type CreateFileRequest struct {
	Id                        uuid.UUID `json:"uuid"`
	EncryptedFile             []byte    `json:"encryptedFile"`
	EncryptedB64EncryptionKey string    `json:"encryptedB64EncryptionKey"`
	B64EncryptionNonce        string    `json:"B64EncryptionNonce"`
	Version                   uint64    `json:"version"`
}

func (fc *FileController) CreateFile(c fuego.ContextWithBody[CreateFileRequest]) (FileInfoResponse, error) {
	body, err := c.Body()
	if err != nil {
		return FileInfoResponse{}, err
	}

	storagePath := filepath.Join(dataRoot, string(body.Id[:]))
	if err := os.Mkdir(storagePath, 0600); err != nil {
		return FileInfoResponse{}, err
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(body.EncryptedB64EncryptionKey)
	if err != nil {
		return FileInfoResponse{}, err
	}

	nonce, err := base64.StdEncoding.DecodeString(body.B64EncryptionNonce)
	if err != nil {
		return FileInfoResponse{}, err
	}

	serverNode := models.ServerNode{
		Id:           body.Id,
		EncryptedKey: encryptedKey,
		Nonce:        nonce,
		Version:      body.Version,
	}

	filePath := filepath.Join(storagePath, "content")

	os.WriteFile(filePath, body.EncryptedFile, 0600)

	return FileInfoResponse{}, nil
}
