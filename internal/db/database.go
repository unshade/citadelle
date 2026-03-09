package db

import (
	"github.com/unshade/citadelle/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(&models.User{}, &models.FileReference{})
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func CreateUser(user *models.User) error {
	return DB.Create(user).Error
}

func CreateFileReference(fileRef *models.FileReference) error {
	return DB.Create(fileRef).Error
}

func GetFileReferencesByUser(userID uint) ([]models.FileReference, error) {
	var fileRefs []models.FileReference
	result := DB.Where("user_id = ?", userID).Find(&fileRefs)
	return fileRefs, result.Error
}

func GetFileReferenceByID(id uint) (*models.FileReference, error) {
	var fileRef models.FileReference
	result := DB.First(&fileRef, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &fileRef, nil
}

func DeleteFileReference(id uint) error {
	return DB.Delete(&models.FileReference{}, id).Error
}
