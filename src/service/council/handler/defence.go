package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/council"
	"thaily/src/pkg/helper"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)



// CreateDefence creates a new Defence record
func (h *Handler) CreateDefence(ctx context.Context, req *pb.CreateDefenceRequest) (*pb.CreateDefenceResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.CouncilCode == "" {
		return nil, status.Error(codes.InvalidArgument, "council_code is required")
	}
	if req.TeacherCode == "" {
		return nil, status.Error(codes.InvalidArgument, "teacher_code is required")
	}
	
	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	
	// Convert Position enum to string
	PositionValue := pb.DefencePosition_PRESIDENT
	
	PositionValue = req.Position
	PositionStr := "president"
	switch PositionValue {
	case pb.DefencePosition_PRESIDENT:
		PositionStr = "president"
	case pb.DefencePosition_SECRETARY:
		PositionStr = "secretary"
	case pb.DefencePosition_REVIEWER:
		PositionStr = "reviewer"
	case pb.DefencePosition_MEMBER:
		PositionStr = "member"
	}
	
	// Insert into database
	query := `
		INSERT INTO Defence (id, title, council_code, teacher_code, position, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		req.CouncilCode,
		req.TeacherCode,
		PositionStr,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "defence already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create defence: %v", err)
	}

	result, err := h.GetDefence(ctx, &pb.GetDefenceRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get defence")
	}
	return &pb.CreateDefenceResponse{
		Defence: result.GetDefence(),
	}, nil
}













// GetDefence retrieves a Defence by ID
func (h *Handler) GetDefence(ctx context.Context, req *pb.GetDefenceRequest) (*pb.GetDefenceResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, council_code, teacher_code, position, created_at, updated_at, created_by, updated_by
		FROM Defence
		WHERE id = ?
	`

	var entity pb.Defence
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var PositionStr string
	
	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.CouncilCode,
		&entity.TeacherCode,
		&PositionStr,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "defence not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get defence: %v", err)
	}

	// Convert Position string to enum
	switch PositionStr {
	case "president":
		entity.Position = pb.DefencePosition_PRESIDENT
	case "secretary":
		entity.Position = pb.DefencePosition_SECRETARY
	case "reviewer":
		entity.Position = pb.DefencePosition_REVIEWER
	case "member":
		entity.Position = pb.DefencePosition_MEMBER
	default:
		entity.Position = pb.DefencePosition_PRESIDENT
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

	return &pb.GetDefenceResponse{
		Defence: &entity,
	}, nil
}













// UpdateDefence updates an existing Defence
func (h *Handler) UpdateDefence(ctx context.Context, req *pb.UpdateDefenceRequest) (*pb.UpdateDefenceResponse, error) {
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
	if req.CouncilCode != nil {
		updateFields = append(updateFields, "council_code = ?")
		args = append(args, *req.CouncilCode)
		
	}
	if req.TeacherCode != nil {
		updateFields = append(updateFields, "teacher_code = ?")
		args = append(args, *req.TeacherCode)
		
	}
	if req.Position != nil {
		updateFields = append(updateFields, "position = ?")
		PositionStr := "president"
		switch *req.Position {
		case pb.DefencePosition_PRESIDENT:
			PositionStr = "president"
		case pb.DefencePosition_SECRETARY:
			PositionStr = "secretary"
		case pb.DefencePosition_REVIEWER:
			PositionStr = "reviewer"
		case pb.DefencePosition_MEMBER:
			PositionStr = "member"
		}
		args = append(args, PositionStr)
		
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
		UPDATE Defence
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update defence: %v", err)
	}

	result, err := h.GetDefence(ctx, &pb.GetDefenceRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get defence")
	}
	return &pb.UpdateDefenceResponse{
		Defence: result.GetDefence(),
	}, nil
}













// DeleteDefence deletes a Defence by ID
func (h *Handler) DeleteDefence(ctx context.Context, req *pb.DeleteDefenceRequest) (*pb.DeleteDefenceResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Defence WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete defence: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "defence not found")
	}

	return &pb.DeleteDefenceResponse{
		Success: true,
	}, nil
}













// ListDefences lists Defences with pagination and filtering
func (h *Handler) ListDefences(ctx context.Context, req *pb.ListDefencesRequest) (*pb.ListDefencesResponse, error) {
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
		"council_code": true,
		"teacher_code": true,
		"position": true,
		
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Defence %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count defences: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, council_code, teacher_code, position, created_at, updated_at, created_by, updated_by
		FROM Defence
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list defences: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Defence{}
	for rows.Next() {
		var entity pb.Defence
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var PositionStr string
		
		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.CouncilCode,
			&entity.TeacherCode,
			&PositionStr,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan defence: %v", err)
		}

		// Convert Position string to enum
		switch PositionStr {
		case "president":
			entity.Position = pb.DefencePosition_PRESIDENT
		case "secretary":
			entity.Position = pb.DefencePosition_SECRETARY
		case "reviewer":
			entity.Position = pb.DefencePosition_REVIEWER
		case "member":
			entity.Position = pb.DefencePosition_MEMBER
		default:
			entity.Position = pb.DefencePosition_PRESIDENT
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
		return nil, status.Errorf(codes.Internal, "error iterating defences: %v", err)
	}

	return &pb.ListDefencesResponse{
		Defences: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}


