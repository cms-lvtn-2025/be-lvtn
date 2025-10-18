package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/user"
	"thaily/src/pkg/tls"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type GRPCUser struct {
	conn        *grpc.ClientConn
	client      pb.UserServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations
	studentCacheTTL = 10 * time.Minute
	teacherCacheTTL = 10 * time.Minute

	// Cache key prefixes
	studentCachePrefix = "user:student:"
	teacherCachePrefix = "user:teacher:"
)

func NewGRPCUser(addr string, redisClient *redis.Client) (*GRPCUser, error) {
	// Load mTLS credentials
	creds, err := tls.LoadClientTLSCredentials("user-service")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
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

// ============================================
// STUDENT METHODS
// ============================================

func (u *GRPCUser) GetStudentsBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListStudentsResponse, error) {
	cacheKey := GenerateCacheKey(studentCachePrefix, search)
	var cached pb.ListStudentsResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for student search")
		return &cached, nil
	}

	log.Printf("Cache MISS for student search")
	resp, err := u.client.ListStudents(ctx, &pb.ListStudentsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, u.redisClient, cacheKey, resp, studentCacheTTL)
	return resp, nil
}

func (u *GRPCUser) GetUserById(ctx context.Context, id string) (*pb.GetStudentResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", studentCachePrefix, id)
	var cached pb.GetStudentResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for student: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for student: %s", id)
	resp, err := u.client.GetStudent(ctx, &pb.GetStudentRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, u.redisClient, cacheKey, resp, studentCacheTTL)
	return resp, nil
}

func (u *GRPCUser) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentResponse, error) {
	resp, err := u.client.UpdateStudent(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", studentCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, u.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, u.redisClient, studentCachePrefix+"*")
	}

	return resp, nil
}

func (u *GRPCUser) DeleteStudent(ctx context.Context, id string) (*pb.DeleteStudentResponse, error) {
	resp, err := u.client.DeleteStudent(ctx, &pb.DeleteStudentRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", studentCachePrefix, id)
	InvalidateCacheByKey(ctx, u.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, u.redisClient, studentCachePrefix+"*")

	return resp, nil
}

func (u *GRPCUser) GetStudentsByIds(ctx context.Context, ids []string) (*pb.ListStudentsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListStudentsResponse{Students: []*pb.Student{}}, nil
	}

	result := &pb.ListStudentsResponse{Students: []*pb.Student{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", studentCachePrefix, id)
		var cached pb.GetStudentResponse

		if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
			if cached.Student != nil {
				result.Students = append(result.Students, cached.Student)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetStudentsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := u.client.ListStudents(ctx, &pb.ListStudentsRequest{
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
		if resp != nil && resp.Students != nil {
			for _, student := range resp.Students {
				if student != nil {
					cacheKey := fmt.Sprintf("%s%s", studentCachePrefix, student.Id)
					SetCachedProto(ctx, u.redisClient, cacheKey, &pb.GetStudentResponse{Student: student}, studentCacheTTL)
					result.Students = append(result.Students, student)
				}
			}
		}
	}

	return result, nil
}

// GetUserByEmail gets students by email (returns list because one email can have multiple semester records)
func (u *GRPCUser) GetUserByEmail(ctx context.Context, email string) (*pb.ListStudentsResponse, error) {
	cacheKey := fmt.Sprintf("%semail:%s", studentCachePrefix, email)
	var cached pb.ListStudentsResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for student email: %s", email)
		return &cached, nil
	}

	log.Printf("Cache MISS for student email: %s", email)
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

	SetCachedProto(ctx, u.redisClient, cacheKey, resp, studentCacheTTL)
	return resp, nil
}

// GetUserByEmailAndSemester gets a specific student by email and semester
func (u *GRPCUser) GetUserByEmailAndSemester(ctx context.Context, email string, semester string) (*pb.ListStudentsResponse, error) {
	cacheKey := GenerateCacheKey(studentCachePrefix, map[string]interface{}{
		"email":    email,
		"semester": semester,
	})

	var cached pb.ListStudentsResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for student email and semester: %s, %s", email, semester)
		return &cached, nil
	}

	log.Printf("Cache MISS for student email and semester: %s, %s", email, semester)
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
					},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, u.redisClient, cacheKey, resp, studentCacheTTL)
	return resp, nil
}

