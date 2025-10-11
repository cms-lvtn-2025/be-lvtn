package client

import (
	"context"
	"fmt"
	pbCommon "thaily/proto/common"
	pb "thaily/proto/user"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCUser struct {
	conn        *grpc.ClientConn
	client      pb.UserServiceClient
	redisClient *redis.Client
}

func NewGRPCUser(addr string) (*GRPCUser, error) {
	fmt.Println(addr)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewUserServiceClient(conn)
	return &GRPCUser{conn: conn, client: client}, nil
}

func (u *GRPCUser) GetUserByEmail(ctx context.Context, email string) (*pb.ListStudentsResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	return u.client.ListStudents(ctx, &pb.ListStudentsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   20,
				SortBy:     "semester_code",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "email",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{email},
						},
					},
				},
			},
		},
	})
}

func (u *GRPCUser) GetUserByEmailAndSemester(ctx context.Context, email string, semester string) (*pb.ListStudentsResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	return u.client.ListStudents(ctx, &pb.ListStudentsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   20,
				SortBy:     "semester_code",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "email",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{email},
						},
					},
				},
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "semester_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{semester},
						},
					}},
			},
		},
	})
}

func (u *GRPCUser) GetUserById(ctx context.Context, id string) (*pb.GetStudentResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	return u.client.GetStudent(ctx, &pb.GetStudentRequest{
		Id: id,
	})
}

func (u *GRPCUser) GetTeacherById(ctx context.Context, id string) (*pb.GetTeacherResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	return u.client.GetTeacher(ctx, &pb.GetTeacherRequest{
		Id: id,
	})
}

func (u *GRPCUser) GetTeacherByEmail(ctx context.Context, email string) (*pb.ListTeachersResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	return u.client.ListTeachers(ctx, &pb.ListTeachersRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   20,
				SortBy:     "semester_code",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "email",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{email},
						},
					},
				},
			},
		},
	})
}
