package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pbCommon "thaily/proto/common"
	pb "thaily/proto/thesis"
	"thaily/src/pkg/tls"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type GRPCthesis struct {
	conn        *grpc.ClientConn
	client      pb.ThesisServiceClient
	redisClient *redis.Client
}

const (
	// Cache TTL configurations
	topicCacheTTL               = 5 * time.Minute
	enrollmentCacheTTL          = 2 * time.Minute
	midtermCacheTTL             = 10 * time.Minute
	finalCacheTTL               = 10 * time.Minute
	gradeReviewCacheTTL         = 10 * time.Minute
	topicCouncilCacheTTL        = 5 * time.Minute
	topicCouncilSupervisorCacheTTL = 5 * time.Minute

	// Cache key prefixes
	topicCachePrefix               = "thesis:topic:"
	enrollmentCachePrefix          = "thesis:enrollment:"
	midtermCachePrefix             = "thesis:midterm:"
	finalCachePrefix               = "thesis:final:"
	gradeReviewCachePrefix         = "thesis:grade_review:"
	topicCouncilCachePrefix        = "thesis:topic_council:"
	topicCouncilSupervisorCachePrefix = "thesis:topic_council_supervisor:"
)

func NewGRPCthesis(addr string, redisClient *redis.Client) (*GRPCthesis, error) {
	// Load mTLS credentials
	creds, err := tls.LoadClientTLSCredentials("thesis-service")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	client := pb.NewThesisServiceClient(conn)
	return &GRPCthesis{
		conn:        conn,
		client:      client,
		redisClient: redisClient,
	}, nil
}

// ============================================
// TOPIC METHODS
// ============================================