// GetStudentsByEmails fetches multiple students by email addresses
func (u *GRPCUser) GetStudentsByEmails(ctx context.Context, emails []string) (*pb.ListStudentsResponse, error) {
	if len(emails) == 0 {
		return &pb.ListStudentsResponse{Students: []*pb.Student{}}, nil
	}

	result := &pb.ListStudentsResponse{Students: []*pb.Student{}}
	missingEmails := []string{}
	cacheHits := 0

	// Check Redis cache for each email
	for _, email := range emails {
		cacheKey := fmt.Sprintf("%semail:%s", studentCachePrefix, email)
		var cached pb.ListStudentsResponse

		if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
			if cached.Students != nil && len(cached.Students) > 0 {
				result.Students = append(result.Students, cached.Students...)
				cacheHits++
			} else {
				missingEmails = append(missingEmails, email)
			}
		} else {
			missingEmails = append(missingEmails, email)
		}
	}

	log.Printf("[GetStudentsByEmails] Total: %d, Cache hits: %d, Database queries needed: %d", len(emails), cacheHits, len(missingEmails))

	// Fetch missing emails from database
	if len(missingEmails) > 0 {
		resp, err := u.client.ListStudents(ctx, &pb.ListStudentsRequest{
			Search: &pbCommon.SearchRequest{
				Pagination: &pbCommon.Pagination{
					Descending: false,
					Page:       1,
					PageSize:   int32(len(missingEmails) * 10), // Each email might have multiple semesters
					SortBy:     "email",
				},
				Filters: []*pbCommon.FilterCriteria{
					{
						Criteria: &pbCommon.FilterCriteria_Condition{
							Condition: &pbCommon.FilterCondition{
								Field:    "email",
								Operator: pbCommon.FilterOperator_IN,
								Values:   missingEmails,
							},
						},
					},
				},
			},
		})

		if err != nil {
			return nil, err
		}

		// Group students by email and store to Redis
		if resp != nil && resp.Students != nil {
			emailMap := make(map[string][]*pb.Student)
			for _, student := range resp.Students {
				if student != nil {
					emailMap[student.Email] = append(emailMap[student.Email], student)
					result.Students = append(result.Students, student)
				}
			}

			// Cache each email group
			for email, students := range emailMap {
				cacheKey := fmt.Sprintf("%semail:%s", studentCachePrefix, email)
				SetCachedProto(ctx, u.redisClient, cacheKey, &pb.ListStudentsResponse{Students: students}, studentCacheTTL)
			}
		}
	}

	return result, nil
}

// ============================================
// TEACHER METHODS
// ============================================

func (u *GRPCUser) GetTeachersBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListTeachersResponse, error) {
	cacheKey := GenerateCacheKey(teacherCachePrefix, search)
	var cached pb.ListTeachersResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for teacher search")
		return &cached, nil
	}

	log.Printf("Cache MISS for teacher search")
	resp, err := u.client.ListTeachers(ctx, &pb.ListTeachersRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, u.redisClient, cacheKey, resp, teacherCacheTTL)
	return resp, nil
}

func (u *GRPCUser) GetTeacherById(ctx context.Context, id string) (*pb.GetTeacherResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", teacherCachePrefix, id)
	var cached pb.GetTeacherResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for teacher: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for teacher: %s", id)
	resp, err := u.client.GetTeacher(ctx, &pb.GetTeacherRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, u.redisClient, cacheKey, resp, teacherCacheTTL)
	return resp, nil
}

func (u *GRPCUser) UpdateTeacher(ctx context.Context, req *pb.UpdateTeacherRequest) (*pb.UpdateTeacherResponse, error) {
	resp, err := u.client.UpdateTeacher(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", teacherCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, u.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, u.redisClient, teacherCachePrefix+"*")
	}

	return resp, nil
}

