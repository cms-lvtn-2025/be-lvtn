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



// CreateGradeReview creates a new GradeReview record
func (h *Handler) CreateGradeReview(ctx context.Context, req *pb.CreateGradeReviewRequest) (*pb.CreateGradeReviewResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.TeacherCode == "" {
		return nil, status.Error(codes.InvalidArgument, "teacher_code is required")
	}
	
	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	ReviewGrade := int32(0)
	if req.ReviewGrade != nil {
		ReviewGrade = *req.ReviewGrade
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
		INSERT INTO GradeReview (id, title, review_grade, teacher_code, status, notes, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		ReviewGrade,
		req.TeacherCode,
		StatusStr,
		Notes,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "gradereview already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create gradereview: %v", err)
	}

	result, err := h.GetGradeReview(ctx, &pb.GetGradeReviewRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get gradereview")
	}
	return &pb.CreateGradeReviewResponse{
		GradeReview: result.GetGradeReview(),
	}, nil
}













// GetGradeReview retrieves a GradeReview by ID
func (h *Handler) GetGradeReview(ctx context.Context, req *pb.GetGradeReviewRequest) (*pb.GetGradeReviewResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, review_grade, teacher_code, status, notes, created_at, updated_at, created_by, updated_by
		FROM GradeReview
		WHERE id = ?
	`

	var entity pb.GradeReview
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string
	
	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.ReviewGrade,
		&entity.TeacherCode,
		&StatusStr,
		&entity.Notes,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "gradereview not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get gradereview: %v", err)
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

	return &pb.GetGradeReviewResponse{
		GradeReview: &entity,
	}, nil
}













// UpdateGradeReview updates an existing GradeReview
func (h *Handler) UpdateGradeReview(ctx context.Context, req *pb.UpdateGradeReviewRequest) (*pb.UpdateGradeReviewResponse, error) {
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
	if req.ReviewGrade != nil {
		updateFields = append(updateFields, "review_grade = ?")
		args = append(args, *req.ReviewGrade)
		
	}
	if req.TeacherCode != nil {
		updateFields = append(updateFields, "teacher_code = ?")
		args = append(args, *req.TeacherCode)
		
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
		UPDATE GradeReview
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update gradereview: %v", err)
	}

	result, err := h.GetGradeReview(ctx, &pb.GetGradeReviewRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get gradereview")
	}
	return &pb.UpdateGradeReviewResponse{
		GradeReview: result.GetGradeReview(),
	}, nil
}













// DeleteGradeReview deletes a GradeReview by ID
func (h *Handler) DeleteGradeReview(ctx context.Context, req *pb.DeleteGradeReviewRequest) (*pb.DeleteGradeReviewResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM GradeReview WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete gradereview: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "gradereview not found")
	}

	return &pb.DeleteGradeReviewResponse{
		Success: true,
	}, nil
}













// ListGradeReviews lists GradeReviews with pagination and filtering
func (h *Handler) ListGradeReviews(ctx context.Context, req *pb.ListGradeReviewsRequest) (*pb.ListGradeReviewsResponse, error) {
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
		"review_grade": true,
		"teacher_code": true,
		"status": true,
		"notes": true,
		
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM GradeReview %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count gradereviews: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, review_grade, teacher_code, status, notes, created_at, updated_at, created_by, updated_by
		FROM GradeReview
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list gradereviews: %v", err)
	}
	defer rows.Close()

	entities := []*pb.GradeReview{}
	for rows.Next() {
		var entity pb.GradeReview
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var StatusStr string
		
		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.ReviewGrade,
			&entity.TeacherCode,
			&StatusStr,
			&entity.Notes,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan gradereview: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating gradereviews: %v", err)
	}

	return &pb.ListGradeReviewsResponse{
		GradeReviews: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}


