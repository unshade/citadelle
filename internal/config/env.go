package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"5432"`
	User     string `env:"USER" envDefault:"citadelle"`
	Password string `env:"PASSWORD,required"`
	Name     string `env:"NAME" envDefault:"citadelle"`
	SSLMode  string `env:"SSL_MODE" envDefault:"disable"`
}

type Config struct {
	Port           string         `env:"PORT" envDefault:"8080"`
	Database       DatabaseConfig `envPrefix:"DB_"`
	JWTSecret      string         `env:"JWT_SECRET,required"`
	AllowedOrigins []string       `env:"ALLOWED_ORIGINS,required" envSeparator:","`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	return &cfg, nil
}
