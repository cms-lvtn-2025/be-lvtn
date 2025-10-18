package convert

import (
	"thaily/proto/council"
	"thaily/proto/thesis"
	"thaily/src/graph/model"
)

// PbMidtermToModel converts protobuf Midterm to GraphQL Midterm
func PbMidtermToModel(pb *thesis.Midterm) *model.Midterm {
	if pb == nil {
		return nil
	}

	result := &model.Midterm{
		ID:     pb.Id,
		Title:  pb.Title,
		Status: PbMidtermStatusToModel(pb.Status),
	}

	// Handle optional Grade
	if pb.Grade != 0 {
		result.Grade = &pb.Grade
	}

	// Handle optional Feedback
	if pb.Feedback != "" {
		result.Feedback = &pb.Feedback
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbFinalToModel converts protobuf Final to GraphQL Final
func PbFinalToModel(pb *thesis.Final) *model.Final {
	if pb == nil {
		return nil
	}

	result := &model.Final{
		ID:     pb.Id,
		Title:  pb.Title,
		Status: PbFinalStatusToModel(pb.Status),
	}

	// Handle optional grades
	if pb.SupervisorGrade != 0 {
		result.SupervisorGrade = &pb.SupervisorGrade
	}
	if pb.DepartmentGrade != 0 {
		result.DepartmentGrade = &pb.DepartmentGrade
	}
	if pb.FinalGrade != 0 {
		result.FinalGrade = &pb.FinalGrade
	}

	// Handle optional Notes
	if pb.Notes != "" {
		result.Notes = &pb.Notes
	}

	// Handle timestamps
	if pb.CompletionDate != nil {
		t := pb.CompletionDate.AsTime()
		result.CompletionDate = &t
	}
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbTopicToModel converts protobuf Topic to GraphQL Topic
func PbTopicToModel(pb *thesis.Topic) *model.Topic {
	if pb == nil {
		return nil
	}

	result := &model.Topic{
		ID:           pb.Id,
		Title:        pb.Title,
		MajorCode:    pb.MajorCode,
		SemesterCode: pb.SemesterCode,
		Status:       PbTopicStatusToModel(pb.Status),
	}

	// Handle optional PercentStage fields
	if pb.PercentStage_1 != nil {
		result.PercentStage1 = pb.PercentStage_1
	}
	if pb.PercentStage_2 != nil {
		result.PercentStage2 = pb.PercentStage_2
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbMidtermStatusToModel converts protobuf MidtermStatus to GraphQL MidtermStatus
func PbMidtermStatusToModel(pb thesis.MidtermStatus) model.MidtermStatus {
	switch pb {
	case thesis.MidtermStatus_NOT_SUBMITTED:
		return model.MidtermStatusNotSubmitted
	case thesis.MidtermStatus_SUBMITTED:
		return model.MidtermStatusSubmitted
	case thesis.MidtermStatus_PASS:
		return model.MidtermStatusPass
	case thesis.MidtermStatus_FAIL:
		return model.MidtermStatusFail
	default:
		return model.MidtermStatusNotSubmitted
	}
}

// PbFinalStatusToModel converts protobuf FinalStatus to GraphQL FinalStatus
func PbFinalStatusToModel(pb thesis.FinalStatus) model.FinalStatus {
	switch pb {
	case thesis.FinalStatus_PENDING:
		return model.FinalStatusPending
	case thesis.FinalStatus_PASSED:
		return model.FinalStatusPassed
	case thesis.FinalStatus_FAILED:
		return model.FinalStatusFailed
	case thesis.FinalStatus_COMPLETED:
		return model.FinalStatusCompleted
	default:
		return model.FinalStatusPending
	}
}

// PbTopicStatusToModel converts protobuf TopicStatus to GraphQL TopicStatus
func PbTopicStatusToModel(pb thesis.TopicStatus) model.TopicStatus {
	switch pb {
	case thesis.TopicStatus_SUBMIT:
		return model.TopicStatusSubmit
	case thesis.TopicStatus_TOPIC_PENDING:
		return model.TopicStatusTopicPending
	case thesis.TopicStatus_APPROVED_1:
		return model.TopicStatusApproved1
	case thesis.TopicStatus_APPROVED_2:
		return model.TopicStatusApproved2
	case thesis.TopicStatus_IN_PROGRESS:
		return model.TopicStatusInProgress
	case thesis.TopicStatus_TOPIC_COMPLETED:
		return model.TopicStatusTopicCompleted
	case thesis.TopicStatus_REJECTED:
		return model.TopicStatusRejected
	default:
		return model.TopicStatusSubmit
	}
}

// PbTopicStageToModel converts protobuf TopicStage to GraphQL TopicStage
func PbTopicStageToModel(pb thesis.TopicStage) model.TopicStage {
	switch pb {
	case thesis.TopicStage_STAGE_DACN:
		return model.TopicStageStageDacn
	case thesis.TopicStage_STAGE_LVTN:
		return model.TopicStageStageLvtn
	default:
		return model.TopicStageStageDacn
	}
}

// PbMidtermsToModel converts array of protobuf Midterms to GraphQL Midterms
func PbMidtermsToModel(pbs []*thesis.Midterm) []*model.Midterm {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Midterm, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbMidtermToModel(pb))
		}
	}
	return result
}

// PbFinalsToModel converts array of protobuf Finals to GraphQL Finals
func PbFinalsToModel(pbs []*thesis.Final) []*model.Final {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Final, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbFinalToModel(pb))
		}
	}
	return result
}

