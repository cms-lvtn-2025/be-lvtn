package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"thaily/src/server/config"
	"thaily/src/server/graph/helper"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		claims, err := helper.ValidateAndParseClaims(authHeader, cfg.AccessSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// Set claims into context
		c.Set(helper.Auth, claims)
		c.Next()
	}
}

// OptionalAuthMiddleware allows requests without token
func OptionalAuthMiddleware(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Only validate if Authorization header exists
		if authHeader != "" {
			claims, err := helper.ValidateAndParseClaims(authHeader, cfg.AccessSecret)
			if err == nil && claims != nil {
				c.Set("claims", claims)
			}
		}

		c.Next()
	}
}
