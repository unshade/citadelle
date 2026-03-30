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

	// AES-GCM nonce used to encrypt the file content blob with the master key.
	// Empty for directories.
	ContentNonce []byte

	IsDirectory bool
	IsFavourite bool
	ParentId    *uuid.UUID
	Parent      *ServerNode

	ProprietaryId uuid.UUID
	Proprietary   User
}
