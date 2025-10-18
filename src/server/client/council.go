package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/council"
	"thaily/src/pkg/tls"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type GRPCCouncil struct {
	conn        *grpc.ClientConn
	client      pb.CouncilServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations
	councilCacheTTL      = 5 * time.Minute
	defenceCacheTTL      = 5 * time.Minute
	scheduleCacheTTL     = 5 * time.Minute
	gradeDefenceCacheTTL = 10 * time.Minute

	// Cache key prefixes
	councilCachePrefix      = "council:council:"
	defenceCachePrefix      = "council:defence:"
	scheduleCachePrefix     = "council:schedule:"
	gradeDefenceCachePrefix = "council:grade_defence:"
)

func NewGRPCCouncil(addr string, redisClient *redis.Client) (*GRPCCouncil, error) {
	// Load mTLS credentials
	creds, err := tls.LoadClientTLSCredentials("council-service")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
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

// ============================================
// COUNCIL METHODS
// ============================================

func (c *GRPCCouncil) GetCouncilBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListCouncilsResponse, error) {
	cacheKey := GenerateCacheKey(councilCachePrefix, search)
	var cached pb.ListCouncilsResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council search")
		return &cached, nil
	}

	log.Printf("Cache MISS for council search")
	resp, err := c.client.ListCouncils(ctx, &pb.ListCouncilsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, c.redisClient, cacheKey, resp, councilCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) GetCouncilById(ctx context.Context, id string) (*pb.GetCouncilResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", councilCachePrefix, id)
	var cached pb.GetCouncilResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for council: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for council: %s", id)
	resp, err := c.client.GetCouncil(ctx, &pb.GetCouncilRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, c.redisClient, cacheKey, resp, councilCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) UpdateCouncil(ctx context.Context, req *pb.UpdateCouncilRequest) (*pb.UpdateCouncilResponse, error) {
	resp, err := c.client.UpdateCouncil(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", councilCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, c.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, c.redisClient, councilCachePrefix+"*")
	}

	return resp, nil
}

func (c *GRPCCouncil) DeleteCouncil(ctx context.Context, id string) (*pb.DeleteCouncilResponse, error) {
	resp, err := c.client.DeleteCouncil(ctx, &pb.DeleteCouncilRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", councilCachePrefix, id)
	InvalidateCacheByKey(ctx, c.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, c.redisClient, councilCachePrefix+"*")

	return resp, nil
}

func (c *GRPCCouncil) GetCouncilsByIds(ctx context.Context, ids []string) (*pb.ListCouncilsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListCouncilsResponse{Councils: []*pb.Council{}}, nil
	}

	result := &pb.ListCouncilsResponse{Councils: []*pb.Council{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", councilCachePrefix, id)
		var cached pb.GetCouncilResponse

		if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
			if cached.Council != nil {
				result.Councils = append(result.Councils, cached.Council)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetCouncilsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := c.client.ListCouncils(ctx, &pb.ListCouncilsRequest{
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
		if resp != nil && resp.Councils != nil {
			for _, council := range resp.Councils {
				if council != nil {
					cacheKey := fmt.Sprintf("%s%s", councilCachePrefix, council.Id)
					SetCachedProto(ctx, c.redisClient, cacheKey, &pb.GetCouncilResponse{Council: council}, councilCacheTTL)
					result.Councils = append(result.Councils, council)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// DEFENCE METHODS
// ============================================

func (c *GRPCCouncil) GetDefencesBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListDefencesResponse, error) {
	cacheKey := GenerateCacheKey(defenceCachePrefix, search)
	var cached pb.ListDefencesResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for defence search")
		return &cached, nil
	}

	log.Printf("Cache MISS for defence search")
	resp, err := c.client.ListDefences(ctx, &pb.ListDefencesRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, c.redisClient, cacheKey, resp, defenceCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) GetDefenceById(ctx context.Context, id string) (*pb.GetDefenceResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", defenceCachePrefix, id)
	var cached pb.GetDefenceResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for defence: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for defence: %s", id)
	resp, err := c.client.GetDefence(ctx, &pb.GetDefenceRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, c.redisClient, cacheKey, resp, defenceCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) UpdateDefence(ctx context.Context, req *pb.UpdateDefenceRequest) (*pb.UpdateDefenceResponse, error) {
	resp, err := c.client.UpdateDefence(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", defenceCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, c.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, c.redisClient, defenceCachePrefix+"*")
	}

	return resp, nil
}

func (c *GRPCCouncil) DeleteDefence(ctx context.Context, id string) (*pb.DeleteDefenceResponse, error) {
	resp, err := c.client.DeleteDefence(ctx, &pb.DeleteDefenceRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", defenceCachePrefix, id)
	InvalidateCacheByKey(ctx, c.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, c.redisClient, defenceCachePrefix+"*")

	return resp, nil
}

func (c *GRPCCouncil) GetDefencesByIds(ctx context.Context, ids []string) (*pb.ListDefencesResponse, error) {
	if len(ids) == 0 {
		return &pb.ListDefencesResponse{Defences: []*pb.Defence{}}, nil
	}

	result := &pb.ListDefencesResponse{Defences: []*pb.Defence{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", defenceCachePrefix, id)
		var cached pb.GetDefenceResponse

		if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
			if cached.Defence != nil {
				result.Defences = append(result.Defences, cached.Defence)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetDefencesByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := c.client.ListDefences(ctx, &pb.ListDefencesRequest{
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
		if resp != nil && resp.Defences != nil {
			for _, defence := range resp.Defences {
				if defence != nil {
					cacheKey := fmt.Sprintf("%s%s", defenceCachePrefix, defence.Id)
					SetCachedProto(ctx, c.redisClient, cacheKey, &pb.GetDefenceResponse{Defence: defence}, defenceCacheTTL)
					result.Defences = append(result.Defences, defence)
				}
			}
		}
	}

	return result, nil
}

func (c *GRPCCouncil) GetDefencesByCouncilCode(ctx context.Context, councilCode string) (*pb.ListDefencesResponse, error) {
	cacheKey := fmt.Sprintf("%scouncil:%s", defenceCachePrefix, councilCode)
	var cached pb.ListDefencesResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for defences by council: %s", councilCode)
		return &cached, nil
	}

	log.Printf("Cache MISS for defences by council: %s", councilCode)
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

// ============================================
// GRADE DEFENCE METHODS
// ============================================

func (c *GRPCCouncil) GetGradeDefenceBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListGradeDefencesResponse, error) {
	cacheKey := GenerateCacheKey(gradeDefenceCachePrefix, search)
	var cached pb.ListGradeDefencesResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for grade defence search")
		return &cached, nil
	}

	log.Printf("Cache MISS for grade defence search")
	resp, err := c.client.ListGradeDefences(ctx, &pb.ListGradeDefencesRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, c.redisClient, cacheKey, resp, gradeDefenceCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) GetGradeById(ctx context.Context, id string) (*pb.GetGradeDefenceResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", gradeDefenceCachePrefix, id)
	var cached pb.GetGradeDefenceResponse
	if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for grade defence: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for grade defence: %s", id)
	resp, err := c.client.GetGradeDefence(ctx, &pb.GetGradeDefenceRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, c.redisClient, cacheKey, resp, gradeDefenceCacheTTL)
	return resp, nil
}

func (c *GRPCCouncil) UpdateGradeDefence(ctx context.Context, req *pb.UpdateGradeDefenceRequest) (*pb.UpdateGradeDefenceResponse, error) {
	resp, err := c.client.UpdateGradeDefence(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", gradeDefenceCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, c.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, c.redisClient, gradeDefenceCachePrefix+"*")
	}

	return resp, nil
}

func (c *GRPCCouncil) DeleteGradeDefence(ctx context.Context, id string) (*pb.DeleteGradeDefenceResponse, error) {
	resp, err := c.client.DeleteGradeDefence(ctx, &pb.DeleteGradeDefenceRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", gradeDefenceCachePrefix, id)
	InvalidateCacheByKey(ctx, c.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, c.redisClient, gradeDefenceCachePrefix+"*")

	return resp, nil
}

func (c *GRPCCouncil) GetGradeDefencesByIds(ctx context.Context, ids []string) (*pb.ListGradeDefencesResponse, error) {
	if len(ids) == 0 {
		return &pb.ListGradeDefencesResponse{GradeDefences: []*pb.GradeDefence{}}, nil
	}

	result := &pb.ListGradeDefencesResponse{GradeDefences: []*pb.GradeDefence{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", gradeDefenceCachePrefix, id)
		var cached pb.GetGradeDefenceResponse

		if hit, _ := GetCachedProto(ctx, c.redisClient, cacheKey, &cached); hit {
			if cached.GradeDefence != nil {
				result.GradeDefences = append(result.GradeDefences, cached.GradeDefence)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetGradeDefencesByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := c.client.ListGradeDefences(ctx, &pb.ListGradeDefencesRequest{
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
		if resp != nil && resp.GradeDefences != nil {
			for _, gradeDefence := range resp.GradeDefences {
				if gradeDefence != nil {
					cacheKey := fmt.Sprintf("%s%s", gradeDefenceCachePrefix, gradeDefence.Id)
					SetCachedProto(ctx, c.redisClient, cacheKey, &pb.GetGradeDefenceResponse{GradeDefence: gradeDefence}, gradeDefenceCacheTTL)
					result.GradeDefences = append(result.GradeDefences, gradeDefence)
				}
			}
		}
	}

	return result, nil
}
