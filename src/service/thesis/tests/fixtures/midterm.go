package fixtures

import (
	"time"

	pb "thaily/proto/thesis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Common test values
	CreatedBy = "test-user"
	UpdatedBy = "test-admin"
)

var (
	// Midterm test data
	MidtermID1 = "midterm-uuid-1"
	MidtermID2 = "midterm-uuid-2"

	MidtermTitle1 = "Midterm Report - AI System"
	MidtermTitle2 = "Midterm Report - Web Platform"

	MidtermGrade1    = int32(85)
	MidtermGrade2    = int32(90)
	MidtermFeedback1 = "Good progress, needs improvement on methodology"
	MidtermFeedback2 = "Excellent work, on track"

	// Midterm entities
	Midterm1 = &pb.Midterm{
		Id:        MidtermID1,
		Title:     MidtermTitle1,
		Grade:     MidtermGrade1,
		Status:    pb.MidtermStatus_SUBMITTED,
		Feedback:  MidtermFeedback1,
		CreatedAt: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		CreatedBy: CreatedBy,
		UpdatedBy: UpdatedBy,
	}

	Midterm2 = &pb.Midterm{
		Id:        MidtermID2,
		Title:     MidtermTitle2,
		Grade:     MidtermGrade2,
		Status:    pb.MidtermStatus_PASS,
		Feedback:  MidtermFeedback2,
		CreatedAt: timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		UpdatedAt: timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		CreatedBy: CreatedBy,
		UpdatedBy: UpdatedBy,
	}

	Midterms = []*pb.Midterm{Midterm1, Midterm2}
)
