package api

import (
	"thaily/src/pkg/response"

	"github.com/gin-gonic/gin"
)

type GoogleLoginRequest struct {
	RedirectURI string `json:"redirect_uri"`
}

type GoogleCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// GoogleLogin xử lý đăng nhập với Google
func (h *APIHandler) GoogleLogin(c *gin.Context) {
	if h.UserClient == nil {
		response.InternalError(c, "User service not available")
		return
	}

	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Nếu không có body, dùng default
	}

	// TODO: Implement Google OAuth flow
	// 1. Generate OAuth URL
	// 2. Return URL để frontend redirect
	oauthURL := "https://accounts.google.com/o/oauth2/v2/auth?..."

	response.Success(c, gin.H{
		"url": oauthURL,
	})
}

// GoogleCallback xử lý callback từ Google
func (h *APIHandler) GoogleCallback(c *gin.Context) {
	if h.UserClient == nil {
		response.InternalError(c, "User service not available")
		return
	}

	var req GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Implement callback logic
	// 1. Exchange code for token
	// 2. Get user info from Google
	// 3. Create/update user via gRPC UserClient
	// 4. Generate JWT token

	response.SuccessWithMessage(c, "Login successful", gin.H{
		"token": "jwt_token_here",
	})
}
