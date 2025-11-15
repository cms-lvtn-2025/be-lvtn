package fixtures

import (
	"time"

	pb "thaily/proto/user"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Common test values
	CreatedBy = "test-user"
	UpdatedBy = "test-admin"
)

var (
	// Student test data
	StudentID1 = "student-uuid-1"
	StudentID2 = "student-uuid-2"

	StudentEmail1        = "student1@example.com"
	StudentEmail2        = "student2@example.com"
	StudentPhone1        = "0123456789"
	StudentPhone2        = "0987654321"
	StudentUsername1     = "student_one"
	StudentUsername2     = "student_two"
	StudentMajorCode1    = "CS"
	StudentMajorCode2    = "SE"
	StudentClassCode1    = "CS2021"
	StudentClassCode2    = "SE2021"
	StudentSemesterCode1 = "2024-1"
	StudentSemesterCode2 = "2024-2"

	// Student entities
	Student1 = &pb.Student{
		Id:           StudentID1,
		Email:        StudentEmail1,
		Phone:        StudentPhone1,
		Username:     StudentUsername1,
		Gender:       pb.Gender_MALE,
		MajorCode:    StudentMajorCode1,
		ClassCode:    StudentClassCode1,
		SemesterCode: StudentSemesterCode1,
		CreatedAt:    timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:    timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		CreatedBy:    CreatedBy,
		UpdatedBy:    UpdatedBy,
	}

	Student2 = &pb.Student{
		Id:           StudentID2,
		Email:        StudentEmail2,
		Phone:        StudentPhone2,
		Username:     StudentUsername2,
		Gender:       pb.Gender_FEMALE,
		MajorCode:    StudentMajorCode2,
		ClassCode:    StudentClassCode2,
		SemesterCode: StudentSemesterCode2,
		CreatedAt:    timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:    timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		CreatedBy:    CreatedBy,
		UpdatedBy:    UpdatedBy,
	}

	Students = []*pb.Student{Student1, Student2}
)
