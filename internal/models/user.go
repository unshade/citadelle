package models

import (
	"github.com/google/uuid"
)

type User struct {
	Id                 uuid.UUID
	Salt               []byte `gorm:"not null"`
	EncryptedMasterKey []byte `gorm:"not null"`

	// TODO : find a way to authenticate without storing a challenge in clear
	EncryptedChallenge []byte
	ClearChallenge     string
}
