package client

import (
	pb "thaily/proto/council"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCCouncil struct {
	conn        *grpc.ClientConn
	client      pb.CouncilServiceClient
	redisClient *redis.Client
}

func NewGRPCCouncil(addr string, redis *redis.Client) (*GRPCCouncil, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewCouncilServiceClient(conn)
	return &GRPCCouncil{conn: conn, client: client}, nil
}
