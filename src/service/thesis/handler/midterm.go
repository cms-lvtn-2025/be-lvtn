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



// CreateMidterm creates a new Midterm record
func (h *Handler) CreateMidterm(ctx context.Context, req *pb.CreateMidtermRequest) (*pb.CreateMidtermResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	
	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	Grade := int32(0)
	if req.Grade != nil {
		Grade = *req.Grade
	}
	Feedback := ""
	if req.Feedback != nil {
		Feedback = *req.Feedback
	}
	
	// Convert Status enum to string
	StatusValue := pb.MidtermStatus_NOT_SUBMITTED
	
	StatusValue = req.Status
	StatusStr := "not_submitted"
	switch StatusValue {
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
		INSERT INTO Midterm (id, title, grade, status, feedback, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		Grade,
		StatusStr,
		Feedback,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "midterm already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create midterm: %v", err)
	}

	result, err := h.GetMidterm(ctx, &pb.GetMidtermRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get midterm")
	}
	return &pb.CreateMidtermResponse{
		Midterm: result.GetMidterm(),
	}, nil
}













// GetMidterm retrieves a Midterm by ID
func (h *Handler) GetMidterm(ctx context.Context, req *pb.GetMidtermRequest) (*pb.GetMidtermResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, grade, status, feedback, created_at, updated_at, created_by, updated_by
		FROM Midterm
		WHERE id = ?
	`

	var entity pb.Midterm
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string
	
	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
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

	return &pb.GetMidtermResponse{
		Midterm: &entity,
	}, nil
}













// UpdateMidterm updates an existing Midterm
func (h *Handler) UpdateMidterm(ctx context.Context, req *pb.UpdateMidtermRequest) (*pb.UpdateMidtermResponse, error) {
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
	if req.Grade != nil {
		updateFields = append(updateFields, "grade = ?")
		args = append(args, *req.Grade)
		
	}
	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		StatusStr := "not_submitted"
		switch *req.Status {
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
		
	}
	if req.Feedback != nil {
		updateFields = append(updateFields, "feedback = ?")
		args = append(args, *req.Feedback)
		
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
		UPDATE Midterm
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update midterm: %v", err)
	}

	result, err := h.GetMidterm(ctx, &pb.GetMidtermRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get midterm")
	}
	return &pb.UpdateMidtermResponse{
		Midterm: result.GetMidterm(),
	}, nil
}













// DeleteMidterm deletes a Midterm by ID
func (h *Handler) DeleteMidterm(ctx context.Context, req *pb.DeleteMidtermRequest) (*pb.DeleteMidtermResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Midterm WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
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
		"title": true,
		"grade": true,
		"status": true,
		"feedback": true,
		
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Midterm %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count midterms: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, grade, status, feedback, created_at, updated_at, created_by, updated_by
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


