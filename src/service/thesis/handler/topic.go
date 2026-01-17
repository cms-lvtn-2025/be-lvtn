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

// CreateTopic creates a new Topic record
func (h *Handler) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.TopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.GetTopic().GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetTopic().GetTitleEn() == "" {
		return nil, status.Error(codes.InvalidArgument, "title_en is required")
	}
	if req.GetTopic().GetDescription() == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}
	if req.GetTopic().GetCurriculum() == "" {
		return nil, status.Error(codes.InvalidArgument, "curriculum is required")
	}
	if req.GetTopic().GetMajorCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "major_code is required")
	}
	if req.GetTopic().GetSemesterCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "semester_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Convert Status enum to string
	StatusStr := "submit"
	switch req.GetTopic().GetStatus() {
	case pb.TopicStatus_SUBMIT:
		StatusStr = "submit"
	case pb.TopicStatus_TOPIC_PENDING:
		StatusStr = "topic_pending"
	case pb.TopicStatus_APPROVED_1:
		StatusStr = "approved_1"
	case pb.TopicStatus_APPROVED_2:
		StatusStr = "approved_2"
	case pb.TopicStatus_IN_PROGRESS:
		StatusStr = "in_progress"
	case pb.TopicStatus_TOPIC_COMPLETED:
		StatusStr = "topic_completed"
	case pb.TopicStatus_REJECTED:
		StatusStr = "rejected"
	}

	// Insert into database
	query := `
		INSERT INTO Topic (id, title, title_en, description, curriculum, major_code, semester_code, status, percent_stage_1, percent_stage_2, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := h.execQuery(ctx, query,
		id,
		req.GetTopic().GetTitle(),
		req.GetTopic().GetTitleEn(),
		req.GetTopic().GetDescription(),
		req.GetTopic().GetCurriculum(),
		req.GetTopic().GetMajorCode(),
		req.GetTopic().GetSemesterCode(),
		StatusStr,
		req.GetTopic().GetPercentStage_1(),
		req.GetTopic().GetPercentStage_2(),
		req.GetTopic().GetCreatedBy(),
		req.GetTopic().GetCreatedBy(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "topic already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create topic: %v", err)
	}

	return h.GetTopic(ctx, &pb.GetTopicRequest{Id: id})
}

// GetTopic retrieves a Topic by ID
func (h *Handler) GetTopic(ctx context.Context, req *pb.GetTopicRequest) (*pb.TopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"title":           true,
		"title_en":        true,
		"description":     true,
		"curriculum":      true,
		"major_code":      true,
		"semester_code":   true,
		"status":          true,
		"percent_stage_1": true,
		"percent_stage_2": true,
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
		SELECT id, title, title_en, description, curriculum, major_code, semester_code, status, percent_stage_1, percent_stage_2, created_at, updated_at, created_by, updated_by
		FROM Topic
		%s
	`, whereClause)

	var entity pb.Topic
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string

	err := h.queryRow(ctx, query, args...).Scan(
		&entity.Id,
		&entity.Title,
		&entity.TitleEn,
		&entity.Description,
		&entity.Curriculum,
		&entity.MajorCode,
		&entity.SemesterCode,
		&StatusStr,
		&entity.PercentStage_1,
		&entity.PercentStage_2,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "topic not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get topic: %v", err)
	}

	// Convert Status string to enum
	switch StatusStr {
	case "submit":
		entity.Status = pb.TopicStatus_SUBMIT
	case "topic_pending":
		entity.Status = pb.TopicStatus_TOPIC_PENDING
	case "approved_1":
		entity.Status = pb.TopicStatus_APPROVED_1
	case "approved_2":
		entity.Status = pb.TopicStatus_APPROVED_2
	case "in_progress":
		entity.Status = pb.TopicStatus_IN_PROGRESS
	case "topic_completed":
		entity.Status = pb.TopicStatus_TOPIC_COMPLETED
	case "rejected":
		entity.Status = pb.TopicStatus_REJECTED
	default:
		entity.Status = pb.TopicStatus_SUBMIT
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

	return &pb.TopicResponse{
		Topic: &entity,
	}, nil
}

