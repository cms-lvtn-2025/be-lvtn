package router

import (
	"context"
	"time"

	"thaily/src/api"
	"thaily/src/config"
	"thaily/src/graph/controller"
	"thaily/src/graph/dataloader"
	"thaily/src/graph/generated"
	"thaily/src/graph/helper"
	"thaily/src/graph/resolver"
	"thaily/src/pkg/container"

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
	setupCORS(r)

	// Setup GraphQL
	setupGraphQL(r, c)

	// Setup REST API
	setupRestAPI(r, c)

	return r
}

func setupCORS(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func setupGraphQL(r *gin.Engine, c *container.Container) {
	// Create controller với tất cả clients
	ctrl := controller.NewController(
		c.Clients.Academic,
		c.Clients.Council,
		c.Clients.File,
		c.Clients.Role,
		c.Clients.Thesis,
		c.Clients.User,
	)

	// Create GraphQL handler
	srv := handler.New(generated.NewExecutableSchema(
		generated.Config{Resolvers: &resolver.Resolver{Ctrl: ctrl}},
	))

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
		dataloaderMiddleware(c), // Inject dataloaders first
		//graphqlAuthMiddleware(c.Config.JWT, ctrl), // Then handle auth
		gin.WrapH(srv))
}

func setupRestAPI(r *gin.Engine, c *container.Container) {
	// Create API handler với clients cần thiết
	apiHandler := api.NewAPIHandler(
		api.WithConfig(c.Config),
		api.WithUserClient(c.Clients.User),
		api.WithFileClient(c.Clients.File),
		api.WithAcademicClient(c.Clients.Academic),
		api.WithRedisClient(c.Clients.Redis),
		api.WithMongoClient(c.Clients.MongoDB),
		api.WithMimIo(c.Clients.MinIO),
	)

	// Register routes
	apiV1 := r.Group("/api/v1")
	apiHandler.RegisterRoutes(apiV1)
}

// dataloaderMiddleware injects dataloaders into the context
func dataloaderMiddleware(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Create new loaders for each request
		loaders := dataloader.NewLoaders(
			c.Clients.User,
			c.Clients.Thesis,
			c.Clients.Council,
		)

		// Inject loaders into context
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
