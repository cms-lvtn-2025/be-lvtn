package fixtures

import (
	common "thaily/proto/common"
	pb "thaily/proto/thesis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Enrollment test data
var (
	TestEnrollmentID1 = "test-enrollment-id-1"
	TestEnrollmentID2 = "test-enrollment-id-2"
)

// GetTestEnrollment returns a test Enrollment
func GetTestEnrollment() *pb.Enrollment {
	finalCode := "FINAL001"
	midtermCode := "MID001"
	return &pb.Enrollment{
		Id:               TestEnrollmentID1,
		Title:            "Test Enrollment",
		StudentCode:      "STU001",
		TopicCouncilCode: "TC001",
		FinalCode:        &finalCode,
		GradeReviewCode:  nil,
		MidtermCode:      &midtermCode,
		CreatedBy:        "test-user",
		CreatedAt:        timestamppb.New(TestTime),
		UpdatedAt:        timestamppb.New(TestTime),
	}
}

// GetTestEnrollment2 returns a second test Enrollment
func GetTestEnrollment2() *pb.Enrollment {
	gradeReviewCode := "GR001"
	return &pb.Enrollment{
		Id:               TestEnrollmentID2,
		Title:            "Test Enrollment 2",
		StudentCode:      "STU002",
		TopicCouncilCode: "TC002",
		FinalCode:        nil,
		GradeReviewCode:  &gradeReviewCode,
		MidtermCode:      nil,
		CreatedBy:        "test-user",
		CreatedAt:        timestamppb.New(TestTime),
		UpdatedAt:        timestamppb.New(TestTime),
	}
}

// GetTestCreateEnrollmentRequest returns a test create request
func GetTestCreateEnrollmentRequest() *pb.CreateEnrollmentRequest {
	finalCode := "FINAL001"
	midtermCode := "MID001"
	return &pb.CreateEnrollmentRequest{
		Title:            "Test Enrollment",
		StudentCode:      "STU001",
		TopicCouncilCode: "TC001",
		FinalCode:        &finalCode,
		GradeReviewCode:  nil,
		MidtermCode:      &midtermCode,
		CreatedBy:        "test-user",
	}
}

// GetTestUpdateEnrollmentRequest returns a test update request
func GetTestUpdateEnrollmentRequest() *pb.UpdateEnrollmentRequest {
	title := "Updated Enrollment"
	studentCode := "STU002"
	topicCouncilCode := "TC002"
	finalCode := "FINAL002"
	gradeReviewCode := "GR002"
	return &pb.UpdateEnrollmentRequest{
		Id:               TestEnrollmentID1,
		Title:            &title,
		StudentCode:      &studentCode,
		TopicCouncilCode: &topicCouncilCode,
		FinalCode:        &finalCode,
		GradeReviewCode:  &gradeReviewCode,
		MidtermCode:      nil,
		UpdatedBy:        "test-user",
	}
}

// GetTestListEnrollmentsRequest returns a test list request
func GetTestListEnrollmentsRequest() *pb.ListEnrollmentsRequest {
	return &pb.ListEnrollmentsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}
}