// PbTopicsToModel converts array of protobuf Topics to GraphQL Topics
func PbTopicsToModel(pbs []*thesis.Topic) []*model.Topic {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Topic, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbTopicToModel(pb))
		}
	}
	return result
}

// PbEnrollmentToModel converts protobuf Enrollment to GraphQL Enrollment
func PbEnrollmentToModel(pb *thesis.Enrollment) *model.Enrollment {
	if pb == nil {
		return nil
	}

	result := &model.Enrollment{
		ID:               pb.Id,
		Title:            pb.Title,
		StudentCode:      pb.StudentCode,
		TopicCouncilCode: pb.TopicCouncilCode,
	}

	// Handle optional codes
	if pb.FinalCode != nil {
		result.FinalCode = pb.FinalCode
	}
	if pb.GradeReviewCode != nil {
		result.GradeReviewCode = pb.GradeReviewCode
	}
	if pb.MidtermCode != nil {
		result.MidtermCode = pb.MidtermCode
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbEnrollmentsToModel converts array of protobuf Enrollments to GraphQL Enrollments
func PbEnrollmentsToModel(pbs []*thesis.Enrollment) []*model.Enrollment {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Enrollment, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbEnrollmentToModel(pb))
		}
	}
	return result
}

// PbTopicCouncilToModel converts protobuf TopicCouncil to GraphQL TopicCouncil
func PbTopicCouncilToModel(pb *thesis.TopicCouncil) *model.TopicCouncil {
	if pb == nil {
		return nil
	}

	result := &model.TopicCouncil{
		ID:        pb.Id,
		Title:     pb.Title,
		Stage:     PbTopicStageToModel(pb.Stage),
		TopicCode: pb.TopicCode,
	}

	// Handle optional CouncilCode
	if pb.CouncilCode != nil {
		result.CouncilCode = pb.CouncilCode
	}

	// Handle required timestamps (convert from *timestamppb to time.Time)
	if pb.TimeStart != nil {
		t := pb.TimeStart.AsTime()
		result.TimeStart = t
	}
	if pb.TimeEnd != nil {
		t := pb.TimeEnd.AsTime()
		result.TimeEnd = t
	}

	// Handle optional timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbTopicCouncilsToModel converts array of protobuf TopicCouncils to GraphQL TopicCouncils
func PbTopicCouncilsToModel(pbs []*thesis.TopicCouncil) []*model.TopicCouncil {
	if pbs == nil {
		return nil
	}

	result := make([]*model.TopicCouncil, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbTopicCouncilToModel(pb))
		}
	}
	return result
}

