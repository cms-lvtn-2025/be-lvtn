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

// CreateEnrollment creates a new Enrollment record
func (h *Handler) CreateEnrollment(ctx context.Context, req *pb.CreateEnrollmentRequest) (*pb.EnrollmentResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.GetEnrollment().GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetEnrollment().GetStudentCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "student_code is required")
	}
	if req.GetEnrollment().GetTopicCouncilCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "topic_council_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Insert into database
	query := `
		INSERT INTO Enrollment (id, title, student_code, topic_council_code, final_code, grade_review_code, midterm_code, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GetEnrollment().GetTitle(),
		req.GetEnrollment().GetStudentCode(),
		req.GetEnrollment().GetTopicCouncilCode(),
		req.GetEnrollment().GetFinalCode(),
		req.GetEnrollment().GetGradeReviewCode(),
		req.GetEnrollment().GetMidtermCode(),
		req.GetEnrollment().GetCreatedBy(),
		req.GetEnrollment().GetCreatedBy(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "enrollment already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create enrollment: %v", err)
	}

	return h.GetEnrollment(ctx, &pb.GetEnrollmentRequest{Id: id})
}

// GetEnrollment retrieves a Enrollment by ID
func (h *Handler) GetEnrollment(ctx context.Context, req *pb.GetEnrollmentRequest) (*pb.EnrollmentResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":                 true,
		"title":              true,
		"student_code":       true,
		"topic_council_code": true,
		"final_code":         true,
		"grade_review_code":  true,
		"midterm_code":       true,
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
		SELECT id, title, student_code, topic_council_code, final_code, grade_review_code, midterm_code, created_at, updated_at, created_by, updated_by
		FROM Enrollment
		%s
	`, whereClause)

	var entity pb.Enrollment
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	err := h.queryRow(ctx, query, args...).Scan(
		&entity.Id,
		&entity.Title,
		&entity.StudentCode,
		&entity.TopicCouncilCode,
		&entity.FinalCode,
		&entity.GradeReviewCode,
		&entity.MidtermCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "enrollment not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get enrollment: %v", err)
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

	return &pb.EnrollmentResponse{
		Enrollment: &entity,
	}, nil
}

// UpdateEnrollment updates an existing Enrollment
func (h *Handler) UpdateEnrollment(ctx context.Context, req *pb.UpdateEnrollmentRequest) (*pb.EnrollmentResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GetEnrollment().GetTitle() != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, req.GetEnrollment().GetTitle())
	}
	if req.GetEnrollment().GetStudentCode() != "" {
		updateFields = append(updateFields, "student_code = ?")
		args = append(args, req.GetEnrollment().GetStudentCode())
	}
	if req.GetEnrollment().GetTopicCouncilCode() != "" {
		updateFields = append(updateFields, "topic_council_code = ?")
		args = append(args, req.GetEnrollment().GetTopicCouncilCode())
	}
	if req.GetEnrollment().GetFinalCode() != "" {
		updateFields = append(updateFields, "final_code = ?")
		args = append(args, req.GetEnrollment().GetFinalCode())
	}
	if req.GetEnrollment().GetGradeReviewCode() != "" {
		updateFields = append(updateFields, "grade_review_code = ?")
		args = append(args, req.GetEnrollment().GetGradeReviewCode())
	}
	if req.GetEnrollment().GetMidtermCode() != "" {
		updateFields = append(updateFields, "midterm_code = ?")
		args = append(args, req.GetEnrollment().GetMidtermCode())
	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.GetEnrollment().GetUpdatedBy())
	updateFields = append(updateFields, "updated_at = NOW()")

	// Build WHERE clause from filter
	whiteMap := map[string]bool{
		"id":                 true,
		"title":              true,
		"student_code":       true,
		"topic_council_code": true,
		"final_code":         true,
		"grade_review_code":  true,
		"midterm_code":       true,
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
		UPDATE Enrollment
		SET %s
		%s
	`, strings.Join(updateFields, ", "), whereClause)

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update enrollment: %v", err)
	}

	return h.GetEnrollment(ctx, &pb.GetEnrollmentRequest{Id: req.Id})
}

// DeleteEnrollment deletes a Enrollment by ID
func (h *Handler) DeleteEnrollment(ctx context.Context, req *pb.DeleteEnrollmentRequest) (*pb.DeleteEnrollmentResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":                 true,
		"title":              true,
		"student_code":       true,
		"topic_council_code": true,
		"final_code":         true,
		"grade_review_code":  true,
		"midterm_code":       true,
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

	query := fmt.Sprintf(`DELETE FROM Enrollment %s`, whereClause)

	result, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete enrollment: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "enrollment not found")
	}

	return &pb.DeleteEnrollmentResponse{
		Success: true,
	}, nil
}

// ListEnrollments lists Enrollments with pagination and filtering
func (h *Handler) ListEnrollments(ctx context.Context, req *pb.ListEnrollmentsRequest) (*pb.ListEnrollmentsResponse, error) {
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
		"id":                 true,
		"title":              true,
		"student_code":       true,
		"topic_council_code": true,
		"final_code":         true,
		"grade_review_code":  true,
		"midterm_code":       true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Enrollment %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count enrollments: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, student_code, topic_council_code, final_code, grade_review_code, midterm_code, created_at, updated_at, created_by, updated_by
		FROM Enrollment
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list enrollments: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Enrollment{}
	for rows.Next() {
		var entity pb.Enrollment
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.StudentCode,
			&entity.TopicCouncilCode,
			&entity.FinalCode,
			&entity.GradeReviewCode,
			&entity.MidtermCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan enrollment: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating enrollments: %v", err)
	}

	return &pb.ListEnrollmentsResponse{
		Enrollments: entities,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}
