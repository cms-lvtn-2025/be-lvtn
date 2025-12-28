package api

import (
	"github.com/gin-gonic/gin"

	"thaily/src/server/api/handler"
	"thaily/src/server/api/middleware"
	"thaily/src/server/client"
	"thaily/src/server/config"
)

// Router handles API route registration
type Router struct {
	handler *handler.Handler
	config  *config.Config
}

// NewRouter creates a new Router
func NewRouter(cfg *config.Config, h *handler.Handler) *Router {
	return &Router{
		handler: h,
		config:  cfg,
	}
}

// SetupRouterWithClients creates handler and router with all clients
func SetupRouterWithClients(
	cfg *config.Config,
	userClient *client.GRPCUser,
	roleClient *client.GRPCRole,
	fileClient *client.GRPCfile,
	thesisClient *client.GRPCthesis,
	redis *client.RedisClient,
	mongodb *client.MongoClient,
	minIO *client.ServiceMinIo,
) *Router {
	h := handler.NewHandler(
		handler.WithConfig(cfg),
		handler.WithServices(
			cfg,
			userClient,
			roleClient,
			fileClient,
			thesisClient,
			redis,
			mongodb,
			minIO,
		),
	)

	return NewRouter(cfg, h)
}

// RegisterRoutes registers all REST API routes
func (r *Router) RegisterRoutes(g *gin.RouterGroup) {
	// Auth routes - no authentication required
	auth := g.Group("/auth")
	{
		auth.POST("/google/login", r.handler.GoogleLogin)
		auth.POST("/google/callback", r.handler.GoogleCallback)
		auth.POST("/refresh", r.handler.RefreshToken)
		auth.POST("/logout", r.handler.Logout)
	}

	// File routes - authentication required
	files := g.Group("/files")
	{
		// Upload endpoints
		files.POST("/upload/grade-defence", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadGradeDefenceFile)
		files.POST("/upload/topic-council-for-department", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadTopicCouncilForDepartmentFile)
		files.POST("/upload/council-for-department", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadCouncilForDepartmentFile)
		files.POST("/upload/council-for-affair", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadCouncilForAffairFile)
		files.POST("/upload/final", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadFinalFile)
		files.POST("/upload/midterm", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadMidtermFile)
		files.POST("/upload/student-for-affair", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadUserForAffairFile)
		files.POST("/upload/teacher-for-affair", middleware.AuthMiddleware(r.config.JWT), r.handler.UploadTeacherForAffairFile)

		// Get presigned download URL
		files.GET("/:id/url", middleware.AuthMiddleware(r.config.JWT), r.handler.GetFileURL)

		// Delete file
		files.DELETE("/:id", middleware.AuthMiddleware(r.config.JWT), r.handler.DeleteFile)

		// List excel files
		files.GET("/list/excel", middleware.AuthMiddleware(r.config.JWT), r.handler.ListFilesExcel)
	}
}
