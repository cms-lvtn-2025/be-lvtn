package api

import (
	"thaily/src/pkg/response"

	"github.com/gin-gonic/gin"
)

type UpdateProfileRequest struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// GetCurrentUser lấy thông tin user hiện tại
func (h *APIHandler) GetCurrentUser(c *gin.Context) {
	if h.UserClient == nil {
		response.InternalError(c, "User service not available")
		return
	}

	// Lấy claims từ middleware
	claims, exists := c.Get("claims")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	// TODO: Gọi UserClient để lấy thông tin user
	// user, err := h.UserClient.GetUser(ctx, &pb.GetUserRequest{...})

	response.Success(c, gin.H{
		"claims": claims,
	})
}

// UpdateProfile cập nhật profile user
func (h *APIHandler) UpdateProfile(c *gin.Context) {
	if h.UserClient == nil {
		response.InternalError(c, "User service not available")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Implement update logic với UserClient

	response.SuccessWithMessage(c, "Profile updated successfully", nil)
}
