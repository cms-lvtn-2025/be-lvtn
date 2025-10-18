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

// CreateGradeDefenceCriterion creates a new GradeDefenceCriterion record
func (h *Handler) CreateGradeDefenceCriterion(ctx context.Context, req *pb.CreateGradeDefenceCriterionRequest) (*pb.CreateGradeDefenceCriterionResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields
	if req.GradeDefenceCode == "" {
		return nil, status.Error(codes.InvalidArgument, "grade_defence_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare optional fields
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	score := ""
	if req.Score != nil {
		score = *req.Score
	}
	maxScore := ""
	if req.MaxScore != nil {
		maxScore = *req.MaxScore
	}

	// Insert into database
	query := `
		INSERT INTO Grade_defence_criterion (id, grade_defence_code, name, score, maxScore, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GradeDefenceCode,
		name,
		score,
		maxScore,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "grade defence criterion already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create grade defence criterion: %v", err)
	}

	result, err := h.GetGradeDefenceCriterion(ctx, &pb.GetGradeDefenceCriterionRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get grade defence criterion")
	}
	return &pb.CreateGradeDefenceCriterionResponse{
		GradeDefenceCriterion: result.GetGradeDefenceCriterion(),
	}, nil
}

// GetGradeDefenceCriterion retrieves a GradeDefenceCriterion by ID
func (h *Handler) GetGradeDefenceCriterion(ctx context.Context, req *pb.GetGradeDefenceCriterionRequest) (*pb.GetGradeDefenceCriterionResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, grade_defence_code, name, score, maxScore, created_at, updated_at, created_by, updated_by
		FROM Grade_defence_criterion
		WHERE id = ?
	`

	var entity pb.GradeDefenceCriterion
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.GradeDefenceCode,
		&entity.Name,
		&entity.Score,
		&entity.MaxScore,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "grade defence criterion not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get grade defence criterion: %v", err)
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

	return &pb.GetGradeDefenceCriterionResponse{
		GradeDefenceCriterion: &entity,
	}, nil
}

// UpdateGradeDefenceCriterion updates an existing GradeDefenceCriterion
func (h *Handler) UpdateGradeDefenceCriterion(ctx context.Context, req *pb.UpdateGradeDefenceCriterionRequest) (*pb.UpdateGradeDefenceCriterionResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GradeDefenceCode != nil {
		updateFields = append(updateFields, "grade_defence_code = ?")
		args = append(args, *req.GradeDefenceCode)
	}
	if req.Name != nil {
		updateFields = append(updateFields, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Score != nil {
		updateFields = append(updateFields, "score = ?")
		args = append(args, *req.Score)
	}
	if req.MaxScore != nil {
		updateFields = append(updateFields, "maxScore = ?")
		args = append(args, *req.MaxScore)
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
		UPDATE Grade_defence_criterion
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update grade defence criterion: %v", err)
	}

	result, err := h.GetGradeDefenceCriterion(ctx, &pb.GetGradeDefenceCriterionRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get grade defence criterion")
	}
	return &pb.UpdateGradeDefenceCriterionResponse{
		GradeDefenceCriterion: result.GetGradeDefenceCriterion(),
	}, nil
}

// DeleteGradeDefenceCriterion deletes a GradeDefenceCriterion by ID
func (h *Handler) DeleteGradeDefenceCriterion(ctx context.Context, req *pb.DeleteGradeDefenceCriterionRequest) (*pb.DeleteGradeDefenceCriterionResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Grade_defence_criterion WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete grade defence criterion: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "grade defence criterion not found")
	}

	return &pb.DeleteGradeDefenceCriterionResponse{
		Success: true,
	}, nil
}

// ListGradeDefenceCriteria lists GradeDefenceCriteria with pagination and filtering
func (h *Handler) ListGradeDefenceCriteria(ctx context.Context, req *pb.ListGradeDefenceCriteriaRequest) (*pb.ListGradeDefenceCriteriaResponse, error) {
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
		"grade_defence_code": true,
		"name":               true,
		"score":              true,
		"maxScore":           true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Grade_defence_criterion %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count grade defence criteria: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, grade_defence_code, name, score, maxScore, created_at, updated_at, created_by, updated_by
		FROM Grade_defence_criterion
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list grade defence criteria: %v", err)
	}
	defer rows.Close()

	entities := []*pb.GradeDefenceCriterion{}
	for rows.Next() {
		var entity pb.GradeDefenceCriterion
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&entity.Id,
			&entity.GradeDefenceCode,
			&entity.Name,
			&entity.Score,
			&entity.MaxScore,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan grade defence criterion: %v", err)
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
		return nil, status.Errorf(codes.Internal, "error iterating grade defence criteria: %v", err)
	}

	return &pb.ListGradeDefenceCriteriaResponse{
		GradeDefenceCriteria: entities,
		Total:                total,
		Page:                 page,
		PageSize:             pageSize,
	}, nil
}
