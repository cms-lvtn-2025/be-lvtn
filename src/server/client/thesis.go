package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/thesis"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCthesis struct {
	conn        *grpc.ClientConn
	client      pb.ThesisServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations
	topicCacheTTL      = 5 * time.Minute  // Topics change less frequently
	enrollmentCacheTTL = 2 * time.Minute  // Enrollments might change more often
	midtermCacheTTL    = 10 * time.Minute // Midterm results are more stable
	finalCacheTTL      = 10 * time.Minute // Final results are more stable

	// Cache key prefixes
	topicCachePrefix      = "thesis:topic:"
	enrollmentCachePrefix = "thesis:enrollment:"
	midtermCachePrefix    = "thesis:midterm:"
	finalCachePrefix      = "thesis:final:"
)

func NewGRPCthesis(addr string, redisClient *redis.Client) (*GRPCthesis, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewThesisServiceClient(conn)
	return &GRPCthesis{
		conn:        conn,
		client:      client,
		redisClient: redisClient,
	}, nil
}

// InvalidateTopicCache invalidates topic-related cache
func (t *GRPCthesis) InvalidateTopicCache(ctx context.Context, topicCode string) error {
	pattern := fmt.Sprintf("%s*%s*", topicCachePrefix, topicCode)
	return InvalidateCacheByPattern(ctx, t.redisClient, pattern)
}

// InvalidateEnrollmentCache invalidates enrollment-related cache
func (t *GRPCthesis) InvalidateEnrollmentCache(ctx context.Context, topicCode string) error {
	pattern := fmt.Sprintf("%s*%s*", enrollmentCachePrefix, topicCode)
	return InvalidateCacheByPattern(ctx, t.redisClient, pattern)
}

// InvalidateMidtermCache invalidates midterm cache by ID
func (t *GRPCthesis) InvalidateMidtermCache(ctx context.Context, midtermCode string) error {
	cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, midtermCode)
	return InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
}

// InvalidateFinalCache invalidates final cache by ID
func (t *GRPCthesis) InvalidateFinalCache(ctx context.Context, finalCode string) error {
	cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, finalCode)
	return InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
}

// InvalidateAllTopicCache invalidates all topic-related caches
func (t *GRPCthesis) InvalidateAllTopicCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", topicCachePrefix)
	return InvalidateCacheByPattern(ctx, t.redisClient, pattern)
}

// InvalidateAllEnrollmentCache invalidates all enrollment-related caches
func (t *GRPCthesis) InvalidateAllEnrollmentCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", enrollmentCachePrefix)
	return InvalidateCacheByPattern(ctx, t.redisClient, pattern)
}

func (t *GRPCthesis) GetTopicBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListTopicsResponse, error) {

	// Generate cache key based on search parameters
	cacheKey := GenerateCacheKey(topicCachePrefix, search)
	// Try to get from cache
	var cached pb.ListTopicsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic search")
		return &cached, nil
	}

	log.Printf("Cache MISS for topic search")

	// Cache miss - fetch from database
	resp, err := t.client.ListTopics(ctx, &pb.ListTopicsRequest{
		Search: search,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, t.redisClient, cacheKey, resp, topicCacheTTL)

	return resp, nil
}

func (t *GRPCthesis) GetTopicById(ctx context.Context, id string) (*pb.GetTopicResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", topicCachePrefix, id)
	var cached pb.GetTopicResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic search")
		return &cached, nil
	}
	log.Printf("Cache MISS for topic search")
	topic, err := t.client.GetTopic(ctx, &pb.GetTopicRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	SetCachedProto(ctx, t.redisClient, cacheKey, topic, topicCacheTTL)
	return topic, nil
}

