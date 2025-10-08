package api

import (
	"net/http"
	"strings"
	"thaily/src/graph/helper"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware kiểm tra JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
			claims, err := helper.ParseJWT(token)
			if claims == nil || err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
				})
				c.Abort()
				return
			}

			// Set claims vào context
			c.Set("claims", claims)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization format",
			})
			c.Abort()
		}
	}
}

// OptionalAuthMiddleware cho phép request không có token
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token != "" && len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
			claims, err := helper.ParseJWT(token)
			if claims != nil && err == nil {
				c.Set("claims", claims)
			}
		}

		c.Next()
	}
}
