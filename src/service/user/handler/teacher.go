package handler

import (
	"context"
	"database/sql"
	"strings"
	pb "thaily/proto/user"
	"thaily/src/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) CreateTeacher(ctx context.Context, req *pb.CreateTeacherRequest) (*pb.CreateTeacherResponse, error) {
	defer logger.TraceFunction(ctx)()

	// Validate require field
	if req.Email == "" || req.MajorCode == "" || req.Username == "" || req.MajorCode == "" || req.SemesterCode == "" {
		return nil, status.Error(codes.InvalidArgument, "email, username, major_code, class_code, and semester_code are required")
	}

	// Generate UUID for teacher
	teacherID := uuid.New().String()

	// Set default gender if not provided
	gender := req.Gender
	// Convert gender enum to string
	genderStr := "male"
	switch gender {
	case pb.Gender_MALE:
		genderStr = "male"
	case pb.Gender_FEMALE:
		genderStr = "female"
	case pb.Gender_OTHER:
		genderStr = "other"
	}

	query := `
		INSERT INTO Teacher (id, email, username, gender, major_code, semester_code, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	if _, err := h.execQuery(ctx, query,
		teacherID,
		req.Email,
		req.Username,
		genderStr,
		req.MajorCode,
		req.SemesterCode,
		req.CreatedBy,
	); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, status.Error(codes.AlreadyExists, "teacher with this email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create teacher: %v", err)
	}

	// Retrieve the created teacher
	getQuery := `
		SELECT id, email, username, gender, major_code, semester_code, created_at, updated_at, created_by, updated_by
		FROM Teacher
		WHERE id = ?
	`

	var teacher pb.Teacher
	var createdAt, updatedAt sql.NullTime
	var updatedBy sql.NullString

	if err := h.queryRow(ctx, getQuery, teacherID).Scan(
		&teacher.Id,
		&teacher.Email,
		&teacher.Username,
		&genderStr,
		&teacher.MajorCode,
		&teacher.SemesterCode,
		&createdAt,
		&updatedAt,
		&teacher.CreatedBy,
		&updatedBy,
	); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created teacher: %v", err)
	}

	// Convert gender string to enum
	switch genderStr {
	case "male":
		teacher.Gender = pb.Gender_MALE
	case "female":
		teacher.Gender = pb.Gender_FEMALE
	case "other":
		teacher.Gender = pb.Gender_OTHER
	default:
		teacher.Gender = pb.Gender_MALE
	}
	if createdAt.Valid {
		teacher.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		teacher.UpdatedAt = timestamppb.New(updatedAt.Time)
	}
	if updatedBy.Valid {
		teacher.UpdatedBy = updatedBy.String
	}

	return &pb.CreateTeacherResponse{
		Teacher: &teacher,
	}, nil
}

func (h *Handler) GetTeacher(ctx context.Context, req *pb.GetTeacherRequest) (*pb.GetTeacherResponse, error) {
	// TODO: Implement GetTeacher
	return &pb.GetTeacherResponse{}, nil
}

func (h *Handler) UpdateTeacher(ctx context.Context, req *pb.UpdateTeacherRequest) (*pb.UpdateTeacherResponse, error) {
	// TODO: Implement UpdateTeacher
	return &pb.UpdateTeacherResponse{}, nil
}

func (h *Handler) DeleteTeacher(ctx context.Context, req *pb.DeleteTeacherRequest) (*pb.DeleteTeacherResponse, error) {
	// TODO: Implement DeleteTeacher
	return &pb.DeleteTeacherResponse{}, nil
}

func (h *Handler) ListTeachers(ctx context.Context, req *pb.ListTeachersRequest) (*pb.ListTeachersResponse, error) {
	// TODO: Implement ListTeachers
	return &pb.ListTeachersResponse{}, nil
}
