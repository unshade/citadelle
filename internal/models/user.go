package models

import (
	"github.com/google/uuid"
)

type User struct {
	Id uuid.UUID

	// PBKDF2 salt — kept separate from the encrypted master key by design.
	Salt []byte `gorm:"not null"`

	// Master key sealed with the KEK (derived from password + salt).
	// Nonce and ciphertext are stored in separate columns so no parsing
	// of concatenated bytes is needed on the client or server.
	MasterKeyNonce     []byte `gorm:"not null"`
	EncryptedMasterKey []byte `gorm:"not null"`

	// Challenge used for zero-knowledge authentication.
	// The server issues EncryptedChallenge; the client decrypts it and
	// returns ClearChallenge to prove knowledge of the password.
	ChallengeNonce     []byte
	EncryptedChallenge []byte
	ClearChallenge     string
}