func (t *GRPCthesis) GetTopicBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListTopicsResponse, error) {
	cacheKey := GenerateCacheKey(topicCachePrefix, search)
	var cached pb.ListTopicsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic search")
		return &cached, nil
	}

	log.Printf("Cache MISS for topic search")
	resp, err := t.client.ListTopics(ctx, &pb.ListTopicsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, topicCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) GetTopicById(ctx context.Context, id string) (*pb.GetTopicResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", topicCachePrefix, id)
	var cached pb.GetTopicResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for topic: %s", id)
	resp, err := t.client.GetTopic(ctx, &pb.GetTopicRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, topicCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) UpdateTopic(ctx context.Context, req *pb.UpdateTopicRequest) (*pb.UpdateTopicResponse, error) {
	resp, err := t.client.UpdateTopic(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", topicCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, t.redisClient, topicCachePrefix+"*")
	}

	return resp, nil
}

func (t *GRPCthesis) DeleteTopic(ctx context.Context, id string) (*pb.DeleteTopicResponse, error) {
	resp, err := t.client.DeleteTopic(ctx, &pb.DeleteTopicRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", topicCachePrefix, id)
	InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, t.redisClient, topicCachePrefix+"*")

	return resp, nil
}

func (t *GRPCthesis) GetTopicsByIds(ctx context.Context, ids []string) (*pb.ListTopicsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListTopicsResponse{Topics: []*pb.Topic{}}, nil
	}

	result := &pb.ListTopicsResponse{Topics: []*pb.Topic{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", topicCachePrefix, id)
		var cached pb.GetTopicResponse

		if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
			if cached.Topic != nil {
				result.Topics = append(result.Topics, cached.Topic)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetTopicsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := t.client.ListTopics(ctx, &pb.ListTopicsRequest{
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
		if resp != nil && resp.Topics != nil {
			for _, topic := range resp.Topics {
				if topic != nil {
					cacheKey := fmt.Sprintf("%s%s", topicCachePrefix, topic.Id)
					SetCachedProto(ctx, t.redisClient, cacheKey, &pb.GetTopicResponse{Topic: topic}, topicCacheTTL)
					result.Topics = append(result.Topics, topic)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// ENROLLMENT METHODS
// ============================================

func (t *GRPCthesis) GetEnrollmentBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListEnrollmentsResponse, error) {
	cacheKey := GenerateCacheKey(enrollmentCachePrefix, search)
	var cached pb.ListEnrollmentsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for enrollment search")
		return &cached, nil
	}

	log.Printf("Cache MISS for enrollment search")
	resp, err := t.client.ListEnrollments(ctx, &pb.ListEnrollmentsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, enrollmentCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) GetEnrollmentById(ctx context.Context, id string) (*pb.GetEnrollmentResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", enrollmentCachePrefix, id)
	var cached pb.GetEnrollmentResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for enrollment: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for enrollment: %s", id)
	resp, err := t.client.GetEnrollment(ctx, &pb.GetEnrollmentRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, enrollmentCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) UpdateEnrollment(ctx context.Context, req *pb.UpdateEnrollmentRequest) (*pb.UpdateEnrollmentResponse, error) {
	resp, err := t.client.UpdateEnrollment(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", enrollmentCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, t.redisClient, enrollmentCachePrefix+"*")
	}

	return resp, nil
}

func (t *GRPCthesis) DeleteEnrollment(ctx context.Context, id string) (*pb.DeleteEnrollmentResponse, error) {
	resp, err := t.client.DeleteEnrollment(ctx, &pb.DeleteEnrollmentRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", enrollmentCachePrefix, id)
	InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, t.redisClient, enrollmentCachePrefix+"*")

	return resp, nil
}

func (t *GRPCthesis) GetEnrollmentsByIds(ctx context.Context, ids []string) (*pb.ListEnrollmentsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListEnrollmentsResponse{Enrollments: []*pb.Enrollment{}}, nil
	}

	result := &pb.ListEnrollmentsResponse{Enrollments: []*pb.Enrollment{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", enrollmentCachePrefix, id)
		var cached pb.GetEnrollmentResponse

		if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
			if cached.Enrollment != nil {
				result.Enrollments = append(result.Enrollments, cached.Enrollment)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetEnrollmentsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := t.client.ListEnrollments(ctx, &pb.ListEnrollmentsRequest{
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
		if resp != nil && resp.Enrollments != nil {
			for _, enrollment := range resp.Enrollments {
				if enrollment != nil {
					cacheKey := fmt.Sprintf("%s%s", enrollmentCachePrefix, enrollment.Id)
					SetCachedProto(ctx, t.redisClient, cacheKey, &pb.GetEnrollmentResponse{Enrollment: enrollment}, enrollmentCacheTTL)
					result.Enrollments = append(result.Enrollments, enrollment)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// MIDTERM METHODS
// ============================================

func (t *GRPCthesis) GetMidtermBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListMidtermsResponse, error) {
	cacheKey := GenerateCacheKey(midtermCachePrefix, search)
	var cached pb.ListMidtermsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for midterm search")
		return &cached, nil
	}

	log.Printf("Cache MISS for midterm search")
	resp, err := t.client.ListMidterms(ctx, &pb.ListMidtermsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, midtermCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) GetMidtermById(ctx context.Context, id string) (*pb.GetMidtermResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, id)
	var cached pb.GetMidtermResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for midterm: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for midterm: %s", id)
	resp, err := t.client.GetMidterm(ctx, &pb.GetMidtermRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, midtermCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) UpdateMidterm(ctx context.Context, req *pb.UpdateMidtermRequest) (*pb.UpdateMidtermResponse, error) {
	resp, err := t.client.UpdateMidterm(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, t.redisClient, midtermCachePrefix+"*")
	}

	return resp, nil
}

func (t *GRPCthesis) DeleteMidterm(ctx context.Context, id string) (*pb.DeleteMidtermResponse, error) {
	resp, err := t.client.DeleteMidterm(ctx, &pb.DeleteMidtermRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, id)
	InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, t.redisClient, midtermCachePrefix+"*")

	return resp, nil
}

func (t *GRPCthesis) GetMidtermsByIds(ctx context.Context, ids []string) (*pb.ListMidtermsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListMidtermsResponse{Midterms: []*pb.Midterm{}}, nil
	}

	result := &pb.ListMidtermsResponse{Midterms: []*pb.Midterm{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, id)
		var cached pb.GetMidtermResponse

		if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
			if cached.Midterm != nil {
				result.Midterms = append(result.Midterms, cached.Midterm)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetMidtermsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := t.client.ListMidterms(ctx, &pb.ListMidtermsRequest{
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
		if resp != nil && resp.Midterms != nil {
			for _, midterm := range resp.Midterms {
				if midterm != nil {
					cacheKey := fmt.Sprintf("%s%s", midtermCachePrefix, midterm.Id)
					SetCachedProto(ctx, t.redisClient, cacheKey, &pb.GetMidtermResponse{Midterm: midterm}, midtermCacheTTL)
					result.Midterms = append(result.Midterms, midterm)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// FINAL METHODS
// ============================================

func (t *GRPCthesis) GetFinalBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListFinalsResponse, error) {
	cacheKey := GenerateCacheKey(finalCachePrefix, search)
	var cached pb.ListFinalsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for final search")
		return &cached, nil
	}

	log.Printf("Cache MISS for final search")
	resp, err := t.client.ListFinals(ctx, &pb.ListFinalsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, finalCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) GetFinalById(ctx context.Context, id string) (*pb.GetFinalResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, id)
	var cached pb.GetFinalResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for final: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for final: %s", id)
	resp, err := t.client.GetFinal(ctx, &pb.GetFinalRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, finalCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) UpdateFinal(ctx context.Context, req *pb.UpdateFinalRequest) (*pb.UpdateFinalResponse, error) {
	resp, err := t.client.UpdateFinal(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, t.redisClient, finalCachePrefix+"*")
	}

	return resp, nil
}

func (t *GRPCthesis) DeleteFinal(ctx context.Context, id string) (*pb.DeleteFinalResponse, error) {
	resp, err := t.client.DeleteFinal(ctx, &pb.DeleteFinalRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, id)
	InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, t.redisClient, finalCachePrefix+"*")

	return resp, nil
}

func (t *GRPCthesis) GetFinalsByIds(ctx context.Context, ids []string) (*pb.ListFinalsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListFinalsResponse{Finals: []*pb.Final{}}, nil
	}

	result := &pb.ListFinalsResponse{Finals: []*pb.Final{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, id)
		var cached pb.GetFinalResponse

		if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
			if cached.Final != nil {
				result.Finals = append(result.Finals, cached.Final)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetFinalsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := t.client.ListFinals(ctx, &pb.ListFinalsRequest{
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
		if resp != nil && resp.Finals != nil {
			for _, final := range resp.Finals {
				if final != nil {
					cacheKey := fmt.Sprintf("%s%s", finalCachePrefix, final.Id)
					SetCachedProto(ctx, t.redisClient, cacheKey, &pb.GetFinalResponse{Final: final}, finalCacheTTL)
					result.Finals = append(result.Finals, final)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// GRADE REVIEW METHODS
// ============================================

func (t *GRPCthesis) GetGradeReviewBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListGradeReviewsResponse, error) {
	cacheKey := GenerateCacheKey(gradeReviewCachePrefix, search)
	var cached pb.ListGradeReviewsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for grade review search")
		return &cached, nil
	}

	log.Printf("Cache MISS for grade review search")
	resp, err := t.client.ListGradeReviews(ctx, &pb.ListGradeReviewsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, gradeReviewCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) GetGradeReviewById(ctx context.Context, id string) (*pb.GetGradeReviewResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", gradeReviewCachePrefix, id)
	var cached pb.GetGradeReviewResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for grade review: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for grade review: %s", id)
	resp, err := t.client.GetGradeReview(ctx, &pb.GetGradeReviewRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, gradeReviewCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) UpdateGradeReview(ctx context.Context, req *pb.UpdateGradeReviewRequest) (*pb.UpdateGradeReviewResponse, error) {
	resp, err := t.client.UpdateGradeReview(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", gradeReviewCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, t.redisClient, gradeReviewCachePrefix+"*")
	}

	return resp, nil
}

func (t *GRPCthesis) DeleteGradeReview(ctx context.Context, id string) (*pb.DeleteGradeReviewResponse, error) {
	resp, err := t.client.DeleteGradeReview(ctx, &pb.DeleteGradeReviewRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", gradeReviewCachePrefix, id)
	InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, t.redisClient, gradeReviewCachePrefix+"*")

	return resp, nil
}

func (t *GRPCthesis) GetGradeReviewsByIds(ctx context.Context, ids []string) (*pb.ListGradeReviewsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListGradeReviewsResponse{GradeReviews: []*pb.GradeReview{}}, nil
	}

	result := &pb.ListGradeReviewsResponse{GradeReviews: []*pb.GradeReview{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", gradeReviewCachePrefix, id)
		var cached pb.GetGradeReviewResponse

		if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
			if cached.GradeReview != nil {
				result.GradeReviews = append(result.GradeReviews, cached.GradeReview)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetGradeReviewsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := t.client.ListGradeReviews(ctx, &pb.ListGradeReviewsRequest{
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
		if resp != nil && resp.GradeReviews != nil {
			for _, gradeReview := range resp.GradeReviews {
				if gradeReview != nil {
					cacheKey := fmt.Sprintf("%s%s", gradeReviewCachePrefix, gradeReview.Id)
					SetCachedProto(ctx, t.redisClient, cacheKey, &pb.GetGradeReviewResponse{GradeReview: gradeReview}, gradeReviewCacheTTL)
					result.GradeReviews = append(result.GradeReviews, gradeReview)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// TOPIC COUNCIL METHODS
// ============================================

func (t *GRPCthesis) GetTopicCouncilBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListTopicCouncilsResponse, error) {
	cacheKey := GenerateCacheKey(topicCouncilCachePrefix, search)
	var cached pb.ListTopicCouncilsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic council search")
		return &cached, nil
	}

	log.Printf("Cache MISS for topic council search")
	resp, err := t.client.ListTopicCouncils(ctx, &pb.ListTopicCouncilsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, topicCouncilCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) GetTopicCouncilById(ctx context.Context, id string) (*pb.GetTopicCouncilResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", topicCouncilCachePrefix, id)
	var cached pb.GetTopicCouncilResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic council: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for topic council: %s", id)
	resp, err := t.client.GetTopicCouncil(ctx, &pb.GetTopicCouncilRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, topicCouncilCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) UpdateTopicCouncil(ctx context.Context, req *pb.UpdateTopicCouncilRequest) (*pb.UpdateTopicCouncilResponse, error) {
	resp, err := t.client.UpdateTopicCouncil(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", topicCouncilCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, t.redisClient, topicCouncilCachePrefix+"*")
	}

	return resp, nil
}

func (t *GRPCthesis) DeleteTopicCouncil(ctx context.Context, id string) (*pb.DeleteTopicCouncilResponse, error) {
	resp, err := t.client.DeleteTopicCouncil(ctx, &pb.DeleteTopicCouncilRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", topicCouncilCachePrefix, id)
	InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, t.redisClient, topicCouncilCachePrefix+"*")

	return resp, nil
}

func (t *GRPCthesis) GetTopicCouncilsByIds(ctx context.Context, ids []string) (*pb.ListTopicCouncilsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListTopicCouncilsResponse{TopicCouncils: []*pb.TopicCouncil{}}, nil
	}

	result := &pb.ListTopicCouncilsResponse{TopicCouncils: []*pb.TopicCouncil{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", topicCouncilCachePrefix, id)
		var cached pb.GetTopicCouncilResponse

		if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
			if cached.TopicCouncil != nil {
				result.TopicCouncils = append(result.TopicCouncils, cached.TopicCouncil)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetTopicCouncilsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := t.client.ListTopicCouncils(ctx, &pb.ListTopicCouncilsRequest{
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
		if resp != nil && resp.TopicCouncils != nil {
			for _, topicCouncil := range resp.TopicCouncils {
				if topicCouncil != nil {
					cacheKey := fmt.Sprintf("%s%s", topicCouncilCachePrefix, topicCouncil.Id)
					SetCachedProto(ctx, t.redisClient, cacheKey, &pb.GetTopicCouncilResponse{TopicCouncil: topicCouncil}, topicCouncilCacheTTL)
					result.TopicCouncils = append(result.TopicCouncils, topicCouncil)
				}
			}
		}
	}

	return result, nil
}

// ============================================
// TOPIC COUNCIL SUPERVISOR METHODS
// ============================================

func (t *GRPCthesis) GetTopicCouncilSupervisorBySearch(ctx context.Context, search *pbCommon.SearchRequest) (*pb.ListTopicCouncilSupervisorsResponse, error) {
	cacheKey := GenerateCacheKey(topicCouncilSupervisorCachePrefix, search)
	var cached pb.ListTopicCouncilSupervisorsResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic council supervisor search")
		return &cached, nil
	}

	log.Printf("Cache MISS for topic council supervisor search")
	resp, err := t.client.ListTopicCouncilSupervisors(ctx, &pb.ListTopicCouncilSupervisorsRequest{Search: search})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, topicCouncilSupervisorCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) GetTopicCouncilSupervisorById(ctx context.Context, id string) (*pb.GetTopicCouncilSupervisorResponse, error) {
	cacheKey := fmt.Sprintf("%s%s", topicCouncilSupervisorCachePrefix, id)
	var cached pb.GetTopicCouncilSupervisorResponse
	if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
		log.Printf("Cache HIT for topic council supervisor: %s", id)
		return &cached, nil
	}

	log.Printf("Cache MISS for topic council supervisor: %s", id)
	resp, err := t.client.GetTopicCouncilSupervisor(ctx, &pb.GetTopicCouncilSupervisorRequest{Id: id})
	if err != nil {
		return nil, err
	}

	SetCachedProto(ctx, t.redisClient, cacheKey, resp, topicCouncilSupervisorCacheTTL)
	return resp, nil
}

func (t *GRPCthesis) UpdateTopicCouncilSupervisor(ctx context.Context, req *pb.UpdateTopicCouncilSupervisorRequest) (*pb.UpdateTopicCouncilSupervisorResponse, error) {
	resp, err := t.client.UpdateTopicCouncilSupervisor(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if req.Id != "" {
		cacheKey := fmt.Sprintf("%s%s", topicCouncilSupervisorCachePrefix, req.Id)
		InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
		InvalidateCacheByPattern(ctx, t.redisClient, topicCouncilSupervisorCachePrefix+"*")
	}

	return resp, nil
}

func (t *GRPCthesis) DeleteTopicCouncilSupervisor(ctx context.Context, id string) (*pb.DeleteTopicCouncilSupervisorResponse, error) {
	resp, err := t.client.DeleteTopicCouncilSupervisor(ctx, &pb.DeleteTopicCouncilSupervisorRequest{Id: id})
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", topicCouncilSupervisorCachePrefix, id)
	InvalidateCacheByKey(ctx, t.redisClient, cacheKey)
	InvalidateCacheByPattern(ctx, t.redisClient, topicCouncilSupervisorCachePrefix+"*")

	return resp, nil
}

func (t *GRPCthesis) GetTopicCouncilSupervisorsByIds(ctx context.Context, ids []string) (*pb.ListTopicCouncilSupervisorsResponse, error) {
	if len(ids) == 0 {
		return &pb.ListTopicCouncilSupervisorsResponse{TopicCouncilSupervisors: []*pb.TopicCouncilSupervisor{}}, nil
	}

	result := &pb.ListTopicCouncilSupervisorsResponse{TopicCouncilSupervisors: []*pb.TopicCouncilSupervisor{}}
	missingIds := []string{}
	cacheHits := 0

	// Check Redis cache for each ID
	for _, id := range ids {
		cacheKey := fmt.Sprintf("%s%s", topicCouncilSupervisorCachePrefix, id)
		var cached pb.GetTopicCouncilSupervisorResponse

		if hit, _ := GetCachedProto(ctx, t.redisClient, cacheKey, &cached); hit {
			if cached.TopicCouncilSupervisor != nil {
				result.TopicCouncilSupervisors = append(result.TopicCouncilSupervisors, cached.TopicCouncilSupervisor)
				cacheHits++
			} else {
				missingIds = append(missingIds, id)
			}
		} else {
			missingIds = append(missingIds, id)
		}
	}

	log.Printf("[GetTopicCouncilSupervisorsByIds] Total: %d, Cache hits: %d, Database queries needed: %d", len(ids), cacheHits, len(missingIds))

	// Fetch missing IDs from database
	if len(missingIds) > 0 {
		resp, err := t.client.ListTopicCouncilSupervisors(ctx, &pb.ListTopicCouncilSupervisorsRequest{
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
		if resp != nil && resp.TopicCouncilSupervisors != nil {
			for _, supervisor := range resp.TopicCouncilSupervisors {
				if supervisor != nil {
					cacheKey := fmt.Sprintf("%s%s", topicCouncilSupervisorCachePrefix, supervisor.Id)
					SetCachedProto(ctx, t.redisClient, cacheKey, &pb.GetTopicCouncilSupervisorResponse{TopicCouncilSupervisor: supervisor}, topicCouncilSupervisorCacheTTL)
					result.TopicCouncilSupervisors = append(result.TopicCouncilSupervisors, supervisor)
				}
			}
		}
	}

	return result, nil
}
