package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/academic"
	"thaily/src/service/pkg/helper"
	"thaily/src/service/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateFaculty creates a new Faculty record
func (h *Handler) CreateFaculty(ctx context.Context, req *pb.CreateFacultyRequest) (*pb.FacultyResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.GetFaculty().GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Use provided ID
	id := req.GetFaculty().GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Prepare fields

	// Insert into database
	query := `
		INSERT INTO Faculty (id, title, ms, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GetFaculty().GetTitle(),
		req.GetFaculty().GetMs(),
		req.GetFaculty().GetCreatedBy(),
		req.GetFaculty().GetCreatedBy(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "faculty already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create faculty: %v", err)
	}

	result, err := h.GetFaculty(ctx, &pb.GetFacultyRequest{Id: req.GetFaculty().GetMs()})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get faculty")
	}
	return &pb.FacultyResponse{
		Faculty: result.GetFaculty(),
	}, nil
}

// GetFaculty retrieves a Faculty by ID
func (h *Handler) GetFaculty(ctx context.Context, req *pb.GetFacultyRequest) (*pb.FacultyResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	whereClause := ""
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":    true,
		"title": true,
		"ms":    true,
	}
	if req.Filter != nil && len(req.Filter) > 0 {
		whereClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
		whereClause = "WHERE " + whereClause + " AND id = ?"
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.Id)

	query := fmt.Sprintf(`
		SELECT id, title, ms, created_at, updated_at, created_by, updated_by
		FROM Faculty
		%s
	`, whereClause)

	var entity pb.Faculty
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	err := h.queryRow(ctx, query, args...).Scan(
		&entity.Id,
		&entity.Title,
		&entity.Ms,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "faculty not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get faculty: %v", err)
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

	return &pb.FacultyResponse{
		Faculty: &entity,
	}, nil
}

// UpdateFaculty updates an existing Faculty
func (h *Handler) UpdateFaculty(ctx context.Context, req *pb.UpdateFacultyRequest) (*pb.FacultyResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GetFaculty().GetTitle() != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, req.GetFaculty().GetTitle())
	}
	if req.GetFaculty().GetMs() != "" {
		updateFields = append(updateFields, "ms = ?")
		args = append(args, req.GetFaculty().GetMs())
	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.GetFaculty().GetUpdatedBy())
	updateFields = append(updateFields, "updated_at = NOW()")

	// Build WHERE clause from filters
	whereClause := ""
	whiteMap := map[string]bool{
		"id":    true,
		"title": true,
		"ms":    true,
	}
	if req.Filter != nil && len(req.Filter) > 0 {
		whereClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
		whereClause = "WHERE " + whereClause + " AND id = ?"
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.GetId())

	query := fmt.Sprintf(`
		UPDATE Faculty
		SET %s
		%s
	`, strings.Join(updateFields, ", "), whereClause)

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update faculty: %v", err)
	}

	result, err := h.GetFaculty(ctx, &pb.GetFacultyRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get faculty")
	}
	return &pb.FacultyResponse{
		Faculty: result.GetFaculty(),
	}, nil
}

// DeleteFaculty deletes a Faculty by ID
func (h *Handler) DeleteFaculty(ctx context.Context, req *pb.DeleteFacultyRequest) (*pb.DeleteFacultyResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filters
	whereClause := ""
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":    true,
		"title": true,
		"ms":    true,
	}
	if req.Filter != nil && len(req.Filter) > 0 {
		whereClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
		whereClause = "WHERE " + whereClause + " AND id = ?"
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.Id)

	query := fmt.Sprintf(`DELETE FROM Faculty %s`, whereClause)

	result, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete faculty: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "faculty not found")
	}

	return &pb.DeleteFacultyResponse{
		Success: true,
	}, nil
}

// ListFaculties lists Facultys with pagination and filtering
func (h *Handler) ListFaculties(ctx context.Context, req *pb.ListFacultiesRequest) (*pb.ListFacultiesResponse, error) {
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
		"id":    true,
		"title": true,
		"ms":    true,
	}
	if req.Search != nil && len(req.Search.Filters) > 0 {
		whereClause = helper.BuildWhereClause(req.Search.Filters, &args, whiteMap, true)
	}

	// Build ORDER BY clause
	sortDirection := "ASC"
	if descending {
		sortDirection = "DESC"
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Faculty %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count facultys: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, created_at, updated_at, created_by, updated_by
		FROM Faculty
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list facultys: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Faculty{}
	for rows.Next() {
		var entity pb.Faculty
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan faculty: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating facultys: %v", err)
	}

	return &pb.ListFacultiesResponse{
		Faculties: entities,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}
