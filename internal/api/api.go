// @title GophKeeper API
// @version 1.0
// @description API for GophKeeper password manager
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/yourusername/gophkeeper/docs"
	"github.com/yourusername/gophkeeper/internal/database"
	"github.com/yourusername/gophkeeper/internal/models"
	"github.com/yourusername/gophkeeper/internal/repository"
)

type API struct {
	router     *gin.Engine
	db         database.DBInterface
	userRepo   repository.UserRepository
	secretRepo repository.SecretRepository
	jwtSecret  string
}

func NewAPI(db database.DBInterface, jwtSecret string) *API {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Panic recovery
	router.Use(gin.Recovery())

	// Logger middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8080", "http://localhost:3000"} // Добавьте нужные домены
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"Accept",
	}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	router.Use(cors.New(corsConfig))

	api := &API{
		router:     router,
		db:         db,
		userRepo:   repository.NewUserRepository(db.DB()),
		secretRepo: repository.NewSecretRepository(db.DB()),
		jwtSecret:  jwtSecret,
	}

	// Register routes
	api.registerRoutes()

	// Статические файлы
	router.Static("/static", "./static")

	// Swagger UI с настройкой авторизации
	router.GET("/swagger/*any", func(c *gin.Context) {
		// Если запрашиваем скрипт авторизации
		if c.Request.URL.Path == "/swagger/auth.js" {
			c.File("./static/js/swagger-ui-auth.js")
			return
		}

		// Стандартный обработчик Swagger UI
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	// Инъекция скрипта в Swagger UI
	router.GET("/swagger/index.html", func(c *gin.Context) {
		// Добавляем скрипт авторизации к стандартной странице Swagger
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Swagger UI</title>
				<link rel="stylesheet" type="text/css" href="./swagger-ui.css" />
				<link rel="stylesheet" type="text/css" href="./index.css" />
				<link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
				<link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
			</head>
			<body>
				<div id="swagger-ui"></div>
				<script src="./swagger-ui-bundle.js" charset="UTF-8"> </script>
				<script src="./swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
				<script src="./auth.js" charset="UTF-8"> </script>
				<script>
				window.onload = function() {
					window.ui = SwaggerUIBundle({
						url: "./doc.json",
						dom_id: '#swagger-ui',
						deepLinking: true,
						presets: [
							SwaggerUIBundle.presets.apis,
							SwaggerUIStandalonePreset
						],
						plugins: [
							SwaggerUIBundle.plugins.DownloadUrl
						],
						layout: "StandaloneLayout",
						persistAuthorization: true
					});
				};
				</script>
			</body>
			</html>
		`)
	})

	return api
}

// Router returns the Gin router
func (a *API) Router() *gin.Engine {
	return a.router
}

// registerRoutes registers all API routes
func (a *API) registerRoutes() {
	// Public routes
	public := a.router.Group("/api")
	{
		public.POST("/register", a.validateRegisterRequest(), a.Register)
		public.POST("/login", a.validateLoginRequest(), a.Login)
	}

	// Protected routes
	protected := a.router.Group("/api")
	protected.Use(a.AuthMiddleware())
	{
		protected.GET("/secrets", a.GetSecrets)
		protected.POST("/secrets", a.validateSecretRequest(), a.CreateSecret)
		protected.GET("/secrets/:id", a.GetSecret)
		protected.PUT("/secrets/:id", a.validateSecretRequest(), a.UpdateSecret)
		protected.DELETE("/secrets/:id", a.DeleteSecret)
	}
}

func (a *API) Run(addr string) error {
	log.Printf("Server starting on address: %s\n", addr)
	return a.router.Run(addr)
}

// Middleware for validating the registration request
func (a *API) validateRegisterRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
			c.Abort()
			return
		}

		var req models.RegisterRequest
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Next()

	}
}

// Middleware for validating the login request
func (a *API) validateLoginRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
			c.Abort()
			return
		}

		var req models.LoginRequest
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Next()
	}
}

// Middleware for validating the secret request
func (a *API) validateSecretRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SecretRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			c.Abort()
			return
		}

		// Check secret type
		validType := false
		for _, t := range models.SecretTypes {
			if req.Type == t {
				validType = true
				break
			}
		}
		if !validType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret type"})
			c.Abort()
			return
		}

		c.Next()
	}
}
