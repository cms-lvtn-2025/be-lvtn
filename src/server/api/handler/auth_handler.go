package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"thaily/src/server/api/model"
	"thaily/src/server/response"
)

// GoogleLogin creates URL to redirect to Google OAuth2
// POST /api/auth/google/login
func (h *Handler) GoogleLogin(c *gin.Context) {
	authURL := h.AuthService.GetAuthURL()

	response.Success(c, model.GoogleLoginResponse{
		AuthURL: authURL,
	})
}

// GoogleCallback handles callback from Google after user login
// POST /api/auth/google/callback
func (h *Handler) GoogleCallback(c *gin.Context) {
	var req model.GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	tokenResp, err := h.AuthService.HandleGoogleCallback(c.Request.Context(), req, userAgent, ipAddress)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Login successful", gin.H{
		"google_user":   tokenResp.GoogleUser,
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
		"expires_in":    tokenResp.ExpiresIn,
		"token_type":    tokenResp.TokenType,
	})
}

// RefreshToken refreshes access token using refresh token
// POST /api/auth/refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	tokenResp, err := h.AuthService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	response.Success(c, gin.H{
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
		"expires_in":    tokenResp.ExpiresIn,
		"token_type":    tokenResp.TokenType,
	})
}

// Logout logs out and removes session
// POST /api/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var req model.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.AuthService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.InternalError(c, "Failed to logout: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "Logout successful", nil)
}
