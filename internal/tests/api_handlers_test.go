package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sanek1/GophKeeper/internal/api"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

// Создание тестового токена
func createTestToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("your-super-secret-key-change-in-production"))
	return tokenString
}

// Тесты для API
func TestAPI(t *testing.T) {
	// Отключаем реальные запросы к Swagger и инициализацию маршрутов
	gin.SetMode(gin.TestMode)

	// Чтобы не было проблем с конфликтами маршрутов Swagger в тестах,
	// используем NewTestAPI вместо NewAPI
	t.Run("API_CanBeCreated", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Проверяем, что API создано успешно
		assert.NotNil(t, apiInstance)
	})

	// Тест эндпоинта Register
	t.Run("Register_Handler", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Данные для регистрации
		registrationData := models.RegisterRequest{
			Login:    "test@example.com",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(registrationData)

		// Создаем тестовый запрос
		req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Создаем роутер для тестирования
		router := gin.New()
		router.POST("/api/register", apiInstance.Register)

		// Выполняем запрос
		router.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", response.Login)
		assert.NotEmpty(t, response.ID)
	})

	// Тест эндпоинта Login
	t.Run("Login_Handler", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Сначала регистрируем пользователя
		registrationData := models.RegisterRequest{
			Login:    "login_test@example.com",
			Password: "password123",
		}
		jsonRegistrationData, _ := json.Marshal(registrationData)

		regReq := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonRegistrationData))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)
		router.ServeHTTP(regW, regReq)

		assert.Equal(t, http.StatusCreated, regW.Code)

		// Теперь пытаемся залогиниться
		loginData := models.LoginRequest{
			Login:    "login_test@example.com",
			Password: "password123",
		}
		jsonLoginData, _ := json.Marshal(loginData)

		loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonLoginData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()

		router.POST("/api/login", apiInstance.Login)
		router.ServeHTTP(loginW, loginReq)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, loginW.Code)

		var response map[string]string
		err := json.Unmarshal(loginW.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["token"])
	})

	// Тест неудачного логина (неверный пароль)
	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Сначала регистрируем пользователя
		registrationData := models.RegisterRequest{
			Login:    "invalid_login@example.com",
			Password: "password123",
		}
		jsonRegistrationData, _ := json.Marshal(registrationData)

		regReq := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonRegistrationData))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)
		router.ServeHTTP(regW, regReq)

		assert.Equal(t, http.StatusCreated, regW.Code)

		// Пытаемся залогиниться с неверным паролем
		loginData := models.LoginRequest{
			Login:    "invalid_login@example.com",
			Password: "wrong_password",
		}
		jsonLoginData, _ := json.Marshal(loginData)

		loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonLoginData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()

		router.POST("/api/login", apiInstance.Login)
		router.ServeHTTP(loginW, loginReq)

		// Проверяем результат - должен быть отказ
		assert.Equal(t, http.StatusUnauthorized, loginW.Code)
		assert.Contains(t, loginW.Body.String(), "invalid credentials")
	})

	// Тест создания секрета
	t.Run("CreateSecret_Handler", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Сначала регистрируем пользователя
		registrationData := models.RegisterRequest{
			Login:    "secret_test@example.com",
			Password: "password123",
		}
		jsonRegistrationData, _ := json.Marshal(registrationData)

		regReq := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonRegistrationData))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)
		router.ServeHTTP(regW, regReq)

		// Получаем ID пользователя
		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()

		// Создаем токен для авторизации
		token := createTestToken(userID)

		// Данные секрета
		secretData := models.SecretRequest{
			Type:     "password",
			Data:     []byte("encrypted_data"),
			Metadata: "Test Password",
		}
		jsonSecretData, _ := json.Marshal(secretData)

		// Создаем запрос на создание секрета
		createReq := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecretData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createW := httptest.NewRecorder()

		// Создаем роутер с middleware
		secureRouter := gin.New()
		secureRouter.Use(apiInstance.GetAuthMiddleware())
		secureRouter.POST("/api/secrets", apiInstance.CreateSecret)

		// Выполняем запрос
		secureRouter.ServeHTTP(createW, createReq)

		// Проверяем результат
		assert.Equal(t, http.StatusCreated, createW.Code)

		var secret models.Secret
		err := json.Unmarshal(createW.Body.Bytes(), &secret)
		assert.NoError(t, err)
		assert.Equal(t, "password", secret.Type)
		assert.Equal(t, []byte("encrypted_data"), secret.Data)
		assert.Equal(t, "Test Password", secret.Metadata)
	})

	// Тест получения списка секретов
	t.Run("GetSecrets_Handler", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Регистрируем пользователя
		registrationData := models.RegisterRequest{
			Login:    "get_secrets@example.com",
			Password: "password123",
		}
		jsonRegistrationData, _ := json.Marshal(registrationData)

		regReq := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonRegistrationData))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)
		router.ServeHTTP(regW, regReq)

		// Получаем ID пользователя
		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()

		// Создаем токен для авторизации
		token := createTestToken(userID)

		// Создаем несколько секретов
		secretRouter := gin.New()
		secretRouter.Use(apiInstance.GetAuthMiddleware())
		secretRouter.POST("/api/secrets", apiInstance.CreateSecret)

		// Первый секрет
		secret1 := models.SecretRequest{
			Type:     "password",
			Data:     []byte("data1"),
			Metadata: "Password 1",
		}
		jsonSecret1, _ := json.Marshal(secret1)
		createReq1 := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecret1))
		createReq1.Header.Set("Content-Type", "application/json")
		createReq1.Header.Set("Authorization", "Bearer "+token)
		createW1 := httptest.NewRecorder()
		secretRouter.ServeHTTP(createW1, createReq1)

		// Второй секрет
		secret2 := models.SecretRequest{
			Type:     "card",
			Data:     []byte("data2"),
			Metadata: "Card 1",
		}
		jsonSecret2, _ := json.Marshal(secret2)
		createReq2 := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecret2))
		createReq2.Header.Set("Content-Type", "application/json")
		createReq2.Header.Set("Authorization", "Bearer "+token)
		createW2 := httptest.NewRecorder()
		secretRouter.ServeHTTP(createW2, createReq2)

		// Теперь получаем все секреты
		getReq := httptest.NewRequest("GET", "/api/secrets", nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getW := httptest.NewRecorder()

		secretRouter.GET("/api/secrets", apiInstance.GetSecrets)
		secretRouter.ServeHTTP(getW, getReq)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, getW.Code)

		var secrets []*models.Secret
		err := json.Unmarshal(getW.Body.Bytes(), &secrets)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(secrets))
	})

	// Тест получения одного секрета по ID
	t.Run("GetSecret_Handler", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Регистрируем пользователя
		registrationData := models.RegisterRequest{
			Login:    "get_one_secret@example.com",
			Password: "password123",
		}
		jsonRegistrationData, _ := json.Marshal(registrationData)

		regReq := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonRegistrationData))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)
		router.ServeHTTP(regW, regReq)

		// Получаем ID пользователя
		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()

		// Создаем токен для авторизации
		token := createTestToken(userID)

		// Создаем секрет
		secretRouter := gin.New()
		secretRouter.Use(apiInstance.GetAuthMiddleware())
		secretRouter.POST("/api/secrets", apiInstance.CreateSecret)

		secretData := models.SecretRequest{
			Type:     "note",
			Data:     []byte("note content"),
			Metadata: "Important Note",
		}
		jsonSecretData, _ := json.Marshal(secretData)
		createReq := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecretData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createW := httptest.NewRecorder()
		secretRouter.ServeHTTP(createW, createReq)

		// Получаем созданный секрет
		var createdSecret models.Secret
		_ = json.Unmarshal(createW.Body.Bytes(), &createdSecret)
		secretID := createdSecret.ID.String()

		// Запрашиваем секрет по ID
		getReq := httptest.NewRequest("GET", "/api/secrets/"+secretID, nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getW := httptest.NewRecorder()

		secretRouter.GET("/api/secrets/:id", apiInstance.GetSecret)
		secretRouter.ServeHTTP(getW, getReq)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, getW.Code)

		var secret models.Secret
		err := json.Unmarshal(getW.Body.Bytes(), &secret)
		assert.NoError(t, err)
		assert.Equal(t, secretID, secret.ID.String())
		assert.Equal(t, "note", secret.Type)
		assert.Equal(t, []byte("note content"), secret.Data)
		assert.Equal(t, "Important Note", secret.Metadata)
	})

	// Тест обновления секрета
	t.Run("UpdateSecret_Handler", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Регистрируем пользователя
		registrationData := models.RegisterRequest{
			Login:    "update_secret@example.com",
			Password: "password123",
		}
		jsonRegistrationData, _ := json.Marshal(registrationData)

		regReq := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonRegistrationData))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)
		router.ServeHTTP(regW, regReq)

		// Получаем ID пользователя
		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()

		// Создаем токен для авторизации
		token := createTestToken(userID)

		// Создаем секрет
		secretRouter := gin.New()
		secretRouter.Use(apiInstance.GetAuthMiddleware())
		secretRouter.POST("/api/secrets", apiInstance.CreateSecret)

		secretData := models.SecretRequest{
			Type:     "text",
			Data:     []byte("old content"),
			Metadata: "Old Text",
		}
		jsonSecretData, _ := json.Marshal(secretData)
		createReq := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecretData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createW := httptest.NewRecorder()
		secretRouter.ServeHTTP(createW, createReq)

		// Получаем созданный секрет
		var createdSecret models.Secret
		_ = json.Unmarshal(createW.Body.Bytes(), &createdSecret)
		secretID := createdSecret.ID.String()

		// Обновляем секрет
		updateData := models.SecretRequest{
			Type:     "text",
			Data:     []byte("new content"),
			Metadata: "Updated Text",
		}
		jsonUpdateData, _ := json.Marshal(updateData)
		updateReq := httptest.NewRequest("PUT", "/api/secrets/"+secretID, bytes.NewBuffer(jsonUpdateData))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateW := httptest.NewRecorder()

		secretRouter.PUT("/api/secrets/:id", apiInstance.UpdateSecret)
		secretRouter.ServeHTTP(updateW, updateReq)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, updateW.Code)

		var updatedSecret models.Secret
		err := json.Unmarshal(updateW.Body.Bytes(), &updatedSecret)
		assert.NoError(t, err)
		assert.Equal(t, secretID, updatedSecret.ID.String())
		assert.Equal(t, "text", updatedSecret.Type)
		assert.Equal(t, []byte("new content"), updatedSecret.Data)
		assert.Equal(t, "Updated Text", updatedSecret.Metadata)
	})

	// Тест удаления секрета
	t.Run("DeleteSecret_Handler", func(t *testing.T) {
		// Создаем мок-базу данных
		db := NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		// Инициализируем тестовый API
		apiInstance := api.NewTestAPI(db, jwtSecret)

		// Регистрируем пользователя
		registrationData := models.RegisterRequest{
			Login:    "delete_secret@example.com",
			Password: "password123",
		}
		jsonRegistrationData, _ := json.Marshal(registrationData)

		regReq := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonRegistrationData))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)
		router.ServeHTTP(regW, regReq)

		// Получаем ID пользователя
		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()

		// Создаем токен для авторизации
		token := createTestToken(userID)

		// Создаем секрет
		secretRouter := gin.New()
		secretRouter.Use(apiInstance.GetAuthMiddleware())
		secretRouter.POST("/api/secrets", apiInstance.CreateSecret)

		secretData := models.SecretRequest{
			Type:     "binary",
			Data:     []byte{0x01, 0x02, 0x03},
			Metadata: "Binary Data",
		}
		jsonSecretData, _ := json.Marshal(secretData)
		createReq := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecretData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createW := httptest.NewRecorder()
		secretRouter.ServeHTTP(createW, createReq)

		// Получаем созданный секрет
		var createdSecret models.Secret
		_ = json.Unmarshal(createW.Body.Bytes(), &createdSecret)
		secretID := createdSecret.ID.String()

		// Удаляем секрет
		deleteReq := httptest.NewRequest("DELETE", "/api/secrets/"+secretID, nil)
		deleteReq.Header.Set("Authorization", "Bearer "+token)
		deleteW := httptest.NewRecorder()

		secretRouter.DELETE("/api/secrets/:id", apiInstance.DeleteSecret)
		secretRouter.ServeHTTP(deleteW, deleteReq)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, deleteW.Code)

		// Пробуем получить удаленный секрет - должен быть код 404
		getReq := httptest.NewRequest("GET", "/api/secrets/"+secretID, nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getW := httptest.NewRecorder()

		secretRouter.GET("/api/secrets/:id", apiInstance.GetSecret)
		secretRouter.ServeHTTP(getW, getReq)

		assert.Equal(t, http.StatusNotFound, getW.Code)
	})
}
