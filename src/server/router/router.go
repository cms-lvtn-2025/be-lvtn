package router

import (
	"context"
	"net/http"
	"time"

	"thaily/src/server/api"
	"thaily/src/server/config"
	dataloader2 "thaily/src/server/graph/dataloader"
	"thaily/src/service/pkg/container"
	"thaily/src/service/pkg/metrics"

	"thaily/src/server/graph/controller"
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vektah/gqlparser/v2/ast"
)

var serverStartTime = time.Now()

// Setup khởi tạo và cấu hình router
func Setup(cfg *config.Config, c *container.Container) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.Default()

	// Setup CORS
	setupCORS(r)

	// System endpoints (health, metrics)
	setupSystemRoutes(r, c)

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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "x-semester"},
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
		graphqlAuthMiddleware(c.Config, ctrl), // Then handle auth

		dataloaderMiddleware(c), // Inject dataloaders first
		gin.WrapH(srv))
}

func setupSystemRoutes(r *gin.Engine, c *container.Container) {
	r.GET("/health", func(ctx *gin.Context) {
		uptime := time.Since(serverStartTime).Round(time.Second)
		metrics.SetServiceHealth("api-gateway", true)
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"uptime":  uptime.String(),
			"service": "api-gateway",
		})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
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
		loaders := dataloader2.NewLoaders(
			c.Clients.User,
			c.Clients.Thesis,
			c.Clients.Council,
			c.Clients.Academic,
			c.Clients.Role,
			c.Clients.File,
		)

		// Inject loaders into context
		requestCtx := dataloader.WithLoaders(ctx.Request.Context(), loaders)
		ctx.Request = ctx.Request.WithContext(requestCtx)
		ctx.Next()
	}
}

// graphqlAuthMiddleware xử lý authentication cho GraphQL và inject dataloaders
func graphqlAuthMiddleware(cfg *config.Config, ctrl *controller.Controller) gin.HandlerFunc {
	return func(c *gin.Context) {
		semester := c.GetHeader("x-semester")
		if cfg.Server.DisableGraphQLAuth {
			defaultClaims := jwt.MapClaims{
				"role": "teacher",
				"ids":  "test-semester-teacher",
			}
			ctx := context.WithValue(c.Request.Context(), helper.Auth, defaultClaims)
			ctx = context.WithValue(ctx, "semester", semester)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		claims, err := helper.ValidateAndParseClaims(authHeader, cfg.JWT.AccessSecret)
		if err != nil {
			c.JSON(401, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), helper.Auth, claims)
		ctx = context.WithValue(ctx, "semester", semester)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
