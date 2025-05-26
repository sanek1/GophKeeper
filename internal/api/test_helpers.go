package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sanek1/GophKeeper/internal/database"
	"github.com/sanek1/GophKeeper/internal/repository/mocks"
)

// NewTestAPI creates an API instance for testing without registering Swagger and routes
func NewTestAPI(db database.DBInterface, jwtSecret string) *API {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Panic recovery
	router.Use(gin.Recovery())

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Create mock repositories for tests
	userRepo := mocks.NewMockUserRepository()
	secretRepo := mocks.NewMockSecretRepository()

	api := &API{
		router:     router,
		db:         db,
		userRepo:   userRepo,
		secretRepo: secretRepo,
		jwtSecret:  jwtSecret,
	}

	return api
}

// GetAuthMiddleware returns only middleware for testing
func (a *API) GetAuthMiddleware() gin.HandlerFunc {
	return a.AuthMiddleware()
}
