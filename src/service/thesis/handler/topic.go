package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/thesis"
	"thaily/src/pkg/helper"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateTopic creates a new Topic record
func (h *Handler) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.MajorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "major_code is required")
	}
	if req.SemesterCode == "" {
		return nil, status.Error(codes.InvalidArgument, "semester_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	PercentStage_1 := int32(0)
	if req.PercentStage_1 != nil {
		PercentStage_1 = *req.PercentStage_1
	}
	PercentStage_2 := int32(0)
	if req.PercentStage_2 != nil {
		PercentStage_2 = *req.PercentStage_2
	}

	// Convert Status enum to string
	StatusValue := pb.TopicStatus_SUBMIT

	StatusValue = req.Status
	StatusStr := "submit"
	switch StatusValue {
	case pb.TopicStatus_SUBMIT:
		StatusStr = "submit"
	case pb.TopicStatus_TOPIC_PENDING:
		StatusStr = "topic_pending"
	case pb.TopicStatus_APPROVED_1:
		StatusStr = "approved_1"
	case pb.TopicStatus_APPROVED_2:
		StatusStr = "approved_2"
	case pb.TopicStatus_IN_PROGRESS:
		StatusStr = "in_progress"
	case pb.TopicStatus_TOPIC_COMPLETED:
		StatusStr = "topic_completed"
	case pb.TopicStatus_REJECTED:
		StatusStr = "rejected"
	}

	// Insert into database
	query := `
		INSERT INTO Topic (id, title, major_code, semester_code, status, percent_stage_1, percent_stage_2, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		req.MajorCode,
		req.SemesterCode,
		StatusStr,
		PercentStage_1,
		PercentStage_2,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "topic already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create topic: %v", err)
	}

	result, err := h.GetTopic(ctx, &pb.GetTopicRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topic")
	}
	return &pb.CreateTopicResponse{
		Topic: result.GetTopic(),
	}, nil
}

// GetTopic retrieves a Topic by ID
func (h *Handler) GetTopic(ctx context.Context, req *pb.GetTopicRequest) (*pb.GetTopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, major_code, semester_code, status, percent_stage_1, percent_stage_2, created_at, updated_at, created_by, updated_by
		FROM Topic
		WHERE id = ?
	`

	var entity pb.Topic
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.MajorCode,
		&entity.SemesterCode,
		&StatusStr,
		&entity.PercentStage_1,
		&entity.PercentStage_2,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "topic not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get topic: %v", err)
	}

	// Convert Status string to enum
	switch StatusStr {
	case "submit":
		entity.Status = pb.TopicStatus_SUBMIT
	case "topic_pending":
		entity.Status = pb.TopicStatus_TOPIC_PENDING
	case "approved_1":
		entity.Status = pb.TopicStatus_APPROVED_1
	case "approved_2":
		entity.Status = pb.TopicStatus_APPROVED_2
	case "in_progress":
		entity.Status = pb.TopicStatus_IN_PROGRESS
	case "topic_completed":
		entity.Status = pb.TopicStatus_TOPIC_COMPLETED
	case "rejected":
		entity.Status = pb.TopicStatus_REJECTED
	default:
		entity.Status = pb.TopicStatus_SUBMIT
	}

	if createdAt.Valid {
		entity.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		entity.UpdatedAt = timestamppb.New(updatedAt.Time)
	}
	if updatedBy.Valid {
		entity.UpdatedBy = updatedBy.String
	}

	return &pb.GetTopicResponse{
		Topic: &entity,
	}, nil
}

