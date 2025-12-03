package main

import (
	"log"
	"net"
	"os"
	"thaily/src/service/pkg/database"
	logger2 "thaily/src/service/pkg/logger"
	"thaily/src/service/pkg/metrics"
	"thaily/src/service/pkg/tls"

	pb "thaily/proto/user"
	"thaily/src/service/user/handler"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const serviceName = "user"

func main() {
	// Load environment variables
	if err := godotenv.Load("./user.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize file logger
	if err := logger2.InitFileLogger("user-service", "log"); err != nil {
		log.Fatalf("Failed to initialize file logger: %v", err)
	}
	defer logger2.GetFileLogger().Close()

	// Start metrics server
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9101"
	}
	metricsServer := metrics.NewMetricsServer(serviceName, metricsPort)
	if err := metricsServer.Start(); err != nil {
		log.Fatalf("Failed to start metrics server: %v", err)
	}
	defer metricsServer.Stop()
	log.Printf("Metrics server started on port %s", metricsPort)

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Verify TLS certificates exist
	if err := tls.VerifyCertificatesExist("user"); err != nil {
		log.Fatalf("TLS certificate verification failed: %v", err)
	}

	// Load TLS credentials
	creds, err := tls.LoadServerTLSCredentials("user")
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}

	// Start gRPC server
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "50056"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			logger2.UnaryServerInterceptor(),
			metrics.UnaryServerInterceptor(serviceName),
		),
	)

	h := handler.NewHandler(database.GetDB())
	pb.RegisterUserServiceServer(grpcServer, h)

	log.Printf("UserService listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
