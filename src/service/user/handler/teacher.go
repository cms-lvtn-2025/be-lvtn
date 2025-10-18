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

// CreateTeacher creates a new Teacher record
func (h *Handler) CreateTeacher(ctx context.Context, req *pb.CreateTeacherRequest) (*pb.CreateTeacherResponse, error) {
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
	if req.SemesterCode == "" {
		return nil, status.Error(codes.InvalidArgument, "semester_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields

	// Convert Gender enum to string
	GenderValue := pb.Gender_MALE

	GenderValue = req.Gender
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
		INSERT INTO Teacher (id, email, username, gender, major_code, semester_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Email,
		req.Username,
		GenderStr,
		req.MajorCode,
		req.SemesterCode,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "teacher already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create teacher: %v", err)
	}

	result, err := h.GetTeacher(ctx, &pb.GetTeacherRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get teacher")
	}
	return &pb.CreateTeacherResponse{
		Teacher: result.GetTeacher(),
	}, nil
}

// GetTeacher retrieves a Teacher by ID
func (h *Handler) GetTeacher(ctx context.Context, req *pb.GetTeacherRequest) (*pb.GetTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, email, username, gender, major_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Teacher
		WHERE id = ?
	`

	var entity pb.Teacher
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var GenderStr string

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Email,
		&entity.Username,
		&GenderStr,
		&entity.MajorCode,
		&entity.SemesterCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "teacher not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get teacher: %v", err)
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

	return &pb.GetTeacherResponse{
		Teacher: &entity,
	}, nil
}

// UpdateTeacher updates an existing Teacher
func (h *Handler) UpdateTeacher(ctx context.Context, req *pb.UpdateTeacherRequest) (*pb.UpdateTeacherResponse, error) {
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
		UPDATE Teacher
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update teacher: %v", err)
	}

	result, err := h.GetTeacher(ctx, &pb.GetTeacherRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get teacher")
	}
	return &pb.UpdateTeacherResponse{
		Teacher: result.GetTeacher(),
	}, nil
}

// DeleteTeacher deletes a Teacher by ID
func (h *Handler) DeleteTeacher(ctx context.Context, req *pb.DeleteTeacherRequest) (*pb.DeleteTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Teacher WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete teacher: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "teacher not found")
	}

	return &pb.DeleteTeacherResponse{
		Success: true,
	}, nil
}

// ListTeachers lists Teachers with pagination and filtering
func (h *Handler) ListTeachers(ctx context.Context, req *pb.ListTeachersRequest) (*pb.ListTeachersResponse, error) {
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
		"username":      true,
		"gender":        true,
		"major_code":    true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Teacher %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count teachers: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, email, username, gender, major_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Teacher
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list teachers: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Teacher{}
	for rows.Next() {
		var entity pb.Teacher
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var GenderStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Email,
			&entity.Username,
			&GenderStr,
			&entity.MajorCode,
			&entity.SemesterCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan teacher: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating teachers: %v", err)
	}

	return &pb.ListTeachersResponse{
		Teachers: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
