package client

import (
	"context"
	"fmt"
	"log"
	"thaily/src/server/pkg/tls"
	"time"

	pb "thaily/proto/academic"
	pbCommon "thaily/proto/common"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type GRPCAcadamicClient struct {
	conn        *grpc.ClientConn
	client      pb.AcademicServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations
	majorCacheTTL    = 30 * time.Minute // Majors are very stable
	semesterCacheTTL = 15 * time.Minute // Semesters are relatively stable
	facultyCacheTTL  = 15 * time.Minute // Faculties are relatively stable
	// Cache key prefixes
	majorCachePrefix    = "academic:major:"
	semesterCachePrefix = "academic:semester:"
	facultyCachePrefix  = "academic:faculty:"
)

func NewGRPCAcadamicClient(addr string, redisClient *redis.Client) (*GRPCAcadamicClient, error) {
	// Load mTLS credentials
	creds, err := tls.LoadClientTLSCredentials("academic-service")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
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

// ============================================
// MAJOR METHODS
// ============================================

func (g *GRPCAcadamicClient) GetMajorsBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListMajorsResponse, error) {
	cacheKey := GenerateCacheKey(majorCachePrefix, search)
	var cached pb.ListMajorsResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for major search")
		return &cached, nil
	}

	log.Printf("Cache MISS for major search")
	resp, err := g.client.ListMajors(ctx, &pb.ListMajorsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, g.redisClient, cacheKey, resp, majorCacheTTL)
	return resp, nil
}

func (g *GRPCAcadamicClient) GetMajorById(ctx context.Context, id string) (*pb.GetMajorResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", majorCachePrefix, id)
	var cached pb.GetMajorResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for major: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for major: %s", id)
	resp, err := g.client.GetMajor(ctx, &pb.GetMajorRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, g.redisClient, cacheKey, resp, majorCacheTTL)
	return resp, nil
}

