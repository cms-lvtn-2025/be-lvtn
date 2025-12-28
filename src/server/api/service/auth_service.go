package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	pbCommon "thaily/proto/common"
	pbRole "thaily/proto/role"
	"thaily/src/server/api/model"
	"thaily/src/server/auth"
	"thaily/src/server/client"
	"thaily/src/server/config"
	"thaily/src/server/graph/helper"
)

// AuthService handles authentication business logic
type AuthService struct {
	Config     *config.Config
	UserClient *client.GRPCUser
	RoleClient *client.GRPCRole
	Redis      *client.RedisClient
	Mongodb    *client.MongoClient
}

// NewAuthService creates a new AuthService
func NewAuthService(cfg *config.Config, userClient *client.GRPCUser, roleClient *client.GRPCRole, redis *client.RedisClient, mongodb *client.MongoClient) *AuthService {
	return &AuthService{
		Config:     cfg,
		UserClient: userClient,
		RoleClient: roleClient,
		Redis:      redis,
		Mongodb:    mongodb,
	}
}

// GetAuthURL generates Google OAuth URL
func (s *AuthService) GetAuthURL() string {
	authService := auth.NewService(s.Config, s.Redis, s.Mongodb, s.UserClient)
	return authService.GetAuthURL("")
}

// HandleGoogleCallback processes Google OAuth callback
func (s *AuthService) HandleGoogleCallback(ctx context.Context, req model.GoogleCallbackRequest, userAgent, ipAddress string) (*model.TokenResponse, error) {
	authService := auth.NewService(s.Config, s.Redis, s.Mongodb, s.UserClient)

	// Exchange code for user info
	googleUser, err := authService.ExchangeCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	var ids string

	switch req.Role {
	case "student":
		user, err := s.UserClient.GetUserByEmail(ctx, googleUser.Email)
		if err != nil || user == nil || len(user.GetStudents()) == 0 {
			return nil, fmt.Errorf("failed to get student: %w", err)
		}
		for _, student := range user.GetStudents() {
			ids += student.GetSemesterCode() + ":" + student.GetId() + ","
		}

	case "teacher":
		teachers, err := s.UserClient.GetTeacherByEmail(ctx, googleUser.Email)
		if err != nil || teachers == nil || len(teachers.GetTeachers()) == 0 {
			return nil, fmt.Errorf("failed to get teacher: %w", err)
		}
		for _, teacher := range teachers.GetTeachers() {
			ids += teacher.GetSemesterCode() + ":" + teacher.GetId() + ","
		}

	default:
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	// Generate token pair
	tokenPair, err := authService.GenerateTokenPair(ctx, ids, req.Role, googleUser, userAgent, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &model.TokenResponse{
		GoogleUser:   googleUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	}, nil
}

// RefreshToken refreshes access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, error) {
	authService := auth.NewService(s.Config, s.Redis, s.Mongodb, s.UserClient)

	tokenPair, err := authService.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	return &model.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	}, nil
}

// Logout invalidates the refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	authService := auth.NewService(s.Config, s.Redis, s.Mongodb, s.UserClient)
	return authService.Logout(ctx, refreshToken)
}

// ExtractUserInfo extracts user information from gin context
func (s *AuthService) ExtractUserInfo(c *gin.Context) (*model.UserInfo, error) {
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

	// Get semester from header
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

	return &model.UserInfo{
		Role:     role,
		Semester: semester,
		UserID:   myID,
		IDs:      idsArr,
	}, nil
}

// ExtractRole gets user roles from role service
func (s *AuthService) ExtractRole(ctx context.Context, userID string) ([]pbRole.RoleType, error) {
	if s.RoleClient == nil {
		return nil, fmt.Errorf("role client not initialized")
	}

	search := &pbCommon.SearchRequest{
		Filters: []*pbCommon.FilterCriteria{
			{
				Criteria: &pbCommon.FilterCriteria_Condition{
					Condition: &pbCommon.FilterCondition{
						Field:    "teacher_code",
						Operator: pbCommon.FilterOperator_EQUAL,
						Values:   []string{userID},
					},
				},
			},
		},
	}

	roles, err := s.RoleClient.GetRoleBySearch(ctx, search)
	if err != nil {
		return nil, err
	}

	if roles == nil || roles.RoleSystems == nil {
		return nil, fmt.Errorf("role not found")
	}

	result := make([]pbRole.RoleType, 0, len(roles.RoleSystems))
	for _, role := range roles.RoleSystems {
		result = append(result, role.Role)
	}

	return result, nil
}