// UpdateTopic updates an existing Topic
func (h *Handler) UpdateTopic(ctx context.Context, req *pb.UpdateTopicRequest) (*pb.TopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}

	if req.GetTopic().GetTitle() != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, req.GetTopic().GetTitle())
	}
	if req.GetTopic().GetTitleEn() != "" {
		updateFields = append(updateFields, "title_en = ?")
		args = append(args, req.GetTopic().GetTitleEn())
	}
	if req.GetTopic().GetDescription() != "" {
		updateFields = append(updateFields, "description = ?")
		args = append(args, req.GetTopic().GetDescription())
	}
	if req.GetTopic().GetCurriculum() != "" {
		updateFields = append(updateFields, "curriculum = ?")
		args = append(args, req.GetTopic().GetCurriculum())
	}
	if req.GetTopic().GetMajorCode() != "" {
		updateFields = append(updateFields, "major_code = ?")
		args = append(args, req.GetTopic().GetMajorCode())
	}
	if req.GetTopic().GetSemesterCode() != "" {
		updateFields = append(updateFields, "semester_code = ?")
		args = append(args, req.GetTopic().GetSemesterCode())
	}
	// Status enum - always include
	updateFields = append(updateFields, "status = ?")
	StatusStr := "submit"
	switch req.GetTopic().GetStatus() {
	case pb.TopicStatus_SUBMIT:
		StatusStr = "submit"
	case pb.TopicStatus_TOPIC_PENDING:
		StatusStr = "topic_pending"
	case pb.TopicStatus_APPROVED_1:
		StatusStr = "approved_1"
	case pb.TopicStatus_APPROVED_2:
		StatusStr = "approved_2"
	case pb.TopicStatus_IN_PROGRESS:
		StatusStr = "in_progress"
	case pb.TopicStatus_TOPIC_COMPLETED:
		StatusStr = "topic_completed"
	case pb.TopicStatus_REJECTED:
		StatusStr = "rejected"
	}
	args = append(args, StatusStr)

	if req.GetTopic().GetPercentStage_1() != 0 {
		updateFields = append(updateFields, "percent_stage_1 = ?")
		args = append(args, req.GetTopic().GetPercentStage_1())
	}
	if req.GetTopic().GetPercentStage_2() != 0 {
		updateFields = append(updateFields, "percent_stage_2 = ?")
		args = append(args, req.GetTopic().GetPercentStage_2())
	}

	if len(updateFields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// Add updated_by and updated_at
	updateFields = append(updateFields, "updated_by = ?")
	args = append(args, req.GetTopic().GetUpdatedBy())
	updateFields = append(updateFields, "updated_at = NOW()")

	// Build WHERE clause from filter
	whiteMap := map[string]bool{
		"title":           true,
		"title_en":        true,
		"description":     true,
		"curriculum":      true,
		"major_code":      true,
		"semester_code":   true,
		"status":          true,
		"percent_stage_1": true,
		"percent_stage_2": true,
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
		UPDATE Topic
		SET %s
		%s
	`, strings.Join(updateFields, ", "), whereClause)

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update topic: %v", err)
	}

	return h.GetTopic(ctx, &pb.GetTopicRequest{Id: req.Id})
}

// DeleteTopic deletes a Topic by ID
func (h *Handler) DeleteTopic(ctx context.Context, req *pb.DeleteTopicRequest) (*pb.DeleteTopicResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Build WHERE clause from filter
	args := []interface{}{}
	whiteMap := map[string]bool{
		"title":           true,
		"title_en":        true,
		"description":     true,
		"curriculum":      true,
		"major_code":      true,
		"semester_code":   true,
		"status":          true,
		"percent_stage_1": true,
		"percent_stage_2": true,
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

	query := fmt.Sprintf(`DELETE FROM Topic %s`, whereClause)

	result, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete topic: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "topic not found")
	}

	return &pb.DeleteTopicResponse{
		Success: true,
	}, nil
}

// ListTopics lists Topics with pagination and filtering
func (h *Handler) ListTopics(ctx context.Context, req *pb.ListTopicsRequest) (*pb.ListTopicsResponse, error) {
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
		"id":              true,
		"title":           true,
		"title_en":        true,
		"description":     true,
		"curriculum":      true,
		"major_code":      true,
		"semester_code":   true,
		"status":          true,
		"percent_stage_1": true,
		"percent_stage_2": true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Topic %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count topics: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, title_en, description, curriculum, major_code, semester_code, status, percent_stage_1, percent_stage_2, created_at, updated_at, created_by, updated_by
		FROM Topic
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list topics: %v", err)
	}
	defer rows.Close()

	entities := []*pb.Topic{}
	for rows.Next() {
		var entity pb.Topic
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var StatusStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.TitleEn,
			&entity.Description,
			&entity.Curriculum,
			&entity.MajorCode,
			&entity.SemesterCode,
			&StatusStr,
			&entity.PercentStage_1,
			&entity.PercentStage_2,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan topic: %v", err)
		}

		// Convert Status string to enum
		switch StatusStr {
		case "submit":
			entity.Status = pb.TopicStatus_SUBMIT
		case "topic_pending":
			entity.Status = pb.TopicStatus_TOPIC_PENDING
		case "approved_1":
			entity.Status = pb.TopicStatus_APPROVED_1
		case "approved_2":
			entity.Status = pb.TopicStatus_APPROVED_2
		case "in_progress":
			entity.Status = pb.TopicStatus_IN_PROGRESS
		case "topic_completed":
			entity.Status = pb.TopicStatus_TOPIC_COMPLETED
		case "rejected":
			entity.Status = pb.TopicStatus_REJECTED
		default:
			entity.Status = pb.TopicStatus_SUBMIT
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
		return nil, status.Errorf(codes.Internal, "error iterating topics: %v", err)
	}

	return &pb.ListTopicsResponse{
		Topics:   entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
