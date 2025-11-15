package fixtures

import (
	"time"

	common "thaily/proto/common"
	pb "thaily/proto/council"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Council test data
var (
	TestCouncilID1 = "council-test-id-1"
	TestCouncilID2 = "council-test-id-2"

	TestTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

// GetTestCouncil returns a test council
func GetTestCouncil() *pb.Council {
	return &pb.Council{
		Id:           TestCouncilID1,
		Title:        "Test Council",
		MajorCode:    "CS",
		SemesterCode: "2024-1",
		CreatedBy:    "test-user",
		CreatedAt:    timestamppb.New(TestTime),
		UpdatedAt:    timestamppb.New(TestTime),
	}
}

// GetTestCouncil2 returns a second test Council
func GetTestCouncil2() *pb.Council {
	return &pb.Council{
		Id:           TestCouncilID2,
		Title:        "Test Council 2",
		MajorCode:    "EE",
		SemesterCode: "2024-2",
		CreatedBy:    "test-user",
		CreatedAt:    timestamppb.New(TestTime),
		UpdatedAt:    timestamppb.New(TestTime),
	}
}

// GetTestCreateCouncilRequest returns a test create request
func GetTestCreateCouncilRequest() *pb.CreateCouncilRequest {
	return &pb.CreateCouncilRequest{
		Title:        "Test Council",
		MajorCode:    "CS",
		SemesterCode: "2024-1",
		CreatedBy:    "test-user",
	}
}

// GetTestUpdateCouncilRequest returns a test update request
func GetTestUpdateCouncilRequest() *pb.UpdateCouncilRequest {
	title := "Updated Council"
	majorCode := "CS"
	semesterCode := "2024-2"
	return &pb.UpdateCouncilRequest{
		Id:           TestCouncilID1,
		Title:        &title,
		MajorCode:    &majorCode,
		SemesterCode: &semesterCode,
		UpdatedBy:    "test-user",
	}
}

// GetTestListCouncilsRequest returns a test list request
func GetTestListCouncilsRequest() *pb.ListCouncilsRequest {
	return &pb.ListCouncilsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}
}
