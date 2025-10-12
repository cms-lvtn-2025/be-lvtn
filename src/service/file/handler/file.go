package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	pb "thaily/proto/file"
	"thaily/src/pkg/helper"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateFile creates a new File record
func (h *Handler) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate required fields (only string types)
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.File == "" {
		return nil, status.Error(codes.InvalidArgument, "file is required")
	}
	if req.Option == "" {
		return nil, status.Error(codes.InvalidArgument, "option is required")
	}
	if req.TableId == "" {
		return nil, status.Error(codes.InvalidArgument, "table_id is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Prepare fields

	// Convert Status enum to string
	StatusValue := pb.FileStatus_FILE_PENDING

	StatusValue = req.Status
	StatusStr := "pending"
	switch StatusValue {
	case pb.FileStatus_FILE_PENDING:
		StatusStr = "pending"
	case pb.FileStatus_APPROVED:
		StatusStr = "approved"
	case pb.FileStatus_REJECTED:
		StatusStr = "rejected"
	}
	fmt.Print(StatusValue)
	// Convert Table enum to string
	TableValue := pb.TableType_TOPIC

	TableValue = req.Table
	TableStr := "topic"
	switch TableValue {
	case pb.TableType_TOPIC:
		TableStr = "topic"
	case pb.TableType_MIDTERM:
		TableStr = "midterm"
	case pb.TableType_FINAL:
		TableStr = "final"
	case pb.TableType_ORDER:
		TableStr = "order"
	}

	// Insert into database
	query := "INSERT INTO `File` (id, title, file, status, `table`, `option`, table_id, created_by, created_at, updated_at) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())"

	_, err := h.execQuery(ctx, query,
		id,
		req.Title,
		req.File,
		StatusStr,
		TableStr,
		req.Option,
		req.TableId,
		req.CreatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "file already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create file: %v", err)
	}

	result, err := h.GetFile(ctx, &pb.GetFileRequest{Id: id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get file")
	}
	return &pb.CreateFileResponse{
		File: result.GetFile(),
	}, nil
}

// GetFile retrieves a File by ID
func (h *Handler) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	fmt.Print("chay vao day", req.Id)
	query := `
	SELECT id, title, file, status, ` + "`table`, `option`" + `, table_id, created_at, updated_at, created_by, updated_by
	FROM File
	WHERE id = ?
`

	var entity pb.File
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString
	var StatusStr string
	var TableStr string

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.File,
		&StatusStr,
		&TableStr,
		&entity.Option,
		&entity.TableId,
		&createdAt,
		&updatedAt,
		&entity.CreatedBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "file not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get file: %v", err)
	}

	// Convert Status string to enum
	switch StatusStr {
	case "pending", "file_pending": // Support both for backward compatibility
		entity.Status = pb.FileStatus_FILE_PENDING
	case "approved":
		entity.Status = pb.FileStatus_APPROVED
	case "rejected":
		entity.Status = pb.FileStatus_REJECTED
	default:
		entity.Status = pb.FileStatus_FILE_PENDING
	}
	// Convert Table string to enum
	switch TableStr {
	case "topic":
		entity.Table = pb.TableType_TOPIC
	case "midterm":
		entity.Table = pb.TableType_MIDTERM
	case "final":
		entity.Table = pb.TableType_FINAL
	case "order":
		entity.Table = pb.TableType_ORDER
	default:
		entity.Table = pb.TableType_TOPIC
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

	return &pb.GetFileResponse{
		File: &entity,
	}, nil
}

// UpdateFile updates an existing File
func (h *Handler) UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileResponse, error) {
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
	if req.File != nil {
		updateFields = append(updateFields, "file = ?")
		args = append(args, *req.File)

	}
	if req.Status != nil {
		updateFields = append(updateFields, "status = ?")
		StatusStr := "pending"
		switch *req.Status {
		case pb.FileStatus_FILE_PENDING:
			StatusStr = "pending"
		case pb.FileStatus_APPROVED:
			StatusStr = "approved"
		case pb.FileStatus_REJECTED:
			StatusStr = "rejected"
		}
		args = append(args, StatusStr)

	}
	if req.Table != nil {
		updateFields = append(updateFields, "table = ?")
		TableStr := "topic"
		switch *req.Table {
		case pb.TableType_TOPIC:
			TableStr = "topic"
		case pb.TableType_MIDTERM:
			TableStr = "midterm"
		case pb.TableType_FINAL:
			TableStr = "final"
		case pb.TableType_ORDER:
			TableStr = "order"
		}
		args = append(args, TableStr)

	}
	if req.Option != nil {
		updateFields = append(updateFields, "option = ?")
		args = append(args, *req.Option)

	}
	if req.TableId != nil {
		updateFields = append(updateFields, "table_id = ?")
		args = append(args, *req.TableId)

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
		UPDATE File
		SET %s
		WHERE id = ?
	`, strings.Join(updateFields, ", "))

	_, err := h.execQuery(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update file: %v", err)
	}

	result, err := h.GetFile(ctx, &pb.GetFileRequest{Id: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get file")
	}
	return &pb.UpdateFileResponse{
		File: result.GetFile(),
	}, nil
}

// DeleteFile deletes a File by ID
func (h *Handler) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	defer logger.TraceFunction(ctx)()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM File WHERE id = ?`

	result, err := h.execQuery(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete file: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	return &pb.DeleteFileResponse{
		Success: true,
	}, nil
}

// ListFiles lists Files with pagination and filtering
func (h *Handler) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
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
		"title":    true,
		"file":     true,
		"status":   true,
		"table":    true,
		"option":   true,
		"table_id": true,
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM File %s", whereClause)
	var total int32
	err := h.queryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count files: %v", err)
	}

	// Get entities with pagination
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, title, file, status, table, option, table_id, created_at, updated_at, created_by, updated_by
		FROM File
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortDirection)

	rows, err := h.query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list files: %v", err)
	}
	defer rows.Close()

	entities := []*pb.File{}
	for rows.Next() {
		var entity pb.File
		var createdAt, updatedAt sql.NullTime
		var updatedBy sql.NullString
		var StatusStr string
		var TableStr string

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.File,
			&StatusStr,
			&TableStr,
			&entity.Option,
			&entity.TableId,
			&createdAt,
			&updatedAt,
			&entity.CreatedBy,
			&updatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan file: %v", err)
		}

		// Convert Status string to enum
		switch StatusStr {
		case "pending", "file_pending": // Support both for backward compatibility
			entity.Status = pb.FileStatus_FILE_PENDING
		case "approved":
			entity.Status = pb.FileStatus_APPROVED
		case "rejected":
			entity.Status = pb.FileStatus_REJECTED
		default:
			entity.Status = pb.FileStatus_FILE_PENDING
		}
		// Convert Table string to enum
		switch TableStr {
		case "topic":
			entity.Table = pb.TableType_TOPIC
		case "midterm":
			entity.Table = pb.TableType_MIDTERM
		case "final":
			entity.Table = pb.TableType_FINAL
		case "order":
			entity.Table = pb.TableType_ORDER
		default:
			entity.Table = pb.TableType_TOPIC
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
		return nil, status.Errorf(codes.Internal, "error iterating files: %v", err)
	}

	return &pb.ListFilesResponse{
		Files:    entities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
