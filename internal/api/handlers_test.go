package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/mocks"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-jwt-secret-for-handlers"

func setupTestAPI() *API {
	gin.SetMode(gin.TestMode)
	db := mocks.NewMockDatabase()
	return NewTestAPI(db, testJWTSecret)
}

func createTestJWT(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     2147483647, // Far future
	})
	tokenString, _ := token.SignedString([]byte(testJWTSecret))
	return tokenString
}

func TestAPI_Register(t *testing.T) {
	api := setupTestAPI()
	router := gin.New()
	router.POST("/register", api.Register)

	t.Run("SuccessfulRegistration", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Login:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("EmptyLogin", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Login:    "",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ShortPassword", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Login:    "test@example.com",
			Password: "123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DuplicateUser", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Login:    "duplicate@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		req = httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestAPI_Login(t *testing.T) {
	api := setupTestAPI()
	router := gin.New()
	router.POST("/register", api.Register)
	router.POST("/login", api.Login)

	reqBody := models.RegisterRequest{
		Login:    "logintest@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	t.Run("SuccessfulLogin", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Login:    "logintest@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["token"])
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Login:    "logintest@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Login:    "nonexistent@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAPI_SecretOperations(t *testing.T) {
	api := setupTestAPI()
	router := gin.New()
	router.Use(api.GetAuthMiddleware())
	router.POST("/secrets", api.CreateSecret)
	router.GET("/secrets", api.GetSecrets)
	router.GET("/secrets/:id", api.GetSecret)
	router.PUT("/secrets/:id", api.UpdateSecret)
	router.DELETE("/secrets/:id", api.DeleteSecret)

	userID := uuid.New().String()
	token := createTestJWT(userID)

	t.Run("CreateSecret_Success", func(t *testing.T) {
		secretReq := models.SecretRequest{
			Type:     "password",
			Data:     []byte("encrypted_data"),
			Metadata: "Test Secret",
		}
		body, _ := json.Marshal(secretReq)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Secret
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "password", response.Type)
		assert.Equal(t, "Test Secret", response.Metadata)
	})

	t.Run("CreateSecret_InvalidType", func(t *testing.T) {
		secretReq := models.SecretRequest{
			Type:     "invalid_type",
			Data:     []byte("encrypted_data"),
			Metadata: "Test Secret",
		}
		body, _ := json.Marshal(secretReq)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateSecret_EmptyData", func(t *testing.T) {
		secretReq := models.SecretRequest{
			Type:     "password",
			Data:     []byte{},
			Metadata: "Test Secret",
		}
		body, _ := json.Marshal(secretReq)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetSecrets_Success", func(t *testing.T) {
		secretReq := models.SecretRequest{
			Type:     "note",
			Data:     []byte("note_data"),
			Metadata: "Test Note",
		}
		body, _ := json.Marshal(secretReq)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		req = httptest.NewRequest("GET", "/secrets", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var secrets []models.Secret
		err := json.Unmarshal(w.Body.Bytes(), &secrets)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(secrets), 1)
	})

	t.Run("GetSecret_ByID", func(t *testing.T) {
		secretReq := models.SecretRequest{
			Type:     "text",
			Data:     []byte("text_data"),
			Metadata: "Get Test",
		}
		body, _ := json.Marshal(secretReq)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var createdSecret models.Secret
		err := json.Unmarshal(w.Body.Bytes(), &createdSecret)
		require.NoError(t, err)

		req = httptest.NewRequest("GET", "/secrets/"+createdSecret.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var secret models.Secret
		err = json.Unmarshal(w.Body.Bytes(), &secret)
		require.NoError(t, err)
		assert.Equal(t, createdSecret.ID, secret.ID)
	})

	t.Run("GetSecret_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New().String()

		req := httptest.NewRequest("GET", "/secrets/"+nonExistentID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("UpdateSecret_Success", func(t *testing.T) {
		secretReq := models.SecretRequest{
			Type:     "card",
			Data:     []byte("card_data"),
			Metadata: "Original Card",
		}
		body, _ := json.Marshal(secretReq)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var createdSecret models.Secret
		err := json.Unmarshal(w.Body.Bytes(), &createdSecret)
		require.NoError(t, err)

		updateReq := models.SecretRequest{
			Type:     "card",
			Data:     []byte("updated_card_data"),
			Metadata: "Updated Card",
		}
		body, _ = json.Marshal(updateReq)

		req = httptest.NewRequest("PUT", "/secrets/"+createdSecret.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedSecret models.Secret
		err = json.Unmarshal(w.Body.Bytes(), &updatedSecret)
		require.NoError(t, err)
		assert.Equal(t, "Updated Card", updatedSecret.Metadata)
	})

	t.Run("UpdateSecret_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New().String()

		updateReq := models.SecretRequest{
			Type:     "password",
			Data:     []byte("new_data"),
			Metadata: "New Metadata",
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest("PUT", "/secrets/"+nonExistentID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("DeleteSecret_Success", func(t *testing.T) {
		// Создаем секрет для удаления
		secretReq := models.SecretRequest{
			Type:     "binary",
			Data:     []byte{0x01, 0x02, 0x03},
			Metadata: "Delete Test",
		}
		body, _ := json.Marshal(secretReq)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var createdSecret models.Secret
		err := json.Unmarshal(w.Body.Bytes(), &createdSecret)
		require.NoError(t, err)

		// Удаляем секрет
		req = httptest.NewRequest("DELETE", "/secrets/"+createdSecret.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем, что секрет удален
		req = httptest.NewRequest("GET", "/secrets/"+createdSecret.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("DeleteSecret_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New().String()

		req := httptest.NewRequest("DELETE", "/secrets/"+nonExistentID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAPI_ValidationMiddleware(t *testing.T) {
	api := setupTestAPI()

	t.Run("ValidateRegisterRequest", func(t *testing.T) {
		router := gin.New()
		router.POST("/register", api.validateRegisterRequest(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Valid request
		reqBody := models.RegisterRequest{
			Login:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Invalid request - malformed JSON
		req = httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidateSecretRequest", func(t *testing.T) {
		router := gin.New()
		router.POST("/secrets", api.validateSecretRequest(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Valid request
		reqBody := models.SecretRequest{
			Type:     "password",
			Data:     []byte("test_data"),
			Metadata: "Test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Invalid request - invalid type
		reqBody.Type = "invalid"
		body, _ = json.Marshal(reqBody)

		req = httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAPI_ErrorHandling(t *testing.T) {
	api := setupTestAPI()

	t.Run("InvalidSecretID", func(t *testing.T) {
		router := gin.New()
		router.Use(api.GetAuthMiddleware())
		router.GET("/secrets/:id", api.GetSecret)

		userID := uuid.New().String()
		token := createTestJWT(userID)

		req := httptest.NewRequest("GET", "/secrets/invalid-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UnauthorizedAccess", func(t *testing.T) {
		router := gin.New()
		router.Use(api.GetAuthMiddleware())
		router.GET("/secrets", api.GetSecrets)

		req := httptest.NewRequest("GET", "/secrets", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidJSONRequest", func(t *testing.T) {
		router := gin.New()
		router.POST("/register", api.Register)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DatabaseErrors", func(t *testing.T) {
		// Test various database error scenarios
		router := gin.New()
		router.POST("/register", api.Register)

		// Test duplicate registration
		reqBody := models.RegisterRequest{
			Login:    "duplicate@test.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		// First registration
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Second registration (should fail)
		req = httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}
