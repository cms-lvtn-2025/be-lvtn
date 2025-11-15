package fixtures

import (
	"time"

	pb "thaily/proto/thesis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// TopicCouncil test data
	TopicCouncilID1 = "topiccouncil-uuid-1"
	TopicCouncilID2 = "topiccouncil-uuid-2"

	TopicCouncilTitle1       = "Topic Council - DACN Stage"
	TopicCouncilTitle2       = "Topic Council - LVTN Stage"
	TopicCouncilTopicCode1   = "TOPIC001"
	TopicCouncilTopicCode2   = "TOPIC002"
	TopicCouncilCouncilCode1 = "COUNCIL001"
	TopicCouncilCouncilCode2 = "COUNCIL002"

	// TopicCouncil entities
	TopicCouncil1 = &pb.TopicCouncil{
		Id:          TopicCouncilID1,
		Title:       TopicCouncilTitle1,
		Stage:       pb.TopicStage_STAGE_DACN,
		TopicCode:   TopicCouncilTopicCode1,
		CouncilCode: &TopicCouncilCouncilCode1, // optional field - pointer
		TimeStart:   timestamppb.New(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)),
		TimeEnd:     timestamppb.New(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)),
		CreatedAt:   timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:   timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		CreatedBy:   CreatedBy,
		UpdatedBy:   UpdatedBy,
	}

	TopicCouncil2 = &pb.TopicCouncil{
		Id:          TopicCouncilID2,
		Title:       TopicCouncilTitle2,
		Stage:       pb.TopicStage_STAGE_LVTN,
		TopicCode:   TopicCouncilTopicCode2,
		CouncilCode: &TopicCouncilCouncilCode2, // optional field - pointer
		TimeStart:   timestamppb.New(time.Date(2024, 6, 2, 9, 0, 0, 0, time.UTC)),
		TimeEnd:     timestamppb.New(time.Date(2024, 6, 2, 12, 0, 0, 0, time.UTC)),
		CreatedAt:   timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:   timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		CreatedBy:   CreatedBy,
		UpdatedBy:   UpdatedBy,
	}

	TopicCouncils = []*pb.TopicCouncil{TopicCouncil1, TopicCouncil2}
)
