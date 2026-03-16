package models

import (
	"github.com/google/uuid"
)

type ServerNode struct {
	Id            uuid.UUID
	Version       uint64
	B64Sha256Path string `gorm:"index"`

	EncryptedName []byte
	EncryptedKey  []byte
	Nonce         []byte
}
