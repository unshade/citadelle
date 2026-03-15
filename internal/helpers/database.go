package helpers

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitServerDb(dbPath string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
}

