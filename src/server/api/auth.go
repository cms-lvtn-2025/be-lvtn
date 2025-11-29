package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	pbCommon "thaily/proto/common"
	"thaily/src/server/auth"
	"thaily/src/server/graph/helper"
	"thaily/src/server/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Request/Response models

type GoogleLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

type GoogleCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
	Role  string `json:"role"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// GoogleLogin tạo URL để redirect đến Google OAuth2
func (h *APIHandler) GoogleLogin(c *gin.Context) {
	authService := auth.NewService(h.Config, h.Redis, h.Mongodb, h.UserClient)

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

	authService := auth.NewService(h.Config, h.Redis, h.Mongodb, h.UserClient)

	// Exchange code để lấy user info
	googleUser, err := authService.ExchangeCode(c.Request.Context(), req.Code)
	if err != nil {
		response.InternalError(c, "Failed to exchange code: "+err.Error())
		return
	}
	if req.Role == "student" {
		user, err := h.UserClient.GetUserByEmail(context.Background(), googleUser.Email)
		if err != nil && user == nil && len(user.GetStudents()) == 0 {
			response.InternalError(c, "Failed to get user: "+err.Error())
		}
		ids := ""
		for _, user := range user.GetStudents() {
			ids += user.GetSemesterCode() + ":" + user.GetId() + ","
		}

		// Generate token pair (access + refresh token)

		// NOTE: Không cần tạo user ngay, sẽ xử lý sau ở user service
		userAgent := c.Request.UserAgent()
		ipAddress := c.ClientIP()

		tokenPair, err := authService.GenerateTokenPair(c.Request.Context(), ids, "student", googleUser, userAgent, ipAddress)
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
	} else if req.Role == "teacher" {
		teachers, err := h.UserClient.GetTeacherByEmail(context.Background(), googleUser.Email)
		if err != nil && teachers == nil && len(teachers.GetTeachers()) == 0 {
			response.InternalError(c, "Failed to get user: "+err.Error())
		}
		ids := ""
		fmt.Print(teachers)
		for _, teacher := range teachers.GetTeachers() {
			ids += teacher.GetSemesterCode() + ":" + teacher.GetId() + ","
		}
		fmt.Println(ids)
		// Generate token pair (access + refresh token)

		// NOTE: Không cần tạo user ngay, sẽ xử lý sau ở user service
		userAgent := c.Request.UserAgent()
		ipAddress := c.ClientIP()

		tokenPair, err := authService.GenerateTokenPair(c.Request.Context(), ids, "teacher", googleUser, userAgent, ipAddress)
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

}

// RefreshToken làm mới access token bằng refresh token
func (h *APIHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	authService := auth.NewService(h.Config, h.Redis, h.Mongodb, h.UserClient)

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

	authService := auth.NewService(h.Config, h.Redis, h.Mongodb, h.UserClient)

	if err := authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.InternalError(c, "Failed to logout: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "Logout successful", nil)
}

func (h *APIHandler) extractRole(c *gin.Context, id string) ([]string, error) {
	if h.RoleClient == nil {
		return nil, fmt.Errorf("role client not initialized")
	}
	newSearch := &pbCommon.SearchRequest{
		Filters: []*pbCommon.FilterCriteria{
			{
				Criteria: &pbCommon.FilterCriteria_Condition{
					Condition: &pbCommon.FilterCondition{
						Field:    "teacher_code",
						Operator: pbCommon.FilterOperator_EQUAL,
						Values:   []string{id},
					},
				},
			},
		},
	}
	roles, err := h.RoleClient.GetRoleBySearch(c.Request.Context(), newSearch)
	if err != nil {
		return nil, err
	}
	if roles == nil || roles.RoleSystems == nil {
		return nil, fmt.Errorf("role not found")
	}
	result := []string{}
	for _, role := range roles.RoleSystems {
		result = append(result, string(role.Role))
	}
	return result, nil
}

// extractUserInfo extracts user information from context
func (h *APIHandler) extractUserInfo(c *gin.Context) (*UserInfo, error) {
	// Use c.Get() instead of c.Value() for gin.Context
	claimsValue, exists := c.Get(helper.Auth)
	if !exists {
		return nil, fmt.Errorf("not authorized - claims not found")
	}

	claims, ok := claimsValue.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("not authorized - invalid claims type")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("role not found in claims")
	}
	// Get semester from context (might be empty)
	semester := c.GetHeader("x-semester")

	// Parse IDs from claims
	idsStr, ok := claims["ids"].(string)
	if !ok {
		return nil, fmt.Errorf("ids not found in claims")
	}

	idsArr := strings.Split(idsStr, ",")
	myID := ""

	if semester == "" {
		// If no semester specified, use first ID
		if len(idsArr) > 0 {
			parts := strings.Split(idsArr[0], ":")
			if len(parts) > 1 {
				myID = parts[1]
				semester = parts[0]
			}
		}
	} else {
		// Find ID matching the semester
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+":") {
				parts := strings.Split(id, ":")
				if len(parts) > 1 {
					myID = parts[1]
					break
				}
			}
		}
	}

	if myID == "" {
		return nil, fmt.Errorf("no user ID found for semester %s", semester)
	}

	return &UserInfo{
		Role:     role,
		Semester: semester,
		UserID:   myID,
		IDs:      idsArr,
	}, nil
}
