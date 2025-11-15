package mocks

import (
	"context"
	"database/sql"
	"strings"
	pb "thaily/proto/academic"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FacultyMock provides test implementations for Faculty operations
type FacultyMock struct {
	*MockHandler
}

// NewFacultyMock creates a new Faculty mock instance
func NewFacultyMock(db *sql.DB) *FacultyMock {
	return &FacultyMock{
		MockHandler: NewMockHandler(db),
	}
}

// CreateFaculty creates a new Faculty record (Test implementation)
func (m *FacultyMock) CreateFaculty(ctx context.Context, req *pb.CreateFacultyRequest) (*pb.CreateFacultyResponse, error) {
	// Validate required fields
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Insert into database (simplified for testing)
	query := `INSERT INTO Faculty (id, title, created_by) VALUES (?, ?, ?)`
	_, err := m.ExecQuery(ctx, query, id, req.Title, req.CreatedBy)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create faculty: %v", err)
	}

	// Return created faculty
	faculty := &pb.Faculty{
		Id:        id,
		Title:     req.Title,
		CreatedBy: req.CreatedBy,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}

	return &pb.CreateFacultyResponse{Faculty: faculty}, nil
}

// GetFaculty retrieves a Faculty by ID (Test implementation)
func (m *FacultyMock) GetFaculty(ctx context.Context, req *pb.GetFacultyRequest) (*pb.GetFacultyResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `SELECT id, title, created_by FROM Faculty WHERE id = ?`

	var entity pb.Faculty

	err := m.QueryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "faculty not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get faculty: %v", err)
	}

	// Set timestamps for test consistency
	entity.CreatedAt = timestamppb.Now()
	entity.UpdatedAt = timestamppb.Now()

	return &pb.GetFacultyResponse{Faculty: &entity}, nil
}

// ListFaculties lists all faculties (Test implementation)
func (m *FacultyMock) ListFaculties(ctx context.Context, req *pb.ListFacultiesRequest) (*pb.ListFacultiesResponse, error) {
	query := `SELECT id, title, created_by FROM Faculty ORDER BY title`

	rows, err := m.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list faculties: %v", err)
	}
	defer rows.Close()

	var faculties []*pb.Faculty
	for rows.Next() {
		var entity pb.Faculty

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.CreatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan faculty: %v", err)
		}

		// Set timestamps for test consistency
		entity.CreatedAt = timestamppb.Now()
		entity.UpdatedAt = timestamppb.Now()

		faculties = append(faculties, &entity)
	}

	return &pb.ListFacultiesResponse{
		Faculties: faculties,
		Total:     int32(len(faculties)),
	}, nil
}

// UpdateFaculty updates a Faculty (Test implementation)
func (m *FacultyMock) UpdateFaculty(ctx context.Context, req *pb.UpdateFacultyRequest) (*pb.UpdateFacultyResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	var setParts []string
	var args []interface{}

	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}

	if req.UpdatedBy != "" {
		setParts = append(setParts, "updated_by = ?")
		args = append(args, req.UpdatedBy)
	}

	if len(setParts) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, req.Id)

	query := `UPDATE Faculty SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`
	result, err := m.ExecQuery(ctx, query, args...)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update faculty: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "faculty not found")
	}

	// Get updated faculty
	getFaculty, err := m.GetFaculty(ctx, &pb.GetFacultyRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateFacultyResponse{Faculty: getFaculty.Faculty}, nil
}

// DeleteFaculty deletes a Faculty (Test implementation)
func (m *FacultyMock) DeleteFaculty(ctx context.Context, req *pb.DeleteFacultyRequest) (*pb.DeleteFacultyResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Faculty WHERE id = ?`
	result, err := m.ExecQuery(ctx, query, req.Id)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete faculty: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "faculty not found")
	}

	return &pb.DeleteFacultyResponse{Success: true}, nil
}
