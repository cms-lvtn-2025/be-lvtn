package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/thesis"
	"thaily/src/service/pkg/helper"
	"thaily/src/service/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateMidterm creates a new Midterm record
func (h *Handler) CreateMidterm(ctx context.Context, req *pb.CreateMidtermRequest) (*pb.MidtermResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.GetMidterm().GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Convert Status enum to string
	StatusStr := "not_submitted"
	switch req.GetMidterm().GetStatus() {
	case pb.MidtermStatus_NOT_SUBMITTED:
		StatusStr = "not_submitted"
	case pb.MidtermStatus_SUBMITTED:
		StatusStr = "submitted"
	case pb.MidtermStatus_PASS:
		StatusStr = "pass"
	case pb.MidtermStatus_FAIL:
		StatusStr = "fail"
	}

	// Insert into database
	query := `
		INSERT INTO Midterm (id, enrollment_code, title, grade, status, feedback, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GetMidterm().GetEnrollmentCode(),
		req.GetMidterm().GetTitle(),
		req.GetMidterm().GetGrade(),
		StatusStr,
		req.GetMidterm().GetFeedback(),
		req.GetMidterm().GetCreatedBy(),
		req.GetMidterm().GetCreatedBy(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "midterm already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create midterm: %v", err)
	}

	return h.GetMidterm(ctx, &pb.GetMidtermRequest{Id: id})
}

// GetMidterm retrieves a Midterm by ID
func (h *Handler) GetMidterm(ctx context.Context, req *pb.GetMidtermRequest) (*pb.MidtermResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":              true,
		"enrollment_code": true,
		"title":           true,
		"grade":           true,
		"status":          true,
		"feedback":        true,
	}
	filterClause := ""
	if len(req.Filter) > 0 {
		filterClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
	}

	var whereClause string
	if filterClause != "" {
		whereClause = fmt.Sprintf("WHERE %s AND id = ?", filterClause)
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.Id)

	query := fmt.Sprintf(`
		SELECT id, enrollment_code, title, grade, status, feedback, created_at, updated_at, created_by, updated_by
		FROM Midterm
		%s
	`, whereClause)

	var entity pb.Midterm
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string

	err := h.queryRow(ctx, query, args...).Scan(
		&entity.Id,
		&entity.EnrollmentCode,
		&entity.Title,
		&entity.Grade,
		&StatusStr,
		&entity.Feedback,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "midterm not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get midterm: %v", err)
	}

	// Convert Status string to enum
	switch StatusStr {
	case "not_submitted":
		entity.Status = pb.MidtermStatus_NOT_SUBMITTED
	case "submitted":
		entity.Status = pb.MidtermStatus_SUBMITTED
	case "pass":
		entity.Status = pb.MidtermStatus_PASS
	case "fail":
		entity.Status = pb.MidtermStatus_FAIL
	default:
		entity.Status = pb.MidtermStatus_NOT_SUBMITTED
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

	return &pb.MidtermResponse{
		Midterm: &entity,
	}, nil
}

// UpdateMidterm updates an existing Midterm
func (h *Handler) UpdateMidterm(ctx context.Context, req *pb.UpdateMidtermRequest) (*pb.MidtermResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GetMidterm().GetTitle() != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, req.GetMidterm().GetTitle())
	}
	if req.GetMidterm().GetGrade() != 0 {
		updateFields = append(updateFields, "grade = ?")
		args = append(args, req.GetMidterm().GetGrade())
	}
	if req.GetMidterm().GetEnrollmentCode() != "" {
		updateFields = append(updateFields, "enrollment_code = ?")
		args = append(args, req.GetMidterm().GetEnrollmentCode())
	}
	// Status enum - always include
	updateFields = append(updateFields, "status = ?")
	StatusStr := "not_submitted"
	switch req.GetMidterm().GetStatus() {
	case pb.MidtermStatus_NOT_SUBMITTED:
		StatusStr = "not_submitted"
	case pb.MidtermStatus_SUBMITTED:
		StatusStr = "submitted"
	case pb.MidtermStatus_PASS:
		StatusStr = "pass"
	case pb.MidtermStatus_FAIL:
		StatusStr = "fail"
	}
	args = append(args, StatusStr)

	if req.GetMidterm().GetFeedback() != "" {
		updateFields = append(updateFields, "feedback = ?")
		args = append(args, req.GetMidterm().GetFeedback())
	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.GetMidterm().GetUpdatedBy())
	updateFields = append(updateFields, "updated_at = NOW()")

	// Build WHERE clause from filter
	whiteMap := map[string]bool{
		"id":              true,
		"enrollment_code": true,
		"title":           true,
		"grade":           true,
		"status":          true,
		"feedback":        true,
	}
	filterClause := ""
	if len(req.Filter) > 0 {
		filterClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
	}

	var whereClause string
	if filterClause != "" {
		whereClause = fmt.Sprintf("WHERE %s AND id = ?", filterClause)
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.Id)

	query := fmt.Sprintf(`
		UPDATE Midterm
		SET %s
		%s
	`, strings.Join(updateFields, ", "), whereClause)

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update midterm: %v", err)
	}

	return h.GetMidterm(ctx, &pb.GetMidtermRequest{Id: req.Id})
}

// DeleteMidterm deletes a Midterm by ID
func (h *Handler) DeleteMidterm(ctx context.Context, req *pb.DeleteMidtermRequest) (*pb.DeleteMidtermResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":              true,
		"enrollment_code": true,
		"title":           true,
		"grade":           true,
		"status":          true,
		"feedback":        true,
	}
	filterClause := ""
	if len(req.Filter) > 0 {
		filterClause = helper.BuildWhereClause(req.Filter, &args, whiteMap, false)
	}

	var whereClause string
	if filterClause != "" {
		whereClause = fmt.Sprintf("WHERE %s AND id = ?", filterClause)
	} else {
		whereClause = "WHERE id = ?"
	}
	args = append(args, req.Id)

	query := fmt.Sprintf(`DELETE FROM Midterm %s`, whereClause)

	result, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete midterm: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "midterm not found")
	}

	return &pb.DeleteMidtermResponse{
		Success: true,
	}, nil
}

// ListMidterms lists Midterms with pagination and filtering
func (h *Handler) ListMidterms(ctx context.Context, req *pb.ListMidtermsRequest) (*pb.ListMidtermsResponse, error) {
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
		"id":              true,
		"enrollment_code": true,
		"title":           true,
		"grade":           true,
		"status":          true,
		"feedback":        true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Midterm %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count midterms: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, enrollment_code, title, grade, status, feedback, created_at, updated_at, created_by, updated_by
		FROM Midterm
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list midterms: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Midterm{}
	for rows.Next() {
		var entity pb.Midterm
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var StatusStr string

		err := rows.Scan(
			&entity.Id,
			&entity.EnrollmentCode,
			&entity.Title,
			&entity.Grade,
			&StatusStr,
			&entity.Feedback,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan midterm: %v", err)
		}

		// Convert Status string to enum
		switch StatusStr {
		case "not_submitted":
			entity.Status = pb.MidtermStatus_NOT_SUBMITTED
		case "submitted":
			entity.Status = pb.MidtermStatus_SUBMITTED
		case "pass":
			entity.Status = pb.MidtermStatus_PASS
		case "fail":
			entity.Status = pb.MidtermStatus_FAIL
		default:
			entity.Status = pb.MidtermStatus_NOT_SUBMITTED
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
		return nil, status.Errorf(codes.Internal, "error iterating midterms: %v", err)
	}

	return &pb.ListMidtermsResponse{
		Midterms: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