//func (t *GRPCthesis) GetTopicByStudentCode(ctx context.Context, studentCode string, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
//	if t.client == nil {
//		return nil, fmt.Errorf("grpc client not initialized")
//	}
//	return t.client.ListTopics(ctx, &pb.ListTopicsRequest{
//		Search: &pbCommon.SearchRequest{
//			Pagination: &pbCommon.Pagination{
//				Descending: order,
//				Page:       page,
//				PageSize:   pageSize,
//				SortBy:     sortBy,
//			},
//			Filters: []*pbCommon.FilterCriteria{
//				{
//					Criteria: &pbCommon.FilterCriteria_Condition{
//						Condition: &pbCommon.FilterCondition{
//							Field:    "studentCode",
//							Operator: pbCommon.FilterOperator_EQUAL,
//							Values:   []string{studentCode},
//						},
//					},
//				},
//			},
//		},
//	})
//}

//func (t *GRPCthesis) GetTopicByTeacherCode(ctx context.Context, teacerCode string, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
//	if t.client == nil {
//		return nil, fmt.Errorf("grpc client not initialized")
//	}
//	return t.client.ListTopics(ctx, &pb.ListTopicsRequest{
//		Search: &pbCommon.SearchRequest{
//			Pagination: &pbCommon.Pagination{
//				Descending: order,
//				Page:       page,
//				PageSize:   pageSize,
//				SortBy:     sortBy,
//			},
//			Filters: []*pbCommon.FilterCriteria{
//				{
//					Criteria: &pbCommon.FilterCriteria_Condition{
//						Condition: &pbCommon.FilterCondition{
//							Field:    "teacher_supervisor_code",
//							Operator: pbCommon.FilterOperator_EQUAL,
//							Values:   []string{teacerCode},
//						},
//					},
//				},
//			},
//		},
//	})
//}

//func (t *GRPCthesis) GetTopicByMajorCode(ctx context.Context, majorCode string, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
//	if t.client == nil {
//		return nil, fmt.Errorf("grpc client not initialized")
//	}
//	return t.client.ListTopics(ctx, &pb.ListTopicsRequest{
//		Search: &pbCommon.SearchRequest{
//			Pagination: &pbCommon.Pagination{
//				Descending: order,
//				Page:       page,
//				PageSize:   pageSize,
//				SortBy:     sortBy,
//			},
//			Filters: []*pbCommon.FilterCriteria{
//				{
//					Criteria: &pbCommon.FilterCriteria_Condition{
//						Condition: &pbCommon.FilterCondition{
//							Field:    "major_code",
//							Operator: pbCommon.FilterOperator_EQUAL,
//							Values:   []string{majorCode},
//						},
//					},
//				},
//			},
//		},
//	})
//}

//func (t *GRPCthesis) GetTopic(ctx context.Context, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
//	if t.client == nil {
//		return nil, fmt.Errorf("grpc client not initialized")
//	}
//	return t.client.ListTopics(ctx, &pb.ListTopicsRequest{
//		Search: &pbCommon.SearchRequest{
//			Pagination: &pbCommon.Pagination{
//				Descending: order,
//				Page:       page,
//				PageSize:   pageSize,
//				SortBy:     sortBy,
//			},
//		},
//	})
//}

//func (t *GRPCthesis) GetEnrolmentByTopicCodeAndStudentCode(ctx context.Context, topicCode string, studentCode string) (*pb.ListEnrollmentsResponse, error) {
//	if t.client == nil {
//		return nil, fmt.Errorf("grpc client not initialized")
//	}
//	return t.client.ListEnrollments(ctx, &pb.ListEnrollmentsRequest{
//		Search: &pbCommon.SearchRequest{
//			Pagination: &pbCommon.Pagination{
//				Descending: true,
//				Page:       1,
//				PageSize:   1000,
//				SortBy:     "created_at",
//			},
//			Filters: []*pbCommon.FilterCriteria{
//				{
//					Criteria: &pbCommon.FilterCriteria_Condition{
//						Condition: &pbCommon.FilterCondition{
//							Field:    "student_code",
//							Operator: pbCommon.FilterOperator_EQUAL,
//							Values:   []string{studentCode},
//						},
//					},
//				}, {
//					Criteria: &pbCommon.FilterCriteria_Condition{
//						Condition: &pbCommon.FilterCondition{
//							Field:    "topic_code",
//							Operator: pbCommon.FilterOperator_EQUAL,
//							Values:   []string{topicCode},
//						},
//					},
//				},
//			},
//		},
//	})
//}

