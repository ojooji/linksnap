package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServerPort: env("SERVER_PORT", "8080"),
		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "5432"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
	}

	if cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPassword == "" {
		return Config{}, errors.New("DB_NAME, DB_USER and DB_PASSWORD are required")
	}
	return cfg, nil
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(c.DBUser),
		url.QueryEscape(c.DBPassword),
		c.DBHost, c.DBPort, c.DBName,
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
