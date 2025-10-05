package main

import (
	"log"
	"net"
	"thaily/src/pkg/config"
	"thaily/src/pkg/database"
	"thaily/src/service/thesis/handler"
	pb "thaily/proto/thesis"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.Load("env/thesis.env")
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
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterThesisServiceServer(grpcServer, h)

	log.Println("ThesisService running on :50055")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
