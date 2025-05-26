package api

import (
	"net/http"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/models"
)

// Register a new user
// @Summary Register new user
// @Description Register a new user with login and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration info"
// @Success 201 {object} models.User
// @Router /api/register [post]
func (a *API) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь с таким логином
	existingUser, err := a.userRepo.GetByLogin(req.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check user existence"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user with this login already exists"})
		return
	}

	user, err := a.userRepo.Create(req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login user
// @Summary Login user
// @Description Login with credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.LoginRequest true "User login info"
// @Success 200 {string} string "JWT token"
// @Router /api/login [post]
func (a *API) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := a.userRepo.GetByLogin(req.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	if user == nil || !a.userRepo.ValidatePassword(user, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Use hardcoded value of JWT_SECRET
	secretValue := "your-super-secret-key-change-in-production"

	// Debug information
	log.Printf("Generating JWT token for user %s with secret: %s", user.ID.String(), secretValue)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	// Use hardcoded value
	tokenString, err := token.SignedString([]byte(secretValue))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Debug information
	log.Printf("Generated token: %s", tokenString)

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// GetSecrets gets all secrets for authenticated user
// @Summary Get all secrets
// @Description Get all secrets for authenticated user
// @Tags secrets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Secret
// @Router /api/secrets [get]
func (a *API) GetSecrets(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	secrets, err := a.secretRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get secrets"})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// CreateSecret creates a new secret
// @Summary Create secret
// @Description Create new secret
// @Tags secrets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param secret body models.SecretRequest true "Secret info"
// @Success 201 {object} models.Secret
// @Router /api/secrets [post]
func (a *API) CreateSecret(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	var req models.SecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if secret type is valid
	validType := false
	for _, t := range models.SecretTypes {
		if t == req.Type {
			validType = true
			break
		}
	}
	if !validType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret type", "valid_types": models.SecretTypes})
		return
	}

	secret, err := a.secretRepo.Create(userID, req.Type, req.Data, req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create secret"})
		return
	}

	c.JSON(http.StatusCreated, secret)
}

// GetSecret gets secret by ID
// @Summary Get secret
// @Description Get secret by ID
// @Tags secrets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Secret ID"
// @Success 200 {object} models.Secret
// @Router /api/secrets/{id} [get]
func (a *API) GetSecret(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret id"})
		return
	}

	secret, err := a.secretRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get secret"})
		return
	}
	if secret == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
		return
	}
	if secret.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// UpdateSecret updates secret by ID
// @Summary Update secret
// @Description Update secret by ID
// @Tags secrets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Secret ID"
// @Param secret body models.SecretRequest true "Secret info"
// @Success 200 {object} models.Secret
// @Router /api/secrets/{id} [put]
func (a *API) UpdateSecret(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret id"})
		return
	}

	// Check if secret exists and belongs to user
	secret, err := a.secretRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get secret"})
		return
	}
	if secret == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
		return
	}
	if secret.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req models.SecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if secret type is valid
	validType := false
	for _, t := range models.SecretTypes {
		if t == req.Type {
			validType = true
			break
		}
	}
	if !validType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret type", "valid_types": models.SecretTypes})
		return
	}

	// Update secret
	updatedSecret, err := a.secretRepo.Update(id, req.Type, req.Data, req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update secret"})
		return
	}

	c.JSON(http.StatusOK, updatedSecret)
}

// DeleteSecret deletes secret by ID
// @Summary Delete secret
// @Description Delete secret by ID
// @Tags secrets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Secret ID"
// @Success 200 {object} models.Secret
// @Router /api/secrets/{id} [delete]
func (a *API) DeleteSecret(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret id"})
		return
	}

	// Check if secret exists and belongs to user
	secret, err := a.secretRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get secret"})
		return
	}
	if secret == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
		return
	}
	if secret.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Delete secret
	if err := a.secretRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete secret"})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// SyncData processes client data synchronization with server
// @Summary Sync data
// @Description Sync client data with server
// @Tags sync
// @Security BearerAuth
// @Success 200 {object} models.SyncStatus
// @Router /api/sync [post]
func (a *API) SyncData(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Get all secrets for user
	secrets, err := a.secretRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get secrets"})
		return
	}

	// Return synchronization status
	syncStatus := models.SyncStatus{
		LastSyncTime:    time.Now(),
		ItemsDownloaded: len(secrets),
		ItemsUploaded:   0,
		Success:         true,
	}

	c.JSON(http.StatusOK, syncStatus)
}