func (t *GRPCthesis) GetEnrollmentByTopicCode(ctx context.Context, topicCode string) (*pb.ListEnrollmentsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	// Generate cache key
	cacheKey := GenerateCacheKey(enrollmentCachePrefix, map[string]interface{}{
		"topicCode": topicCode,
	})

	// Try to get from cache
	var cached pb.ListEnrollmentsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for enrollments of topic: %s", topicCode)
		return &cached, nil
	}

	log.Printf("Cache MISS for enrollments of topic: %s", topicCode)

	// Cache miss - fetch from database
	resp, err := t.client.ListEnrollments(ctx, &pb.ListEnrollmentsRequest{
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
							Field:    "topic_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{topicCode},
						},
					},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, t.redisClient, cacheKey, resp, enrollmentCacheTTL)

	return resp, nil
}

func (t *GRPCthesis) GetMidtermById(ctx context.Context, midtermCode string) (*pb.GetMidtermResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, midtermCode)

	// Try to get from cache
	var cached pb.GetMidtermResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for midterm: %s", midtermCode)
		return &cached, nil
	}

	log.Printf("Cache MISS for midterm: %s", midtermCode)

	// Cache miss - fetch from database
	resp, err := t.client.GetMidterm(ctx, &pb.GetMidtermRequest{
		Id: midtermCode,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, t.redisClient, cacheKey, resp, midtermCacheTTL)

	return resp, nil
}

func (t *GRPCthesis) GetFinalById(ctx context.Context, finalCode string) (*pb.GetFinalResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, finalCode)

	// Try to get from cache
	var cached pb.GetFinalResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for final: %s", finalCode)
		return &cached, nil
	}

	log.Printf("Cache MISS for final: %s", finalCode)

	// Cache miss - fetch from database
	resp, err := t.client.GetFinal(ctx, &pb.GetFinalRequest{
		Id: finalCode,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, t.redisClient, cacheKey, resp, finalCacheTTL)

	return resp, nil
}

func (t *GRPCthesis) GetEnrollmentBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListEnrollmentsResponse, error) {

	// Generate cache key based on search parameters
	cacheKey := GenerateCacheKey(enrollmentCachePrefix, search)
	// Try to get from cache
	var cached pb.ListEnrollmentsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic search")
		return &cached, nil
	}

	log.Printf("Cache MISS for topic search")

	// Cache miss - fetch from database
	resp, err := t.client.ListEnrollments(ctx, &pb.ListEnrollmentsRequest{
		Search: search,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, t.redisClient, cacheKey, resp, enrollmentCacheTTL)

	return resp, nil
}

// GetTopicsByIds fetches multiple topics using IN operator
// This is for DataLoader batching
func (t *GRPCthesis) GetTopicsByIds(ctx context.Context, ids []string) (*pb.ListTopicsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	resp, err := t.client.ListTopics(ctx, &pb.ListTopicsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: false,
				Page:       1,
				PageSize:   int32(len(ids)),
				SortBy:     "id",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "id",
							Operator: pbCommon.FilterOperator_IN,
							Values:   ids,
						},
					},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetMidtermsByIds fetches multiple midterms using IN operator
// This is for DataLoader batching
func (t *GRPCthesis) GetMidtermsByIds(ctx context.Context, ids []string) (*pb.ListMidtermsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	fmt.Print(ids)
	resp, err := t.client.ListMidterms(ctx, &pb.ListMidtermsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: false,
				Page:       1,
				PageSize:   int32(len(ids)),
				SortBy:     "id",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "id",
							Operator: pbCommon.FilterOperator_IN,
							Values:   ids,
						},
					},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetFinalsByIds fetches multiple finals using IN operator
// This is for DataLoader batching
func (t *GRPCthesis) GetFinalsByIds(ctx context.Context, ids []string) (*pb.ListFinalsResponse, error) {
	if t.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	resp, err := t.client.ListFinals(ctx, &pb.ListFinalsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: false,
				Page:       1,
				PageSize:   int32(len(ids)),
				SortBy:     "id",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "id",
							Operator: pbCommon.FilterOperator_IN,
							Values:   ids,
						},
					},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
