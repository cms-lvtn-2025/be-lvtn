package client

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/thesis"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
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

// generateCacheKey creates a unique cache key from prefix and parameters
func generateCacheKey(prefix string, params interface{}) string {
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s%x", prefix, hash[:16])
}

// getCachedProto retrieves cached protobuf message from Redis
func (t *GRPCthesis) getCachedProto(ctx context.Context, key string, dest proto.Message) (bool, error) {
	if t.redisClient == nil {
		return false, nil
	}

	data, err := t.redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil // Cache miss
	}
	if err != nil {
		log.Printf("Redis get error for key %s: %v", key, err)
		return false, nil // Don't fail on cache errors
	}

	if err := proto.Unmarshal(data, dest); err != nil {
		log.Printf("Proto unmarshal error for key %s: %v", key, err)
		return false, nil
	}

	return true, nil
}

// setCachedProto stores protobuf message in Redis
func (t *GRPCthesis) setCachedProto(ctx context.Context, key string, msg proto.Message, ttl time.Duration) {
	if t.redisClient == nil {
		return
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Proto marshal error for key %s: %v", key, err)
		return
	}

	if err := t.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("Redis set error for key %s: %v", key, err)
	}
}

// InvalidateTopicCache invalidates topic-related cache
func (t *GRPCthesis) InvalidateTopicCache(ctx context.Context, topicCode string) error {
	if t.redisClient == nil {
		return nil
	}

	// Delete all keys matching the pattern
	pattern := fmt.Sprintf("%s*%s*", topicCachePrefix, topicCode)
	iter := t.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := t.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("Failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	return iter.Err()
}

// InvalidateEnrollmentCache invalidates enrollment-related cache
func (t *GRPCthesis) InvalidateEnrollmentCache(ctx context.Context, topicCode string) error {
	if t.redisClient == nil {
		return nil
	}

	pattern := fmt.Sprintf("%s*%s*", enrollmentCachePrefix, topicCode)
	iter := t.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := t.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("Failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	return iter.Err()
}

// InvalidateMidtermCache invalidates midterm cache by ID
func (t *GRPCthesis) InvalidateMidtermCache(ctx context.Context, midtermCode string) error {
	if t.redisClient == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, midtermCode)
	if err := t.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("Failed to delete midterm cache key %s: %v", cacheKey, err)
		return err
	}
	return nil
}

// InvalidateFinalCache invalidates final cache by ID
func (t *GRPCthesis) InvalidateFinalCache(ctx context.Context, finalCode string) error {
	if t.redisClient == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, finalCode)
	if err := t.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("Failed to delete final cache key %s: %v", cacheKey, err)
		return err
	}
	return nil
}

// InvalidateAllTopicCache invalidates all topic-related caches
func (t *GRPCthesis) InvalidateAllTopicCache(ctx context.Context) error {
	if t.redisClient == nil {
		return nil
	}

	pattern := fmt.Sprintf("%s*", topicCachePrefix)
	iter := t.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := t.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("Failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	return iter.Err()
}

// InvalidateAllEnrollmentCache invalidates all enrollment-related caches
func (t *GRPCthesis) InvalidateAllEnrollmentCache(ctx context.Context) error {
	if t.redisClient == nil {
		return nil
	}

	pattern := fmt.Sprintf("%s*", enrollmentCachePrefix)
	iter := t.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := t.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("Failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	return iter.Err()
}

func (t *GRPCthesis) GetTopicBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListTopicsResponse, error) {

	// Generate cache key based on search parameters
	cacheKey := generateCacheKey(topicCachePrefix, search)
	// Try to get from cache
	var cached pb.ListTopicsResponse
	if hit, _ := t.getCachedProto(ctx, cacheKey, &cached); hit {
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
	t.setCachedProto(ctx, cacheKey, resp, topicCacheTTL)

	return resp, nil
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

func (t *GRPCthesis) GetTopicByMajorCode(ctx context.Context, majorCode string, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
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
							Field:    "major_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{majorCode},
						},
					},
				},
			},
		},
	})
}

func (t *GRPCthesis) GetTopic(ctx context.Context, page int32, pageSize int32, sortBy string, order bool) (*pb.ListTopicsResponse, error) {
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

	// Generate cache key
	cacheKey := generateCacheKey(enrollmentCachePrefix, map[string]interface{}{
		"topicCode": topicCode,
	})

	// Try to get from cache
	var cached pb.ListEnrollmentsResponse
	if hit, _ := t.getCachedProto(ctx, cacheKey, &cached); hit {
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
	t.setCachedProto(ctx, cacheKey, resp, enrollmentCacheTTL)

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
	if hit, _ := t.getCachedProto(ctx, cacheKey, &cached); hit {
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
	t.setCachedProto(ctx, cacheKey, resp, midtermCacheTTL)

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
	if hit, _ := t.getCachedProto(ctx, cacheKey, &cached); hit {
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
	t.setCachedProto(ctx, cacheKey, resp, finalCacheTTL)

	return resp, nil
}
