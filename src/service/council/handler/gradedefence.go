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

// CreateGradeDefence creates a new GradeDefence record
func (h *Handler) CreateGradeDefence(ctx context.Context, req *pb.CreateGradeDefenceRequest) (*pb.CreateGradeDefenceResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields

	// Insert into database
	query := `
		INSERT INTO GradeDefence (id, council, secretary, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Council,
		req.Secretary,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "gradedefence already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create gradedefence: %v", err)
	}

	result, err := h.GetGradeDefence(ctx, &pb.GetGradeDefenceRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get gradedefence")
	}
	return &pb.CreateGradeDefenceResponse{
		GradeDefence: result.GetGradeDefence(),
	}, nil
}

// GetGradeDefence retrieves a GradeDefence by ID
func (h *Handler) GetGradeDefence(ctx context.Context, req *pb.GetGradeDefenceRequest) (*pb.GetGradeDefenceResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, council, secretary, created_at, updated_at, created_by, updated_by
		FROM GradeDefence
		WHERE id = ?
	`

	var entity pb.GradeDefence
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Council,
		&entity.Secretary,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "gradedefence not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get gradedefence: %v", err)
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

	return &pb.GetGradeDefenceResponse{
		GradeDefence: &entity,
	}, nil
}

// UpdateGradeDefence updates an existing GradeDefence
func (h *Handler) UpdateGradeDefence(ctx context.Context, req *pb.UpdateGradeDefenceRequest) (*pb.UpdateGradeDefenceResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.Council != nil {
		updateFields = append(updateFields, "council = ?")
		args = append(args, *req.Council)

	}
	if req.Secretary != nil {
		updateFields = append(updateFields, "secretary = ?")
		args = append(args, *req.Secretary)

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
		UPDATE GradeDefence
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update gradedefence: %v", err)
	}

	result, err := h.GetGradeDefence(ctx, &pb.GetGradeDefenceRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get gradedefence")
	}
	return &pb.UpdateGradeDefenceResponse{
		GradeDefence: result.GetGradeDefence(),
	}, nil
}

// DeleteGradeDefence deletes a GradeDefence by ID
func (h *Handler) DeleteGradeDefence(ctx context.Context, req *pb.DeleteGradeDefenceRequest) (*pb.DeleteGradeDefenceResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM GradeDefence WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete gradedefence: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "gradedefence not found")
	}

	return &pb.DeleteGradeDefenceResponse{
		Success: true,
	}, nil
}

// ListGradeDefences lists GradeDefences with pagination and filtering
func (h *Handler) ListGradeDefences(ctx context.Context, req *pb.ListGradeDefencesRequest) (*pb.ListGradeDefencesResponse, error) {
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
		"council":   true,
		"secretary": true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM GradeDefence %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count gradedefences: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, council, secretary, created_at, updated_at, created_by, updated_by
		FROM GradeDefence
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list gradedefences: %v", err)
	}
	defer rows.Close()

	entities := []*pb.GradeDefence{}
	for rows.Next() {
		var entity pb.GradeDefence
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&entity.Id,
			&entity.Council,
			&entity.Secretary,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan gradedefence: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating gradedefences: %v", err)
	}

	return &pb.ListGradeDefencesResponse{
		GradeDefences: entities,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}
