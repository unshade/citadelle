package models

import (
	"time"

	"github.com/google/uuid"
)

type ServerNode struct {
	Id                  uuid.UUID
	Version             uint64
	EncryptedClientPath string
	EncryptedName       string

	EncryptedKey []byte
	Nonce        []byte
}

type ClientNode struct {
	Id      string
	Path    string
	Name    string
	Type    string
	Size    uint64 // in O
	ModTime time.Time
	Version uint64
}