func (u *GRPCUser) DeleteTeacher(ctx context.Context, id string) (*pb.DeleteTeacherResponse, error) {
	resp, err := u.client.DeleteTeacher(ctx, &pb.DeleteTeacherRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", teacherCachePrefix, id)
	InvalidateCacheByKey(ctx, u.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, u.redisClient, teacherCachePrefix+"*")

	return resp, nil
}

func (u *GRPCUser) GetTeachersByIds(ctx context.Context, ids []string) (*pb.ListTeachersResponse, error) {
	if len(ids) == 0 {
		return &pb.ListTeachersResponse{Teachers: []*pb.Teacher{}}, nil
	}

	result := &pb.ListTeachersResponse{Teachers: []*pb.Teacher{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", teacherCachePrefix, id)
		var cached pb.GetTeacherResponse

		if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
			if cached.Teacher != nil {
				result.Teachers = append(result.Teachers, cached.Teacher)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetTeachersByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := u.client.ListTeachers(ctx, &pb.ListTeachersRequest{
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
		if resp != nil && resp.Teachers != nil {
			for _, teacher := range resp.Teachers {
				if teacher != nil {
					cacheKey := fmt.Sprintf("%s%s", teacherCachePrefix, teacher.Id)
					SetCachedProto(ctx, u.redisClient, cacheKey, &pb.GetTeacherResponse{Teacher: teacher}, teacherCacheTTL)
					result.Teachers = append(result.Teachers, teacher)
				}
			}
		}
	}

	return result, nil
}

// GetTeacherByEmail gets teachers by email (returns list because one email can have multiple semester records)
func (u *GRPCUser) GetTeacherByEmail(ctx context.Context, email string) (*pb.ListTeachersResponse, error) {
	cacheKey := fmt.Sprintf("%semail:%s", teacherCachePrefix, email)
	var cached pb.ListTeachersResponse
	if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for teacher email: %s", email)
		return &cached, nil
	}

	log.Printf("Cache MISS for teacher email: %s", email)
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

	SetCachedProto(ctx, u.redisClient, cacheKey, resp, teacherCacheTTL)
	return resp, nil
}

// GetTeachersByEmails fetches multiple teachers by email addresses
func (u *GRPCUser) GetTeachersByEmails(ctx context.Context, emails []string) (*pb.ListTeachersResponse, error) {
	if len(emails) == 0 {
		return &pb.ListTeachersResponse{Teachers: []*pb.Teacher{}}, nil
	}

	result := &pb.ListTeachersResponse{Teachers: []*pb.Teacher{}}
	missingEmails := []string{}
	cacheHits := 0

	// Check Redis cache for each email
	for _, email := range emails {
		cacheKey := fmt.Sprintf("%semail:%s", teacherCachePrefix, email)
		var cached pb.ListTeachersResponse

		if hit, _ := GetCachedProto(ctx, u.redisClient, cacheKey, &cached); hit {
			if cached.Teachers != nil && len(cached.Teachers) > 0 {
				result.Teachers = append(result.Teachers, cached.Teachers...)
				cacheHits++
			} else {
				missingEmails = append(missingEmails, email)
			}
		} else {
			missingEmails = append(missingEmails, email)
		}
	}

	log.Printf("[GetTeachersByEmails] Total: %d, Cache hits: %d, Database queries needed: %d", len(emails), cacheHits, len(missingEmails))

	// Fetch missing emails from database
	if len(missingEmails) > 0 {
		resp, err := u.client.ListTeachers(ctx, &pb.ListTeachersRequest{
			Search: &pbCommon.SearchRequest{
				Pagination: &pbCommon.Pagination{
					Descending: false,
					Page:       1,
					PageSize:   int32(len(missingEmails) * 10), // Each email might have multiple semesters
					SortBy:     "email",
				},
				Filters: []*pbCommon.FilterCriteria{
					{
						Criteria: &pbCommon.FilterCriteria_Condition{
							Condition: &pbCommon.FilterCondition{
								Field:    "email",
								Operator: pbCommon.FilterOperator_IN,
								Values:   missingEmails,
							},
						},
					},
				},
			},
		})

		if err != nil {
			return nil, err
		}

		// Group teachers by email and store to Redis
		if resp != nil && resp.Teachers != nil {
			emailMap := make(map[string][]*pb.Teacher)
			for _, teacher := range resp.Teachers {
				if teacher != nil {
					emailMap[teacher.Email] = append(emailMap[teacher.Email], teacher)
					result.Teachers = append(result.Teachers, teacher)
				}
			}

			// Cache each email group
			for email, teachers := range emailMap {
				cacheKey := fmt.Sprintf("%semail:%s", teacherCachePrefix, email)
				SetCachedProto(ctx, u.redisClient, cacheKey, &pb.ListTeachersResponse{Teachers: teachers}, teacherCacheTTL)
			}
		}
	}

	return result, nil
}
