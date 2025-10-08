package client

import (
	pb "thaily/proto/academic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCAcadamicClient struct {
	conn   *grpc.ClientConn
	client pb.AcademicServiceClient
}

func NewGRPCAcadamicClient(addr string) (*GRPCAcadamicClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewAcademicServiceClient(conn)
	return &GRPCAcadamicClient{conn: conn, client: client}, nil
}
