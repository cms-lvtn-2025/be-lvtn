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

// CreateFinal creates a new Final record
func (h *Handler) CreateFinal(ctx context.Context, req *pb.CreateFinalRequest) (*pb.CreateFinalResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	SupervisorGrade := int32(0)
	if req.SupervisorGrade != nil {
		SupervisorGrade = *req.SupervisorGrade
	}
	DepartmentGrade := int32(0)
	if req.DepartmentGrade != nil {
		DepartmentGrade = *req.DepartmentGrade
	}
	FinalGrade := int32(0)
	if req.FinalGrade != nil {
		FinalGrade = *req.FinalGrade
	}
	Notes := ""
	if req.Notes != nil {
		Notes = *req.Notes
	}

	// Convert Status enum to string
	StatusValue := pb.FinalStatus_PENDING

	StatusValue = req.Status
	StatusStr := "pending"
	switch StatusValue {
	case pb.FinalStatus_PENDING:
		StatusStr = "pending"
	case pb.FinalStatus_PASSED:
		StatusStr = "passed"
	case pb.FinalStatus_FAILED:
		StatusStr = "failed"
	case pb.FinalStatus_COMPLETED:
		StatusStr = "completed"
	}

	// Insert into database
	query := `
		INSERT INTO Final (id, title, supervisor_grade, department_grade, final_grade, status, notes, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		SupervisorGrade,
		DepartmentGrade,
		FinalGrade,
		StatusStr,
		Notes,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "final already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create final: %v", err)
	}

	result, err := h.GetFinal(ctx, &pb.GetFinalRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get final")
	}
	return &pb.CreateFinalResponse{
		Final: result.GetFinal(),
	}, nil
}

// GetFinal retrieves a Final by ID
func (h *Handler) GetFinal(ctx context.Context, req *pb.GetFinalRequest) (*pb.GetFinalResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, supervisor_grade, department_grade, final_grade, status, notes, created_at, updated_at, created_by, updated_by
		FROM Final
		WHERE id = ?
	`

	var entity pb.Final
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.SupervisorGrade,
		&entity.DepartmentGrade,
		&entity.FinalGrade,
		&StatusStr,
		&entity.Notes,
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

	if createdAt.Valid {
		entity.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		entity.UpdatedAt = timestamppb.New(updatedAt.Time)
	}
	if updatedBy.Valid {
		entity.UpdatedBy = updatedBy.String
	}

	return &pb.GetFinalResponse{
		Final: &entity,
	}, nil
}

// UpdateFinal updates an existing Final
func (h *Handler) UpdateFinal(ctx context.Context, req *pb.UpdateFinalRequest) (*pb.UpdateFinalResponse, error) {
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
	if req.SupervisorGrade != nil {
		updateFields = append(updateFields, "supervisor_grade = ?")
		args = append(args, *req.SupervisorGrade)

	}
	if req.DepartmentGrade != nil {
		updateFields = append(updateFields, "department_grade = ?")
		args = append(args, *req.DepartmentGrade)

	}
	if req.FinalGrade != nil {
		updateFields = append(updateFields, "final_grade = ?")
		args = append(args, *req.FinalGrade)

	}
	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		StatusStr := "pending"
		switch *req.Status {
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

	}
	if req.Notes != nil {
		updateFields = append(updateFields, "notes = ?")
		args = append(args, *req.Notes)

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
		UPDATE Final
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update final: %v", err)
	}

	result, err := h.GetFinal(ctx, &pb.GetFinalRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get final")
	}
	return &pb.UpdateFinalResponse{
		Final: result.GetFinal(),
	}, nil
}

// DeleteFinal deletes a Final by ID
func (h *Handler) DeleteFinal(ctx context.Context, req *pb.DeleteFinalRequest) (*pb.DeleteFinalResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Final WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
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
		"id": true,

		"title":            true,
		"supervisor_grade": true,
		"department_grade": true,
		"final_grade":      true,
		"status":           true,
		"notes":            true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Final %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count finals: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, supervisor_grade, department_grade, final_grade, status, notes, created_at, updated_at, created_by, updated_by
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
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var StatusStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.SupervisorGrade,
			&entity.DepartmentGrade,
			&entity.FinalGrade,
			&StatusStr,
			&entity.Notes,
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
