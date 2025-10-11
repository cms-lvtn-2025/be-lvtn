package client

import (
	"context"
	"fmt"
	pbCommon "thaily/proto/common"
	pb "thaily/proto/role"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCRole struct {
	conn        *grpc.ClientConn
	client      pb.RoleServiceClient
	redisClient *redis.Client
}

func NewGRPCRole(addr string, redis *redis.Client) (*GRPCRole, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewRoleServiceClient(conn)
	return &GRPCRole{conn: conn, client: client}, nil
}

func (r *GRPCRole) GetAllRoleByTeacherId(ctx context.Context, teacherId string) (*pb.ListRoleSystemsResponse, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("GRPCRole is nil")
	}
	return r.client.ListRoleSystems(ctx, &pb.ListRoleSystemsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: nil,
			Filters: []*pbCommon.FilterCriteria{

				{Criteria: &pbCommon.FilterCriteria_Condition{
					Condition: &pbCommon.FilterCondition{
						Field:    "teacher_code",
						Operator: pbCommon.FilterOperator_EQUAL,
						Values:   []string{teacherId},
					},
				},
				},
			},
		},
	})

}
