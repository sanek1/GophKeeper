package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sanek1/GophKeeper/internal/mocks"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := mocks.NewMockDatabase()
	api := NewTestAPI(db, "test-secret")

	t.Run("ValidateLoginRequest_Success", func(t *testing.T) {
		router := gin.New()
		router.POST("/login", api.validateLoginRequest(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		reqBody := models.LoginRequest{
			Login:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ValidateLoginRequest_InvalidJSON", func(t *testing.T) {
		router := gin.New()
		router.POST("/login", api.validateLoginRequest(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidateSecretRequest_EmptyData", func(t *testing.T) {
		router := gin.New()
		router.POST("/secrets", api.validateSecretRequest(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		reqBody := models.SecretRequest{
			Type:     "password",
			Data:     []byte{},
			Metadata: "Test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code) // Empty data is allowed
	})
}
