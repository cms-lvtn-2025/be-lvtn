package api

import (
	"thaily/src/server/client"
	"thaily/src/server/config"

	"github.com/gin-gonic/gin"
)

// APIHandler chứa các clients cần thiết cho REST API
type APIHandler struct {
	Config         *config.Config
	UserClient     *client.GRPCUser
	AcademicClient *client.GRPCAcadamicClient
	FileClient     *client.GRPCfile
	Redis          *client.RedisClient
	RoleClient     *client.GRPCRole
	Mongodb        *client.MongoClient
	MimIo          *client.ServiceMinIo
	ThesisClient   *client.GRPCthesis
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

// WithRoleClient inject role client
func WithRoleClient(client *client.GRPCRole) ClientOption {
	return func(h *APIHandler) {
		h.RoleClient = client
	}
}

// WithFileClient inject file client
func WithFileClient(client *client.GRPCfile) ClientOption {
	return func(h *APIHandler) {
		h.FileClient = client
	}
}

// WithRedisClient inject redis client
func WithRedisClient(client *client.RedisClient) ClientOption {
	return func(h *APIHandler) {
		h.Redis = client
	}
}

// WithMongoClient inject mongo client
func WithMongoClient(client *client.MongoClient) ClientOption {
	return func(h *APIHandler) {
		h.Mongodb = client
	}
}

// WithMimIo inject MinIO client
func WithMimIo(client *client.ServiceMinIo) ClientOption {
	return func(h *APIHandler) {
		h.MimIo = client
	}
}

// WithConfig inject config
func WithConfig(cfg *config.Config) ClientOption {
	return func(h *APIHandler) {
		h.Config = cfg
	}
}

func WithThesisClient(client *client.GRPCthesis) ClientOption {
	return func(h *APIHandler) {
		h.ThesisClient = client
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

	// File routes
	files := r.Group("/files")
	{
		// Upload endpoints - require authentication

		files.POST("/upload/grade-supervisor", AuthMiddleware(h.Config.JWT), h.UploadGradeSuppervisorFile)
		files.POST("/upload/grade-defence", AuthMiddleware(h.Config.JWT), h.UploadGradeDefenceFile)
		files.POST("/upload/topic-council-for-department", AuthMiddleware(h.Config.JWT), h.UploadTopicCouncilForDepartmentFile)
		files.POST("/upload/council-for-department", AuthMiddleware(h.Config.JWT), h.UploadCouncilForDepartmentFile)
		files.POST("/upload/council-for-affair", AuthMiddleware(h.Config.JWT), h.UploadCouncilForAffairFile)
		files.POST("/upload/final", AuthMiddleware(h.Config.JWT), h.UploadFinalFile)
		files.POST("/upload/midterm", AuthMiddleware(h.Config.JWT), h.UploadMidtermFile)
		files.POST("/upload/student-for-affair", AuthMiddleware(h.Config.JWT), h.UploadUserForAffairFile)
		files.POST("/upload/teacher-for-affair", AuthMiddleware(h.Config.JWT), h.UploadTeacherForAffairFile)
		// Get presigned download URL
		files.GET("/:id/url", AuthMiddleware(h.Config.JWT), h.GetFileURL)
		// Delete file
		files.DELETE("/:id", AuthMiddleware(h.Config.JWT), h.DeleteFile)
		// Public blob endpoint - uses token in query string
		files.GET("/list/excel", AuthMiddleware(h.Config.JWT), h.ListFilesExcel)
	}
}
