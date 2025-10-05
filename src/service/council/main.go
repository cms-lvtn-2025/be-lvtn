package main

import (
	"log"
	"net"
	"thaily/src/pkg/config"
	"thaily/src/pkg/database"
	"thaily/src/service/council/handler"
	pb "thaily/proto/council"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.Load("env/council.env")
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
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCouncilServiceServer(grpcServer, h)

	log.Println("CouncilService running on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
