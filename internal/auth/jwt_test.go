package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	// Тест создания токена с валидными данными
	userID := "test-user-id"
	expirationHours := 24

	token, err := CreateToken(userID, expirationHours)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Проверяем, что в токене содержатся правильные данные
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["user_id"])

	// Проверяем, что время истечения токена установлено правильно с погрешностью в 1 минуту
	expTime := time.Unix(int64(claims["exp"].(float64)), 0)
	expectedExpTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	timeDiff := expectedExpTime.Sub(expTime)
	assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute)

	// Тест с отрицательным временем истечения (должен использоваться дефолт)
	token, err = CreateToken(userID, -1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestVerifyToken(t *testing.T) {
	// Создаем токен для тестирования
	userID := "test-user-id"
	token, err := CreateToken(userID, 24)
	assert.NoError(t, err)

	// Верификация валидного токена
	extractedUserID, err := VerifyToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)

	// Тест с неверным токеном
	_, err = VerifyToken("invalid.token.string")
	assert.Error(t, err)

	// Тест с пустым токеном
	_, err = VerifyToken("")
	assert.Error(t, err)

	// Тест с просроченным токеном
	expiredToken := createExpiredToken(userID)
	_, err = VerifyToken(expiredToken)
	assert.Error(t, err)
}

func TestExtractUserIDFromToken(t *testing.T) {
	// Создаем токен для тестирования
	userID := "test-user-id"
	token, err := CreateToken(userID, 24)
	assert.NoError(t, err)

	// Извлечение данных из валидного токена
	extractedUserID, exp, err := ExtractUserIDFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
	assert.NotZero(t, exp)

	// Тест с неверным форматом токена
	_, _, err = ExtractUserIDFromToken("invalid.token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неверный формат токена")

	// Тест с пустым токеном
	_, _, err = ExtractUserIDFromToken("")
	assert.Error(t, err)
}

func TestRegenerateToken(t *testing.T) {
	// Создаем токен для тестирования
	userID := "test-user-id"
	originalToken, err := CreateToken(userID, 24)
	assert.NoError(t, err)

	// Регенерируем токен
	newToken, err := RegenerateToken(originalToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)

	// Токены должны отличаться, но не проверяем конкретное значение
	assert.NotEqual(t, originalToken, newToken)

	// Проверяем, что в новом токене те же данные
	extractedUserID, err := VerifyToken(newToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)

	// Тест с неверным токеном
	_, err = RegenerateToken("invalid.token.string")
	assert.Error(t, err)
}

// Вспомогательная функция для создания просроченного токена
func createExpiredToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(-time.Hour).Unix(), // Истекший на 1 час
	})

	tokenString, _ := token.SignedString([]byte(JWTSecret))
	return tokenString
}
