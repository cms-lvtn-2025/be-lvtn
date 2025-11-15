package fixtures

import (
	"time"

	pb "thaily/proto/thesis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// GradeReview test data
	GradeReviewID1 = "gradereview-uuid-1"
	GradeReviewID2 = "gradereview-uuid-2"

	GradeReviewTitle1       = "Grade Review - AI System Project"
	GradeReviewTitle2       = "Grade Review - Web Platform"
	GradeReviewTeacherCode1 = "TEACHER001"
	GradeReviewTeacherCode2 = "TEACHER002"
	GradeReviewGrade1       = int32(85)
	GradeReviewGrade2       = int32(90)
	GradeReviewNotes1       = "Good work, minor improvements needed"
	GradeReviewNotes2       = "Excellent presentation and documentation"

	// GradeReview entities
	GradeReview1 = &pb.GradeReview{
		Id:          GradeReviewID1,
		Title:       GradeReviewTitle1,
		ReviewGrade: &GradeReviewGrade1, // optional field - pointer
		TeacherCode: GradeReviewTeacherCode1,
		Status:      pb.FinalStatus_PENDING,
		Notes:       &GradeReviewNotes1, // optional field - pointer
		CreatedAt:   timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:   timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		CreatedBy:   CreatedBy,
		UpdatedBy:   UpdatedBy,
	}

	GradeReview2 = &pb.GradeReview{
		Id:          GradeReviewID2,
		Title:       GradeReviewTitle2,
		ReviewGrade: &GradeReviewGrade2, // optional field - pointer
		TeacherCode: GradeReviewTeacherCode2,
		Status:      pb.FinalStatus_PASSED,
		Notes:       &GradeReviewNotes2, // optional field - pointer
		CreatedAt:   timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:   timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		CreatedBy:   CreatedBy,
		UpdatedBy:   UpdatedBy,
	}

	GradeReviews = []*pb.GradeReview{GradeReview1, GradeReview2}
)
