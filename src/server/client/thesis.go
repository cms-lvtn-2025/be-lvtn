package client

import (
	"context"
	"fmt"
	pbCommon "thaily/proto/common"
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

func (t *GRPCthesis) GetTopicByStudentCode(ctx context.Context, studentCode string, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	return t.client.ListTopics(ctx, &pb.ListTopicsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: order,
				Page:       page,
				PageSize:   pageSize,
				SortBy:     sortBy,
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "studentCode",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{studentCode},
						},
					},
				},
			},
		},
	})
}

func (t *GRPCthesis) GetTopicByTeacherCode(ctx context.Context, teacerCode string, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	return t.client.ListTopics(ctx, &pb.ListTopicsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: order,
				Page:       page,
				PageSize:   pageSize,
				SortBy:     sortBy,
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "teacher_supervisor_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{teacerCode},
						},
					},
				},
			},
		},
	})
}

func (t *GRPCthesis) GetEnrolmentByTopicCodeAndStudentCode(ctx context.Context, topicCode string, studentCode string) (*pb.ListEnrollmentsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	return t.client.ListEnrollments(ctx, &pb.ListEnrollmentsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   1000,
				SortBy:     "created_at",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "student_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{studentCode},
						},
					},
				}, {
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "topic_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{topicCode},
						},
					},
				},
			},
		},
	})
}

func (t *GRPCthesis) GetEnrollmentByTopicCode(ctx context.Context, topicCode string) (*pb.ListEnrollmentsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	return t.client.ListEnrollments(ctx, &pb.ListEnrollmentsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   1000,
				SortBy:     "created_at",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "student_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{topicCode},
						},
					},
				},
			},
		},
	})
}
