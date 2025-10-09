package api

import (
	"thaily/src/config"
	"thaily/src/server/client"

	"github.com/gin-gonic/gin"
)

// APIHandler chứa các clients cần thiết cho REST API
type APIHandler struct {
	Config         *config.Config
	UserClient     *client.GRPCUser
	AcademicClient *client.GRPCAcadamicClient
	FileClient     *client.GRPCfile
	Redis          *client.RedisClient
	Mongodb        *client.MongoClient
	MimIo          *client.ServiceMinIo
	// Thêm các client khác nếu cần
}

// NewAPIHandler tạo instance mới với các clients được inject
func NewAPIHandler(opts ...ClientOption) *APIHandler {
	h := &APIHandler{}

	// Apply các options để inject clients
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// ClientOption là function type để inject clients
type ClientOption func(*APIHandler)

// WithUserClient inject user client
func WithUserClient(client *client.GRPCUser) ClientOption {
	return func(h *APIHandler) {
		h.UserClient = client
	}
}

// WithAcademicClient inject academic client
func WithAcademicClient(client *client.GRPCAcadamicClient) ClientOption {
	return func(h *APIHandler) {
		h.AcademicClient = client
	}
}

// WithFileClient inject file client
func WithFileClient(client *client.GRPCfile) ClientOption {
	return func(h *APIHandler) {
		h.FileClient = client
	}
}

func WithRedisClient(client *client.RedisClient) ClientOption {
	return func(h *APIHandler) {
		h.Redis = client
	}

}

func WithMongoClient(client *client.MongoClient) ClientOption {
	return func(h *APIHandler) {
		h.Mongodb = client
	}
}

func WithMimIo(client *client.ServiceMinIo) ClientOption {
	return func(h *APIHandler) {
		h.MimIo = client
	}
}

func WithConfig(cfg *config.Config) ClientOption {
	return func(h *APIHandler) {
		h.Config = cfg
	}
}

// RegisterRoutes đăng ký các REST API routes
func (h *APIHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Auth routes - không cần authentication
	auth := r.Group("/auth")
	{
		auth.POST("/google/login", h.GoogleLogin)
		auth.POST("/google/callback", h.GoogleCallback)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
	}

	// User routes - cần authentication
	users := r.Group("/users")
	users.Use(AuthMiddleware())
	{
		users.GET("/me", h.GetCurrentUser)
		users.PUT("/profile", h.UpdateProfile)
	}

	// File routes
	files := r.Group("/files")
	{
		// Upload cần authentication
		files.POST("/upload", AuthMiddleware(), h.UploadFile)
		// Get file có thể public hoặc private tùy logic
		files.GET("/:id", OptionalAuthMiddleware(), h.GetFile)
	}
}
