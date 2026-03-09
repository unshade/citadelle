package models

import "gorm.io/gorm"

type FileReference struct {
	gorm.Model
	DirectoryPath string `gorm:"not null"`
	Name          string `gorm:"not null"`
	EncryptedKey  []byte `gorm:"not null"`
	Nonce         []byte `gorm:"not null"`
	UserID        uint   `gorm:"not null"`
	User          User   `gorm:"constraint:OnDelete:CASCADE;"`
}
