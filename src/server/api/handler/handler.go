package handler

import (
	"thaily/src/server/api/service"
	"thaily/src/server/client"
	"thaily/src/server/config"
)

// Handler contains all API handlers and their dependencies
type Handler struct {
	Config      *config.Config
	AuthService *service.AuthService
	FileService *service.FileService
}

// HandlerOption is function type for injecting dependencies
type HandlerOption func(*Handler)

// NewHandler creates a new Handler with options
func NewHandler(opts ...HandlerOption) *Handler {
	h := &Handler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// WithConfig injects config
func WithConfig(cfg *config.Config) HandlerOption {
	return func(h *Handler) {
		h.Config = cfg
	}
}

// WithServices injects all services with required clients
func WithServices(
	cfg *config.Config,
	userClient *client.GRPCUser,
	roleClient *client.GRPCRole,
	fileClient *client.GRPCfile,
	thesisClient *client.GRPCthesis,
	redis *client.RedisClient,
	mongodb *client.MongoClient,
	minIO *client.ServiceMinIo,
) HandlerOption {
	return func(h *Handler) {
		// Create auth service first
		authService := service.NewAuthService(cfg, userClient, roleClient, redis, mongodb)
		h.AuthService = authService

		// Create file service with auth service dependency
		h.FileService = service.NewFileService(
			cfg,
			fileClient,
			thesisClient,
			mongodb,
			minIO,
			redis,
			authService,
		)
	}
}
