package fixtures

import (
	"time"

	pb "thaily/proto/thesis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// TopicCouncilSupervisor test data
	TopicCouncilSupervisorID1 = "topiccouncilsupervisor-uuid-1"
	TopicCouncilSupervisorID2 = "topiccouncilsupervisor-uuid-2"

	TopicCouncilSupervisorTeacherCode1 = "TEACHER001"
	TopicCouncilSupervisorTeacherCode2 = "TEACHER002"
	TopicCouncilSupervisorCouncilCode1 = "COUNCIL001"
	TopicCouncilSupervisorCouncilCode2 = "COUNCIL002"

	// TopicCouncilSupervisor entities
	TopicCouncilSupervisor1 = &pb.TopicCouncilSupervisor{
		Id:                    TopicCouncilSupervisorID1,
		TeacherSupervisorCode: TopicCouncilSupervisorTeacherCode1,
		TopicCouncilCode:      TopicCouncilSupervisorCouncilCode1,
		CreatedAt:             timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:             timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		CreatedBy:             CreatedBy,
		UpdatedBy:             UpdatedBy,
	}

	TopicCouncilSupervisor2 = &pb.TopicCouncilSupervisor{
		Id:                    TopicCouncilSupervisorID2,
		TeacherSupervisorCode: TopicCouncilSupervisorTeacherCode2,
		TopicCouncilCode:      TopicCouncilSupervisorCouncilCode2,
		CreatedAt:             timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:             timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		CreatedBy:             CreatedBy,
		UpdatedBy:             UpdatedBy,
	}

	TopicCouncilSupervisors = []*pb.TopicCouncilSupervisor{TopicCouncilSupervisor1, TopicCouncilSupervisor2}
)
