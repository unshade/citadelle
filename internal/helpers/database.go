package helpers

import (
	"fmt"

	"github.com/unshade/citadelle/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitServerDb(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

