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

// CreateFinal creates a new Final record
func (h *Handler) CreateFinal(ctx context.Context, req *pb.CreateFinalRequest) (*pb.FinalResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.GetFinal().GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Convert Status enum to string
	StatusStr := "pending"
	switch req.GetFinal().GetStatus() {
	case pb.FinalStatus_PENDING:
		StatusStr = "pending"
	case pb.FinalStatus_PASSED:
		StatusStr = "passed"
	case pb.FinalStatus_FAILED:
		StatusStr = "failed"
	case pb.FinalStatus_COMPLETED:
		StatusStr = "completed"
	}

	// Handle completion_date
	var completionDate interface{}
	if req.GetFinal().GetCompletionDate() != nil {
		completionDate = req.GetFinal().GetCompletionDate().AsTime()
	} else {
		completionDate = nil
	}

	// Insert into database
	query := `
		INSERT INTO Final (id, title,enrollment_code, supervisor_grade, final_grade, status, notes, completion_date, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?,?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GetFinal().GetTitle(),
		req.GetFinal().GetEnrollmentCode(),
		req.GetFinal().GetSupervisorGrade(),
		req.GetFinal().GetFinalGrade(),
		StatusStr,
		req.GetFinal().GetNotes(),
		completionDate,
		req.GetFinal().GetCreatedBy(),
		req.GetFinal().GetCreatedBy(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "final already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create final: %v", err)
	}

	return h.GetFinal(ctx, &pb.GetFinalRequest{Id: id})
}

// GetFinal retrieves a Final by ID
func (h *Handler) GetFinal(ctx context.Context, req *pb.GetFinalRequest) (*pb.FinalResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":               true,
		"title":            true,
		"supervisor_grade": true,
		"enrollment_code":  true,
		"final_grade":      true,
		"status":           true,
		"notes":            true,
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
		SELECT id, title, supervisor_grade, enrollment_code, final_grade, status, notes, completion_date, created_at, updated_at, created_by, updated_by
		FROM Final
		%s
	`, whereClause)

	var entity pb.Final
	var createdAt, updatedAt, completionDate sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string

	err := h.queryRow(ctx, query, args...).Scan(
		&entity.Id,
		&entity.Title,
		&entity.SupervisorGrade,
		&entity.EnrollmentCode,
		&entity.FinalGrade,
		&StatusStr,
		&entity.Notes,
		&completionDate,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "final not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get final: %v", err)
	}

	// Convert Status string to enum
	switch StatusStr {
	case "pending":
		entity.Status = pb.FinalStatus_PENDING
	case "passed":
		entity.Status = pb.FinalStatus_PASSED
	case "failed":
		entity.Status = pb.FinalStatus_FAILED
	case "completed":
		entity.Status = pb.FinalStatus_COMPLETED
	default:
		entity.Status = pb.FinalStatus_PENDING
	}

	if completionDate.Valid {
		entity.CompletionDate = timestamppb.New(completionDate.Time)
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

	return &pb.FinalResponse{
		Final: &entity,
	}, nil
}

// UpdateFinal updates an existing Final
func (h *Handler) UpdateFinal(ctx context.Context, req *pb.UpdateFinalRequest) (*pb.FinalResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GetFinal().GetTitle() != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, req.GetFinal().GetTitle())
	}
	if req.GetFinal().GetSupervisorGrade() != 0 {
		updateFields = append(updateFields, "supervisor_grade = ?")
		args = append(args, req.GetFinal().GetSupervisorGrade())
	}
	if req.GetFinal().GetEnrollmentCode() != "" {
		updateFields = append(updateFields, "enrollment_code = ?")
		args = append(args, req.GetFinal().GetEnrollmentCode())
	}
	if req.GetFinal().GetFinalGrade() != 0 {
		updateFields = append(updateFields, "final_grade = ?")
		args = append(args, req.GetFinal().GetFinalGrade())
	}
	// Status enum - always include
	updateFields = append(updateFields, "status = ?")
	StatusStr := "pending"
	switch req.GetFinal().GetStatus() {
	case pb.FinalStatus_PENDING:
		StatusStr = "pending"
	case pb.FinalStatus_PASSED:
		StatusStr = "passed"
	case pb.FinalStatus_FAILED:
		StatusStr = "failed"
	case pb.FinalStatus_COMPLETED:
		StatusStr = "completed"
	}
	args = append(args, StatusStr)

	if req.GetFinal().GetNotes() != "" {
		updateFields = append(updateFields, "notes = ?")
		args = append(args, req.GetFinal().GetNotes())
	}
	if req.GetFinal().GetCompletionDate() != nil {
		updateFields = append(updateFields, "completion_date = ?")
		args = append(args, req.GetFinal().GetCompletionDate().AsTime())
	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.GetFinal().GetUpdatedBy())
	updateFields = append(updateFields, "updated_at = NOW()")

	// Build WHERE clause from filter
	whiteMap := map[string]bool{
		"id":               true,
		"title":            true,
		"supervisor_grade": true,
		"enrollment_code":  true,
		"final_grade":      true,
		"status":           true,
		"notes":            true,
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
		UPDATE Final
		SET %s
		%s
	`, strings.Join(updateFields, ", "), whereClause)

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update final: %v", err)
	}

	return h.GetFinal(ctx, &pb.GetFinalRequest{Id: req.Id})
}

// DeleteFinal deletes a Final by ID
func (h *Handler) DeleteFinal(ctx context.Context, req *pb.DeleteFinalRequest) (*pb.DeleteFinalResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":               true,
		"title":            true,
		"supervisor_grade": true,
		"enrollment_code":  true,
		"final_grade":      true,
		"status":           true,
		"notes":            true,
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

	query := fmt.Sprintf(`DELETE FROM Final %s`, whereClause)

	result, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete final: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "final not found")
	}

	return &pb.DeleteFinalResponse{
		Success: true,
	}, nil
}

// ListFinals lists Finals with pagination and filtering
func (h *Handler) ListFinals(ctx context.Context, req *pb.ListFinalsRequest) (*pb.ListFinalsResponse, error) {
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
		"id":               true,
		"title":            true,
		"supervisor_grade": true,
		"enrollment_code":  true,
		"final_grade":      true,
		"status":           true,
		"notes":            true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Final %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count finals: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, supervisor_grade, enrollment_code, final_grade, status, notes, completion_date, created_at, updated_at, created_by, updated_by
		FROM Final
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list finals: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Final{}
	for rows.Next() {
		var entity pb.Final
		var createdAt, updatedAt, completionDate sql.NullTime
		var updatedBy sql.NullString
		var StatusStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.SupervisorGrade,
			&entity.EnrollmentCode,
			&entity.FinalGrade,
			&StatusStr,
			&entity.Notes,
			&completionDate,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan final: %v", err)
		}

		// Convert Status string to enum
		switch StatusStr {
		case "pending":
			entity.Status = pb.FinalStatus_PENDING
		case "passed":
			entity.Status = pb.FinalStatus_PASSED
		case "failed":
			entity.Status = pb.FinalStatus_FAILED
		case "completed":
			entity.Status = pb.FinalStatus_COMPLETED
		default:
			entity.Status = pb.FinalStatus_PENDING
		}

		if completionDate.Valid {
			entity.CompletionDate = timestamppb.New(completionDate.Time)
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
		return nil, status.Errorf(codes.Internal, "error iterating finals: %v", err)
	}

	return &pb.ListFinalsResponse{
		Finals:   entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
