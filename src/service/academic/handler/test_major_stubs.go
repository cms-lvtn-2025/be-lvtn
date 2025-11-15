package handler

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

// CreateMajor creates a new Major record (Test stub)
func (h *TestHandler) CreateMajor(ctx context.Context, req *pb.CreateMajorRequest) (*pb.CreateMajorResponse, error) {
	// Validate required fields
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.FacultyCode == "" {
		return nil, status.Error(codes.InvalidArgument, "faculty_code is required")
	}

	// Generate UUID
	id := uuid.New().String()

	// Insert into database
	query := `INSERT INTO Major (id, title, faculty_code, created_by) VALUES (?, ?, ?, ?)`
	_, err := h.execQuery(ctx, query, id, req.Title, req.FacultyCode, req.CreatedBy)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create major: %v", err)
	}

	// Return created major
	major := &pb.Major{
		Id:          id,
		Title:       req.Title,
		FacultyCode: req.FacultyCode,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	}

	return &pb.CreateMajorResponse{Major: major}, nil
}

// GetMajor retrieves a Major by ID (Test stub)
func (h *TestHandler) GetMajor(ctx context.Context, req *pb.GetMajorRequest) (*pb.GetMajorResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `SELECT id, title, faculty_code, created_by FROM Major WHERE id = ?`

	var entity pb.Major

	err := h.queryRow(ctx, query, req.Id).Scan(
		&entity.Id,
		&entity.Title,
		&entity.FacultyCode,
		&entity.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "major not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get major: %v", err)
	}

	// Set timestamps for test consistency
	entity.CreatedAt = timestamppb.Now()
	entity.UpdatedAt = timestamppb.Now()

	return &pb.GetMajorResponse{Major: &entity}, nil
}

// ListMajors lists all majors (Test stub)
func (h *TestHandler) ListMajors(ctx context.Context, req *pb.ListMajorsRequest) (*pb.ListMajorsResponse, error) {
	query := `SELECT id, title, faculty_code, created_by FROM Major ORDER BY title`

	rows, err := h.query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list majors: %v", err)
	}
	defer rows.Close()

	var majors []*pb.Major
	for rows.Next() {
		var entity pb.Major

		err := rows.Scan(
			&entity.Id,
			&entity.Title,
			&entity.FacultyCode,
			&entity.CreatedBy,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan major: %v", err)
		}

		// Set timestamps for test consistency
		entity.CreatedAt = timestamppb.Now()
		entity.UpdatedAt = timestamppb.Now()

		majors = append(majors, &entity)
	}

	return &pb.ListMajorsResponse{
		Majors: majors,
		Total:  int32(len(majors)),
	}, nil
}

// UpdateMajor updates a Major (Test stub)
func (h *TestHandler) UpdateMajor(ctx context.Context, req *pb.UpdateMajorRequest) (*pb.UpdateMajorResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	var setParts []string
	var args []interface{}

	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}

	if req.FacultyCode != nil {
		setParts = append(setParts, "faculty_code = ?")
		args = append(args, *req.FacultyCode)
	}

	if req.UpdatedBy != "" {
		setParts = append(setParts, "updated_by = ?")
		args = append(args, req.UpdatedBy)
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, req.Id)

	query := `UPDATE Major SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`
	result, err := h.execQuery(ctx, query, args...)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update major: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "major not found")
	}

	// Get updated major
	getMajor, err := h.GetMajor(ctx, &pb.GetMajorRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateMajorResponse{Major: getMajor.Major}, nil
}

// DeleteMajor deletes a Major (Test stub)
func (h *TestHandler) DeleteMajor(ctx context.Context, req *pb.DeleteMajorRequest) (*pb.DeleteMajorResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM Major WHERE id = ?`
	result, err := h.execQuery(ctx, query, req.Id)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete major: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "major not found")
	}

	return &pb.DeleteMajorResponse{Success: true}, nil
}
