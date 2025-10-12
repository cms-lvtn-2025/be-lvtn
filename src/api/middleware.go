package api

import (
	"fmt"
	"net/http"
	"thaily/src/config"
	"thaily/src/graph/helper"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware kiểm tra JWT token
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
		fmt.Println(claims)
		// Set claims vào context
		c.Set(helper.Auth, claims)
		c.Next()
	}
}

// OptionalAuthMiddleware cho phép request không có token
func OptionalAuthMiddleware(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Chỉ validate nếu có Authorization header
		if authHeader != "" {
			claims, err := helper.ValidateAndParseClaims(authHeader, cfg.AccessSecret)
			if err == nil && claims != nil {
				c.Set("claims", claims)
			}
		}

		c.Next()
	}
}
