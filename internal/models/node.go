package models

import (
	"github.com/google/uuid"
)

type ServerNode struct {
	Id      uuid.UUID
	Version uint64

	// Encrypted path (indexed for directory lookups).
	// Nonce and ciphertext stored separately — no concatenation.
	B64PathNonce     string
	B64EncryptedPath string `gorm:"index"`

	// Encrypted node name.
	NameNonce     []byte
	EncryptedName []byte

	// Per-node content key sealed with the master key.
	// Empty for directories (no file content).
	KeyNonce     []byte
	EncryptedKey []byte

	// AES-GCM nonce used to encrypt the file content blob.
	// Empty for directories.
	ContentNonce []byte

	IsDirectory bool
	ParentId    *uuid.UUID
	Parent      *ServerNode

	ProprietaryId uuid.UUID
	Proprietary   User
}
