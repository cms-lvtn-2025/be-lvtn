package main

import (
	"fmt"
	"log"

	"thaily/src/config"
	"thaily/src/pkg/container"
	"thaily/src/router"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize dependency container
	c, err := container.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Setup router
	r := router.Setup(cfg, c)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("GraphQL Playground: http://localhost%s/", addr)
	log.Printf("GraphQL Endpoint: http://localhost%s/query", addr)
	log.Printf("REST API: http://localhost%s/api/v1", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
