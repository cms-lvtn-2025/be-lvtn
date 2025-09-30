package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "thaily/proto/common"
	resolver "thaily/src/service/common/resolvers"
	connect "thaily/src/service/database"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	var (
		port   = flag.String("port", "50051", "The server port")
		user   = flag.String("mysql-user", getEnv("MYSQL_USER", "root"), "mysqlDB connection user")
		passwd = flag.String("mysql-passwd", getEnv("MYSQL_PASSWD", "Th@i2004"), "mysqlDB connection passwd")
		host   = flag.String("mysql-host", getEnv("MYSQL_HOST", "localhost"), "mysqlDB connection host")
		portdb = flag.String("mysql-port", getEnv("MYSQL_PORT", "10001"), "mysqlDB connection port")

		mysqldb = flag.String("mysql-db", getEnv("MYSQL_DB", "thesis_management_system"), "mysqlDB database")
	)
	flag.Parse()

	mySql := connect.GetInstanceDB(*user, *passwd, *host, *portdb, *mysqldb)
	defer connect.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	service := resolver.NewCommonService(mySql)
	pb.RegisterCommonServiceServer(grpcServer, service)

	reflection.Register(grpcServer)

	log.Printf("Common service is running on port %s", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
