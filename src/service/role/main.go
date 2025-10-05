package main

import (
	"log"
	"net"
	"thaily/src/pkg/config"
	"thaily/src/pkg/database"
	"thaily/src/service/role/handler"
	pb "thaily/proto/role"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.Load("env/role.env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	grpcServer := grpc.NewServer()
	pb.RegisterRoleServiceServer(grpcServer, h)

	log.Println("RoleService running on :50054")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
