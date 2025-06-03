package auth

import (
	"testing"
	"time"

	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const testJWTSecret = "test-jwt-secret-for-handlers"

func TestCreateToken(t *testing.T) {
	userID := "test-user-id"
	expirationHours := 24

	token, err := CreateToken(userID, expirationHours, testJWTSecret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(testJWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["user_id"])

	expTime := time.Unix(int64(claims["exp"].(float64)), 0)
	expectedExpTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	timeDiff := expectedExpTime.Sub(expTime)
	assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute)

	token, err = CreateToken(userID, -1, testJWTSecret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestVerifyToken(t *testing.T) {
	userID := "test-user-id"
	token, err := CreateToken(userID, 24, testJWTSecret)
	assert.NoError(t, err)

	extractedUserID, err := VerifyToken(token, testJWTSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)

	_, err = VerifyToken("invalid.token.string", testJWTSecret)
	assert.Error(t, err)

	_, err = VerifyToken("", testJWTSecret)
	assert.Error(t, err)

	expiredToken := createExpiredToken(userID)
	_, err = VerifyToken(expiredToken, testJWTSecret)
	assert.Error(t, err)
}

func TestExtractUserIDFromToken(t *testing.T) {
	userID := "test-user-id"
	token, err := CreateToken(userID, 24, testJWTSecret)
	assert.NoError(t, err)

	extractedUserID, exp, err := ExtractUserIDFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
	assert.NotZero(t, exp)

	_, _, err = ExtractUserIDFromToken("invalid.token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	_, _, err = ExtractUserIDFromToken("")
	assert.Error(t, err)
}

func TestRegenerateToken(t *testing.T) {
	userID := "test-user-id"
	originalToken, err := CreateToken(userID, 24, testJWTSecret)
	assert.NoError(t, err)

	newToken, err := RegenerateToken(originalToken, testJWTSecret)
	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)

	assert.NotEqual(t, originalToken, newToken)

	extractedUserID, err := VerifyToken(newToken, testJWTSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)

	_, err = RegenerateToken("invalid.token.string", testJWTSecret)
	assert.Error(t, err)
}

func createExpiredToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(-time.Hour).Unix(), // Истекший на 1 час
	})

	tokenString, _ := token.SignedString([]byte(testJWTSecret))
	return tokenString
}

func TestCreateToken_EdgeCases(t *testing.T) {
	t.Run("EmptyUserID", func(t *testing.T) {
		userID := ""
		secret := "test-secret-key"

		token, err := CreateToken(userID, 24, secret)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("VeryLongSecret", func(t *testing.T) {
		userID := uuid.New().String()
		secret := strings.Repeat("a", 1000) // Very long secret

		token, err := CreateToken(userID, 24, secret)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestVerifyToken_EdgeCases(t *testing.T) {
	secret := "test-secret-key"

	t.Run("ExpiredToken", func(t *testing.T) {
		// Create an expired token manually
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": uuid.New().String(),
			"exp":     time.Now().Add(-1 * time.Hour).Unix(), // Expired
		})
		tokenString, _ := token.SignedString([]byte(secret))

		userID, err := VerifyToken(tokenString, secret)
		assert.Error(t, err)
		assert.Empty(t, userID)
	})

	t.Run("TokenWithoutUserID", func(t *testing.T) {
		// Create a token without user_id claim
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(secret))

		userID, err := VerifyToken(tokenString, secret)
		assert.Error(t, err)
		assert.Empty(t, userID)
	})

	t.Run("TokenWithInvalidUserID", func(t *testing.T) {
		// Create a token with invalid user_id
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "invalid-uuid",
			"exp":     time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(secret))

		userID, err := VerifyToken(tokenString, secret)
		assert.NoError(t, err) // VerifyToken doesn't validate UUID format
		assert.Equal(t, "invalid-uuid", userID)
	})
}

func TestExtractUserIDFromToken_EdgeCases(t *testing.T) {
	secret := "test-secret-key"

	t.Run("TokenWithNonStringUserID", func(t *testing.T) {
		// Create a token with numeric user_id
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": 12345, // Numeric instead of string
			"exp":     time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(secret))

		userID, exp, err := ExtractUserIDFromToken(tokenString)
		assert.Error(t, err) // Should fail because userID is not a string
		assert.Empty(t, userID)
		assert.Equal(t, int64(0), exp)
	})
}

func TestRegenerateToken_EdgeCases(t *testing.T) {
	secret := "test-secret-key"

	t.Run("RegenerateExpiredToken", func(t *testing.T) {
		// Create an expired token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": uuid.New().String(),
			"exp":     time.Now().Add(-1 * time.Hour).Unix(), // Expired
		})
		oldTokenString, _ := token.SignedString([]byte(secret))

		newToken, err := RegenerateToken(oldTokenString, secret)
		assert.NoError(t, err) // Should work even if token is expired
		assert.NotEmpty(t, newToken)
	})

	t.Run("RegenerateWithDifferentSecret", func(t *testing.T) {
		userID := uuid.New().String()

		// Create token with one secret
		oldToken, _ := CreateToken(userID, 24, "old-secret")

		// Try to regenerate with different secret
		newToken, err := RegenerateToken(oldToken, "new-secret")
		assert.NoError(t, err) // Should work because ExtractUserIDFromToken doesn't verify signature
		assert.NotEmpty(t, newToken)
	})
}
