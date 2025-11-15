package fixtures

import (
	"time"

	pb "thaily/proto/user"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// Teacher test data
	TeacherID1 = "teacher-uuid-1"
	TeacherID2 = "teacher-uuid-2"

	TeacherEmail1        = "teacher1@example.com"
	TeacherEmail2        = "teacher2@example.com"
	TeacherUsername1     = "teacher_one"
	TeacherUsername2     = "teacher_two"
	TeacherMajorCode1    = "CS"
	TeacherMajorCode2    = "SE"
	TeacherSemesterCode1 = "2024-1"
	TeacherSemesterCode2 = "2024-2"

	// Teacher entities
	Teacher1 = &pb.Teacher{
		Id:           TeacherID1,
		Email:        TeacherEmail1,
		Username:     TeacherUsername1,
		Gender:       pb.Gender_MALE,
		MajorCode:    TeacherMajorCode1,
		SemesterCode: TeacherSemesterCode1,
		CreatedAt:    timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:    timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		CreatedBy:    CreatedBy,
		UpdatedBy:    UpdatedBy,
	}

	Teacher2 = &pb.Teacher{
		Id:           TeacherID2,
		Email:        TeacherEmail2,
		Username:     TeacherUsername2,
		Gender:       pb.Gender_FEMALE,
		MajorCode:    TeacherMajorCode2,
		SemesterCode: TeacherSemesterCode2,
		CreatedAt:    timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:    timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		CreatedBy:    CreatedBy,
		UpdatedBy:    UpdatedBy,
	}

	Teachers = []*pb.Teacher{Teacher1, Teacher2}
)