func (g *GRPCAcadamicClient) CreateMajor(ctx context.Context, req *pb.CreateMajorRequest) (*pb.CreateMajorResponse, error) {
	resp, err := g.client.CreateMajor(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	InvalidateCacheByPattern(ctx, g.redisClient, majorCachePrefix+"*")

	return resp, nil
}

func (g *GRPCAcadamicClient) UpdateMajor(ctx context.Context, req *pb.UpdateMajorRequest) (*pb.UpdateMajorResponse, error) {
	resp, err := g.client.UpdateMajor(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", majorCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, g.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, g.redisClient, majorCachePrefix+"*")
	}

	return resp, nil
}

func (g *GRPCAcadamicClient) DeleteMajor(ctx context.Context, id string) (*pb.DeleteMajorResponse, error) {
	resp, err := g.client.DeleteMajor(ctx, &pb.DeleteMajorRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", majorCachePrefix, id)
	InvalidateCacheByKey(ctx, g.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, g.redisClient, majorCachePrefix+"*")

	return resp, nil
}

func (g *GRPCAcadamicClient) GetMajorsByIds(ctx context.Context, ids []string) (*pb.ListMajorsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListMajorsResponse{Majors: []*pb.Major{}}, nil
	}

	result := &pb.ListMajorsResponse{Majors: []*pb.Major{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", majorCachePrefix, id)
		var cached pb.GetMajorResponse

		if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
			if cached.Major != nil {
				result.Majors = append(result.Majors, cached.Major)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetMajorsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := g.client.ListMajors(ctx, &pb.ListMajorsRequest{
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
		if resp != nil && resp.Majors != nil {
			for _, major := range resp.Majors {
				if major != nil {
					cacheKey := fmt.Sprintf("%s%s", majorCachePrefix, major.Id)
					SetCachedProto(ctx, g.redisClient, cacheKey, &pb.GetMajorResponse{Major: major}, majorCacheTTL)
					result.Majors = append(result.Majors, major)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// SEMESTER METHODS
// ============================================

func (g *GRPCAcadamicClient) GetSemestersBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListSemestersResponse, error) {
	cacheKey := GenerateCacheKey(semesterCachePrefix, search)
	var cached pb.ListSemestersResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for semester search")
		return &cached, nil
	}

	log.Printf("Cache MISS for semester search")
	resp, err := g.client.ListSemesters(ctx, &pb.ListSemestersRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, g.redisClient, cacheKey, resp, semesterCacheTTL)
	return resp, nil
}

func (g *GRPCAcadamicClient) GetSemesterById(ctx context.Context, id string) (*pb.GetSemesterResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", semesterCachePrefix, id)
	var cached pb.GetSemesterResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for semester: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for semester: %s", id)
	resp, err := g.client.GetSemester(ctx, &pb.GetSemesterRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, g.redisClient, cacheKey, resp, semesterCacheTTL)
	return resp, nil
}

func (g *GRPCAcadamicClient) CreateSemester(ctx context.Context, req *pb.CreateSemesterRequest) (*pb.CreateSemesterResponse, error) {
	resp, err := g.client.CreateSemester(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	InvalidateCacheByPattern(ctx, g.redisClient, semesterCachePrefix+"*")

	return resp, nil
}

func (g *GRPCAcadamicClient) UpdateSemester(ctx context.Context, req *pb.UpdateSemesterRequest) (*pb.UpdateSemesterResponse, error) {
	resp, err := g.client.UpdateSemester(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", semesterCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, g.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, g.redisClient, semesterCachePrefix+"*")
	}

	return resp, nil
}

func (g *GRPCAcadamicClient) DeleteSemester(ctx context.Context, id string) (*pb.DeleteSemesterResponse, error) {
	resp, err := g.client.DeleteSemester(ctx, &pb.DeleteSemesterRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", semesterCachePrefix, id)
	InvalidateCacheByKey(ctx, g.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, g.redisClient, semesterCachePrefix+"*")

	return resp, nil
}

func (g *GRPCAcadamicClient) GetSemestersByIds(ctx context.Context, ids []string) (*pb.ListSemestersResponse, error) {
	if len(ids) == 0 {
		return &pb.ListSemestersResponse{Semesters: []*pb.Semester{}}, nil
	}

	result := &pb.ListSemestersResponse{Semesters: []*pb.Semester{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", semesterCachePrefix, id)
		var cached pb.GetSemesterResponse

		if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
			if cached.Semester != nil {
				result.Semesters = append(result.Semesters, cached.Semester)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetSemestersByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := g.client.ListSemesters(ctx, &pb.ListSemestersRequest{
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
		if resp != nil && resp.Semesters != nil {
			for _, semester := range resp.Semesters {
				if semester != nil {
					cacheKey := fmt.Sprintf("%s%s", semesterCachePrefix, semester.Id)
					SetCachedProto(ctx, g.redisClient, cacheKey, &pb.GetSemesterResponse{Semester: semester}, semesterCacheTTL)
					result.Semesters = append(result.Semesters, semester)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// FACULTY METHODS
// ============================================

func (g *GRPCAcadamicClient) GetFacultiesBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListFacultiesResponse, error) {
	cacheKey := GenerateCacheKey(majorCachePrefix, search)
	var cached pb.ListFacultiesResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for major search")
		return &cached, nil
	}

	log.Printf("Cache MISS for major search")
	resp, err := g.client.ListFaculties(ctx, &pb.ListFacultiesRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, g.redisClient, cacheKey, resp, majorCacheTTL)
	return resp, nil
}

func (g *GRPCAcadamicClient) GetFacultyById(ctx context.Context, id string) (*pb.GetFacultyResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", facultyCachePrefix, id)
	var cached pb.GetFacultyResponse
	if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for Faculty: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for Faculty: %s", id)
	resp, err := g.client.GetFaculty(ctx, &pb.GetFacultyRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, g.redisClient, cacheKey, resp, facultyCacheTTL)
	return resp, nil
}

func (g *GRPCAcadamicClient) CreateFaculty(ctx context.Context, req *pb.CreateFacultyRequest) (*pb.CreateFacultyResponse, error) {
	resp, err := g.client.CreateFaculty(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	InvalidateCacheByPattern(ctx, g.redisClient, facultyCachePrefix+"*")

	return resp, nil
}

func (g *GRPCAcadamicClient) UpdateFaculty(ctx context.Context, req *pb.UpdateFacultyRequest) (*pb.UpdateFacultyResponse, error) {
	resp, err := g.client.UpdateFaculty(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", facultyCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, g.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, g.redisClient, facultyCachePrefix+"*")
	}

	return resp, nil
}

func (g *GRPCAcadamicClient) DeleteFaculty(ctx context.Context, id string) (*pb.DeleteFacultyResponse, error) {
	resp, err := g.client.DeleteFaculty(ctx, &pb.DeleteFacultyRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", facultyCachePrefix, id)
	InvalidateCacheByKey(ctx, g.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, g.redisClient, facultyCachePrefix+"*")

	return resp, nil
}

func (g *GRPCAcadamicClient) GetFacultiesByIds(ctx context.Context, ids []string) (*pb.ListFacultiesResponse, error) {
	if len(ids) == 0 {
		return &pb.ListFacultiesResponse{Faculties: []*pb.Faculty{}}, nil
	}

	result := &pb.ListFacultiesResponse{Faculties: []*pb.Faculty{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", majorCachePrefix, id)
		var cached pb.GetFacultyResponse

		if hit, _ := GetCachedProto(ctx, g.redisClient, cacheKey, &cached); hit {
			if cached.Faculty != nil {
				result.Faculties = append(result.Faculties, cached.Faculty)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetFacultiesByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := g.client.ListFaculties(ctx, &pb.ListFacultiesRequest{
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
		if resp != nil && resp.Faculties != nil {
			for _, faculty := range resp.Faculties {
				if faculty != nil {
					cacheKey := fmt.Sprintf("%s%s", facultyCachePrefix, faculty.Id)
					SetCachedProto(ctx, g.redisClient, cacheKey, &pb.GetFacultyResponse{Faculty: faculty}, facultyCacheTTL)
					result.Faculties = append(result.Faculties, faculty)
				}
			}
		}
	}

	return result, nil
}
