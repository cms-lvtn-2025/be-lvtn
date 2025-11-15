package fixtures

import (
	"time"

	common "thaily/proto/common"
	pb "thaily/proto/thesis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Topic test data
var (
	TestTopicID1 = "test-topic-id-1"
	TestTopicID2 = "test-topic-id-2"
	TestTime     = time.Now()
)

// GetTestTopic returns a test Topic
func GetTestTopic() *pb.Topic {
	percent1 := int32(50)
	percent2 := int32(50)
	return &pb.Topic{
		Id:             TestTopicID1,
		Title:          "Test Topic",
		MajorCode:      "CS",
		SemesterCode:   "2024-1",
		Status:         pb.TopicStatus_SUBMIT,
		PercentStage_1: &percent1,
		PercentStage_2: &percent2,
		CreatedBy:      "test-user",
		CreatedAt:      timestamppb.New(TestTime),
		UpdatedAt:      timestamppb.New(TestTime),
	}
}

// GetTestTopic2 returns a second test Topic
func GetTestTopic2() *pb.Topic {
	percent1 := int32(40)
	percent2 := int32(60)
	return &pb.Topic{
		Id:             TestTopicID2,
		Title:          "Test Topic 2",
		MajorCode:      "EE",
		SemesterCode:   "2024-2",
		Status:         pb.TopicStatus_IN_PROGRESS,
		PercentStage_1: &percent1,
		PercentStage_2: &percent2,
		CreatedBy:      "test-user",
		CreatedAt:      timestamppb.New(TestTime),
		UpdatedAt:      timestamppb.New(TestTime),
	}
}

// GetTestCreateTopicRequest returns a test create request
func GetTestCreateTopicRequest() *pb.CreateTopicRequest {
	percent1 := int32(50)
	percent2 := int32(50)
	return &pb.CreateTopicRequest{
		Title:          "Test Topic",
		MajorCode:      "CS",
		SemesterCode:   "2024-1",
		Status:         pb.TopicStatus_SUBMIT,
		PercentStage_1: &percent1,
		PercentStage_2: &percent2,
		CreatedBy:      "test-user",
	}
}

// GetTestUpdateTopicRequest returns a test update request
func GetTestUpdateTopicRequest() *pb.UpdateTopicRequest {
	title := "Updated Topic"
	majorCode := "CS"
	semesterCode := "2024-2"
	status := pb.TopicStatus_APPROVED_1
	percent1 := int32(60)
	percent2 := int32(40)
	return &pb.UpdateTopicRequest{
		Id:             TestTopicID1,
		Title:          &title,
		MajorCode:      &majorCode,
		SemesterCode:   &semesterCode,
		Status:         &status,
		PercentStage_1: &percent1,
		PercentStage_2: &percent2,
		UpdatedBy:      "test-user",
	}
}

// GetTestListTopicsRequest returns a test list request
func GetTestListTopicsRequest() *pb.ListTopicsRequest {
	return &pb.ListTopicsRequest{
		Search: &common.SearchRequest{
			Pagination: &common.Pagination{
				Page:     1,
				PageSize: 10,
			},
		},
	}
}
