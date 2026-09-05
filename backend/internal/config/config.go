package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAppEnv             = "development"
	defaultServerPort         = "8080"
	defaultFrontendOrigin     = "http://localhost:5173"
	defaultJWTExpiresIn       = "168h"
	defaultCookieSecure       = "false"
	defaultAIProvider         = "local"
	defaultSiliconFlowBaseURL = "https://api.siliconflow.cn/v1"
	defaultSiliconFlowModel   = "Pro/zai-org/GLM-5.1"
	defaultAIRequestTimeout   = "60s"
)

type Config struct {
	AppEnv             string
	ServerPort         string
	FrontendOrigin     string
	DatabaseDSN        string
	JWTSecret          string
	JWTExpiresIn       time.Duration
	CookieSecure       bool
	AIProvider         string
	SiliconFlowAPIKey  string
	SiliconFlowBaseURL string
	SiliconFlowModel   string
	SiliconFlowModels  []string
	AIRequestTimeout   time.Duration
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

	aiProvider := strings.ToLower(
		strings.TrimSpace(
			getEnv("AI_PROVIDER", defaultAIProvider),
		),
	)

	switch aiProvider {
	case "local", "siliconflow":
		// 当前只允许已实现的 Provider，避免配置拼错后静默回退。
	default:
		return Config{}, fmt.Errorf(
			"unsupported AI_PROVIDER: %s",
			aiProvider,
		)
	}

	aiRequestTimeout, err := time.ParseDuration(
		getEnv("AI_REQUEST_TIMEOUT", defaultAIRequestTimeout),
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"parse AI_REQUEST_TIMEOUT: %w",
			err,
		)
	}

	if aiRequestTimeout <= 0 {
		return Config{}, errors.New(
			"AI_REQUEST_TIMEOUT must be greater than zero",
		)
	}

	siliconFlowAPIKey := strings.TrimSpace(
		os.Getenv("SILICONFLOW_API_KEY"),
	)

	if aiProvider == "siliconflow" && siliconFlowAPIKey == "" {
		return Config{}, errors.New(
			"SILICONFLOW_API_KEY is required when AI_PROVIDER is siliconflow",
		)
	}

	siliconFlowModel := strings.TrimSpace(
		getEnv("SILICONFLOW_MODEL", defaultSiliconFlowModel),
	)
	if siliconFlowModel == "" {
		siliconFlowModel = defaultSiliconFlowModel
	}

	// 默认模型始终排在第一位，前端可直接将它作为初始选项。
	siliconFlowModels := []string{siliconFlowModel}
	for _, modelID := range parseModelIDs(os.Getenv("SILICONFLOW_MODELS")) {
		if modelID != siliconFlowModel {
			siliconFlowModels = append(siliconFlowModels, modelID)
		}
	}

	return Config{
		AppEnv:            getEnv("APP_ENV", defaultAppEnv),
		ServerPort:        getEnv("SERVER_PORT", defaultServerPort),
		FrontendOrigin:    getEnv("FRONTEND_ORIGIN", defaultFrontendOrigin),
		DatabaseDSN:       databaseDSN,
		JWTSecret:         jwtSecret,
		JWTExpiresIn:      jwtExpiresIn,
		CookieSecure:      cookieSecure,
		AIProvider:        aiProvider,
		SiliconFlowAPIKey: siliconFlowAPIKey,
		SiliconFlowBaseURL: getEnv(
			"SILICONFLOW_BASE_URL",
			defaultSiliconFlowBaseURL,
		),
		SiliconFlowModel:  siliconFlowModel,
		SiliconFlowModels: siliconFlowModels,
		AIRequestTimeout:  aiRequestTimeout,
	}, nil
}

func parseModelIDs(value string) []string {
	modelIDs := make([]string, 0)
	seen := make(map[string]struct{})

	for _, item := range strings.Split(value, ",") {
		modelID := strings.TrimSpace(item)
		if modelID == "" {
			continue
		}

		if _, exists := seen[modelID]; exists {
			continue
		}

		seen[modelID] = struct{}{}
		modelIDs = append(modelIDs, modelID)
	}

	return modelIDs
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