// UpdateTopic updates an existing Topic
func (h *Handler) UpdateTopic(ctx context.Context, req *pb.UpdateTopicRequest) (*pb.UpdateTopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.Title != nil {
		updateFields = append(updateFields, "title = ?")
		args = append(args, *req.Title)

	}
	if req.MajorCode != nil {
		updateFields = append(updateFields, "major_code = ?")
		args = append(args, *req.MajorCode)

	}
	if req.SemesterCode != nil {
		updateFields = append(updateFields, "semester_code = ?")
		args = append(args, *req.SemesterCode)

	}
	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		StatusStr := "submit"
		switch *req.Status {
		case pb.TopicStatus_SUBMIT:
			StatusStr = "submit"
		case pb.TopicStatus_TOPIC_PENDING:
			StatusStr = "topic_pending"
		case pb.TopicStatus_APPROVED_1:
			StatusStr = "approved_1"
		case pb.TopicStatus_APPROVED_2:
			StatusStr = "approved_2"
		case pb.TopicStatus_IN_PROGRESS:
			StatusStr = "in_progress"
		case pb.TopicStatus_TOPIC_COMPLETED:
			StatusStr = "topic_completed"
		case pb.TopicStatus_REJECTED:
			StatusStr = "rejected"
		}
		args = append(args, StatusStr)

	}
	if req.PercentStage_1 != nil {
		updateFields = append(updateFields, "percent_stage_1 = ?")
		args = append(args, *req.PercentStage_1)

	}
	if req.PercentStage_2 != nil {
		updateFields = append(updateFields, "percent_stage_2 = ?")
		args = append(args, *req.PercentStage_2)

	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.UpdatedBy)
	updateFields = append(updateFields, "updated_at = NOW()")

	// Add id as last parameter
	args = append(args, req.Id)

	query := fmt.Sprintf(`
		UPDATE Topic
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update topic: %v", err)
	}

	result, err := h.GetTopic(ctx, &pb.GetTopicRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topic")
	}
	return &pb.UpdateTopicResponse{
		Topic: result.GetTopic(),
	}, nil
}

// DeleteTopic deletes a Topic by ID
func (h *Handler) DeleteTopic(ctx context.Context, req *pb.DeleteTopicRequest) (*pb.DeleteTopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Topic WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete topic: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "topic not found")
	}

	return &pb.DeleteTopicResponse{
		Success: true,
	}, nil
}

// ListTopics lists Topics with pagination and filtering
func (h *Handler) ListTopics(ctx context.Context, req *pb.ListTopicsRequest) (*pb.ListTopicsResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Default pagination
	page := int32(1)
	pageSize := int32(10)
	sortBy := "created_at"
	descending := true
	if req.Search != nil && req.Search.Pagination != nil {
		if req.Search.Pagination.Page > 0 {
			page = req.Search.Pagination.Page
		}
		if req.Search.Pagination.PageSize > 0 {
			pageSize = req.Search.Pagination.PageSize
		}
		if req.Search.Pagination.SortBy != "" {
			sortBy = req.Search.Pagination.SortBy
		}
		descending = req.Search.Pagination.Descending
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Build WHERE clause from filters
	whereClause := ""
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id": true,

		"title":           true,
		"major_code":      true,
		"semester_code":   true,
		"status":          true,
		"percent_stage_1": true,
		"percent_stage_2": true,
	}
	if req.Search != nil && len(req.Search.Filters) > 0 {
		whereConditions := []string{}
		for _, filter := range req.Search.Filters {
			if filter.GetCondition() != nil {
				condition := filter.GetCondition()
				if _, ok := whiteMap[condition.Field]; !ok {
					continue
				}
				whereConditions = append(whereConditions, helper.BuildFilterCondition(condition, &args))
			}
		}
		if len(whereConditions) > 0 {
			whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
		}
	}

	// Build ORDER BY clause
	sortDirection := "ASC"
	if descending {
		sortDirection = "DESC"
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Topic %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count topics: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, major_code, semester_code, status, percent_stage_1, percent_stage_2, created_at, updated_at, created_by, updated_by
		FROM Topic
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list topics: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Topic{}
	for rows.Next() {
		var entity pb.Topic
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var StatusStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.MajorCode,
			&entity.SemesterCode,
			&StatusStr,
			&entity.PercentStage_1,
			&entity.PercentStage_2,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan topic: %v", err)
		}

		// Convert Status string to enum
		switch StatusStr {
		case "submit":
			entity.Status = pb.TopicStatus_SUBMIT
		case "topic_pending":
			entity.Status = pb.TopicStatus_TOPIC_PENDING
		case "approved_1":
			entity.Status = pb.TopicStatus_APPROVED_1
		case "approved_2":
			entity.Status = pb.TopicStatus_APPROVED_2
		case "in_progress":
			entity.Status = pb.TopicStatus_IN_PROGRESS
		case "topic_completed":
			entity.Status = pb.TopicStatus_TOPIC_COMPLETED
		case "rejected":
			entity.Status = pb.TopicStatus_REJECTED
		default:
			entity.Status = pb.TopicStatus_SUBMIT
		}

		if createdAt.Valid {
			entity.CreatedAt = timestamppb.New(createdAt.Time)
		}
		if updatedAt.Valid {
			entity.UpdatedAt = timestamppb.New(updatedAt.Time)
		}
		if updatedBy.Valid {
			entity.UpdatedBy = updatedBy.String
		}

		entities = append(entities, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error iterating topics: %v", err)
	}

	return &pb.ListTopicsResponse{
		Topics:   entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
