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



// CreateTopicCouncil creates a new TopicCouncil record
func (h *Handler) CreateTopicCouncil(ctx context.Context, req *pb.CreateTopicCouncilRequest) (*pb.CreateTopicCouncilResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.TopicCode == "" {
		return nil, status.Error(codes.InvalidArgument, "topic_code is required")
	}
	
	// Generate UUID
	id := uuid.New().String()

	// Prepare fields
	CouncilCode := ""
	if req.CouncilCode != nil {
		CouncilCode = *req.CouncilCode
	}
	
	// Convert Stage enum to string
	StageValue := pb.TopicStage_STAGE_DACN
	
	StageValue = req.Stage
	StageStr := "stage_dacn"
	switch StageValue {
	case pb.TopicStage_STAGE_DACN:
		StageStr = "stage_dacn"
	case pb.TopicStage_STAGE_LVTN:
		StageStr = "stage_lvtn"
	}
	
	// Insert into database
	query := `
		INSERT INTO TopicCouncil (id, title, stage, topic_code, council_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		StageStr,
		req.TopicCode,
		CouncilCode,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "topiccouncil already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create topiccouncil: %v", err)
	}

	result, err := h.GetTopicCouncil(ctx, &pb.GetTopicCouncilRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topiccouncil")
	}
	return &pb.CreateTopicCouncilResponse{
		TopicCouncil: result.GetTopicCouncil(),
	}, nil
}













// GetTopicCouncil retrieves a TopicCouncil by ID
func (h *Handler) GetTopicCouncil(ctx context.Context, req *pb.GetTopicCouncilRequest) (*pb.GetTopicCouncilResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `
		SELECT id, title, stage, topic_code, council_code, created_at, updated_at, created_by, updated_by
		FROM TopicCouncil
		WHERE id = ?
	`

	var entity pb.TopicCouncil
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StageStr string
	
	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&StageStr,
		&entity.TopicCode,
		&entity.CouncilCode,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "topiccouncil not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get topiccouncil: %v", err)
	}

	// Convert Stage string to enum
	switch StageStr {
	case "stage_dacn":
		entity.Stage = pb.TopicStage_STAGE_DACN
	case "stage_lvtn":
		entity.Stage = pb.TopicStage_STAGE_LVTN
	default:
		entity.Stage = pb.TopicStage_STAGE_DACN
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

	return &pb.GetTopicCouncilResponse{
		TopicCouncil: &entity,
	}, nil
}













// UpdateTopicCouncil updates an existing TopicCouncil
func (h *Handler) UpdateTopicCouncil(ctx context.Context, req *pb.UpdateTopicCouncilRequest) (*pb.UpdateTopicCouncilResponse, error) {
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
	if req.Stage != nil {
		updateFields = append(updateFields, "stage = ?")
		StageStr := "stage_dacn"
		switch *req.Stage {
		case pb.TopicStage_STAGE_DACN:
			StageStr = "stage_dacn"
		case pb.TopicStage_STAGE_LVTN:
			StageStr = "stage_lvtn"
		}
		args = append(args, StageStr)
		
	}
	if req.TopicCode != nil {
		updateFields = append(updateFields, "topic_code = ?")
		args = append(args, *req.TopicCode)
		
	}
	if req.CouncilCode != nil {
		updateFields = append(updateFields, "council_code = ?")
		args = append(args, *req.CouncilCode)
		
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
		UPDATE TopicCouncil
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update topiccouncil: %v", err)
	}

	result, err := h.GetTopicCouncil(ctx, &pb.GetTopicCouncilRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get topiccouncil")
	}
	return &pb.UpdateTopicCouncilResponse{
		TopicCouncil: result.GetTopicCouncil(),
	}, nil
}













// DeleteTopicCouncil deletes a TopicCouncil by ID
func (h *Handler) DeleteTopicCouncil(ctx context.Context, req *pb.DeleteTopicCouncilRequest) (*pb.DeleteTopicCouncilResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM TopicCouncil WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete topiccouncil: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "topiccouncil not found")
	}

	return &pb.DeleteTopicCouncilResponse{
		Success: true,
	}, nil
}













// ListTopicCouncils lists TopicCouncils with pagination and filtering
func (h *Handler) ListTopicCouncils(ctx context.Context, req *pb.ListTopicCouncilsRequest) (*pb.ListTopicCouncilsResponse, error) {
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
		"stage": true,
		"topic_code": true,
		"council_code": true,
		
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM TopicCouncil %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count topiccouncils: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, stage, topic_code, council_code, created_at, updated_at, created_by, updated_by
		FROM TopicCouncil
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list topiccouncils: %v", err)
	}
	defer rows.Close()

	entities := []*pb.TopicCouncil{}
	for rows.Next() {
		var entity pb.TopicCouncil
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var StageStr string
		
		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&StageStr,
			&entity.TopicCode,
			&entity.CouncilCode,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan topiccouncil: %v", err)
		}

		// Convert Stage string to enum
		switch StageStr {
		case "stage_dacn":
			entity.Stage = pb.TopicStage_STAGE_DACN
		case "stage_lvtn":
			entity.Stage = pb.TopicStage_STAGE_LVTN
		default:
			entity.Stage = pb.TopicStage_STAGE_DACN
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
		return nil, status.Errorf(codes.Internal, "error iterating topiccouncils: %v", err)
	}

	return &pb.ListTopicCouncilsResponse{
		TopicCouncils: entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}


