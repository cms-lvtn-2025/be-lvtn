package client

import (
	pb "thaily/proto/file"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCfile struct {
	conn   *grpc.ClientConn
	client pb.FileServiceClient
}

func NewGRPCfile(addr string) (*GRPCfile, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewFileServiceClient(conn)
	return &GRPCfile{conn: conn, client: client}, nil
}