// PbTopicCouncilSupervisorToModel converts protobuf TopicCouncilSupervisor to GraphQL TopicCouncilSupervisor
func PbTopicCouncilSupervisorToModel(pb *thesis.TopicCouncilSupervisor) *model.TopicCouncilSupervisor {
	if pb == nil {
		return nil
	}

	result := &model.TopicCouncilSupervisor{
		ID:                    pb.Id,
		TeacherSupervisorCode: pb.TeacherSupervisorCode,
		TopicCouncilCode:      pb.TopicCouncilCode,
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbTopicCouncilSupervisorsToModel converts array of protobuf TopicCouncilSupervisors to GraphQL TopicCouncilSupervisors
func PbTopicCouncilSupervisorsToModel(pbs []*thesis.TopicCouncilSupervisor) []*model.TopicCouncilSupervisor {
	if pbs == nil {
		return nil
	}

	result := make([]*model.TopicCouncilSupervisor, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbTopicCouncilSupervisorToModel(pb))
		}
	}
	return result
}

// PbGradeReviewToModel converts protobuf GradeReview to GraphQL GradeReview
func PbGradeReviewToModel(pb *thesis.GradeReview) *model.GradeReview {
	if pb == nil {
		return nil
	}

	result := &model.GradeReview{
		ID:          pb.Id,
		Title:       pb.Title,
		TeacherCode: pb.TeacherCode,
		Status:      PbFinalStatusToModel(pb.Status),
	}

	// Handle optional ReviewGrade
	if pb.ReviewGrade != nil {
		result.ReviewGrade = pb.ReviewGrade
	}

	// Handle optional Notes
	if pb.Notes != nil {
		result.Notes = pb.Notes
	}

	// Handle timestamps
	if pb.CompletionDate != nil {
		t := pb.CompletionDate.AsTime()
		result.CompletionDate = &t
	}
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbGradeReviewsToModel converts array of protobuf GradeReviews to GraphQL GradeReviews
func PbGradeReviewsToModel(pbs []*thesis.GradeReview) []*model.GradeReview {
	if pbs == nil {
		return nil
	}

	result := make([]*model.GradeReview, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbGradeReviewToModel(pb))
		}
	}
	return result
}

// ============================================
// STUDENT VIEW CONVERTERS
// ============================================

