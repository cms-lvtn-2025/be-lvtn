package main

import (
	"log"
	"net"
	pb "thaily/proto/role"
	"thaily/src/pkg/config"
	"thaily/src/pkg/database"
	"thaily/src/pkg/logger"
	"thaily/src/service/role/handler"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.Load("../../../env/role.env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize file logger
	if err := logger.InitFileLogger("role-service", "log"); err != nil {
		log.Fatalf("Failed to initialize file logger: %v", err)
	}
	defer logger.GetFileLogger().Close()

	// Connect to database
	db, err := database.Connect(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create handler with database connection
	h := handler.NewHandler(db)

	// Setup gRPC server
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryServerInterceptor()),
	)
	pb.RegisterRoleServiceServer(grpcServer, h)

	log.Println("RoleService running on :50054")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
