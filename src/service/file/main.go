package main

import (
	"log"
	"net"
	"os"

	pb "thaily/proto/file"
	"thaily/src/pkg/database"
	"thaily/src/pkg/logger"
	"thaily/src/service/file/handler"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("env/file.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize file logger
	if err := logger.InitFileLogger("file-service", "log"); err != nil {
		log.Fatalf("Failed to initialize file logger: %v", err)
	}
	defer logger.GetFileLogger().Close()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

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
		logger.UnaryServerInterceptor(),
	)

	h := handler.NewHandler()
	pb.RegisterFileServiceServer(grpcServer, h)

	log.Printf("FileService listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
