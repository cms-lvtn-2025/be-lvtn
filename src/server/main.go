package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"thaily/src/graph/controller"
	"thaily/src/graph/generated"
	"thaily/src/graph/resolver"
	"thaily/src/helper"

	"thaily/src/server/client"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"

	"encoding/json"
	"io"
)

const defaultPort = "8081"

func main() {

	gin.SetMode(gin.ReleaseMode)

	// Tải biến môi trường
	godotenv.Load("env/server.env")

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Chỉ cho phép origin cụ thể
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Set port
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	serviceCommon := os.Getenv("SERVICE_COMMON")
	if serviceCommon == "" {
		log.Fatal("SERVICE_COMMON environment variable is not set")
	}
	log.Printf("Connecting to gRPC service at: %s", serviceCommon)

	common, err := client.NewGRPCCommonClient(serviceCommon)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	ctrl := controller.NewController(common)
	// Create GraphQL handler
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{Ctrl: ctrl}}))

	// Add GraphQL transports (Options, GET, POST)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{})
	// Set query cache
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Gin routes foyground and query handler
	r.GET("/", gin.WrapH(playground.Handler("GraphQL Playground", "/query")))

	// JWT Middleware with selective auth
	r.POST("/query", AuthMiddleware(), gin.WrapH(srv))

	// Run the server
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start the server:", err)
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// List of operations that don't require authentication
		publicOperations := map[string]bool{
			"login":              true,
			"semesters":          true,
			"introspectionquery": true, // Allow GraphQL introspection (lowercase)
		}

		// Read the request body to check operation name
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(400, gin.H{"message": "Failed to read request body"})
			c.Abort()
			return
		}

		// Restore body for the actual GraphQL handler
		c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

		// Parse GraphQL request to get operation name
		var gqlRequest struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}

		if err := json.Unmarshal(body, &gqlRequest); err != nil {
			c.JSON(400, gin.H{"message": "Invalid GraphQL request"})
			c.Abort()
			return
		}

		// Check if operation is public
		operationName := gqlRequest.OperationName

		// Extract from query if operation name not provided
		query := strings.TrimSpace(gqlRequest.Query)

		// Check for introspection first
		if strings.Contains(query, "__schema") || strings.Contains(query, "__type") || 
		   strings.Contains(query, "IntrospectionQuery") {
			operationName = "introspectionquery"
		} else if operationName == "" {
			// Try to extract operation name from query
			if strings.HasPrefix(query, "query") {
				// Format: query OperationName { ... }
				parts := strings.Fields(query)
				if len(parts) > 1 {
					operationName = strings.ToLower(parts[1])
				}
			} else if strings.Contains(query, "login") {
				operationName = "login"
			} else if strings.Contains(query, "semesters") {
				operationName = "semesters"
			}
		}

		// Convert to lowercase for comparison
		operationName = strings.ToLower(operationName)

		// If it's a public operation, allow without auth
		if publicOperations[operationName] {
			c.Next()
			return
		}

		// Otherwise, check for authentication
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"message": "Authorization header required"})
			c.Abort()
			return
		}

		if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
			Claims, err := helper.ParseJWT(token)
			if Claims == nil || err != nil {
				c.JSON(401, gin.H{"message": "Invalid Authorization header"})
				c.Abort()
				return
			}
			ctx := context.WithValue(c.Request.Context(), helper.Auth, Claims)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