// PbEnrollmentToStudentEnrollment converts protobuf Enrollment to GraphQL StudentEnrollment
func PbEnrollmentToStudentEnrollment(pb *thesis.Enrollment) *model.StudentEnrollment {
	if pb == nil {
		return nil
	}

	result := &model.StudentEnrollment{
		ID:               pb.Id,
		Title:            pb.Title,
		StudentCode:      pb.StudentCode,
		TopicCouncilCode: pb.TopicCouncilCode,
	}

	// Handle optional codes
	if pb.FinalCode != nil {
		result.FinalCode = pb.FinalCode
	}
	if pb.GradeReviewCode != nil {
		result.GradeReviewCode = pb.GradeReviewCode
	}
	if pb.MidtermCode != nil {
		result.MidtermCode = pb.MidtermCode
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbTopicCouncilToStudentTopicCouncil converts protobuf TopicCouncil to GraphQL StudentTopicCouncil
func PbTopicCouncilToStudentTopicCouncil(pb *thesis.TopicCouncil) *model.StudentTopicCouncil {
	if pb == nil {
		return nil
	}

	result := &model.StudentTopicCouncil{
		ID:        pb.Id,
		Title:     pb.Title,
		Stage:     PbTopicStageToModel(pb.Stage),
		TopicCode: pb.TopicCode,
	}

	// Handle optional CouncilCode
	if pb.CouncilCode != nil {
		result.CouncilCode = pb.CouncilCode
	}

	// Handle required timestamps
	if pb.TimeStart != nil {
		result.TimeStart = pb.TimeStart.AsTime()
	}
	if pb.TimeEnd != nil {
		result.TimeEnd = pb.TimeEnd.AsTime()
	}

	// Handle optional timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	return result
}

// PbTopicToStudentTopic converts protobuf Topic to GraphQL StudentTopic
func PbTopicToStudentTopic(pb *thesis.Topic) *model.StudentTopic {
	if pb == nil {
		return nil
	}

	result := &model.StudentTopic{
		ID:           pb.Id,
		Title:        pb.Title,
		MajorCode:    pb.MajorCode,
		SemesterCode: pb.SemesterCode,
		Status:       PbTopicStatusToModel(pb.Status),
	}

	// Handle optional PercentStage fields
	if pb.PercentStage_1 != nil {
		result.PercentStage1 = pb.PercentStage_1
	}
	if pb.PercentStage_2 != nil {
		result.PercentStage2 = pb.PercentStage_2
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	return result
}

// PbTopicCouncilSupervisorToStudentTopicSupervisor converts protobuf TopicCouncilSupervisor to GraphQL StudentTopicSupervisor
func PbTopicCouncilSupervisorToStudentTopicSupervisor(pb *thesis.TopicCouncilSupervisor) *model.StudentTopicSupervisor {
	if pb == nil {
		return nil
	}

	return &model.StudentTopicSupervisor{
		ID:                    pb.Id,
		TeacherSupervisorCode: pb.TeacherSupervisorCode,
		TopicCouncilCode:      pb.TopicCouncilCode,
	}
}

// PbCouncilToStudentCouncil converts protobuf Council to GraphQL StudentCouncil
func PbCouncilToStudentCouncil(pb *council.Council) *model.StudentCouncil {
	if pb == nil {
		return nil
	}

	result := &model.StudentCouncil{
		ID:           pb.Id,
		Title:        pb.Title,
		MajorCode:    pb.MajorCode,
		SemesterCode: pb.SemesterCode,
	}

	// Handle optional TimeStart
	if pb.TimeStart != nil {
		t := pb.TimeStart.AsTime()
		result.TimeStart = &t
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	return result
}

// ============================================
// SUPERVISOR VIEW CONVERTERS
// ============================================

// PbEnrollmentToSupervisorEnrollment converts protobuf Enrollment to GraphQL SupervisorEnrollment
func PbEnrollmentToSupervisorEnrollment(pb *thesis.Enrollment) *model.SupervisorEnrollment {
	if pb == nil {
		return nil
	}

	result := &model.SupervisorEnrollment{
		ID:               pb.Id,
		Title:            pb.Title,
		StudentCode:      pb.StudentCode,
		TopicCouncilCode: pb.TopicCouncilCode,
	}

	// Handle optional codes
	if pb.FinalCode != nil {
		result.FinalCode = pb.FinalCode
	}
	if pb.GradeReviewCode != nil {
		result.GradeReviewCode = pb.GradeReviewCode
	}
	if pb.MidtermCode != nil {
		result.MidtermCode = pb.MidtermCode
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbTopicCouncilToSupervisorTopicCouncil converts protobuf TopicCouncil to GraphQL SupervisorTopicCouncil
func PbTopicCouncilToSupervisorTopicCouncil(pb *thesis.TopicCouncil) *model.SupervisorTopicCouncil {
	if pb == nil {
		return nil
	}

	result := &model.SupervisorTopicCouncil{
		ID:        pb.Id,
		Title:     pb.Title,
		Stage:     PbTopicStageToModel(pb.Stage),
		TopicCode: pb.TopicCode,
	}

	// Handle optional CouncilCode
	if pb.CouncilCode != nil {
		result.CouncilCode = pb.CouncilCode
	}

	// Handle required timestamps
	if pb.TimeStart != nil {
		result.TimeStart = pb.TimeStart.AsTime()
	}
	if pb.TimeEnd != nil {
		result.TimeEnd = pb.TimeEnd.AsTime()
	}

	// Handle optional timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	return result
}

// PbTopicToSupervisorTopic converts protobuf Topic to GraphQL SupervisorTopic
func PbTopicToSupervisorTopic(pb *thesis.Topic) *model.SupervisorTopic {
	if pb == nil {
		return nil
	}

	result := &model.SupervisorTopic{
		ID:           pb.Id,
		Title:        pb.Title,
		MajorCode:    pb.MajorCode,
		SemesterCode: pb.SemesterCode,
		Status:       PbTopicStatusToModel(pb.Status),
	}

	// Handle optional PercentStage fields
	if pb.PercentStage_1 != nil {
		result.PercentStage1 = pb.PercentStage_1
	}
	if pb.PercentStage_2 != nil {
		result.PercentStage2 = pb.PercentStage_2
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbTopicCouncilSupervisorToSupervisorTopicCouncilAssignment converts protobuf TopicCouncilSupervisor to GraphQL SupervisorTopicCouncilAssignment
func PbTopicCouncilSupervisorToSupervisorTopicCouncilAssignment(pb *thesis.TopicCouncilSupervisor) *model.SupervisorTopicCouncilAssignment {
	if pb == nil {
		return nil
	}

	result := &model.SupervisorTopicCouncilAssignment{
		ID:                    pb.Id,
		TeacherSupervisorCode: pb.TeacherSupervisorCode,
		TopicCouncilCode:      pb.TopicCouncilCode,
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	return result
}

// ============================================
// REVIEWER VIEW CONVERTERS
// ============================================

// PbEnrollmentToReviewerEnrollment converts protobuf Enrollment to GraphQL ReviewerEnrollment
func PbEnrollmentToReviewerEnrollment(pb *thesis.Enrollment) *model.ReviewerEnrollment {
	if pb == nil {
		return nil
	}

	result := &model.ReviewerEnrollment{
		ID:               pb.Id,
		Title:            pb.Title,
		StudentCode:      pb.StudentCode,
		TopicCouncilCode: pb.TopicCouncilCode,
	}

	// Handle optional codes
	if pb.GradeReviewCode != nil {
		result.GradeReviewCode = pb.GradeReviewCode
	}
	if pb.MidtermCode != nil {
		result.MidtermCode = pb.MidtermCode
	}
	if pb.FinalCode != nil {
		result.FinalCode = pb.FinalCode
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	return result
}

// PbTopicCouncilToReviewerTopicCouncil converts protobuf TopicCouncil to GraphQL ReviewerTopicCouncil
func PbTopicCouncilToReviewerTopicCouncil(pb *thesis.TopicCouncil) *model.ReviewerTopicCouncil {
	if pb == nil {
		return nil
	}

	result := &model.ReviewerTopicCouncil{
		ID:        pb.Id,
		Title:     pb.Title,
		Stage:     PbTopicStageToModel(pb.Stage),
		TopicCode: pb.TopicCode,
	}

	// Handle required timestamps
	if pb.TimeStart != nil {
		result.TimeStart = pb.TimeStart.AsTime()
	}
	if pb.TimeEnd != nil {
		result.TimeEnd = pb.TimeEnd.AsTime()
	}

	return result
}

// PbTopicToReviewerTopic converts protobuf Topic to GraphQL ReviewerTopic
func PbTopicToReviewerTopic(pb *thesis.Topic) *model.ReviewerTopic {
	if pb == nil {
		return nil
	}

	return &model.ReviewerTopic{
		ID:        pb.Id,
		Title:     pb.Title,
		Status:    PbTopicStatusToModel(pb.Status),
		MajorCode: pb.MajorCode,
	}
}

// PbGradeReviewToReviewerGradeReview converts protobuf GradeReview to GraphQL ReviewerGradeReview
func PbGradeReviewToReviewerGradeReview(pb *thesis.GradeReview) *model.ReviewerGradeReview {
	if pb == nil {
		return nil
	}

	result := &model.ReviewerGradeReview{
		ID:          pb.Id,
		Title:       pb.Title,
		TeacherCode: pb.TeacherCode,
		Status:      PbFinalStatusToModel(pb.Status),
	}

	// Handle optional ReviewGrade
	if pb.ReviewGrade != nil {
		result.ReviewGrade = pb.ReviewGrade
	}

	// Handle optional Notes
	if pb.Notes != nil {
		result.Notes = pb.Notes
	}

	// Handle timestamps
	if pb.CompletionDate != nil {
		t := pb.CompletionDate.AsTime()
		result.CompletionDate = &t
	}
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	return result
}

// ============================================
// LIST RESPONSE FACTORY FUNCTIONS
// ============================================

// CreateMidtermListResponse creates a MidtermListResponse
func CreateMidtermListResponse(midterms []*model.Midterm, total int32) *model.MidtermListResponse {
	return &model.MidtermListResponse{
		Data:  midterms,
		Total: total,
	}
}

// CreateFinalListResponse creates a FinalListResponse
func CreateFinalListResponse(finals []*model.Final, total int32) *model.FinalListResponse {
	return &model.FinalListResponse{
		Data:  finals,
		Total: total,
	}
}

// CreateTopicListResponse creates a TopicListResponse
func CreateTopicListResponse(topics []*model.Topic, total int32) *model.TopicListResponse {
	return &model.TopicListResponse{
		Data:  topics,
		Total: total,
	}
}

// CreateEnrollmentListResponse creates an EnrollmentListResponse
func CreateEnrollmentListResponse(enrollments []*model.Enrollment, total int32) *model.EnrollmentListResponse {
	return &model.EnrollmentListResponse{
		Data:  enrollments,
		Total: total,
	}
}

// CreateTopicCouncilListResponse creates a TopicCouncilListResponse
func CreateTopicCouncilListResponse(topicCouncils []*model.TopicCouncil, total int32) *model.TopicCouncilListResponse {
	return &model.TopicCouncilListResponse{
		Data:  topicCouncils,
		Total: total,
	}
}

// CreateTopicCouncilSupervisorListResponse creates a TopicCouncilSupervisorListResponse
func CreateTopicCouncilSupervisorListResponse(supervisors []*model.TopicCouncilSupervisor, total int32) *model.TopicCouncilSupervisorListResponse {
	return &model.TopicCouncilSupervisorListResponse{
		Data:  supervisors,
		Total: total,
	}
}

// CreateGradeReviewListResponse creates a GradeReviewListResponse
func CreateGradeReviewListResponse(gradeReviews []*model.GradeReview, total int32) *model.GradeReviewListResponse {
	return &model.GradeReviewListResponse{
		Data:  gradeReviews,
		Total: total,
	}
}

// CreateStudentEnrollmentListResponse creates a StudentEnrollmentListResponse
func CreateStudentEnrollmentListResponse(enrollments []*model.StudentEnrollment, total int32) *model.StudentEnrollmentListResponse {
	return &model.StudentEnrollmentListResponse{
		Data:  enrollments,
		Total: total,
	}
}

// CreateStudentTopicSupervisorListResponse creates a StudentTopicSupervisorListResponse
func CreateStudentTopicSupervisorListResponse(supervisors []*model.StudentTopicSupervisor, total int32) *model.StudentTopicSupervisorListResponse {
	return &model.StudentTopicSupervisorListResponse{
		Data:  supervisors,
		Total: total,
	}
}

// CreateSupervisorEnrollmentListResponse creates a SupervisorEnrollmentListResponse
func CreateSupervisorEnrollmentListResponse(enrollments []*model.SupervisorEnrollment, total int32) *model.SupervisorEnrollmentListResponse {
	return &model.SupervisorEnrollmentListResponse{
		Data:  enrollments,
		Total: total,
	}
}

// CreateSupervisorTopicCouncilListResponse creates a SupervisorTopicCouncilListResponse
func CreateSupervisorTopicCouncilListResponse(topicCouncils []*model.SupervisorTopicCouncil, total int32) *model.SupervisorTopicCouncilListResponse {
	return &model.SupervisorTopicCouncilListResponse{
		Data:  topicCouncils,
		Total: total,
	}
}

// CreateSupervisorTopicCouncilAssignmentListResponse creates a SupervisorTopicCouncilAssignmentListResponse
func CreateSupervisorTopicCouncilAssignmentListResponse(assignments []*model.SupervisorTopicCouncilAssignment, total int32) *model.SupervisorTopicCouncilAssignmentListResponse {
	return &model.SupervisorTopicCouncilAssignmentListResponse{
		Data:  assignments,
		Total: total,
	}
}

// CreateReviewerGradeReviewListResponse creates a ReviewerGradeReviewListResponse
func CreateReviewerGradeReviewListResponse(gradeReviews []*model.ReviewerGradeReview, total int32) *model.ReviewerGradeReviewListResponse {
	return &model.ReviewerGradeReviewListResponse{
		Data:  gradeReviews,
		Total: total,
	}
}
