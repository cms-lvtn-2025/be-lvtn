package main

import (
	"log"
	"net"
	"os"

	pb "thaily/proto/role"
	"thaily/src/pkg/database"
	"thaily/src/pkg/logger"
	"thaily/src/service/role/handler"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("env/role.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize file logger
	if err := logger.InitFileLogger("role-service", "log"); err != nil {
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
		port = "50054"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		logger.UnaryServerInterceptor(),
	)

	h := handler.NewHandler()
	pb.RegisterRoleServiceServer(grpcServer, h)

	log.Printf("RoleService listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
