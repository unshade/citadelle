package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port      string `env:"PORT" envDefault:"8080"`
	DBPath    string `env:"DB_PATH" envDefault:"citadelle.db"`
	JWTSecret string `env:"JWT_SECRET,required"`
}

func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	return &cfg, nil
}
