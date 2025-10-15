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



// CreateTopicSupervisor creates a new TopicSupervisor record
func (h *Handler) CreateTopicSupervisor(ctx context.Context, req *pb.CreateTopicSupervisorRequest) (*pb.CreateTopicSupervisorResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.TeacherSupervisorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "teacher_supervisor_code is required")
	}
	if req.TopicCode == "" {
		return nil, status.Error(codes.InvalidArgument, "topic_code is required")
	}
	
	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	
	
	// Insert into database
	query := `
		INSERT INTO TopicSupervisor (id, teacher_supervisor_code, topic_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.TeacherSupervisorCode,
		req.TopicCode,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "topicsupervisor already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create topicsupervisor: %v", err)
	}

	result, err := h.GetTopicSupervisor(ctx, &pb.GetTopicSupervisorRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topicsupervisor")
	}
	return &pb.CreateTopicSupervisorResponse{
		TopicSupervisor: result.GetTopicSupervisor(),
	}, nil
}













// GetTopicSupervisor retrieves a TopicSupervisor by ID
func (h *Handler) GetTopicSupervisor(ctx context.Context, req *pb.GetTopicSupervisorRequest) (*pb.GetTopicSupervisorResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, teacher_supervisor_code, topic_code, created_at, updated_at, created_by, updated_by
		FROM TopicSupervisor
		WHERE id = ?
	`

	var entity pb.TopicSupervisor
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	
	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.TeacherSupervisorCode,
		&entity.TopicCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "topicsupervisor not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get topicsupervisor: %v", err)
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

	return &pb.GetTopicSupervisorResponse{
		TopicSupervisor: &entity,
	}, nil
}













// UpdateTopicSupervisor updates an existing TopicSupervisor
func (h *Handler) UpdateTopicSupervisor(ctx context.Context, req *pb.UpdateTopicSupervisorRequest) (*pb.UpdateTopicSupervisorResponse, error) {
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
	if req.TopicCode != nil {
		updateFields = append(updateFields, "topic_code = ?")
		args = append(args, *req.TopicCode)
		
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
		UPDATE TopicSupervisor
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update topicsupervisor: %v", err)
	}

	result, err := h.GetTopicSupervisor(ctx, &pb.GetTopicSupervisorRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topicsupervisor")
	}
	return &pb.UpdateTopicSupervisorResponse{
		TopicSupervisor: result.GetTopicSupervisor(),
	}, nil
}













// DeleteTopicSupervisor deletes a TopicSupervisor by ID
func (h *Handler) DeleteTopicSupervisor(ctx context.Context, req *pb.DeleteTopicSupervisorRequest) (*pb.DeleteTopicSupervisorResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM TopicSupervisor WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete topicsupervisor: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "topicsupervisor not found")
	}

	return &pb.DeleteTopicSupervisorResponse{
		Success: true,
	}, nil
}













// ListTopicSupervisors lists TopicSupervisors with pagination and filtering
func (h *Handler) ListTopicSupervisors(ctx context.Context, req *pb.ListTopicSupervisorsRequest) (*pb.ListTopicSupervisorsResponse, error) {
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
		"teacher_supervisor_code": true,
		"topic_code": true,
		
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM TopicSupervisor %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count topicsupervisors: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, teacher_supervisor_code, topic_code, created_at, updated_at, created_by, updated_by
		FROM TopicSupervisor
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list topicsupervisors: %v", err)
	}
	defer rows.Close()

	entities := []*pb.TopicSupervisor{}
	for rows.Next() {
		var entity pb.TopicSupervisor
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		
		err := rows.Scan(
			&entity.Id,
			&entity.TeacherSupervisorCode,
			&entity.TopicCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan topicsupervisor: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating topicsupervisors: %v", err)
	}

	return &pb.ListTopicSupervisorsResponse{
		TopicSupervisors: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}


