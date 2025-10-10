package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	pb "thaily/proto/user"
	"time"

	"thaily/src/config"
	"thaily/src/server/client"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	GoogleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	RedisKeyPrefix    = "session:"
)

type Service struct {
	config      *config.Config
	redis       *client.RedisClient
	mongodb     *client.MongoClient
	user        *client.GRPCUser
	oauthConfig *oauth2.Config
}

// NewService tạo auth service mới
func NewService(cfg *config.Config, redis *client.RedisClient, mongodb *client.MongoClient, user *client.GRPCUser) *Service {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &Service{
		config:      cfg,
		redis:       redis,
		mongodb:     mongodb,
		oauthConfig: oauthConfig,
		user:        user,
	}
}

// GetAuthURL tạo URL để redirect user đến Google login
func (s *Service) GetAuthURL(state string) string {
	if state == "" {
		state = s.generateState()
	}
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode đổi authorization code lấy token và user info
func (s *Service) ExchangeCode(ctx context.Context, code string) (*GoogleUserInfo, error) {
	// Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	userInfo, err := s.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo, nil
}

// GenerateTokenPair tạo access token và refresh token
// NOTE: User data sẽ được xử lý ở service user sau, giờ chỉ cần lưu session
func (s *Service) GenerateTokenPair(ctx context.Context, user *pb.ListStudentsResponse, googleUser *GoogleUserInfo, userAgent, ipAddress string) (*TokenPair, error) {
	// Tạo access token (JWT) với email từ Google
	ids := ""
	for _, student := range user.GetStudents() {
		ids += student.GetSemesterCode() + "-" + student.GetId() + ","
	}

	accessToken, err := s.createAccessToken(googleUser.Email, googleUser.Name, googleUser.ID, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Tạo refresh token
	refreshToken := s.generateRefreshToken()

	// Lưu session vào MongoDB (chỉ lưu session info, không lưu user)
	session := Session{
		UserID:       googleUser.ID, // Google ID tạm thời, sau sẽ được map với user service
		Email:        googleUser.Email,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(time.Duration(s.config.JWT.RefreshTokenExpiry) * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	collection := s.mongodb.GetCollection("sessions")
	result, err := collection.InsertOne(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	sessionID := result.InsertedID.(primitive.ObjectID)

	// Lưu refresh token vào Redis với TTL
	claims := RefreshTokenClaims{
		UserID:    googleUser.ID,
		SessionID: sessionID.Hex(),
		ExpiresAt: session.ExpiresAt,
	}

	claimsJSON, _ := json.Marshal(claims)
	redisKey := RedisKeyPrefix + refreshToken
	ttl := time.Duration(s.config.JWT.RefreshTokenExpiry) * 24 * time.Hour

	if err := s.redis.Set(ctx, redisKey, claimsJSON, ttl); err != nil {
		return nil, fmt.Errorf("failed to save refresh token to redis: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.config.JWT.AccessTokenExpiry * 60, // convert minutes to seconds
		TokenType:    "Bearer",
	}, nil
}

// RefreshAccessToken làm mới access token bằng refresh token
func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Check if user client is initialized
	if s.user == nil {
		return nil, fmt.Errorf("user service client is not initialized")
	}

	// Lấy claims từ Redis
	redisKey := RedisKeyPrefix + refreshToken
	claimsJSON, err := s.redis.Get(ctx, redisKey)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	var claims RefreshTokenClaims
	if err := json.Unmarshal([]byte(claimsJSON), &claims); err != nil {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Kiểm tra expiry
	if time.Now().After(claims.ExpiresAt) {
		s.redis.Del(ctx, redisKey)
		return nil, fmt.Errorf("refresh token expired")
	}

	// Lấy session từ MongoDB để có email
	sessionID, err := primitive.ObjectIDFromHex(claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session id")
	}

	var session Session
	collection := s.mongodb.GetCollection("sessions")
	if err := collection.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session); err != nil {
		return nil, fmt.Errorf("session not found")
	}

	user, err := s.user.GetUserByEmail(ctx, session.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid user email")
	}
	ids := ""
	for _, student := range user.GetStudents() {
		ids += student.GetSemesterCode() + "-" + student.GetId() + ","
	}

	// Tạo access token mới
	accessToken, err := s.createAccessToken(session.Email, "", claims.UserID, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Giữ nguyên refresh token
		ExpiresIn:    s.config.JWT.AccessTokenExpiry * 60,
		TokenType:    "Bearer",
	}, nil
}

// Logout xóa session và refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	// Xóa từ Redis
	redisKey := RedisKeyPrefix + refreshToken
	claimsJSON, err := s.redis.Get(ctx, redisKey)
	if err == nil {
		var claims RefreshTokenClaims
		if err := json.Unmarshal([]byte(claimsJSON), &claims); err == nil {
			// Xóa session từ MongoDB
			sessionID, _ := primitive.ObjectIDFromHex(claims.SessionID)
			collection := s.mongodb.GetCollection("sessions")
			collection.DeleteOne(ctx, bson.M{"_id": sessionID})
		}
	}

	// Xóa refresh token từ Redis
	s.redis.Del(ctx, redisKey)

	return nil
}

// Private helper methods

func (s *Service) createAccessToken(email, name, googleID string, ids string) (string, error) {
	claims := jwt.MapClaims{
		"ids":       ids,
		"email":     email,
		"name":      name,
		"google_id": googleID,
		"exp":       time.Now().Add(time.Duration(s.config.JWT.AccessTokenExpiry) * time.Minute).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.AccessSecret))
}

func (s *Service) generateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *Service) generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *Service) getUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", GoogleUserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
