package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/role"
	"thaily/src/pkg/helper"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateRoleSystem creates a new RoleSystem record
func (h *Handler) CreateRoleSystem(ctx context.Context, req *pb.CreateRoleSystemRequest) (*pb.CreateRoleSystemResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.TeacherCode == "" {
		return nil, status.Error(codes.InvalidArgument, "teacher_code is required")
	}
	if req.SemesterCode == "" {
		return nil, status.Error(codes.InvalidArgument, "semester_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields

	// Convert Role enum to string
	RoleValue := pb.RoleType_ACADEMIC_AFFAIRS_STAFF

	RoleValue = req.Role
	RoleStr := "academic_affairs_staff"
	switch RoleValue {
	case pb.RoleType_ACADEMIC_AFFAIRS_STAFF:
		RoleStr = "academic_affairs_staff"
	case pb.RoleType_DEPARTMENT_LECTURER:
		RoleStr = "department_lecturer"
	case pb.RoleType_TEACHER:
		RoleStr = "teacher"
	}

	// Insert into database
	query := `
		INSERT INTO RoleSystem (id, title, teacher_code, role, semester_code, activate, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		req.TeacherCode,
		RoleStr,
		req.SemesterCode,
		req.Activate,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "rolesystem already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create rolesystem: %v", err)
	}

	result, err := h.GetRoleSystem(ctx, &pb.GetRoleSystemRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get rolesystem")
	}
	return &pb.CreateRoleSystemResponse{
		RoleSystem: result.GetRoleSystem(),
	}, nil
}

// GetRoleSystem retrieves a RoleSystem by ID
func (h *Handler) GetRoleSystem(ctx context.Context, req *pb.GetRoleSystemRequest) (*pb.GetRoleSystemResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, teacher_code, role, semester_code, activate, created_at, updated_at, created_by, updated_by
		FROM RoleSystem
		WHERE id = ?
	`

	var entity pb.RoleSystem
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var RoleStr string

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.TeacherCode,
		&RoleStr,
		&entity.SemesterCode,
		&entity.Activate,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "rolesystem not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get rolesystem: %v", err)
	}

	// Convert Role string to enum
	switch RoleStr {
	case "academic_affairs_staff":
		entity.Role = pb.RoleType_ACADEMIC_AFFAIRS_STAFF
	case "department_lecturer":
		entity.Role = pb.RoleType_DEPARTMENT_LECTURER
	case "teacher":
		entity.Role = pb.RoleType_TEACHER
	default:
		entity.Role = pb.RoleType_TEACHER
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

	return &pb.GetRoleSystemResponse{
		RoleSystem: &entity,
	}, nil
}

// UpdateRoleSystem updates an existing RoleSystem
func (h *Handler) UpdateRoleSystem(ctx context.Context, req *pb.UpdateRoleSystemRequest) (*pb.UpdateRoleSystemResponse, error) {
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
	if req.TeacherCode != nil {
		updateFields = append(updateFields, "teacher_code = ?")
		args = append(args, *req.TeacherCode)

	}
	if req.Role != nil {
		updateFields = append(updateFields, "role = ?")
		RoleStr := "academic_affairs_staff"
		switch *req.Role {
		case pb.RoleType_ACADEMIC_AFFAIRS_STAFF:
			RoleStr = "academic_affairs_staff"
		case pb.RoleType_DEPARTMENT_LECTURER:
			RoleStr = "department_lecturer"
		case pb.RoleType_TEACHER:
			RoleStr = "teacher"
		}
		args = append(args, RoleStr)

	}
	if req.SemesterCode != nil {
		updateFields = append(updateFields, "semester_code = ?")
		args = append(args, *req.SemesterCode)

	}
	if req.Activate != nil {
		updateFields = append(updateFields, "activate = ?")
		args = append(args, *req.Activate)

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
		UPDATE RoleSystem
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update rolesystem: %v", err)
	}

	result, err := h.GetRoleSystem(ctx, &pb.GetRoleSystemRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get rolesystem")
	}
	return &pb.UpdateRoleSystemResponse{
		RoleSystem: result.GetRoleSystem(),
	}, nil
}

// DeleteRoleSystem deletes a RoleSystem by ID
func (h *Handler) DeleteRoleSystem(ctx context.Context, req *pb.DeleteRoleSystemRequest) (*pb.DeleteRoleSystemResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM RoleSystem WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete rolesystem: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "rolesystem not found")
	}

	return &pb.DeleteRoleSystemResponse{
		Success: true,
	}, nil
}

// ListRoleSystems lists RoleSystems with pagination and filtering
func (h *Handler) ListRoleSystems(ctx context.Context, req *pb.ListRoleSystemsRequest) (*pb.ListRoleSystemsResponse, error) {
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

		"title":         true,
		"teacher_code":  true,
		"role":          true,
		"semester_code": true,
		"activate":      true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM RoleSystem %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count rolesystems: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, teacher_code, role, semester_code, activate, created_at, updated_at, created_by, updated_by
		FROM RoleSystem
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list rolesystems: %v", err)
	}
	defer rows.Close()

	entities := []*pb.RoleSystem{}
	for rows.Next() {
		var entity pb.RoleSystem
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var RoleStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.TeacherCode,
			&RoleStr,
			&entity.SemesterCode,
			&entity.Activate,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan rolesystem: %v", err)
		}

		// Convert Role string to enum
		switch RoleStr {
		case "academic_affairs_staff":
			entity.Role = pb.RoleType_ACADEMIC_AFFAIRS_STAFF
		case "department_lecturer":
			entity.Role = pb.RoleType_DEPARTMENT_LECTURER
		case "teacher":
			entity.Role = pb.RoleType_TEACHER
		default:
			entity.Role = pb.RoleType_TEACHER
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
		return nil, status.Errorf(codes.Internal, "error iterating rolesystems: %v", err)
	}

	return &pb.ListRoleSystemsResponse{
		RoleSystems: entities,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}
