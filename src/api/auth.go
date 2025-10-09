package api

import (
	"net/http"
	"thaily/src/auth"
	"thaily/src/pkg/response"

	"github.com/gin-gonic/gin"
)

// Request/Response models

type GoogleLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

type GoogleCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// GoogleLogin tạo URL để redirect đến Google OAuth2
func (h *APIHandler) GoogleLogin(c *gin.Context) {
	authService := auth.NewService(h.Config, h.Redis, h.Mongodb)

	// Generate auth URL (state sẽ được tạo tự động bên trong)
	authURL := authService.GetAuthURL("")

	response.Success(c, GoogleLoginResponse{
		AuthURL: authURL,
	})
}

// GoogleCallback xử lý callback từ Google sau khi user đăng nhập
func (h *APIHandler) GoogleCallback(c *gin.Context) {
	var req GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	authService := auth.NewService(h.Config, h.Redis, h.Mongodb)

	// Exchange code để lấy user info
	googleUser, err := authService.ExchangeCode(c.Request.Context(), req.Code)
	if err != nil {
		response.InternalError(c, "Failed to exchange code: "+err.Error())
		return
	}

	// Generate token pair (access + refresh token)
	// NOTE: Không cần tạo user ngay, sẽ xử lý sau ở user service
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	tokenPair, err := authService.GenerateTokenPair(c.Request.Context(), googleUser, userAgent, ipAddress)
	if err != nil {
		response.InternalError(c, "Failed to generate tokens: "+err.Error())
		return
	}

	// Trả về token và thông tin Google user
	// NOTE: User data sẽ được xử lý sau ở user service
	response.SuccessWithMessage(c, "Login successful", gin.H{
		"google_user":   googleUser,
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
		"token_type":    tokenPair.TokenType,
	})
}

// RefreshToken làm mới access token bằng refresh token
func (h *APIHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	authService := auth.NewService(h.Config, h.Redis, h.Mongodb)

	tokenPair, err := authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	response.Success(c, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
		"token_type":    tokenPair.TokenType,
	})
}

// Logout đăng xuất và xóa session
func (h *APIHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	authService := auth.NewService(h.Config, h.Redis, h.Mongodb)

	if err := authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.InternalError(c, "Failed to logout: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "Logout successful", nil)
}
