package resolver

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	pb "thaily/proto/common"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type CommonService struct {
	pb.UnimplementedCommonServiceServer
	db *sql.DB
}

func NewCommonService(db *sql.DB) *CommonService {
	return &CommonService{
		db: db,
	}
}

// Procedure thực thi stored procedure với câu query trực tiếp - Tối ưu performance
func (s *CommonService) Procedure(ctx context.Context, req *pb.ProcedureRequest) (*pb.ProcedureResponse, error) {
	// Kiểm tra procedureQuery có được cung cấp không
	if req.ProcedureQuery == "" {
		return nil, status.Error(codes.InvalidArgument, "procedureQuery is required")
	}

	// Thực thi stored procedure trực tiếp với query đã chuẩn bị sẵn
	rows, err := s.db.QueryContext(ctx, req.ProcedureQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute procedure: %v", err)
	}
	defer rows.Close()

	// Kiểm tra xem có result set nào cần xử lý không
	columns, err := rows.Columns()
	if err != nil || len(columns) == 0 {
		// Không có result set (procedure UPDATE/INSERT) - trả về response ngay
		return &pb.ProcedureResponse{
			Success: true,
			Message: "Procedure executed successfully",
			Data:    nil,
		}, nil
	}

	// Pre-allocate memory cho performance tốt hơn
	results := make([]map[string]interface{}, 0, 100) // Giả định tối đa 100 rows
	colCount := len(columns)
	
	// Pre-allocate slice cho việc scan - tái sử dụng memory
	values := make([]interface{}, colCount)
	valuePtrs := make([]interface{}, colCount)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Xử lý từng row với minimal allocation
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			// Skip invalid rows silently for better performance
			continue
		}

		// Tạo row map với capacity đã biết
		row := make(map[string]interface{}, colCount)
		for i, col := range columns {
			// Xử lý byte arrays hiệu quả hơn
			if b, ok := values[i].([]byte); ok {
				// Chỉ convert khi cần thiết và sử dụng unsafe convert nếu data lớn
				if len(b) < 1024 {
					row[col] = string(b)
				} else {
					row[col] = string(b) // Có thể optimize thêm bằng unsafe package
				}
			} else {
				row[col] = values[i]
			}
		}
		results = append(results, row)
	}

	// Kiểm tra lỗi iteration
	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error iterating results: %v", err)
	}

	// Tạo response struct một lần duy nhất
	if len(results) == 0 {
		return &pb.ProcedureResponse{
			Success: true,
			Message: "Procedure executed successfully",
			Data:    nil,
		}, nil
	}

	resultStruct, err := structpb.NewStruct(map[string]interface{}{
		"results":      convertResultsToListValue(results),
		"total_rows":   len(results),
		"column_count": colCount,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create result struct: %v", err)
	}

	return &pb.ProcedureResponse{
		Success: true,
		Message: "Procedure executed successfully",
		Data:    resultStruct,
	}, nil
}

func (s *CommonService) Create(ctx context.Context, req *pb.GenericRequest) (*pb.GenericResponse, error) {
	if req.TableName == "" {
		return nil, status.Error(codes.InvalidArgument, "table name is required")
	}

	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}

	data := req.Data.AsMap()

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for key, value := range data {
		columns = append(columns, key)
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		req.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create record: %v", err)
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get last insert id: %v", err)
	}

	data["id"] = lastInsertId
	responseData, err := structpb.NewStruct(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create response struct: %v", err)
	}

	return &pb.GenericResponse{
		Success: true,
		Message: "Record created successfully",
		Data:    responseData,
	}, nil
}

