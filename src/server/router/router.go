package router

import (
	"context"
	"thaily/src/server/api"
	"thaily/src/server/config"
	"thaily/src/server/graph/controller"
	dataloader2 "thaily/src/server/graph/dataloader"
	"thaily/src/server/graph/directive"
	"thaily/src/service/pkg/container"
	"time"

	"thaily/src/server/graph/dataloader"
	"thaily/src/server/graph/generated"
	"thaily/src/server/graph/helper"
	"thaily/src/server/graph/resolver"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
)

// Setup khởi tạo và cấu hình router
func Setup(cfg *config.Config, c *container.Container) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.Default()

	// Setup CORS
	setupCORS(r, cfg.CORS.AllowOrigins)

	// Setup GraphQL
	setupGraphQL(r, c)

	// Setup REST API
	setupRestAPI(r, c)

	return r
}

func setupCORS(r *gin.Engine, allowOrigins []string) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "x-semester", "x-role"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func setupGraphQL(r *gin.Engine, c *container.Container) {
	// Create SINGLETON loaders - shared across all requests
	// This prevents goroutine leaks from creating new loaders per request
	sharedLoaders := dataloader2.NewLoaders(
		c.Clients.User,
		c.Clients.Thesis,
		c.Clients.Council,
		c.Clients.Academic,
		c.Clients.Role,
		c.Clients.File,
	)

	// Create controller với tất cả clients
	ctrl := controller.NewController(
		c.Clients.Academic,
		c.Clients.Council,
		c.Clients.File,
		c.Clients.Role,
		c.Clients.Thesis,
		c.Clients.User,
	)

	// Create GraphQL handler with directive
	cfg := generated.Config{
		Resolvers: &resolver.Resolver{Ctrl: ctrl},
		Directives: generated.DirectiveRoot{
			RbacRole: directive.RbacRoleDirective,
		},
	}
	srv := handler.New(generated.NewExecutableSchema(cfg))

	// Add field middleware for RBAC propagation
	srv.AroundFields(directive.RbacFieldMiddleware)

	// Configure transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{})

	// Configure cache and extensions
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Routes
	r.GET("/", gin.WrapH(playground.Handler("GraphQL Playground", "/query")))
	r.Any("/query",
		graphqlAuthMiddleware(c.Config.JWT, ctrl), // Then handle auth
		dataloaderMiddleware(sharedLoaders),       // Inject shared dataloaders
		gin.WrapH(srv))
}

func setupRestAPI(r *gin.Engine, c *container.Container) {
	// Create router with all clients
	router := api.SetupRouterWithClients(
		c.Config,
		c.Clients.User,
		c.Clients.Role,
		c.Clients.File,
		c.Clients.Thesis,
		c.Clients.Redis,
		c.Clients.MongoDB,
		c.Clients.MinIO,
	)

	// Register routes
	apiV1 := r.Group("/api/v1")
	router.RegisterRoutes(apiV1)
}

// dataloaderMiddleware injects shared dataloaders into the context
func dataloaderMiddleware(loaders *dataloader2.Loaders) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Use shared loaders instead of creating new ones per request
		// This prevents goroutine leaks from cleanup goroutines
		requestCtx := dataloader.WithLoaders(ctx.Request.Context(), loaders)
		ctx.Request = ctx.Request.WithContext(requestCtx)
		ctx.Next()
	}
}

// graphqlAuthMiddleware xử lý authentication cho GraphQL và inject dataloaders
func graphqlAuthMiddleware(cfg config.JWTConfig, ctrl *controller.Controller) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		semester := c.GetHeader("x-semester")
		claims, err := helper.ValidateAndParseClaims(authHeader, cfg.AccessSecret)
		if err != nil {
			c.JSON(401, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		// Set claims vào context cho GraphQL resolver
		ctx := context.WithValue(c.Request.Context(), helper.Auth, claims)
		ctx = context.WithValue(ctx, "semester", semester)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
