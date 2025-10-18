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

// CreateTopicCouncilSupervisor creates a new TopicCouncilSupervisor record
func (h *Handler) CreateTopicCouncilSupervisor(ctx context.Context, req *pb.CreateTopicCouncilSupervisorRequest) (*pb.CreateTopicCouncilSupervisorResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.TeacherSupervisorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "teacher_supervisor_code is required")
	}
	if req.TopicCouncilCode == "" {
		return nil, status.Error(codes.InvalidArgument, "topic_council_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields

	// Insert into database
	query := `
		INSERT INTO TopicCouncilSupervisor (id, teacher_supervisor_code, topic_council_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.TeacherSupervisorCode,
		req.TopicCouncilCode,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "topiccouncilsupervisor already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create topiccouncilsupervisor: %v", err)
	}

	result, err := h.GetTopicCouncilSupervisor(ctx, &pb.GetTopicCouncilSupervisorRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topiccouncilsupervisor")
	}
	return &pb.CreateTopicCouncilSupervisorResponse{
		TopicCouncilSupervisor: result.GetTopicCouncilSupervisor(),
	}, nil
}

// GetTopicCouncilSupervisor retrieves a TopicCouncilSupervisor by ID
func (h *Handler) GetTopicCouncilSupervisor(ctx context.Context, req *pb.GetTopicCouncilSupervisorRequest) (*pb.GetTopicCouncilSupervisorResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, teacher_supervisor_code, topic_council_code, created_at, updated_at, created_by, updated_by
		FROM TopicCouncilSupervisor
		WHERE id = ?
	`

	var entity pb.TopicCouncilSupervisor
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.TeacherSupervisorCode,
		&entity.TopicCouncilCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "topiccouncilsupervisor not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get topiccouncilsupervisor: %v", err)
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

	return &pb.GetTopicCouncilSupervisorResponse{
		TopicCouncilSupervisor: &entity,
	}, nil
}

// UpdateTopicCouncilSupervisor updates an existing TopicCouncilSupervisor
func (h *Handler) UpdateTopicCouncilSupervisor(ctx context.Context, req *pb.UpdateTopicCouncilSupervisorRequest) (*pb.UpdateTopicCouncilSupervisorResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.TeacherSupervisorCode != nil {
		updateFields = append(updateFields, "teacher_supervisor_code = ?")
		args = append(args, *req.TeacherSupervisorCode)

	}
	if req.TopicCouncilCode != nil {
		updateFields = append(updateFields, "topic_council_code = ?")
		args = append(args, *req.TopicCouncilCode)

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
		UPDATE TopicCouncilSupervisor
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update topiccouncilsupervisor: %v", err)
	}

	result, err := h.GetTopicCouncilSupervisor(ctx, &pb.GetTopicCouncilSupervisorRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topiccouncilsupervisor")
	}
	return &pb.UpdateTopicCouncilSupervisorResponse{
		TopicCouncilSupervisor: result.GetTopicCouncilSupervisor(),
	}, nil
}

// DeleteTopicCouncilSupervisor deletes a TopicCouncilSupervisor by ID
func (h *Handler) DeleteTopicCouncilSupervisor(ctx context.Context, req *pb.DeleteTopicCouncilSupervisorRequest) (*pb.DeleteTopicCouncilSupervisorResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM TopicCouncilSupervisor WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete topiccouncilsupervisor: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "topiccouncilsupervisor not found")
	}

	return &pb.DeleteTopicCouncilSupervisorResponse{
		Success: true,
	}, nil
}

// ListTopicCouncilSupervisors lists TopicCouncilSupervisors with pagination and filtering
func (h *Handler) ListTopicCouncilSupervisors(ctx context.Context, req *pb.ListTopicCouncilSupervisorsRequest) (*pb.ListTopicCouncilSupervisorsResponse, error) {
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

		"teacher_supervisor_code": true,
		"topic_council_code":      true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM TopicCouncilSupervisor %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count topiccouncilsupervisors: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, teacher_supervisor_code, topic_council_code, created_at, updated_at, created_by, updated_by
		FROM TopicCouncilSupervisor
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list topiccouncilsupervisors: %v", err)
	}
	defer rows.Close()

	entities := []*pb.TopicCouncilSupervisor{}
	for rows.Next() {
		var entity pb.TopicCouncilSupervisor
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&entity.Id,
			&entity.TeacherSupervisorCode,
			&entity.TopicCouncilCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan topiccouncilsupervisor: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating topiccouncilsupervisors: %v", err)
	}

	return &pb.ListTopicCouncilSupervisorsResponse{
		TopicCouncilSupervisors: entities,
		Total:                   total,
		Page:                    page,
		PageSize:                pageSize,
	}, nil
}
