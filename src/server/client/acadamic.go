package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/academic"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCAcadamicClient struct {
	conn        *grpc.ClientConn
	client      pb.AcademicServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations for Academic
	majorCacheTTL    = 30 * time.Minute // Majors are very stable
	semesterCacheTTL = 15 * time.Minute // Semesters are relatively stable

	// Cache key prefixes for Academic
	majorCachePrefix    = "academic:major:"
	semesterCachePrefix = "academic:semester:"
)

func NewGRPCAcadamicClient(addr string, redisClient *redis.Client) (*GRPCAcadamicClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewAcademicServiceClient(conn)
	return &GRPCAcadamicClient{
		conn:        conn,
		client:      client,
		redisClient: redisClient,
	}, nil
}

// InvalidateMajorCache invalidates major-related cache
func (g *GRPCAcadamicClient) InvalidateMajorCache(ctx context.Context, majorCode string) error {
	pattern := fmt.Sprintf("%s*%s*", majorCachePrefix, majorCode)
	return InvalidateCacheByPattern(ctx, g.redisClient, pattern)
}

// InvalidateSemesterCache invalidates semester-related cache
func (g *GRPCAcadamicClient) InvalidateSemesterCache(ctx context.Context, semesterCode string) error {
	pattern := fmt.Sprintf("%s*%s*", semesterCachePrefix, semesterCode)
	return InvalidateCacheByPattern(ctx, g.redisClient, pattern)
}

// InvalidateAllMajorCache invalidates all major-related caches
func (g *GRPCAcadamicClient) InvalidateAllMajorCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", majorCachePrefix)
	return InvalidateCacheByPattern(ctx, g.redisClient, pattern)
}

// InvalidateAllSemesterCache invalidates all semester-related caches
func (g *GRPCAcadamicClient) InvalidateAllSemesterCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", semesterCachePrefix)
	return InvalidateCacheByPattern(ctx, g.redisClient, pattern)
}

func (g *GRPCAcadamicClient) GetMajorById(ctx context.Context, id string) (*pb.GetMajorResponse, error) {
	if g.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", majorCachePrefix, id)

	// Try to get from cache
	var cached pb.GetMajorResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for major: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for major: %s", id)

	// Cache miss - fetch from database
	resp, err := g.client.GetMajor(ctx, &pb.GetMajorRequest{
		Id: id,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, g.redisClient, cacheKey, resp, majorCacheTTL)

	return resp, nil
}

func (g *GRPCAcadamicClient) GetSemesterById(ctx context.Context, id string) (*pb.GetSemesterResponse, error) {
	if g.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", semesterCachePrefix, id)

	// Try to get from cache
	var cached pb.GetSemesterResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for semester: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for semester: %s", id)

	// Cache miss - fetch from database
	resp, err := g.client.GetSemester(ctx, &pb.GetSemesterRequest{
		Id: id,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, g.redisClient, cacheKey, resp, semesterCacheTTL)

	return resp, nil
}

// GetMajorsBySearch retrieves majors by search with caching
func (g *GRPCAcadamicClient) GetMajorsBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListMajorsResponse, error) {
	// Generate cache key based on search parameters
	cacheKey := GenerateCacheKey(majorCachePrefix, search)

	// Try to get from cache
	var cached pb.ListMajorsResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for major search")
		return &cached, nil
	}

	log.Printf("Cache MISS for major search")

	// Cache miss - fetch from database
	resp, err := g.client.ListMajors(ctx, &pb.ListMajorsRequest{
		Search: search,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, g.redisClient, cacheKey, resp, majorCacheTTL)

	return resp, nil
}

// GetSemestersBySearch retrieves semesters by search with caching
func (g *GRPCAcadamicClient) GetSemestersBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListSemestersResponse, error) {
	// Generate cache key based on search parameters
	cacheKey := GenerateCacheKey(semesterCachePrefix, search)

	// Try to get from cache
	var cached pb.ListSemestersResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for semester search")
		return &cached, nil
	}

	log.Printf("Cache MISS for semester search")

	// Cache miss - fetch from database
	resp, err := g.client.ListSemesters(ctx, &pb.ListSemestersRequest{
		Search: search,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, g.redisClient, cacheKey, resp, semesterCacheTTL)

	return resp, nil
}
