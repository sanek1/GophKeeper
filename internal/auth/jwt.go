package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// claims for JWT tokens
type JWTClaims struct {
	UserID string `json:"user_id"`
	Exp    int64  `json:"exp"`
}

func CreateToken(userID string, expirationHours int, jwtSecret string) (string, error) {
	if expirationHours <= 0 {
		expirationHours = 24
	}

	if jwtSecret == "" {
		return "", fmt.Errorf("JWT secret is required")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("error creating token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken checks the validity of JWT token
func VerifyToken(tokenString string, jwtSecret string) (string, error) {
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT secret is required")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("error verifying token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID in token")
	}

	return userID, nil
}

// ExtractUserIDFromToken extracts user ID from token without verification (useful for token regeneration)
func ExtractUserIDFromToken(tokenString string) (string, int64, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", 0, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("error decoding token: %w", err)
	}

	var claims struct {
		UserID string `json:"user_id"`
		Exp    int64  `json:"exp"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", 0, fmt.Errorf("error parsing token data: %w", err)
	}

	return claims.UserID, claims.Exp, nil
}

// RegenerateToken creates a new token based on the data from the existing token with updated expiration time
func RegenerateToken(tokenString string, jwtSecret string) (string, error) {
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT secret is required")
	}

	userID, _, err := ExtractUserIDFromToken(tokenString)
	if err != nil {
		return "", err
	}

	// create a new token with updated expiration time (25 hours instead of 24)
	// this ensures that the new token will be different from the old one
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(25 * time.Hour).Unix(),
	})

	newTokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("error creating new token: %w", err)
	}

	return newTokenString, nil
}
