package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/user"
	"thaily/src/pkg/helper"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateStudent creates a new Student record
func (h *Handler) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.MajorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "major_code is required")
	}
	if req.ClassCode == "" {
		return nil, status.Error(codes.InvalidArgument, "class_code is required")
	}
	if req.SemesterCode == "" {
		return nil, status.Error(codes.InvalidArgument, "semester_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	Phone := ""
	if req.Phone != nil {
		Phone = *req.Phone
	}

	// Convert Gender enum to string
	GenderValue := pb.Gender_MALE
	if req.Gender != nil {
		GenderValue = *req.Gender
	}
	GenderStr := "male"
	switch GenderValue {
	case pb.Gender_MALE:
		GenderStr = "male"
	case pb.Gender_FEMALE:
		GenderStr = "female"
	case pb.Gender_OTHER:
		GenderStr = "other"
	}

	// Insert into database
	query := `
		INSERT INTO Student (id, email, phone, username, gender, major_code, class_code, semester_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Email,
		Phone,
		req.Username,
		GenderStr,
		req.MajorCode,
		req.ClassCode,
		req.SemesterCode,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "student already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create student: %v", err)
	}

	result, err := h.GetStudent(ctx, &pb.GetStudentRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get student")
	}
	return &pb.CreateStudentResponse{
		Student: result.GetStudent(),
	}, nil
}

// GetStudent retrieves a Student by ID
func (h *Handler) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.GetStudentResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, email, phone, username, gender, major_code, class_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Student
		WHERE id = ?
	`

	var entity pb.Student
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var GenderStr string

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Email,
		&entity.Phone,
		&entity.Username,
		&GenderStr,
		&entity.MajorCode,
		&entity.ClassCode,
		&entity.SemesterCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "student not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get student: %v", err)
	}

	// Convert Gender string to enum
	switch GenderStr {
	case "male":
		entity.Gender = pb.Gender_MALE
	case "female":
		entity.Gender = pb.Gender_FEMALE
	case "other":
		entity.Gender = pb.Gender_OTHER
	default:
		entity.Gender = pb.Gender_MALE
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

	return &pb.GetStudentResponse{
		Student: &entity,
	}, nil
}

// UpdateStudent updates an existing Student
func (h *Handler) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.Email != nil {
		updateFields = append(updateFields, "email = ?")
		args = append(args, *req.Email)

	}
	if req.Phone != nil {
		updateFields = append(updateFields, "phone = ?")
		args = append(args, *req.Phone)

	}
	if req.Username != nil {
		updateFields = append(updateFields, "username = ?")
		args = append(args, *req.Username)

	}
	if req.Gender != nil {
		updateFields = append(updateFields, "gender = ?")
		GenderStr := "male"
		switch *req.Gender {
		case pb.Gender_MALE:
			GenderStr = "male"
		case pb.Gender_FEMALE:
			GenderStr = "female"
		case pb.Gender_OTHER:
			GenderStr = "other"
		}
		args = append(args, GenderStr)

	}
	if req.MajorCode != nil {
		updateFields = append(updateFields, "major_code = ?")
		args = append(args, *req.MajorCode)

	}
	if req.ClassCode != nil {
		updateFields = append(updateFields, "class_code = ?")
		args = append(args, *req.ClassCode)

	}
	if req.SemesterCode != nil {
		updateFields = append(updateFields, "semester_code = ?")
		args = append(args, *req.SemesterCode)

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
		UPDATE Student
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update student: %v", err)
	}

	result, err := h.GetStudent(ctx, &pb.GetStudentRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get student")
	}
	return &pb.UpdateStudentResponse{
		Student: result.GetStudent(),
	}, nil
}

// DeleteStudent deletes a Student by ID
func (h *Handler) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*pb.DeleteStudentResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Student WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete student: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "student not found")
	}

	return &pb.DeleteStudentResponse{
		Success: true,
	}, nil
}

// ListStudents lists Students with pagination and filtering
func (h *Handler) ListStudents(ctx context.Context, req *pb.ListStudentsRequest) (*pb.ListStudentsResponse, error) {
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

		"email":         true,
		"phone":         true,
		"username":      true,
		"gender":        true,
		"major_code":    true,
		"class_code":    true,
		"semester_code": true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Student %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count students: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, email, phone, username, gender, major_code, class_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Student
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list students: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Student{}
	for rows.Next() {
		var entity pb.Student
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var GenderStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Email,
			&entity.Phone,
			&entity.Username,
			&GenderStr,
			&entity.MajorCode,
			&entity.ClassCode,
			&entity.SemesterCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan student: %v", err)
		}

		// Convert Gender string to enum
		switch GenderStr {
		case "male":
			entity.Gender = pb.Gender_MALE
		case "female":
			entity.Gender = pb.Gender_FEMALE
		case "other":
			entity.Gender = pb.Gender_OTHER
		default:
			entity.Gender = pb.Gender_MALE
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
		return nil, status.Errorf(codes.Internal, "error iterating students: %v", err)
	}

	return &pb.ListStudentsResponse{
		Students: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
