package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/gophkeeper/internal/api"
)

// Тест middleware без инициализации полного API
func TestAuthMiddleware(t *testing.T) {
	// Устанавливаем режим gin
	gin.SetMode(gin.TestMode)

	// Создаем тестовый API с мок-базой данных
	db := NewMockDatabase()
	jwtSecret := "your-super-secret-key-change-in-production"

	// Используем NewTestAPI вместо NewAPI, чтобы избежать регистрации Swagger
	testAPI := api.NewTestAPI(db, jwtSecret)

	// Тесты для AuthMiddleware
	t.Run("MissingAuthHeader", func(t *testing.T) {
		// Создаем тестовый запрос
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Применяем middleware
		testAPI.GetAuthMiddleware()(ctx)

		// Проверяем результат
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authorization header is required")
	})

	t.Run("InvalidAuthHeaderFormat", func(t *testing.T) {
		// Создаем тестовый запрос с неверным форматом заголовка
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Применяем middleware
		testAPI.GetAuthMiddleware()(ctx)

		// Проверяем результат
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid authorization header format")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Создаем тестовый запрос с неверным токеном
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Применяем middleware
		testAPI.GetAuthMiddleware()(ctx)

		// Проверяем результат
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ValidToken", func(t *testing.T) {
		// Создаем валидный токен с нашим секретом
		userID := "test-user-id"
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"exp":     2147483647, // Очень далекая дата истечения
		})
		tokenString, err := token.SignedString([]byte("your-super-secret-key-change-in-production"))
		assert.NoError(t, err)

		// Создаем запрос с валидным токеном
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		// Создаем минимальный тестовый роутер
		router := gin.New()

		// Создаем handler, который будет вызван после middleware
		var userIDFromContext string
		handler := func(c *gin.Context) {
			id, exists := c.Get("user_id")
			if exists {
				userIDFromContext = id.(string)
			}
			c.Status(http.StatusOK)
		}

		// Используем middleware в минимальном роутере
		router.GET("/", testAPI.GetAuthMiddleware(), handler)
		router.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, userID, userIDFromContext)
	})
}
