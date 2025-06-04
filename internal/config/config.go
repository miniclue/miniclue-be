package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	DBHost      string `envconfig:"DB_HOST" required:"true"`
	DBPort      int    `envconfig:"DB_PORT" default:"5432"`
	DBUser      string `envconfig:"DB_USER" required:"true"`
	DBPassword  string `envconfig:"DB_PASSWORD" required:"true"`
	DBName      string `envconfig:"DB_NAME" required:"true"`
	JWTSecret   string `envconfig:"JWT_SECRET"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
