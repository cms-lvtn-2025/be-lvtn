package client

import (
	"context"
	pb "thaily/proto/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCCommonClient struct {
	conn   *grpc.ClientConn
	client pb.CommonServiceClient
}

func NewGRPCCommonClient(addr string) (*GRPCCommonClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewCommonServiceClient(conn)
	return &GRPCCommonClient{conn: conn, client: client}, nil
}

func (c *GRPCCommonClient) CallProcerdure(ctx context.Context, procedure string) (*pb.ProcedureResponse, error) {
	in := &pb.ProcedureRequest{
		ProcedureQuery: procedure,
	}
	return c.client.Procedure(ctx, in)
}
