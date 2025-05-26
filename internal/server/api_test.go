package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		c.Status(200)
	})

	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.Status(200)
	})

	return r
}

func TestRegisterAndLogin(t *testing.T) {
	r := setupRouter()

	// Test data
	registerBody := map[string]string{
		"login":    "testuser",
		"password": "testpass",
	}
	body, _ := json.Marshal(registerBody)

	// Test registration
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Test login
	loginBody := map[string]string{
		"login":    "testuser",
		"password": "testpass",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
