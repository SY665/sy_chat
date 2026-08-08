package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenIssuer       = "sy-chat"
	SessionCookieName = "sy_chat_session"
)

var ErrInvalidToken = errors.New("invalid authentication token")

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret    []byte
	expiresIn time.Duration
}

func NewTokenManager(secret string, expiresIn time.Duration) *TokenManager {
	return &TokenManager{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

// Generate 创建包含用户 ID 和过期时间的 JWT。
func (manager *TokenManager) Generate(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", ErrUserIDRequired
	}

	now := time.Now()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(manager.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(manager.secret)
	if err != nil {
		return "", fmt.Errorf("sign authentication token: %w", err)
	}

	return signedToken, nil
}

// Parse 验证 JWT 的签名、签发者和过期时间，并返回用户 ID。
func (manager *TokenManager) Parse(tokenString string) (string, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrInvalidToken
			}
			return manager.secret, nil
		},

		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.UserID == "" {
		return "", ErrInvalidToken
	}
	return claims.UserID, nil

}
