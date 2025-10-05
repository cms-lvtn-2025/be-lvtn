package main

import (
	"log"
	"net"
	"thaily/src/pkg/config"
	"thaily/src/pkg/database"
	"thaily/src/service/file/handler"
	pb "thaily/proto/file"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.Load("env/file.env")
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
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, h)

	log.Println("FileService running on :50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
