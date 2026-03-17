package models

import (
	"github.com/google/uuid"
)

type ServerNode struct {
	Id               uuid.UUID
	Version          uint64
	B64EncryptedPath string `gorm:"index"`

	EncryptedName []byte
	EncryptedKey  []byte
	Nonce         []byte

	IsDirectory bool
	ParentId    uuid.UUID
	Parent      *ServerNode
}
