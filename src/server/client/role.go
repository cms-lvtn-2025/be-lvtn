package client

import (
	pb "thaily/proto/role"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCRole struct {
	conn   *grpc.ClientConn
	client pb.RoleServiceClient
}

func NewGRPCRole(addr string) (*GRPCRole, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewRoleServiceClient(conn)
	return &GRPCRole{conn: conn, client: client}, nil
}
