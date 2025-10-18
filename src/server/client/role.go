package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/role"
	"thaily/src/pkg/tls"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type GRPCRole struct {
	conn        *grpc.ClientConn
	client      pb.RoleServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations
	roleSystemCacheTTL = 10 * time.Minute // Roles are relatively stable

	// Cache key prefixes
	roleSystemCachePrefix = "role:role_system:"
)

func NewGRPCRole(addr string, redisClient *redis.Client) (*GRPCRole, error) {
	// Load mTLS credentials
	creds, err := tls.LoadClientTLSCredentials("role-service")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
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

// ============================================
// ROLE SYSTEM METHODS
// ============================================

func (r *GRPCRole) GetRoleBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListRoleSystemsResponse, error) {
	cacheKey := GenerateCacheKey(roleSystemCachePrefix, search)
	var cached pb.ListRoleSystemsResponse
	if hit, _ := GetCachedProto(ctx, r.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for role search")
		return &cached, nil
	}

	log.Printf("Cache MISS for role search")
	resp, err := r.client.ListRoleSystems(ctx, &pb.ListRoleSystemsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, r.redisClient, cacheKey, resp, roleSystemCacheTTL)
	return resp, nil
}

func (r *GRPCRole) GetRoleById(ctx context.Context, id string) (*pb.GetRoleSystemResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", roleSystemCachePrefix, id)
	var cached pb.GetRoleSystemResponse
	if hit, _ := GetCachedProto(ctx, r.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for role: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for role: %s", id)
	resp, err := r.client.GetRoleSystem(ctx, &pb.GetRoleSystemRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, r.redisClient, cacheKey, resp, roleSystemCacheTTL)
	return resp, nil
}

func (r *GRPCRole) UpdateRole(ctx context.Context, req *pb.UpdateRoleSystemRequest) (*pb.UpdateRoleSystemResponse, error) {
	resp, err := r.client.UpdateRoleSystem(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", roleSystemCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, r.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, r.redisClient, roleSystemCachePrefix+"*")

		// Also invalidate teacher-specific cache if teacher_code changed
		if req.TeacherCode != nil && *req.TeacherCode != "" {
			InvalidateCacheByPattern(ctx, r.redisClient, fmt.Sprintf("%steacher:%s*", roleSystemCachePrefix, *req.TeacherCode))
		}
	}

	return resp, nil
}

func (r *GRPCRole) DeleteRole(ctx context.Context, id string) (*pb.DeleteRoleSystemResponse, error) {
	resp, err := r.client.DeleteRoleSystem(ctx, &pb.DeleteRoleSystemRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", roleSystemCachePrefix, id)
	InvalidateCacheByKey(ctx, r.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, r.redisClient, roleSystemCachePrefix+"*")

	return resp, nil
}

func (r *GRPCRole) GetRolesByIds(ctx context.Context, ids []string) (*pb.ListRoleSystemsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListRoleSystemsResponse{RoleSystems: []*pb.RoleSystem{}}, nil
	}

	result := &pb.ListRoleSystemsResponse{RoleSystems: []*pb.RoleSystem{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", roleSystemCachePrefix, id)
		var cached pb.GetRoleSystemResponse

		if hit, _ := GetCachedProto(ctx, r.redisClient, cacheKey, &cached); hit {
			if cached.RoleSystem != nil {
				result.RoleSystems = append(result.RoleSystems, cached.RoleSystem)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetRolesByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := r.client.ListRoleSystems(ctx, &pb.ListRoleSystemsRequest{
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
		if resp != nil && resp.RoleSystems != nil {
			for _, role := range resp.RoleSystems {
				if role != nil {
					cacheKey := fmt.Sprintf("%s%s", roleSystemCachePrefix, role.Id)
					SetCachedProto(ctx, r.redisClient, cacheKey, &pb.GetRoleSystemResponse{RoleSystem: role}, roleSystemCacheTTL)
					result.RoleSystems = append(result.RoleSystems, role)
				}
			}
		}
	}

	return result, nil
}

// GetAllRoleByTeacherId retrieves all roles for a teacher with caching
func (r *GRPCRole) GetAllRoleByTeacherId(ctx context.Context, teacherId string) (*pb.ListRoleSystemsResponse, error) {
	cacheKey := fmt.Sprintf("%steacher:%s", roleSystemCachePrefix, teacherId)
	var cached pb.ListRoleSystemsResponse
	if hit, _ := GetCachedProto(ctx, r.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for teacher roles: %s", teacherId)
		return &cached, nil
	}

	log.Printf("Cache MISS for teacher roles: %s", teacherId)
	resp, err := r.client.ListRoleSystems(ctx, &pb.ListRoleSystemsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: nil,
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
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

	SetCachedProto(ctx, r.redisClient, cacheKey, resp, roleSystemCacheTTL)
	return resp, nil
}

// GetRolesByTeacherIds fetches roles for multiple teachers
func (r *GRPCRole) GetRolesByTeacherIds(ctx context.Context, teacherIds []string) (*pb.ListRoleSystemsResponse, error) {
	if len(teacherIds) == 0 {
		return &pb.ListRoleSystemsResponse{RoleSystems: []*pb.RoleSystem{}}, nil
	}

	result := &pb.ListRoleSystemsResponse{RoleSystems: []*pb.RoleSystem{}}
	missingTeacherIds := []string{}
	cacheHits := 0

	// Check Redis cache for each teacher ID
	for _, teacherId := range teacherIds {
		cacheKey := fmt.Sprintf("%steacher:%s", roleSystemCachePrefix, teacherId)
		var cached pb.ListRoleSystemsResponse

		if hit, _ := GetCachedProto(ctx, r.redisClient, cacheKey, &cached); hit {
			if cached.RoleSystems != nil && len(cached.RoleSystems) > 0 {
				result.RoleSystems = append(result.RoleSystems, cached.RoleSystems...)
				cacheHits++
			} else {
				missingTeacherIds = append(missingTeacherIds, teacherId)
			}
		} else {
			missingTeacherIds = append(missingTeacherIds, teacherId)
		}
	}

	log.Printf("[GetRolesByTeacherIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(teacherIds), cacheHits, len(missingTeacherIds))

	// Fetch missing teacher IDs from database
	if len(missingTeacherIds) > 0 {
		resp, err := r.client.ListRoleSystems(ctx, &pb.ListRoleSystemsRequest{
			Search: &pbCommon.SearchRequest{
				Pagination: &pbCommon.Pagination{
					Descending: false,
					Page:       1,
					PageSize:   int32(len(missingTeacherIds) * 10), // Each teacher may have multiple roles
					SortBy:     "teacher_code",
				},
				Filters: []*pbCommon.FilterCriteria{
					{
						Criteria: &pbCommon.FilterCriteria_Condition{
							Condition: &pbCommon.FilterCondition{
								Field:    "teacher_code",
								Operator: pbCommon.FilterOperator_IN,
								Values:   missingTeacherIds,
							},
						},
					},
				},
			},
		})

		if err != nil {
			return nil, err
		}

		// Group roles by teacher_code and store to Redis
		if resp != nil && resp.RoleSystems != nil {
			teacherMap := make(map[string][]*pb.RoleSystem)
			for _, role := range resp.RoleSystems {
				if role != nil {
					teacherMap[role.TeacherCode] = append(teacherMap[role.TeacherCode], role)
					result.RoleSystems = append(result.RoleSystems, role)
				}
			}

			// Cache each teacher's roles
			for teacherId, roles := range teacherMap {
				cacheKey := fmt.Sprintf("%steacher:%s", roleSystemCachePrefix, teacherId)
				SetCachedProto(ctx, r.redisClient, cacheKey, &pb.ListRoleSystemsResponse{RoleSystems: roles}, roleSystemCacheTTL)
			}
		}
	}

	return result, nil
}
