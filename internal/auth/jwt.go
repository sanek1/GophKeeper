package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Фиксированный секрет для подписи JWT токенов
const JWTSecret = "your-super-secret-key-change-in-production"

// Структура для хранения данных в JWT
type JWTClaims struct {
	UserID string `json:"user_id"`
	Exp    int64  `json:"exp"`
}

// CreateToken создает новый JWT токен для пользователя
func CreateToken(userID string, expirationHours int) (string, error) {
	if expirationHours <= 0 {
		expirationHours = 24 // По умолчанию 24 часа
	}

	// Создаем стандартный JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix(),
	})

	// Подписываем фиксированным секретом
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", fmt.Errorf("ошибка создания токена: %w", err)
	}

	return tokenString, nil
}

// VerifyToken проверяет валидность JWT токена
func VerifyToken(tokenString string) (string, error) {
	// Проверяем токен с помощью библиотеки jwt
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что используется правильный алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("ошибка проверки токена: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("недействительный токен")
	}

	// Извлекаем идентификатор пользователя из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("недействительные данные токена")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("недействительный идентификатор пользователя в токене")
	}

	return userID, nil
}

// ExtractUserIDFromToken извлекает идентификатор пользователя из токена
// без проверки подписи (полезно для перегенерации токена)
func ExtractUserIDFromToken(tokenString string) (string, int64, error) {
	// Разделяем токен на части
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", 0, fmt.Errorf("неверный формат токена")
	}

	// Декодируем данные (вторую часть) токена
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("ошибка декодирования токена: %w", err)
	}

	// Разбираем JSON с данными
	var claims struct {
		UserID string `json:"user_id"`
		Exp    int64  `json:"exp"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", 0, fmt.Errorf("ошибка разбора данных токена: %w", err)
	}

	return claims.UserID, claims.Exp, nil
}

// RegenerateToken создает новый токен на основе данных из существующего
// с обновленным временем истечения
func RegenerateToken(tokenString string) (string, error) {
	// Извлекаем данные из существующего токена
	userID, _, err := ExtractUserIDFromToken(tokenString)
	if err != nil {
		return "", err
	}

	// Создаем новый токен с обновленным временем истечения (25 часов вместо 24)
	// Это гарантирует, что новый токен будет отличаться от старого
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(25 * time.Hour).Unix(),
	})

	// Подписываем новый токен нашим фиксированным секретом
	newTokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", fmt.Errorf("ошибка создания нового токена: %w", err)
	}

	return newTokenString, nil
}
