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

// SemesterMock provides test implementations for Semester operations
type SemesterMock struct {
	*MockHandler
}

// NewSemesterMock creates a new Semester mock instance
func NewSemesterMock(db *sql.DB) *SemesterMock {
	return &SemesterMock{
		MockHandler: NewMockHandler(db),
	}
}

// CreateSemester creates a new Semester record (Test implementation)
func (m *SemesterMock) CreateSemester(ctx context.Context, req *pb.CreateSemesterRequest) (*pb.CreateSemesterResponse, error) {
	// Validate required fields
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Insert into database
	query := `INSERT INTO Semester (id, title, created_by) VALUES (?, ?, ?)`
	_, err := m.ExecQuery(ctx, query, id, req.Title, req.CreatedBy)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create semester: %v", err)
	}

	// Return created semester
	semester := &pb.Semester{
		Id:        id,
		Title:     req.Title,
		CreatedBy: req.CreatedBy,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}

	return &pb.CreateSemesterResponse{Semester: semester}, nil
}

// GetSemester retrieves a Semester by ID (Test implementation)
func (m *SemesterMock) GetSemester(ctx context.Context, req *pb.GetSemesterRequest) (*pb.GetSemesterResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `SELECT id, title, created_by FROM Semester WHERE id = ?`

	var entity pb.Semester

	err := m.QueryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "semester not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get semester: %v", err)
	}

	// Set timestamps for test consistency
	entity.CreatedAt = timestamppb.Now()
	entity.UpdatedAt = timestamppb.Now()

	return &pb.GetSemesterResponse{Semester: &entity}, nil
}

// ListSemesters lists all semesters (Test implementation)
func (m *SemesterMock) ListSemesters(ctx context.Context, req *pb.ListSemestersRequest) (*pb.ListSemestersResponse, error) {
	query := `SELECT id, title, created_by FROM Semester ORDER BY title`

	rows, err := m.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list semesters: %v", err)
	}
	defer rows.Close()

	var semesters []*pb.Semester
	for rows.Next() {
		var entity pb.Semester

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.CreatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan semester: %v", err)
		}

		// Set timestamps for test consistency
		entity.CreatedAt = timestamppb.Now()
		entity.UpdatedAt = timestamppb.Now()

		semesters = append(semesters, &entity)
	}

	return &pb.ListSemestersResponse{
		Semesters: semesters,
		Total:     int32(len(semesters)),
	}, nil
}

// UpdateSemester updates a Semester (Test implementation)
func (m *SemesterMock) UpdateSemester(ctx context.Context, req *pb.UpdateSemesterRequest) (*pb.UpdateSemesterResponse, error) {
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

	query := `UPDATE Semester SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`
	result, err := m.ExecQuery(ctx, query, args...)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update semester: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "semester not found")
	}

	// Get updated semester
	getSemester, err := m.GetSemester(ctx, &pb.GetSemesterRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateSemesterResponse{Semester: getSemester.Semester}, nil
}

// DeleteSemester deletes a Semester (Test implementation)
func (m *SemesterMock) DeleteSemester(ctx context.Context, req *pb.DeleteSemesterRequest) (*pb.DeleteSemesterResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Semester WHERE id = ?`
	result, err := m.ExecQuery(ctx, query, req.Id)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete semester: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "semester not found")
	}

	return &pb.DeleteSemesterResponse{Success: true}, nil
}
