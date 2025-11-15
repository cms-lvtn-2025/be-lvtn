package fixtures

import (
	"time"

	common "thaily/proto/common"
	pb "thaily/proto/role"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// RoleSystem test data
var (
	TestRoleSystemID1 = "test-role-id-1"
	TestRoleSystemID2 = "test-role-id-2"
	TestTime          = time.Now()
)

// GetTestRoleSystem returns a test RoleSystem
func GetTestRoleSystem() *pb.RoleSystem {
	return &pb.RoleSystem{
		Id:           TestRoleSystemID1,
		Title:        "Test Role",
		TeacherCode:  "T001",
		Role:         pb.RoleType_TEACHER,
		SemesterCode: "2024-1",
		Activate:     true,
		CreatedBy:    "test-user",
		CreatedAt:    timestamppb.New(TestTime),
		UpdatedAt:    timestamppb.New(TestTime),
	}
}

// GetTestRoleSystem2 returns a second test RoleSystem
func GetTestRoleSystem2() *pb.RoleSystem {
	return &pb.RoleSystem{
		Id:           TestRoleSystemID2,
		Title:        "Test Role 2",
		TeacherCode:  "T002",
		Role:         pb.RoleType_DEPARTMENT_LECTURER,
		SemesterCode: "2024-2",
		Activate:     false,
		CreatedBy:    "test-user",
		CreatedAt:    timestamppb.New(TestTime),
		UpdatedAt:    timestamppb.New(TestTime),
	}
}

// GetTestCreateRoleSystemRequest returns a test create request
func GetTestCreateRoleSystemRequest() *pb.CreateRoleSystemRequest {
	return &pb.CreateRoleSystemRequest{
		Title:        "Test Role",
		TeacherCode:  "T001",
		Role:         pb.RoleType_TEACHER,
		SemesterCode: "2024-1",
		Activate:     true,
		CreatedBy:    "test-user",
	}
}

// GetTestUpdateRoleSystemRequest returns a test update request
func GetTestUpdateRoleSystemRequest() *pb.UpdateRoleSystemRequest {
	title := "Updated Role"
	teacherCode := "T002"
	role := pb.RoleType_DEPARTMENT_LECTURER
	semesterCode := "2024-2"
	activate := false
	return &pb.UpdateRoleSystemRequest{
		Id:           TestRoleSystemID1,
		Title:        &title,
		TeacherCode:  &teacherCode,
		Role:         &role,
		SemesterCode: &semesterCode,
		Activate:     &activate,
		UpdatedBy:    "test-user",
	}
}

// GetTestListRoleSystemsRequest returns a test list request
func GetTestListRoleSystemsRequest() *pb.ListRoleSystemsRequest {
	return &pb.ListRoleSystemsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}
}