func (s *CommonService) GetById(ctx context.Context, req *pb.GetByIdRequest) (*pb.GenericResponse, error) {
	if req.TableName == "" {
		return nil, status.Error(codes.InvalidArgument, "table name is required")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", req.TableName)
	row := s.db.QueryRowContext(ctx, query, req.Id)

	columns, err := s.getTableColumns(ctx, req.TableName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get table columns: %v", err)
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := row.Scan(valuePtrs...); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to scan record: %v", err)
	}

	data := make(map[string]interface{})
	for i, col := range columns {
		// Handle byte arrays and convert to string
		if b, ok := values[i].([]byte); ok {
			data[col] = string(b)
		} else {
			data[col] = values[i]
		}
	}

	responseData, err := structpb.NewStruct(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create response struct: %v", err)
	}

	return &pb.GenericResponse{
		Success: true,
		Message: "Record retrieved successfully",
		Data:    responseData,
	}, nil
}

func (s *CommonService) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	startTime := time.Now()

	// Execute data query and count query in parallel
	var wg sync.WaitGroup
	var dataErr, countErr error
	var results []map[string]interface{}
	var totalCount int64

	wg.Add(2)

	// Data query goroutine
	go func() {
		defer wg.Done()
		rows, err := s.db.QueryContext(ctx, req.Query)
		if err != nil {
			dataErr = err
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			dataErr = err
			return
		}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				// Handle byte arrays and convert to string
				if b, ok := values[i].([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = values[i]
				}
			}
			results = append(results, row)
		}
	}()

	// Count query goroutine
	go func() {
		defer wg.Done()
		// Use custom count query if provided, otherwise skip count
		if req.QueryCount != "" {
			var count sql.NullInt64
			err := s.db.QueryRowContext(ctx, req.QueryCount).Scan(&count)
			if err != nil {
				log.Printf("Error executing count query: %v", err)
				countErr = err
				return
			}
			if count.Valid {
				totalCount = count.Int64
			}
		}
	}()

	wg.Wait()

	if dataErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute query: %v", dataErr)
	}

	executionTime := time.Since(startTime).Milliseconds()

	// If count query failed, use results length as fallback
	if countErr != nil || totalCount == 0 {
		totalCount = int64(len(results))
	}

	// Convert results array to structpb-compatible format
	responseData, err := structpb.NewStruct(map[string]interface{}{
		"results": convertResultsToListValue(results),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create response struct: %v", err)
	}

	return &pb.QueryResponse{
		Success:         true,
		Message:         "Query executed successfully",
		Data:            responseData,
		TotalItems:      totalCount,
		ExecutionTimeMs: executionTime,
	}, nil
}

func (s *CommonService) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.GenericResponse, error) {
	if req.TableName == "" {
		return nil, status.Error(codes.InvalidArgument, "table name is required")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}

	data := req.Data.AsMap()

	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+1)

	for key, value := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}
	values = append(values, req.Id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		req.TableName,
		strings.Join(setClauses, ", "))

	result, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update record: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	data["id"] = req.Id
	responseData, err := structpb.NewStruct(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create response struct: %v", err)
	}

	return &pb.GenericResponse{
		Success: true,
		Message: "Record updated successfully",
		Data:    responseData,
	}, nil
}

func (s *CommonService) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if req.TableName == "" {
		return nil, status.Error(codes.InvalidArgument, "table name is required")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", req.TableName)
	result, err := s.db.ExecContext(ctx, query, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete record: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	return &pb.DeleteResponse{
		Success:      true,
		Message:      "Record deleted successfully",
		DeletedCount: int32(rowsAffected),
	}, nil
}

func (s *CommonService) DeleteMany(ctx context.Context, req *pb.DeleteManyRequest) (*pb.DeleteManyResponse, error) {
	if req.Table == "" {
		return nil, status.Error(codes.InvalidArgument, "table name is required")
	}

	if len(req.Ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one id is required")
	}

	placeholders := make([]string, len(req.Ids))
	values := make([]interface{}, len(req.Ids))
	for i, id := range req.Ids {
		placeholders[i] = "?"
		values[i] = id
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)",
		req.Table,
		strings.Join(placeholders, ", "))

	result, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete records: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	var failedIds []string
	if int(rowsAffected) < len(req.Ids) {
		for _, id := range req.Ids {
			var exists bool
			checkQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ?)", req.Table)
			err := s.db.QueryRowContext(ctx, checkQuery, id).Scan(&exists)
			if err == nil && !exists {
				failedIds = append(failedIds, id)
			}
		}
	}

	return &pb.DeleteManyResponse{
		Success:      true,
		Message:      fmt.Sprintf("Deleted %d records", rowsAffected),
		DeletedCount: int32(rowsAffected),
		FailedIds:    failedIds,
	}, nil
}

func (s *CommonService) getTableColumns(ctx context.Context, tableName string) ([]string, error) {
	query := `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS 
			  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
			  ORDER BY ORDINAL_POSITION`

	rows, err := s.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

func convertResultsToListValue(results []map[string]interface{}) []interface{} {
	listValue := make([]interface{}, len(results))
	for i, result := range results {
		listValue[i] = result
	}
	return listValue
}
