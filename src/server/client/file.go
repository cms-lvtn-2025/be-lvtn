package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/file"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCfile struct {
	conn        *grpc.ClientConn
	client      pb.FileServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations for File
	fileCacheTTL = 10 * time.Minute // Files are more stable

	// Cache key prefixes for File
	fileCachePrefix = "file:file:"
)

func NewGRPCfile(addr string, redisClient *redis.Client) (*GRPCfile, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewFileServiceClient(conn)
	return &GRPCfile{
		conn:        conn,
		client:      client,
		redisClient: redisClient,
	}, nil
}

// InvalidateFileCache invalidates file-related cache
func (f *GRPCfile) InvalidateFileCache(ctx context.Context, fileCode string) error {
	pattern := fmt.Sprintf("%s*%s*", fileCachePrefix, fileCode)
	return InvalidateCacheByPattern(ctx, f.redisClient, pattern)
}

// InvalidateAllFileCache invalidates all file-related caches
func (f *GRPCfile) InvalidateAllFileCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", fileCachePrefix)
	return InvalidateCacheByPattern(ctx, f.redisClient, pattern)
}

// GetFileBySearch retrieves files by search with caching
func (f *GRPCfile) GetFileBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListFilesResponse, error) {
	// Generate cache key based on search parameters
	cacheKey := GenerateCacheKey(fileCachePrefix, search)

	// Try to get from cache
	var cached pb.ListFilesResponse
	if hit, _ := GetCachedProto(ctx, f.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for file search")
		return &cached, nil
	}

	log.Printf("Cache MISS for file search")

	// Cache miss - fetch from database
	resp, err := f.client.ListFiles(ctx, &pb.ListFilesRequest{
		Search: search,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, f.redisClient, cacheKey, resp, fileCacheTTL)

	return resp, nil
}

// GetFileById retrieves a file by ID with caching
func (f *GRPCfile) GetFileById(ctx context.Context, fileCode string) (*pb.GetFileResponse, error) {
	if f.client == nil {
		return nil, fmt.Errorf("grpc client not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", fileCachePrefix, fileCode)

	// Try to get from cache
	var cached pb.GetFileResponse
	if hit, _ := GetCachedProto(ctx, f.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for file: %s", fileCode)
		return &cached, nil
	}

	log.Printf("Cache MISS for file: %s", fileCode)

	// Cache miss - fetch from database
	resp, err := f.client.GetFile(ctx, &pb.GetFileRequest{
		Id: fileCode,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, f.redisClient, cacheKey, resp, fileCacheTTL)

	return resp, nil
}
