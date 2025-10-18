package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/file"
	"thaily/src/pkg/tls"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type GRPCfile struct {
	conn        *grpc.ClientConn
	client      pb.FileServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations
	fileCacheTTL = 15 * time.Minute // Files are relatively stable once uploaded

	// Cache key prefixes
	fileCachePrefix = "file:file:"
)

func NewGRPCfile(addr string, redisClient *redis.Client) (*GRPCfile, error) {
	// Load mTLS credentials
	creds, err := tls.LoadClientTLSCredentials("file-service")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
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

// ============================================
// FILE METHODS
// ============================================

func (f *GRPCfile) GetFileBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListFilesResponse, error) {
	cacheKey := GenerateCacheKey(fileCachePrefix, search)
	var cached pb.ListFilesResponse
	if hit, _ := GetCachedProto(ctx, f.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for file search")
		return &cached, nil
	}

	log.Printf("Cache MISS for file search")
	resp, err := f.client.ListFiles(ctx, &pb.ListFilesRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, f.redisClient, cacheKey, resp, fileCacheTTL)
	return resp, nil
}

func (f *GRPCfile) GetFileById(ctx context.Context, id string) (*pb.GetFileResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", fileCachePrefix, id)
	var cached pb.GetFileResponse
	if hit, _ := GetCachedProto(ctx, f.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for file: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for file: %s", id)
	resp, err := f.client.GetFile(ctx, &pb.GetFileRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, f.redisClient, cacheKey, resp, fileCacheTTL)
	return resp, nil
}

func (f *GRPCfile) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	resp, err := f.client.CreateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate search caches (new file added)
	InvalidateCacheByPattern(ctx, f.redisClient, fileCachePrefix+"search:*")

	return resp, nil
}

func (f *GRPCfile) UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileResponse, error) {
	resp, err := f.client.UpdateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", fileCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, f.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, f.redisClient, fileCachePrefix+"*")
	}

	return resp, nil
}

func (f *GRPCfile) DeleteFile(ctx context.Context, id string) (*pb.DeleteFileResponse, error) {
	resp, err := f.client.DeleteFile(ctx, &pb.DeleteFileRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", fileCachePrefix, id)
	InvalidateCacheByKey(ctx, f.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, f.redisClient, fileCachePrefix+"*")

	return resp, nil
}

func (f *GRPCfile) GetFilesByIds(ctx context.Context, ids []string) (*pb.ListFilesResponse, error) {
	if len(ids) == 0 {
		return &pb.ListFilesResponse{Files: []*pb.File{}}, nil
	}

	result := &pb.ListFilesResponse{Files: []*pb.File{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", fileCachePrefix, id)
		var cached pb.GetFileResponse

		if hit, _ := GetCachedProto(ctx, f.redisClient, cacheKey, &cached); hit {
			if cached.File != nil {
				result.Files = append(result.Files, cached.File)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetFilesByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := f.client.ListFiles(ctx, &pb.ListFilesRequest{
			Search: &pbCommon.SearchRequest{
				Pagination: &pbCommon.Pagination{
					Descending: false,
					Page:       1,
					PageSize:   int32(len(missingIds)),
					SortBy:     "id",
				},
				Filters: []*pbCommon.FilterCriteria{
					{
						Criteria: &pbCommon.FilterCriteria_Condition{
							Condition: &pbCommon.FilterCondition{
								Field:    "id",
								Operator: pbCommon.FilterOperator_IN,
								Values:   missingIds,
							},
						},
					},
				},
			},
		})

		if err != nil {
			return nil, err
		}

		// Store fetched items to Redis and add to result
		if resp != nil && resp.Files != nil {
			for _, file := range resp.Files {
				if file != nil {
					cacheKey := fmt.Sprintf("%s%s", fileCachePrefix, file.Id)
					SetCachedProto(ctx, f.redisClient, cacheKey, &pb.GetFileResponse{File: file}, fileCacheTTL)
					result.Files = append(result.Files, file)
				}
			}
		}
	}

	return result, nil
}
