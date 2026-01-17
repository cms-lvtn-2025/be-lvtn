package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/academic"
	"thaily/src/service/pkg/helper"
	"thaily/src/service/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateSemester creates a new Semester record
func (h *Handler) CreateSemester(ctx context.Context, req *pb.CreateSemesterRequest) (*pb.SemesterResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.GetSemester().GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields

	// Insert into database
	query := `
		INSERT INTO Semester (id, title, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GetSemester().GetTitle(),
		req.GetSemester().GetCreatedBy(),
		req.GetSemester().GetCreatedBy(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "semester already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create semester: %v", err)
	}
	var slug string
	slugQuery := "SELECT create_slug(?) AS slug"

	// Sử dụng QueryRowContext để thực thi và lấy 1 hàng kết quả
	row := h.db.QueryRowContext(ctx, slugQuery, req.GetSemester().GetTitle())

	// Hứng kết quả (slug) vào biến đã khai báo
	err = row.Scan(&slug)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate slug: %v", err)
	}

	// Kiểm tra giá trị slug (ví dụ: nếu nó rỗng)
	if slug == "" {
		return nil, status.Error(codes.Internal, "generated slug is empty")
	}

	return h.GetSemester(ctx, &pb.GetSemesterRequest{Id: slug})
}

// GetSemester retrieves a Semester by ID
func (h *Handler) GetSemester(ctx context.Context, req *pb.GetSemesterRequest) (*pb.SemesterResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	whereClause := ""
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":    true,
		"title": true,
	}
	if req.Filter != nil && len(req.Filter) > 0 {
		whereClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
		whereClause = "WHERE " + whereClause + " AND id = ?"
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.Id)

	query := fmt.Sprintf(`
	SELECT id, title, created_at, updated_at, created_by, updated_by
	FROM Semester
	%s
`, whereClause)

	var entity pb.Semester
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	err := h.queryRow(ctx, query, args...).Scan(
		&entity.Id,
		&entity.Title,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "semester not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get semester: %v", err)
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

	return &pb.SemesterResponse{
		Semester: &entity,
	}, nil
}

// UpdateSemester updates an existing Semester
func (h *Handler) UpdateSemester(ctx context.Context, req *pb.UpdateSemesterRequest) (*pb.SemesterResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.GetSemester().GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GetSemester().GetTitle() != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, req.GetSemester().GetTitle())
	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.GetSemester().GetUpdatedBy())
	updateFields = append(updateFields, "updated_at = NOW()")

	// Build WHERE clause from filters
	whereClause := ""
	whiteMap := map[string]bool{
		"id":    true,
		"title": true,
	}
	if req.Filter != nil && len(req.Filter) > 0 {
		whereClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
		whereClause = "WHERE " + whereClause + " AND id = ?"
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.GetId())

	query := fmt.Sprintf(`
		UPDATE Semester
		SET %s
		%s
	`, strings.Join(updateFields, ", "), whereClause)

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update semester: %v", err)
	}

	result, err := h.GetSemester(ctx, &pb.GetSemesterRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get semester")
	}
	return &pb.SemesterResponse{
		Semester: result.GetSemester(),
	}, nil
}

// DeleteSemester deletes a Semester by ID
func (h *Handler) DeleteSemester(ctx context.Context, req *pb.DeleteSemesterRequest) (*pb.DeleteSemesterResponse, error) {
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
	}
	if req.Filter != nil && len(req.Filter) > 0 {
		whereClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
		whereClause = "WHERE " + whereClause + " AND id = ?"
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.Id)

	query := fmt.Sprintf(`DELETE FROM Semester %s`, whereClause)

	result, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete semester: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "semester not found")
	}

	return &pb.DeleteSemesterResponse{
		Success: true,
	}, nil
}

// ListSemesters lists Semesters with pagination and filtering
func (h *Handler) ListSemesters(ctx context.Context, req *pb.ListSemestersRequest) (*pb.ListSemestersResponse, error) {
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Semester %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count semesters: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, created_at, updated_at, created_by, updated_by
		FROM Semester
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list semesters: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Semester{}
	for rows.Next() {
		var entity pb.Semester
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
			return nil, status.Errorf(codes.Internal, "failed to scan semester: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating semesters: %v", err)
	}

	return &pb.ListSemestersResponse{
		Semesters: entities,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}
