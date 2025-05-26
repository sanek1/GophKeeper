package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const jwtSecretValue = "your-super-secret-key-change-in-production"

func (a *API) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		var tokenString string
		// Check format - with Bearer or without
		if strings.HasPrefix(authHeader, "Bearer ") {
			// Standard format "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else {
			// Token passed directly without "Bearer " - accept, but log a warning
			log.Printf("WARNING: Token passed without 'Bearer' prefix. Use format 'Bearer <token>'")
			tokenString = authHeader
		}

		// Debug information
		log.Printf("Checking token, JWT_SECRET: %s, token: %s", jwtSecretValue, tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check signature algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Use hardcoded value instead of a.jwtSecret
			return []byte(jwtSecretValue), nil
		})

		if err != nil {
			log.Printf("Error checking token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Printf("Token is invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
