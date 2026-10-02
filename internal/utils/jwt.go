package utils

import (
	"errors"
	"fmt"
	"time"

	"taskhub/internal/configs"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents custom claims stored in the JWT token
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates an individual token with specified duration and secret
func GenerateToken(userID uint, role string, secret string, duration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "taskhub-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateTokenPair creates both access token and refresh token
func GenerateTokenPair(userID uint, role string, cfg *configs.Config) (accessToken string, refreshToken string, expiresIn string, err error) {
	accessDuration, err := time.ParseDuration(cfg.JWTExpiresIn)
	if err != nil {
		accessDuration = 24 * time.Hour
	}

	refreshDuration, err := time.ParseDuration(cfg.JWTRefreshExpiresIn)
	if err != nil {
		refreshDuration = 7 * 24 * time.Hour
	}

	accessToken, err = GenerateToken(userID, role, cfg.JWTSecret, accessDuration)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = GenerateToken(userID, role, cfg.JWTRefreshSecret, refreshDuration)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, cfg.JWTExpiresIn, nil
}

// ValidateToken parses and validates a JWT token string
func ValidateToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
