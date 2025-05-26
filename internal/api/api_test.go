package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sanek1/GophKeeper/internal/mocks"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func createTestToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("your-super-secret-key-change-in-production"))
	return tokenString
}

// Tests for API
func TestAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("API_CanBeCreated", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"
		apiInstance := NewTestAPI(db, jwtSecret)
		assert.NotNil(t, apiInstance)
	})

	t.Run("Register_Handler", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"
		apiInstance := NewTestAPI(db, jwtSecret)
		registrationData := models.RegisterRequest{
			Login:    "test@example.com",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(registrationData)
		req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/api/register", apiInstance.Register)

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", response.Login)
		assert.NotEmpty(t, response.ID)
	})

	t.Run("Login_Handler", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"
		apiInstance := NewTestAPI(db, jwtSecret)

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

		assert.Equal(t, http.StatusOK, loginW.Code)

		var response map[string]string
		err := json.Unmarshal(loginW.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		apiInstance := NewTestAPI(db, jwtSecret)
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

		assert.Equal(t, http.StatusUnauthorized, loginW.Code)
		assert.Contains(t, loginW.Body.String(), "invalid credentials")
	})

	t.Run("CreateSecret_Handler", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		apiInstance := NewTestAPI(db, jwtSecret)

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

		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()
		token := createTestToken(userID)

		secretData := models.SecretRequest{
			Type:     "password",
			Data:     []byte("encrypted_data"),
			Metadata: "Test Password",
		}
		jsonSecretData, _ := json.Marshal(secretData)
		createReq := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecretData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createW := httptest.NewRecorder()

		secureRouter := gin.New()
		secureRouter.Use(apiInstance.GetAuthMiddleware())
		secureRouter.POST("/api/secrets", apiInstance.CreateSecret)

		secureRouter.ServeHTTP(createW, createReq)

		assert.Equal(t, http.StatusCreated, createW.Code)

		var secret models.Secret
		err := json.Unmarshal(createW.Body.Bytes(), &secret)
		assert.NoError(t, err)
		assert.Equal(t, "password", secret.Type)
		assert.Equal(t, []byte("encrypted_data"), secret.Data)
		assert.Equal(t, "Test Password", secret.Metadata)
	})

	t.Run("GetSecrets_Handler", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		apiInstance := NewTestAPI(db, jwtSecret)
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

		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()
		token := createTestToken(userID)

		secretRouter := gin.New()
		secretRouter.Use(apiInstance.GetAuthMiddleware())
		secretRouter.POST("/api/secrets", apiInstance.CreateSecret)

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

		getReq := httptest.NewRequest("GET", "/api/secrets", nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getW := httptest.NewRecorder()

		secretRouter.GET("/api/secrets", apiInstance.GetSecrets)
		secretRouter.ServeHTTP(getW, getReq)

		assert.Equal(t, http.StatusOK, getW.Code)

		var secrets []*models.Secret
		err := json.Unmarshal(getW.Body.Bytes(), &secrets)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(secrets))
	})

	t.Run("GetSecret_Handler", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		apiInstance := NewTestAPI(db, jwtSecret)

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

		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()

		token := createTestToken(userID)

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

		var createdSecret models.Secret
		_ = json.Unmarshal(createW.Body.Bytes(), &createdSecret)
		secretID := createdSecret.ID.String()

		getReq := httptest.NewRequest("GET", "/api/secrets/"+secretID, nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getW := httptest.NewRecorder()

		secretRouter.GET("/api/secrets/:id", apiInstance.GetSecret)
		secretRouter.ServeHTTP(getW, getReq)

		assert.Equal(t, http.StatusOK, getW.Code)

		var secret models.Secret
		err := json.Unmarshal(getW.Body.Bytes(), &secret)
		assert.NoError(t, err)
		assert.Equal(t, secretID, secret.ID.String())
		assert.Equal(t, "note", secret.Type)
		assert.Equal(t, []byte("note content"), secret.Data)
		assert.Equal(t, "Important Note", secret.Metadata)
	})

	t.Run("UpdateSecret_Handler", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"
		apiInstance := NewTestAPI(db, jwtSecret)

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

		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()
		token := createTestToken(userID)

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

		var createdSecret models.Secret
		_ = json.Unmarshal(createW.Body.Bytes(), &createdSecret)
		secretID := createdSecret.ID.String()

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

		assert.Equal(t, http.StatusOK, updateW.Code)

		var updatedSecret models.Secret
		err := json.Unmarshal(updateW.Body.Bytes(), &updatedSecret)
		assert.NoError(t, err)
		assert.Equal(t, secretID, updatedSecret.ID.String())
		assert.Equal(t, "text", updatedSecret.Type)
		assert.Equal(t, []byte("new content"), updatedSecret.Data)
		assert.Equal(t, "Updated Text", updatedSecret.Metadata)
	})

	t.Run("DeleteSecret_Handler", func(t *testing.T) {
		db := mocks.NewMockDatabase()
		jwtSecret := "your-super-secret-key-change-in-production"

		apiInstance := NewTestAPI(db, jwtSecret)
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
		var user models.User
		_ = json.Unmarshal(regW.Body.Bytes(), &user)
		userID := user.ID.String()

		token := createTestToken(userID)

		secretRouter := gin.New()
		secretRouter.Use(apiInstance.GetAuthMiddleware())
		secretRouter.POST("/api/secrets", apiInstance.CreateSecret)

		secretData := models.SecretRequest{
			Type:     "text",
			Data:     []byte("content to delete"),
			Metadata: "Text to Delete",
		}
		jsonSecretData, _ := json.Marshal(secretData)
		createReq := httptest.NewRequest("POST", "/api/secrets", bytes.NewBuffer(jsonSecretData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createW := httptest.NewRecorder()
		secretRouter.ServeHTTP(createW, createReq)

		var createdSecret models.Secret
		_ = json.Unmarshal(createW.Body.Bytes(), &createdSecret)
		secretID := createdSecret.ID.String()

		deleteReq := httptest.NewRequest("DELETE", "/api/secrets/"+secretID, nil)
		deleteReq.Header.Set("Authorization", "Bearer "+token)
		deleteW := httptest.NewRecorder()

		secretRouter.DELETE("/api/secrets/:id", apiInstance.DeleteSecret)
		secretRouter.ServeHTTP(deleteW, deleteReq)

		assert.Equal(t, http.StatusOK, deleteW.Code)

		getReq := httptest.NewRequest("GET", "/api/secrets/"+secretID, nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getW := httptest.NewRecorder()

		secretRouter.GET("/api/secrets/:id", apiInstance.GetSecret)
		secretRouter.ServeHTTP(getW, getReq)

		assert.Equal(t, http.StatusNotFound, getW.Code)
	})
}
