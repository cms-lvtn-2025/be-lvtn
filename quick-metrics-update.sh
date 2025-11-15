#!/bin/bash

# Script to quickly update Council and Role services with metrics

echo "🚀 Updating Council Service..."

# Update Council main.go
cat > src/service/council/main-metrics.go << 'EOF'
package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"thaily/src/service/pkg/database"
	logger2 "thaily/src/service/pkg/logger"
	"thaily/src/service/pkg/metrics"
	"thaily/src/service/pkg/tls"

	pb "thaily/proto/council"
	"thaily/src/service/council/handler"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load("./council.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}
	
	metrics.SetServiceHealth("council", true)

	if err := logger2.InitFileLogger("council-service", "log"); err != nil {
		log.Fatalf("Failed to initialize file logger: %v", err)
	}
	defer logger2.GetFileLogger().Close()

	if err := database.InitDB(); err != nil {
		metrics.SetServiceHealth("council", false)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9092"
	}
	go func() {
		log.Printf("Metrics server starting on port %s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				metrics.Handler().ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		})); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	if err := tls.VerifyCertificatesExist("council"); err != nil {
		log.Fatalf("TLS certificate verification failed: %v", err)
	}

	creds, err := tls.LoadServerTLSCredentials("council")
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			logger2.UnaryServerInterceptor(),
			metrics.UnaryServerInterceptor("council"),
		),
	)

	h := handler.NewHandler(database.GetDB())
	pb.RegisterCouncilServiceServer(grpcServer, h)

	log.Printf("CouncilService listening on port %s", port)
	log.Printf("Metrics available at http://localhost:%s/metrics", metricsPort)
	if err := grpcServer.Serve(lis); err != nil {
		metrics.SetServiceHealth("council", false)
		log.Fatalf("Failed to serve: %v", err)
	}
}
EOF

# Update Role main.go
cat > src/service/role/main-metrics.go << 'EOF'
package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"thaily/src/service/pkg/database"
	logger2 "thaily/src/service/pkg/logger"
	"thaily/src/service/pkg/metrics"
	"thaily/src/service/pkg/tls"

	pb "thaily/proto/role"
	"thaily/src/service/role/handler"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load("./role.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}
	
	metrics.SetServiceHealth("role", true)

	if err := logger2.InitFileLogger("role-service", "log"); err != nil {
		log.Fatalf("Failed to initialize file logger: %v", err)
	}
	defer logger2.GetFileLogger().Close()

	if err := database.InitDB(); err != nil {
		metrics.SetServiceHealth("role", false)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9094"
	}
	go func() {
		log.Printf("Metrics server starting on port %s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				metrics.Handler().ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		})); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	if err := tls.VerifyCertificatesExist("role"); err != nil {
		log.Fatalf("TLS certificate verification failed: %v", err)
	}

	creds, err := tls.LoadServerTLSCredentials("role")
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "50054"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			logger2.UnaryServerInterceptor(),
			metrics.UnaryServerInterceptor("role"),
		),
	)

	h := handler.NewHandler(database.GetDB())
	pb.RegisterRoleServiceServer(grpcServer, h)

	log.Printf("RoleService listening on port %s", port)
	log.Printf("Metrics available at http://localhost:%s/metrics", metricsPort)
	if err := grpcServer.Serve(lis); err != nil {
		metrics.SetServiceHealth("role", false)
		log.Fatalf("Failed to serve: %v", err)
	}
}
EOF

echo "✅ Created main-metrics.go files for Council and Role services"
echo "📝 Please manually copy content to replace original main.go files"
echo "🔧 Also update handler.go files to use metrics.DatabaseHandler"