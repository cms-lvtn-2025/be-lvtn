package main

import (
	"log"
	"net"
	"os"

	pb "thaily/proto/user"
	"thaily/src/pkg/database"
	"thaily/src/pkg/logger"
	"thaily/src/service/user/handler"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("/home/thaily/code/heheheh_be/env/user.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize file logger
	if err := logger.InitFileLogger("user-service", "log"); err != nil {
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
		port = "50056"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryServerInterceptor()),
	)

	h := handler.NewHandler(database.GetDB())
	pb.RegisterUserServiceServer(grpcServer, h)

	log.Printf("UserService listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
