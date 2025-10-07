package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/council"
	"thaily/src/pkg/helper"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateCouncilSchedule creates a new CouncilSchedule record
func (h *Handler) CreateCouncilSchedule(ctx context.Context, req *pb.CreateCouncilScheduleRequest) (*pb.CreateCouncilScheduleResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.CouncilsCode == "" {
		return nil, status.Error(codes.InvalidArgument, "councils_code is required")
	}
	if req.TopicCode == "" {
		return nil, status.Error(codes.InvalidArgument, "topic_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields

	// Insert into database
	query := `
		INSERT INTO CouncilSchedule (id, councils_code, topic_code, status, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.CouncilsCode,
		req.TopicCode,
		req.Status,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "councilschedule already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create councilschedule: %v", err)
	}

	result, err := h.GetCouncilSchedule(ctx, &pb.GetCouncilScheduleRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get councilschedule")
	}
	return &pb.CreateCouncilScheduleResponse{
		CouncilSchedule: result.GetCouncilSchedule(),
	}, nil
}

// GetCouncilSchedule retrieves a CouncilSchedule by ID
func (h *Handler) GetCouncilSchedule(ctx context.Context, req *pb.GetCouncilScheduleRequest) (*pb.GetCouncilScheduleResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, councils_code, topic_code, status, created_at, updated_at, created_by, updated_by
		FROM CouncilSchedule
		WHERE id = ?
	`

	var entity pb.CouncilSchedule
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.CouncilsCode,
		&entity.TopicCode,
		&entity.Status,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "councilschedule not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get councilschedule: %v", err)
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

	return &pb.GetCouncilScheduleResponse{
		CouncilSchedule: &entity,
	}, nil
}

// UpdateCouncilSchedule updates an existing CouncilSchedule
func (h *Handler) UpdateCouncilSchedule(ctx context.Context, req *pb.UpdateCouncilScheduleRequest) (*pb.UpdateCouncilScheduleResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.CouncilsCode != nil {
		updateFields = append(updateFields, "councils_code = ?")
		args = append(args, *req.CouncilsCode)

	}
	if req.TopicCode != nil {
		updateFields = append(updateFields, "topic_code = ?")
		args = append(args, *req.TopicCode)

	}
	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		args = append(args, *req.Status)

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
		UPDATE CouncilSchedule
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update councilschedule: %v", err)
	}

	result, err := h.GetCouncilSchedule(ctx, &pb.GetCouncilScheduleRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get councilschedule")
	}
	return &pb.UpdateCouncilScheduleResponse{
		CouncilSchedule: result.GetCouncilSchedule(),
	}, nil
}

// DeleteCouncilSchedule deletes a CouncilSchedule by ID
func (h *Handler) DeleteCouncilSchedule(ctx context.Context, req *pb.DeleteCouncilScheduleRequest) (*pb.DeleteCouncilScheduleResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM CouncilSchedule WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete councilschedule: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "councilschedule not found")
	}

	return &pb.DeleteCouncilScheduleResponse{
		Success: true,
	}, nil
}

// ListCouncilSchedules lists CouncilSchedules with pagination and filtering
func (h *Handler) ListCouncilSchedules(ctx context.Context, req *pb.ListCouncilSchedulesRequest) (*pb.ListCouncilSchedulesResponse, error) {
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
		"councils_code": true,
		"topic_code":    true,
		"status":        true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM CouncilSchedule %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count councilschedules: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, councils_code, topic_code, status, created_at, updated_at, created_by, updated_by
		FROM CouncilSchedule
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list councilschedules: %v", err)
	}
	defer rows.Close()

	entities := []*pb.CouncilSchedule{}
	for rows.Next() {
		var entity pb.CouncilSchedule
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&entity.Id,
			&entity.CouncilsCode,
			&entity.TopicCode,
			&entity.Status,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan councilschedule: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating councilschedules: %v", err)
	}

	return &pb.ListCouncilSchedulesResponse{
		CouncilSchedules: entities,
		Total:            total,
		Page:             page,
		PageSize:         pageSize,
	}, nil
}
