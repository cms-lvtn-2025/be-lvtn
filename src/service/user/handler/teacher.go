package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/user"
	"thaily/src/service/pkg/helper"
	"thaily/src/service/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateTeacher creates a new Teacher record
func (h *Handler) CreateTeacher(ctx context.Context, req *pb.CreateTeacherRequest) (*pb.TeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.GetTeacher().GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetTeacher().GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.GetTeacher().GetMajorCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "major_code is required")
	}
	if req.GetTeacher().GetSemesterCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "semester_code is required")
	}

	// Use provided ID
	id := req.GetTeacher().GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Convert Gender enum to string
	GenderStr := "male"
	switch req.GetTeacher().GetGender() {
	case pb.Gender_MALE:
		GenderStr = "male"
	case pb.Gender_FEMALE:
		GenderStr = "female"
	case pb.Gender_OTHER:
		GenderStr = "other"
	}

	// Insert into database
	query := `
		INSERT INTO Teacher (id, email, username, gender, major_code, semester_code, msgv, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GetTeacher().GetEmail(),
		req.GetTeacher().GetUsername(),
		GenderStr,
		req.GetTeacher().GetMajorCode(),
		req.GetTeacher().GetSemesterCode(),
		req.GetTeacher().GetMsgv(),
		req.GetTeacher().GetCreatedBy(),
		req.GetTeacher().GetCreatedBy(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "teacher already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create teacher: %v", err)
	}

	id = fmt.Sprint(req.GetTeacher().GetMsgv(), "_", req.GetTeacher().GetSemesterCode())
	return h.GetTeacher(ctx, &pb.GetTeacherRequest{Id: id})
}

// GetTeacher retrieves a Teacher by ID
func (h *Handler) GetTeacher(ctx context.Context, req *pb.GetTeacherRequest) (*pb.TeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":            true,
		"email":         true,
		"username":      true,
		"gender":        true,
		"major_code":    true,
		"semester_code": true,
		"msgv":          true,
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
		SELECT id, email, username, gender, major_code, semester_code, msgv, created_at, updated_at, created_by, updated_by
		FROM Teacher
		%s
	`, whereClause)

	var entity pb.Teacher
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var GenderStr string

	err := h.queryRow(ctx, query, args...).Scan(
		&entity.Id,
		&entity.Email,
		&entity.Username,
		&GenderStr,
		&entity.MajorCode,
		&entity.SemesterCode,
		&entity.Msgv,
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

	return &pb.TeacherResponse{
		Teacher: &entity,
	}, nil
}

// UpdateTeacher updates an existing Teacher
func (h *Handler) UpdateTeacher(ctx context.Context, req *pb.UpdateTeacherRequest) (*pb.TeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GetTeacher().GetEmail() != "" {
		updateFields = append(updateFields, "email = ?")
		args = append(args, req.GetTeacher().GetEmail())
	}
	if req.GetTeacher().GetUsername() != "" {
		updateFields = append(updateFields, "username = ?")
		args = append(args, req.GetTeacher().GetUsername())
	}

	// Gender enum - always include
	updateFields = append(updateFields, "gender = ?")
	GenderStr := "male"
	switch req.GetTeacher().GetGender() {
	case pb.Gender_MALE:
		GenderStr = "male"
	case pb.Gender_FEMALE:
		GenderStr = "female"
	case pb.Gender_OTHER:
		GenderStr = "other"
	}
	args = append(args, GenderStr)

	if req.GetTeacher().GetMajorCode() != "" {
		updateFields = append(updateFields, "major_code = ?")
		args = append(args, req.GetTeacher().GetMajorCode())
	}
	if req.GetTeacher().GetSemesterCode() != "" {
		updateFields = append(updateFields, "semester_code = ?")
		args = append(args, req.GetTeacher().GetSemesterCode())
	}
	if req.GetTeacher().GetMsgv() != "" {
		updateFields = append(updateFields, "msgv = ?")
		args = append(args, req.GetTeacher().GetMsgv())
	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.GetTeacher().GetUpdatedBy())
	updateFields = append(updateFields, "updated_at = NOW()")

	// Build WHERE clause from filter
	whiteMap := map[string]bool{
		"id":            true,
		"email":         true,
		"username":      true,
		"gender":        true,
		"major_code":    true,
		"semester_code": true,
		"msgv":          true,
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
		UPDATE Teacher
		SET %s
		%s
	`, strings.Join(updateFields, ", "), whereClause)

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update teacher: %v", err)
	}

	return h.GetTeacher(ctx, &pb.GetTeacherRequest{Id: req.Id})
}

// DeleteTeacher deletes a Teacher by ID
func (h *Handler) DeleteTeacher(ctx context.Context, req *pb.DeleteTeacherRequest) (*pb.DeleteTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"id":            true,
		"email":         true,
		"username":      true,
		"gender":        true,
		"major_code":    true,
		"semester_code": true,
		"msgv":          true,
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

	query := fmt.Sprintf(`DELETE FROM Teacher %s`, whereClause)

	result, err := h.execQuery(ctx, query, args...)
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
		"id":            true,
		"email":         true,
		"username":      true,
		"gender":        true,
		"major_code":    true,
		"semester_code": true,
		"msgv":          true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Teacher %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count teachers: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, email, msgv, username, gender, major_code, semester_code, created_at, updated_at, created_by, updated_by
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
			&entity.Msgv,
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
