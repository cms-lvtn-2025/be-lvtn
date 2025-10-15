package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/academic"
	"thaily/src/pkg/helper"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)



// CreateMajor creates a new Major record
func (h *Handler) CreateMajor(ctx context.Context, req *pb.CreateMajorRequest) (*pb.CreateMajorResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.FacultyCode == "" {
		return nil, status.Error(codes.InvalidArgument, "faculty_code is required")
	}
	
	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	
	
	// Insert into database
	query := `
		INSERT INTO Major (id, title, faculty_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		req.FacultyCode,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "major already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create major: %v", err)
	}

	result, err := h.GetMajor(ctx, &pb.GetMajorRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get major")
	}
	return &pb.CreateMajorResponse{
		Major: result.GetMajor(),
	}, nil
}













// GetMajor retrieves a Major by ID
func (h *Handler) GetMajor(ctx context.Context, req *pb.GetMajorRequest) (*pb.GetMajorResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, faculty_code, created_at, updated_at, created_by, updated_by
		FROM Major
		WHERE id = ?
	`

	var entity pb.Major
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	
	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.FacultyCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "major not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get major: %v", err)
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

	return &pb.GetMajorResponse{
		Major: &entity,
	}, nil
}













// UpdateMajor updates an existing Major
func (h *Handler) UpdateMajor(ctx context.Context, req *pb.UpdateMajorRequest) (*pb.UpdateMajorResponse, error) {
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
	if req.FacultyCode != nil {
		updateFields = append(updateFields, "faculty_code = ?")
		args = append(args, *req.FacultyCode)
		
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
		UPDATE Major
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update major: %v", err)
	}

	result, err := h.GetMajor(ctx, &pb.GetMajorRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get major")
	}
	return &pb.UpdateMajorResponse{
		Major: result.GetMajor(),
	}, nil
}













// DeleteMajor deletes a Major by ID
func (h *Handler) DeleteMajor(ctx context.Context, req *pb.DeleteMajorRequest) (*pb.DeleteMajorResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Major WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete major: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "major not found")
	}

	return &pb.DeleteMajorResponse{
		Success: true,
	}, nil
}













// ListMajors lists Majors with pagination and filtering
func (h *Handler) ListMajors(ctx context.Context, req *pb.ListMajorsRequest) (*pb.ListMajorsResponse, error) {
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
		"title": true,
		"faculty_code": true,
		
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Major %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count majors: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, faculty_code, created_at, updated_at, created_by, updated_by
		FROM Major
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list majors: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Major{}
	for rows.Next() {
		var entity pb.Major
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		
		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.FacultyCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan major: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating majors: %v", err)
	}

	return &pb.ListMajorsResponse{
		Majors: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}


