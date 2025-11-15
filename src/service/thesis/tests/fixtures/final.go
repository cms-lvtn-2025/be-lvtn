package fixtures

import (
	"time"

	pb "thaily/proto/thesis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Common test values (if not already defined elsewhere in fixtures)
	createdBy = "test-user"
	updatedBy = "test-admin"
)

var (
	// Final test data
	FinalID1 = "final-uuid-1"
	FinalID2 = "final-uuid-2"

	FinalTitle1 = "Final Defense - AI System Implementation"
	FinalTitle2 = "Final Defense - Web Platform Development"

	FinalSupervisorGrade1 = int32(88)
	FinalSupervisorGrade2 = int32(92)
	FinalDepartmentGrade1 = int32(85)
	FinalDepartmentGrade2 = int32(90)
	FinalGrade1           = int32(87)
	FinalGrade2           = int32(91)
	FinalNotes1           = "Strong technical implementation, good presentation"
	FinalNotes2           = "Excellent project, innovative approach"

	// Final entities
	Final1 = &pb.Final{
		Id:              FinalID1,
		Title:           FinalTitle1,
		SupervisorGrade: FinalSupervisorGrade1,
		DepartmentGrade: FinalDepartmentGrade1,
		FinalGrade:      FinalGrade1,
		Status:          pb.FinalStatus_PASSED,
		Notes:           FinalNotes1,
		CreatedAt:       timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:       timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
	}

	Final2 = &pb.Final{
		Id:              FinalID2,
		Title:           FinalTitle2,
		SupervisorGrade: FinalSupervisorGrade2,
		DepartmentGrade: FinalDepartmentGrade2,
		FinalGrade:      FinalGrade2,
		Status:          pb.FinalStatus_COMPLETED,
		Notes:           FinalNotes2,
		CreatedAt:       timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:       timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
	}

	Finals = []*pb.Final{Final1, Final2}
)
