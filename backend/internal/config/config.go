package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultAppEnv         = "development"
	defaultServerPort     = "8080"
	defaultFrontendOrigin = "http://localhost:5173"
)

type Config struct {
	AppEnv         string
	ServerPort     string
	FrontendOrigin string
}

func Load() (Config, error) {
	err := godotenv.Load()

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load environment file: %w", err)
	}

	return Config{
		AppEnv:         getEnv("APP_ENV", defaultAppEnv),
		ServerPort:     getEnv("SERVER_PORT", defaultServerPort),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", defaultFrontendOrigin),
	}, nil

}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
