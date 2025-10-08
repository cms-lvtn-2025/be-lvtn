package client

import (
	pb "thaily/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCUser struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewGRPCUser(addr string) (*GRPCUser, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewUserServiceClient(conn)
	return &GRPCUser{conn: conn, client: client}, nil
}
