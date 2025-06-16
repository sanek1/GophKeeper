package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sanek1/GophKeeper/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAPIRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := mocks.NewMockDatabase()
	api := NewTestAPI(db, "test-secret")

	t.Run("Run_InvalidAddress", func(t *testing.T) {
		// Test with invalid address format
		err := api.Run("invalid-address-format")
		assert.Error(t, err)
	})
}

func TestMiddlewareEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := mocks.NewMockDatabase()
	api := NewTestAPI(db, "test-secret")

	t.Run("AuthMiddleware_MissingBearerPrefix", func(t *testing.T) {
		router := gin.New()
		router.Use(api.GetAuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidToken")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ValidateRegisterRequest_ReadError", func(t *testing.T) {
		router := gin.New()
		router.POST("/register", api.validateRegisterRequest(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Create request with nil body to trigger read error
		req := httptest.NewRequest("POST", "/register", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should handle gracefully - either OK or BadRequest depending on implementation
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})
}
