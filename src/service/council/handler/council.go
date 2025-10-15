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



// CreateCouncil creates a new Council record
func (h *Handler) CreateCouncil(ctx context.Context, req *pb.CreateCouncilRequest) (*pb.CreateCouncilResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
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
	
	
	// Insert into database
	query := `
		INSERT INTO Council (id, title, major_code, semester_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		req.MajorCode,
		req.SemesterCode,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "council already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create council: %v", err)
	}

	result, err := h.GetCouncil(ctx, &pb.GetCouncilRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get council")
	}
	return &pb.CreateCouncilResponse{
		Council: result.GetCouncil(),
	}, nil
}













// GetCouncil retrieves a Council by ID
func (h *Handler) GetCouncil(ctx context.Context, req *pb.GetCouncilRequest) (*pb.GetCouncilResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, major_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Council
		WHERE id = ?
	`

	var entity pb.Council
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	
	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.MajorCode,
		&entity.SemesterCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "council not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get council: %v", err)
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

	return &pb.GetCouncilResponse{
		Council: &entity,
	}, nil
}













// UpdateCouncil updates an existing Council
func (h *Handler) UpdateCouncil(ctx context.Context, req *pb.UpdateCouncilRequest) (*pb.UpdateCouncilResponse, error) {
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
		UPDATE Council
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update council: %v", err)
	}

	result, err := h.GetCouncil(ctx, &pb.GetCouncilRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get council")
	}
	return &pb.UpdateCouncilResponse{
		Council: result.GetCouncil(),
	}, nil
}













// DeleteCouncil deletes a Council by ID
func (h *Handler) DeleteCouncil(ctx context.Context, req *pb.DeleteCouncilRequest) (*pb.DeleteCouncilResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Council WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete council: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "council not found")
	}

	return &pb.DeleteCouncilResponse{
		Success: true,
	}, nil
}













// ListCouncils lists Councils with pagination and filtering
func (h *Handler) ListCouncils(ctx context.Context, req *pb.ListCouncilsRequest) (*pb.ListCouncilsResponse, error) {
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
		"major_code": true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Council %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count councils: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, major_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Council
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list councils: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Council{}
	for rows.Next() {
		var entity pb.Council
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		
		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.MajorCode,
			&entity.SemesterCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan council: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating councils: %v", err)
	}

	return &pb.ListCouncilsResponse{
		Councils: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}


