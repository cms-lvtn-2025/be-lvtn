package client

import (
	"context"
	"fmt"
	"log"
	"time"

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

const (
	// Cache TTL configurations for Role
	roleCacheTTL = 5 * time.Minute

	// Cache key prefixes for Role
	roleCachePrefix        = "role:role:"
	roleTeacherCachePrefix = "role:teacher:"
)

func NewGRPCRole(addr string, redisClient *redis.Client) (*GRPCRole, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewRoleServiceClient(conn)
	return &GRPCRole{
		conn:        conn,
		client:      client,
		redisClient: redisClient,
	}, nil
}

// InvalidateRoleCache invalidates role-related cache
func (r *GRPCRole) InvalidateRoleCache(ctx context.Context, roleCode string) error {
	pattern := fmt.Sprintf("%s*%s*", roleCachePrefix, roleCode)
	return InvalidateCacheByPattern(ctx, r.redisClient, pattern)
}

// InvalidateTeacherRoleCache invalidates teacher role cache
func (r *GRPCRole) InvalidateTeacherRoleCache(ctx context.Context, teacherId string) error {
	pattern := fmt.Sprintf("%s%s*", roleTeacherCachePrefix, teacherId)
	return InvalidateCacheByPattern(ctx, r.redisClient, pattern)
}

// InvalidateAllRoleCache invalidates all role-related caches
func (r *GRPCRole) InvalidateAllRoleCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", roleCachePrefix)
	return InvalidateCacheByPattern(ctx, r.redisClient, pattern)
}

// GetAllRoleByTeacherId retrieves all roles for a teacher with caching
func (r *GRPCRole) GetAllRoleByTeacherId(ctx context.Context, teacherId string) (*pb.ListRoleSystemsResponse, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("GRPCRole is nil")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", roleTeacherCachePrefix, teacherId)

	// Try to get from cache
	var cached pb.ListRoleSystemsResponse
	if hit, _ := GetCachedProto(ctx, r.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for teacher roles: %s", teacherId)
		return &cached, nil
	}

	log.Printf("Cache MISS for teacher roles: %s", teacherId)

	// Cache miss - fetch from database
	resp, err := r.client.ListRoleSystems(ctx, &pb.ListRoleSystemsRequest{
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

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, r.redisClient, cacheKey, resp, roleCacheTTL)

	return resp, nil
}

// GetRoleBySearch retrieves roles by search with caching
func (r *GRPCRole) GetRoleBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListRoleSystemsResponse, error) {
	// Generate cache key based on search parameters
	cacheKey := GenerateCacheKey(roleCachePrefix, search)

	// Try to get from cache
	var cached pb.ListRoleSystemsResponse
	if hit, _ := GetCachedProto(ctx, r.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for role search")
		return &cached, nil
	}

	log.Printf("Cache MISS for role search")

	// Cache miss - fetch from database
	resp, err := r.client.ListRoleSystems(ctx, &pb.ListRoleSystemsRequest{
		Search: search,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, r.redisClient, cacheKey, resp, roleCacheTTL)

	return resp, nil
}
