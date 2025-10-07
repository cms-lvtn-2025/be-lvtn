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

func (h *Handler) CreateTeacher(ctx context.Context, req *pb.CreateTeacherRequest) (*pb.CreateTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate require field
	if req.Email == "" || req.MajorCode == "" || req.Username == "" || req.MajorCode == "" || req.SemesterCode == "" {
		return nil, status.Error(codes.InvalidArgument, "email, username, major_code, class_code, and semester_code are required")
	}

	// Generate UUID for teacher
	teacherID := uuid.New().String()

	// Set default gender if not provided
	gender := req.Gender
	// Convert gender enum to string
	genderStr := "male"
	switch gender {
	case pb.Gender_MALE:
		genderStr = "male"
	case pb.Gender_FEMALE:
		genderStr = "female"
	case pb.Gender_OTHER:
		genderStr = "other"
	}

	query := `
		INSERT INTO Teacher (id, email, username, gender, major_code, semester_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	if _, err := h.execQuery(ctx, query,
		teacherID,
		req.Email,
		req.Username,
		genderStr,
		req.MajorCode,
		req.SemesterCode,
		req.CreatedBy,
	); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "teacher with this email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create teacher: %v", err)
	}

	result, err := h.GetTeacher(ctx, &pb.GetTeacherRequest{
		Id: teacherID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get teacher: %v", err)
	}

	return &pb.CreateTeacherResponse{
		Teacher: result.GetTeacher(),
	}, nil
}

func (h *Handler) GetTeacher(ctx context.Context, req *pb.GetTeacherRequest) (*pb.GetTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, email, username, gender, major_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Teahcer
		WHERE id = ?
	`

	var teacher pb.Teacher
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var genderStr string

	err := h.queryRow(ctx, query, req.Id).Scan(
		&teacher.Id,
		&teacher.Email,
		&teacher.Username,
		&genderStr,
		&teacher.MajorCode,
		&teacher.SemesterCode,
		&createdAt,
		&updatedAt,
		&teacher.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "teacher not found")
		}
		return nil, status.Errorf(codes.Internal, "teacher to get student: %v", err)
	}

	// Convert gender string to enum
	switch genderStr {
	case "male":
		teacher.Gender = pb.Gender_MALE
	case "female":
		teacher.Gender = pb.Gender_FEMALE
	case "other":
		teacher.Gender = pb.Gender_OTHER
	default:
		teacher.Gender = pb.Gender_MALE
	}
	if createdAt.Valid {
		teacher.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		teacher.UpdatedAt = timestamppb.New(updatedAt.Time)
	}
	if updatedBy.Valid {
		teacher.UpdatedBy = updatedBy.String
	}

	return &pb.GetTeacherResponse{
		Teacher: &teacher,
	}, nil
}

func (h *Handler) UpdateTeacher(ctx context.Context, req *pb.UpdateTeacherRequest) (*pb.UpdateTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

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
	if req.MajorCode != nil {
		updateFields = append(updateFields, "major_code = ?")
		args = append(args, *req.MajorCode)
	}
	if req.SemesterCode != nil {
		updateFields = append(updateFields, "semester_code = ?")
		args = append(args, *req.SemesterCode)

	}
	if req.Gender != nil {
		updateFields = append(updateFields, "gender = ?")
		genderStr := "male"
		switch *req.Gender {
		case pb.Gender_MALE:
			genderStr = "male"
		case pb.Gender_FEMALE:
			genderStr = "female"
		case pb.Gender_OTHER:
			genderStr = "other"
		}
		args = append(args, genderStr)
	}
	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "update field is required")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.UpdatedBy)

	updateFields = append(updateFields, "updated_at = NOW()")

	args = append(args, req.Id)

	query := fmt.Sprintf(`
		UPDATE Teacher
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	if _, err := h.execQuery(ctx, query, args...); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update student: %v", err)
	}

	result, err := h.GetTeacher(ctx, &pb.GetTeacherRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get student")
	}
	return &pb.UpdateTeacherResponse{
		Teacher: result.GetTeacher(),
	}, nil
}

func (h *Handler) DeleteTeacher(ctx context.Context, req *pb.DeleteTeacherRequest) (*pb.DeleteTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Teacher WHERE id = $1`

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
		"email":         true,
		"username":      true,
		"gender":        true,
		"semester_code": true,
		"major_code":    true,
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

	// Get students with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, email, phone, username, gender, major_code, class_code, semester_code, created_at, updated_at, created_by, updated_by
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

	teachers := []*pb.Teacher{}
	for rows.Next() {
		var teacher pb.Teacher
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var genderStr string

		err := rows.Scan(
			&teacher.Id,
			&teacher.Email,
			&teacher.Username,
			&genderStr,
			&teacher.MajorCode,
			&teacher.SemesterCode,
			&createdAt,
			&updatedAt,
			&teacher.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan student: %v", err)
		}

		// Convert gender string to enum
		switch genderStr {
		case "male":
			teacher.Gender = pb.Gender_MALE
		case "female":
			teacher.Gender = pb.Gender_FEMALE
		case "other":
			teacher.Gender = pb.Gender_OTHER
		default:
			teacher.Gender = pb.Gender_MALE
		}

		if createdAt.Valid {
			teacher.CreatedAt = timestamppb.New(createdAt.Time)
		}
		if updatedAt.Valid {
			teacher.UpdatedAt = timestamppb.New(updatedAt.Time)
		}
		if updatedBy.Valid {
			teacher.UpdatedBy = updatedBy.String
		}

		teachers = append(teachers, &teacher)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error iterating students: %v", err)
	}

	return &pb.ListTeachersResponse{
		Teachers: teachers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
