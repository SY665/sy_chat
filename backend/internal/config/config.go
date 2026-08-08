package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAppEnv         = "development"
	defaultServerPort     = "8080"
	defaultFrontendOrigin = "http://localhost:5173"
	defaultJWTExpiresIn   = "168h"
	defaultCookieSecure   = "false"
)

type Config struct {
	AppEnv         string
	ServerPort     string
	FrontendOrigin string
	DatabaseDSN    string
	JWTSecret      string
	JWTExpiresIn   time.Duration
	CookieSecure   bool
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load environment file: %w", err)
	}

	databaseDSN := os.Getenv("DATABASE_DSN")
	if databaseDSN == "" {
		return Config{}, errors.New("DATABASE_DSN is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return Config{}, errors.New(
			"JWT_SECRET is required and must contain at least 32 characters",
		)
	}

	jwtExpiresIn, err := time.ParseDuration(
		getEnv("JWT_EXPIRES_IN", defaultJWTExpiresIn),
	)
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_EXPIRES_IN: %w", err)
	}

	cookieSecure, err := strconv.ParseBool(
		getEnv("COOKIE_SECURE", defaultCookieSecure),
	)
	if err != nil {
		return Config{}, fmt.Errorf("parse COOKIE_SECURE: %w", err)
	}

	return Config{
		AppEnv:         getEnv("APP_ENV", defaultAppEnv),
		ServerPort:     getEnv("SERVER_PORT", defaultServerPort),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", defaultFrontendOrigin),
		DatabaseDSN:    databaseDSN,
		JWTSecret:      jwtSecret,
		JWTExpiresIn:   jwtExpiresIn,
		CookieSecure:   cookieSecure,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
