package client

import (
	pb "thaily/proto/thesis"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCthesis struct {
	conn   *grpc.ClientConn
	client pb.ThesisServiceClient
}

func NewGRPCthesis(addr string) (*GRPCthesis, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewThesisServiceClient(conn)
	return &GRPCthesis{conn: conn, client: client}, nil
}
