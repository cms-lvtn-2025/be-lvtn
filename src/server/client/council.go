package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
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

const (
	// Cache TTL configurations for Council
	councilCacheTTL  = 5 * time.Minute
	defenceCacheTTL  = 5 * time.Minute
	scheduleCacheTTL = 5 * time.Minute

	// Cache key prefixes for Council
	councilCachePrefix             = "council:council:"
	defenceCouncilCodeCachePrefix  = "defence:council:"
	scheduleCouncilCodeCachePrefix = "schedule:council:"
	scheduleTopicCodeCachePrefix   = "schedule:topic:"
)

func NewGRPCCouncil(addr string, redisClient *redis.Client) (*GRPCCouncil, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewCouncilServiceClient(conn)
	return &GRPCCouncil{
		conn:        conn,
		client:      client,
		redisClient: redisClient,
	}, nil
}

// InvalidateCouncilCache invalidates council-related cache
func (c *GRPCCouncil) InvalidateCouncilCache(ctx context.Context, councilCode string) error {
	pattern := fmt.Sprintf("%s*%s*", councilCachePrefix, councilCode)
	return InvalidateCacheByPattern(ctx, c.redisClient, pattern)
}

// InvalidateAllCouncilCache invalidates all council-related caches
func (c *GRPCCouncil) InvalidateAllCouncilCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", councilCachePrefix)
	return InvalidateCacheByPattern(ctx, c.redisClient, pattern)
}

// GetCouncilBySearch retrieves councils by search with caching
func (c *GRPCCouncil) GetCouncilBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListCouncilsResponse, error) {
	// Generate cache key based on search parameters
	cacheKey := GenerateCacheKey(councilCachePrefix, search)

	// Try to get from cache
	var cached pb.ListCouncilsResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council search")
		return &cached, nil
	}

	log.Printf("Cache MISS for council search")

	// Cache miss - fetch from database
	resp, err := c.client.ListCouncils(ctx, &pb.ListCouncilsRequest{
		Search: search,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, c.redisClient, cacheKey, resp, councilCacheTTL)

	return resp, nil
}

// GetCouncilById retrieves a council by ID with caching
func (c *GRPCCouncil) GetCouncilById(ctx context.Context, councilCode string) (*pb.GetCouncilResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", councilCachePrefix, councilCode)

	// Try to get from cache
	var cached pb.GetCouncilResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council: %s", councilCode)
		return &cached, nil
	}

	log.Printf("Cache MISS for council: %s", councilCode)

	// Cache miss - fetch from database
	resp, err := c.client.GetCouncil(ctx, &pb.GetCouncilRequest{
		Id: councilCode,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, c.redisClient, cacheKey, resp, councilCacheTTL)

	return resp, nil
}

func (c *GRPCCouncil) GetDefencesBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListDefencesResponse, error) {
	cacheKey := GenerateCacheKey(councilCachePrefix, search)
	var cached pb.ListDefencesResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council search")
		return &cached, nil
	}
	log.Printf("Cache MISS for council search")
	resp, err := c.client.ListDefences(ctx, &pb.ListDefencesRequest{
		Search: search,
	})
	if err != nil {
		return nil, err
	}
	SetCachedProto(ctx, c.redisClient, cacheKey, resp, councilCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) GetDefencesByCouncilCode(ctx context.Context, councilCode string) (*pb.ListDefencesResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	cacheKey := fmt.Sprintf("%s%s", defenceCouncilCodeCachePrefix, councilCode)
	var cached pb.ListDefencesResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council: %s", councilCode)
		return &cached, nil
	}
	log.Printf("Cache MISS for council: %s", councilCode)
	resp, err := c.client.ListDefences(ctx, &pb.ListDefencesRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   100,
				SortBy:     "created_by",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "council_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{councilCode},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	SetCachedProto(ctx, c.redisClient, cacheKey, resp, defenceCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) GetSchedulesByCouncilCode(ctx context.Context, councilCode string) (*pb.ListCouncilSchedulesResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	cacheKey := fmt.Sprintf("%s%s", scheduleCouncilCodeCachePrefix, councilCode)
	var cached pb.ListCouncilSchedulesResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council: %s", councilCode)
		return &cached, nil
	}
	log.Printf("Cache MISS for council: %s", councilCode)
	resp, err := c.client.ListCouncilSchedules(ctx, &pb.ListCouncilSchedulesRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   100,
				SortBy:     "created_by",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "council_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{councilCode},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	SetCachedProto(ctx, c.redisClient, cacheKey, resp, councilCacheTTL)
	return resp, nil

}

func (c *GRPCCouncil) GetScheduleByTopicCode(ctx context.Context, topicCode string) (*pb.ListCouncilSchedulesResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}
	cacheKey := fmt.Sprintf("%s%s", scheduleTopicCodeCachePrefix, topicCode)
	var cached pb.ListCouncilSchedulesResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council: %s", topicCode)
		return &cached, nil
	}
	log.Printf("Cache MISS for council: %s", topicCode)
	resp, err := c.client.ListCouncilSchedules(ctx, &pb.ListCouncilSchedulesRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   100,
				SortBy:     "created_by",
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
	SetCachedProto(ctx, c.redisClient, cacheKey, resp, councilCacheTTL)
	return resp, nil

}
