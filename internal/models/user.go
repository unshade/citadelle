package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email              string `gorm:"uniqueIndex;not null"`
	Salt               []byte `gorm:"not null"`
	EncryptedMasterKey []byte `gorm:"not null"`
}
