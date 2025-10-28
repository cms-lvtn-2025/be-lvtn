package main

import (
	"log"
	"net"
	"os"
	"thaily/src/service/pkg/database"
	logger2 "thaily/src/service/pkg/logger"
	"thaily/src/service/pkg/tls"

	pb "thaily/proto/file"
	"thaily/src/service/file/handler"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("./file.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize file logger
	if err := logger2.InitFileLogger("file-service", "log"); err != nil {
		log.Fatalf("Failed to initialize file logger: %v", err)
	}
	defer logger2.GetFileLogger().Close()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Verify TLS certificates exist
	if err := tls.VerifyCertificatesExist("file"); err != nil {
		log.Fatalf("TLS certificate verification failed: %v", err)
	}

	// Load TLS credentials
	creds, err := tls.LoadServerTLSCredentials("file")
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}

	// Start gRPC server
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "50053"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(logger2.UnaryServerInterceptor()),
	)

	h := handler.NewHandler(database.GetDB())
	pb.RegisterFileServiceServer(grpcServer, h)

	log.Printf("FileService listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
