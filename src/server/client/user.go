package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/user"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCUser struct {
	conn        *grpc.ClientConn
	client      pb.UserServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations for User
	studentCacheTTL = 5 * time.Minute
	teacherCacheTTL = 5 * time.Minute

	// Cache key prefixes for User
	studentCachePrefix        = "user:student:"
	teacherCachePrefix        = "user:teacher:"
	studentEmailCachePrefix   = "user:student:email:"
	teacherEmailCachePrefix   = "user:teacher:email:"
	studentIdCachePrefix      = "user:student:id:"
	teacherIdCachePrefix      = "user:teacher:id:"
)

func NewGRPCUser(addr string, redisClient *redis.Client) (*GRPCUser, error) {
	fmt.Println(addr)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewUserServiceClient(conn)
	return &GRPCUser{
		conn:        conn,
		client:      client,
		redisClient: redisClient,
	}, nil
}

// InvalidateStudentCache invalidates student-related cache
func (u *GRPCUser) InvalidateStudentCache(ctx context.Context, studentCode string) error {
	pattern := fmt.Sprintf("%s*%s*", studentCachePrefix, studentCode)
	return InvalidateCacheByPattern(ctx, u.redisClient, pattern)
}

// InvalidateTeacherCache invalidates teacher-related cache
func (u *GRPCUser) InvalidateTeacherCache(ctx context.Context, teacherCode string) error {
	pattern := fmt.Sprintf("%s*%s*", teacherCachePrefix, teacherCode)
	return InvalidateCacheByPattern(ctx, u.redisClient, pattern)
}

// InvalidateAllStudentCache invalidates all student-related caches
func (u *GRPCUser) InvalidateAllStudentCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", studentCachePrefix)
	return InvalidateCacheByPattern(ctx, u.redisClient, pattern)
}

// InvalidateAllTeacherCache invalidates all teacher-related caches
func (u *GRPCUser) InvalidateAllTeacherCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", teacherCachePrefix)
	return InvalidateCacheByPattern(ctx, u.redisClient, pattern)
}

func (u *GRPCUser) GetUserByEmail(ctx context.Context, email string) (*pb.ListStudentsResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", studentEmailCachePrefix, email)

	// Try to get from cache
	var cached pb.ListStudentsResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for student email: %s", email)
		return &cached, nil
	}

	log.Printf("Cache MISS for student email: %s", email)

	// Cache miss - fetch from database
	resp, err := u.client.ListStudents(ctx, &pb.ListStudentsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   20,
				SortBy:     "semester_code",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "email",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{email},
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
	SetCachedProto(ctx, u.redisClient, cacheKey, resp, studentCacheTTL)

	return resp, nil
}

func (u *GRPCUser) GetUserByEmailAndSemester(ctx context.Context, email string, semester string) (*pb.ListStudentsResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}

	// Generate cache key
	cacheKey := GenerateCacheKey(studentEmailCachePrefix, map[string]interface{}{
		"email":    email,
		"semester": semester,
	})

	// Try to get from cache
	var cached pb.ListStudentsResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for student email and semester: %s, %s", email, semester)
		return &cached, nil
	}

	log.Printf("Cache MISS for student email and semester: %s, %s", email, semester)

	// Cache miss - fetch from database
	resp, err := u.client.ListStudents(ctx, &pb.ListStudentsRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   20,
				SortBy:     "semester_code",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "email",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{email},
						},
					},
				},
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "semester_code",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{semester},
						},
					}},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, u.redisClient, cacheKey, resp, studentCacheTTL)

	return resp, nil
}

func (u *GRPCUser) GetUserById(ctx context.Context, id string) (*pb.GetStudentResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", studentIdCachePrefix, id)

	// Try to get from cache
	var cached pb.GetStudentResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for student id: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for student id: %s", id)

	// Cache miss - fetch from database
	resp, err := u.client.GetStudent(ctx, &pb.GetStudentRequest{
		Id: id,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, u.redisClient, cacheKey, resp, studentCacheTTL)

	return resp, nil
}

func (u *GRPCUser) GetTeacherById(ctx context.Context, id string) (*pb.GetTeacherResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", teacherIdCachePrefix, id)

	// Try to get from cache
	var cached pb.GetTeacherResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for teacher id: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for teacher id: %s", id)

	// Cache miss - fetch from database
	resp, err := u.client.GetTeacher(ctx, &pb.GetTeacherRequest{
		Id: id,
	})

	if err != nil {
		return nil, err
	}

	// Store in cache
	SetCachedProto(ctx, u.redisClient, cacheKey, resp, teacherCacheTTL)

	return resp, nil
}

func (u *GRPCUser) GetTeacherByEmail(ctx context.Context, email string) (*pb.ListTeachersResponse, error) {
	if u == nil {
		return nil, fmt.Errorf("GRPCUser is nil")
	}
	if u.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s%s", teacherEmailCachePrefix, email)

	// Try to get from cache
	var cached pb.ListTeachersResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for teacher email: %s", email)
		return &cached, nil
	}

	log.Printf("Cache MISS for teacher email: %s", email)

	// Cache miss - fetch from database
	resp, err := u.client.ListTeachers(ctx, &pb.ListTeachersRequest{
		Search: &pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Descending: true,
				Page:       1,
				PageSize:   20,
				SortBy:     "semester_code",
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "email",
							Operator: pbCommon.FilterOperator_EQUAL,
							Values:   []string{email},
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
	SetCachedProto(ctx, u.redisClient, cacheKey, resp, teacherCacheTTL)

	return resp, nil
}
