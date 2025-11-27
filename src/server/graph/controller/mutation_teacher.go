package controller

import (
	"context"
	"fmt"

	pb "thaily/proto/common"
	pbCouncil "thaily/proto/council"
	pbFile "thaily/proto/file"
	pbThesis "thaily/proto/thesis"
	pbUser "thaily/proto/user"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================================
// TEACHER MUTATION CONTROLLER
// Handles: TeacherMutation resolvers (profile update)
// ============================================

// UpdateMyTeacherProfile updates teacher's own profile
func (c *Controller) UpdateMyTeacherProfile(ctx context.Context, input model.UpdateTeacherProfileInput) (*model.Teacher, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	req := &pbUser.UpdateTeacherRequest{
		Id:        *myId,
		UpdatedBy: *myId,
	}

	if input.Email != nil {
		req.Email = input.Email
	}
	if input.Username != nil {
		req.Username = input.Username
	}

	resp, err := c.user.UpdateTeacher(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbTeacherToModel(resp.GetTeacher()), nil
}

// ============================================
// SUPERVISOR MUTATION - Grading
// ============================================

// GradeMidterm grades a midterm (supervisor only)
func (c *Controller) GradeMidterm(ctx context.Context, enrollmentID string, input model.GradeMidtermInput) (*model.Midterm, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	enrollment, err := c.thesis.GetEnrollmentById(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}

	topicCouncilId := enrollment.GetEnrollment().GetTopicCouncilCode()
	if !c.verifySupervisor(ctx, *myId, topicCouncilId) {
		return nil, fmt.Errorf("not authorized: not supervisor of this topic")
	}

	midtermCode := enrollment.GetEnrollment().GetMidtermCode()
	if midtermCode == "" {
		return nil, fmt.Errorf("no midterm found for this enrollment")
	}

	var status pbThesis.MidtermStatus
	switch input.Status {
	case model.MidtermStatusNotSubmitted:
		status = pbThesis.MidtermStatus_NOT_SUBMITTED
	case model.MidtermStatusSubmitted:
		status = pbThesis.MidtermStatus_SUBMITTED
	case model.MidtermStatusPass:
		status = pbThesis.MidtermStatus_PASS
	case model.MidtermStatusFail:
		status = pbThesis.MidtermStatus_FAIL
	}

	grade := int32(input.Grade)
	req := &pbThesis.UpdateMidtermRequest{
		Id:        midtermCode,
		Grade:     &grade,
		Status:    &status,
		UpdatedBy: *myId,
	}

	if input.Feedback != nil {
		req.Feedback = input.Feedback
	}

	resp, err := c.thesis.UpdateMidterm(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbMidtermToModel(resp.GetMidterm()), nil
}

// FeedbackMidterm adds feedback to midterm
func (c *Controller) FeedbackMidterm(ctx context.Context, midtermID string, feedback string) (*model.Midterm, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	midterm, err := c.thesis.GetMidtermById(ctx, midtermID)
	if err != nil {
		return nil, err
	}

	enrollmentSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{Page: 1, PageSize: 1},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "midterm_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{midtermID},
					},
				},
			},
		},
	}
	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, &enrollmentSearch)
	if err != nil {
		return nil, err
	}
	if len(enrollments.GetEnrollments()) == 0 {
		return nil, fmt.Errorf("no enrollment found for this midterm")
	}
	enrollment := enrollments.GetEnrollments()[0]

	if !c.verifySupervisor(ctx, *myId, enrollment.GetTopicCouncilCode()) {
		return nil, fmt.Errorf("not authorized")
	}

	req := &pbThesis.UpdateMidtermRequest{
		Id:        midterm.GetMidterm().GetId(),
		Feedback:  &feedback,
		UpdatedBy: *myId,
	}

	resp, err := c.thesis.UpdateMidterm(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbMidtermToModel(resp.GetMidterm()), nil
}

// GradeFinal grades a final (supervisor only)
func (c *Controller) GradeFinal(ctx context.Context, enrollmentID string, input model.GradeFinalInput) (*model.Final, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	enrollment, err := c.thesis.GetEnrollmentById(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}

	if !c.verifySupervisor(ctx, *myId, enrollment.GetEnrollment().GetTopicCouncilCode()) {
		return nil, fmt.Errorf("not authorized")
	}

	finalCode := enrollment.GetEnrollment().GetFinalCode()
	if finalCode == "" {
		return nil, fmt.Errorf("no final found for this enrollment")
	}

	var status pbThesis.FinalStatus
	switch input.Status {
	case model.FinalStatusPending:
		status = pbThesis.FinalStatus_PENDING
	case model.FinalStatusPassed:
		status = pbThesis.FinalStatus_PASSED
	case model.FinalStatusFailed:
		status = pbThesis.FinalStatus_FAILED
	case model.FinalStatusCompleted:
		status = pbThesis.FinalStatus_COMPLETED
	}

	supervisorGrade := int32(input.SupervisorGrade)
	req := &pbThesis.UpdateFinalRequest{
		Id:              finalCode,
		SupervisorGrade: &supervisorGrade,
		Status:          &status,
		UpdatedBy:       *myId,
	}

	if input.Notes != nil {
		req.Notes = input.Notes
	}

	resp, err := c.thesis.UpdateFinal(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbFinalToModel(resp.GetFinal()), nil
}

// FeedbackFinal adds notes to final
func (c *Controller) FeedbackFinal(ctx context.Context, finalID string, notes string) (*model.Final, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	final, err := c.thesis.GetFinalById(ctx, finalID)
	if err != nil {
		return nil, err
	}

	enrollmentSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{Page: 1, PageSize: 1},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "final_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{finalID},
					},
				},
			},
		},
	}
	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, &enrollmentSearch)
	if err != nil {
		return nil, err
	}
	if len(enrollments.GetEnrollments()) == 0 {
		return nil, fmt.Errorf("no enrollment found for this final")
	}
	enrollment := enrollments.GetEnrollments()[0]

	if !c.verifySupervisor(ctx, *myId, enrollment.GetTopicCouncilCode()) {
		return nil, fmt.Errorf("not authorized")
	}

	req := &pbThesis.UpdateFinalRequest{
		Id:        final.GetFinal().GetId(),
		Notes:     &notes,
		UpdatedBy: *myId,
	}

	resp, err := c.thesis.UpdateFinal(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbFinalToModel(resp.GetFinal()), nil
}

// ============================================
// SUPERVISOR MUTATION - File Approval
// ============================================

// ApproveMidtermFile approves a midterm file
func (c *Controller) ApproveMidtermFile(ctx context.Context, fileID string) (*model.File, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	status := pbFile.FileStatus_APPROVED
	req := &pbFile.UpdateFileRequest{
		Id:        fileID,
		Status:    &status,
		UpdatedBy: *myId,
	}

	resp, err := c.file.UpdateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbFileToModel(resp.GetFile()), nil
}

// RejectMidtermFile rejects a midterm file
func (c *Controller) RejectMidtermFile(ctx context.Context, fileID string, reason *string) (*model.File, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	status := pbFile.FileStatus_REJECTED
	req := &pbFile.UpdateFileRequest{
		Id:        fileID,
		Status:    &status,
		UpdatedBy: *myId,
	}

	resp, err := c.file.UpdateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbFileToModel(resp.GetFile()), nil
}

// ApproveFinalFile approves a final file
func (c *Controller) ApproveFinalFile(ctx context.Context, fileID string) (*model.File, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	status := pbFile.FileStatus_APPROVED
	req := &pbFile.UpdateFileRequest{
		Id:        fileID,
		Status:    &status,
		UpdatedBy: *myId,
	}

	resp, err := c.file.UpdateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbFileToModel(resp.GetFile()), nil
}

// RejectFinalFile rejects a final file
func (c *Controller) RejectFinalFile(ctx context.Context, fileID string, reason *string) (*model.File, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	status := pbFile.FileStatus_REJECTED
	req := &pbFile.UpdateFileRequest{
		Id:        fileID,
		Status:    &status,
		UpdatedBy: *myId,
	}

	resp, err := c.file.UpdateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbFileToModel(resp.GetFile()), nil
}

// ============================================
// SUPERVISOR MUTATION - Topic & TopicCouncil Creation
// ============================================

// CreateTopicForSuperVisor creates a topic with topic council and enrollments
func (c *Controller) CreateTopicForSuperVisor(ctx context.Context, input model.CreateTopicInput) (*model.Topic, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}
	if teacher == nil || teacher.GetTeacher() == nil {
		return nil, fmt.Errorf("teacher not found")
	}

	if input.Curriculum == nil {
		return nil, fmt.Errorf("curriculum is required")
	}

	// Track created resources for rollback
	type createdResources struct {
		topicId        string
		topicCouncilId string
		supervisorId   string
		enrollmentIds  []string
	}
	resources := &createdResources{enrollmentIds: make([]string, 0)}
	shouldRollback := true

	rollback := func() {
		for _, enrollmentId := range resources.enrollmentIds {
			_, _ = c.thesis.DeleteEnrollment(ctx, enrollmentId)
		}
		if resources.supervisorId != "" {
			_, _ = c.thesis.DeleteTopicCouncilSupervisor(ctx, resources.supervisorId)
		}
		if resources.topicCouncilId != "" {
			_, _ = c.thesis.DeleteTopicCouncil(ctx, resources.topicCouncilId)
		}
		if resources.topicId != "" {
			_, _ = c.thesis.DeleteTopic(ctx, resources.topicId)
		}
	}

	defer func() {
		if shouldRollback {
			rollback()
		}
	}()

	// Create topic
	topicResp, err := c.thesis.CreateTopic(ctx, &pbThesis.CreateTopicRequest{
		Title:        input.Title,
		TitleEn:      input.TitleEn,
		Description:  input.Description,
		Curriculum:   *input.Curriculum,
		MajorCode:    teacher.GetTeacher().GetMajorCode(),
		SemesterCode: teacher.GetTeacher().GetSemesterCode(),
		Status:       pbThesis.TopicStatus_SUBMIT,
		CreatedBy:    *myId,
	})
	if err != nil {
		return nil, err
	}
	resources.topicId = topicResp.GetTopic().GetId()

	// Create topic council
	stage := pbThesis.TopicStage_STAGE_DACN
	if input.Stage == model.TopicStageStageLvtn {
		stage = pbThesis.TopicStage_STAGE_LVTN
	}

	topicCouncilResp, err := c.thesis.CreateTopicCouncil(ctx, &pbThesis.CreateTopicCouncilRequest{
		Title:     fmt.Sprintf("Topic Council for %s", input.Title),
		TopicCode: resources.topicId,
		Stage:     stage,
		TimeStart: timestamppb.New(input.TimeStart),
		TimeEnd:   timestamppb.New(input.TimeEnd),
		CreatedBy: *myId,
	})
	if err != nil {
		return nil, err
	}
	resources.topicCouncilId = topicCouncilResp.GetTopicCouncil().GetId()

	// Create supervisor
	supervisorResp, err := c.thesis.CreateTopicCouncilSupervisor(ctx, &pbThesis.CreateTopicCouncilSupervisorRequest{
		TopicCouncilCode:      resources.topicCouncilId,
		TeacherSupervisorCode: *myId,
		CreatedBy:             *myId,
	})
	if err != nil {
		return nil, err
	}
	resources.supervisorId = supervisorResp.GetTopicCouncilSupervisor().GetId()

	// Create enrollments
	for _, student := range input.Students {
		idStudent := fmt.Sprint(student, "_", teacher.GetTeacher().GetSemesterCode())
		enrollmentResp, err := c.thesis.CreateEnrollment(ctx, &pbThesis.CreateEnrollmentRequest{
			TopicCouncilCode: resources.topicCouncilId,
			StudentCode:      idStudent,
			CreatedBy:        *myId,
			Title:            fmt.Sprintf("Enrollment for %s", student),
		})
		if err != nil {
			return nil, err
		}
		resources.enrollmentIds = append(resources.enrollmentIds, enrollmentResp.GetEnrollment().GetId())
	}

	shouldRollback = false
	return convert.PbTopicToModel(topicResp.GetTopic()), nil
}

// CreateTopicCouncilForSuperVisor creates a topic council for existing topic
func (c *Controller) CreateTopicCouncilForSuperVisor(ctx context.Context, input model.CreateTopicCouncilInput) (*model.TopicCouncil, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	// Default to DACN stage
	stage := pbThesis.TopicStage_STAGE_DACN

	topicCouncilResp, err := c.thesis.CreateTopicCouncil(ctx, &pbThesis.CreateTopicCouncilRequest{
		Title:     fmt.Sprintf("Topic Council for %s", input.TopicCode),
		TopicCode: input.TopicCode,
		Stage:     stage,
		TimeStart: timestamppb.New(input.TimeStart),
		TimeEnd:   timestamppb.New(input.TimeEnd),
		CreatedBy: *myId,
	})
	if err != nil {
		return nil, err
	}

	// Add supervisor
	_, err = c.thesis.CreateTopicCouncilSupervisor(ctx, &pbThesis.CreateTopicCouncilSupervisorRequest{
		TopicCouncilCode:      topicCouncilResp.GetTopicCouncil().GetId(),
		TeacherSupervisorCode: *myId,
		CreatedBy:             *myId,
	})
	if err != nil {
		return nil, err
	}

	// Create enrollments
	for _, student := range input.Students {
		idStudent := fmt.Sprint(student, "_", teacher.GetTeacher().GetSemesterCode())
		_, err = c.thesis.CreateEnrollment(ctx, &pbThesis.CreateEnrollmentRequest{
			TopicCouncilCode: topicCouncilResp.GetTopicCouncil().GetId(),
			StudentCode:      idStudent,
			CreatedBy:        *myId,
			Title:            fmt.Sprintf("Enrollment for %s", student),
		})
		if err != nil {
			return nil, err
		}
	}

	return convert.PbTopicCouncilToModel(topicCouncilResp.GetTopicCouncil()), nil
}

// ============================================
// COUNCIL MEMBER MUTATION - Grade Defence
// ============================================

// CreateGradeDefence creates a grade defence
func (c *Controller) CreateGradeDefence(ctx context.Context, input model.CreateGradeDefenceInput) (*model.GradeDefence, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	req := &pbCouncil.CreateGradeDefenceRequest{
		DefenceCode:    input.DefenceCode,
		EnrollmentCode: input.EnrollmentCode,
		CreatedBy:      *myId,
	}

	if input.Note != nil {
		req.Note = input.Note
	}
	if input.TotalScore != nil {
		req.TotalScore = input.TotalScore
	}

	resp, err := c.council.CreateGradeDefence(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefenceToModel(resp.GetGradeDefence()), nil
}

// UpdateGradeDefence updates a grade defence
func (c *Controller) UpdateGradeDefence(ctx context.Context, id string, input model.UpdateGradeDefenceInput) (*model.GradeDefence, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	req := &pbCouncil.UpdateGradeDefenceRequest{
		Id:        id,
		UpdatedBy: *myId,
	}

	if input.Note != nil {
		req.Note = input.Note
	}
	if input.TotalScore != nil {
		req.TotalScore = input.TotalScore
	}

	resp, err := c.council.UpdateGradeDefence(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefenceToModel(resp.GetGradeDefence()), nil
}

// AddGradeDefenceCriterion adds a criterion to grade defence
func (c *Controller) AddGradeDefenceCriterion(ctx context.Context, input model.CreateGradeDefenceCriterionInput) (*model.GradeDefenceCriterion, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	resp, err := c.council.CreateGradeDefenceCriterion(ctx, &pbCouncil.CreateGradeDefenceCriterionRequest{
		GradeDefenceCode: input.GradeDefenceCode,
		Name:             &input.Name,
		Score:            &input.Score,
		MaxScore:         &input.MaxScore,
		CreatedBy:        &*myId,
	})
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefenceCriterionToModel(resp.GetGradeDefenceCriterion()), nil
}

// UpdateGradeDefenceCriterion updates a criterion
func (c *Controller) UpdateGradeDefenceCriterion(ctx context.Context, id string, input model.UpdateGradeDefenceCriterionInput) (*model.GradeDefenceCriterion, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	updatedBy := *myId
	req := &pbCouncil.UpdateGradeDefenceCriterionRequest{
		Id:        id,
		UpdatedBy: &updatedBy,
	}

	if input.Name != nil {
		req.Name = input.Name
	}
	if input.Score != nil {
		req.Score = input.Score
	}
	if input.MaxScore != nil {
		req.MaxScore = input.MaxScore
	}

	resp, err := c.council.UpdateGradeDefenceCriterion(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefenceCriterionToModel(resp.GetGradeDefenceCriterion()), nil
}

// DeleteGradeDefenceCriterion deletes a criterion
func (c *Controller) DeleteGradeDefenceCriterion(ctx context.Context, id string) (bool, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return false, err
	}
	if !check || myId == nil {
		return false, fmt.Errorf("not authorized")
	}

	_, err = c.council.DeleteGradeDefenceCriterion(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// verifySupervisor checks if teacher is supervisor of a topic council
func (c *Controller) verifySupervisor(ctx context.Context, teacherId string, topicCouncilId string) bool {
	search := &pb.SearchRequest{
		Pagination: &pb.Pagination{Page: 1, PageSize: 10},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "topic_council_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{topicCouncilId},
					},
				},
			},
		},
	}

	supervisors, err := c.thesis.GetTopicCouncilSupervisorBySearch(ctx, search)
	if err != nil {
		return false
	}

	for _, s := range supervisors.GetTopicCouncilSupervisors() {
		if s.GetTeacherSupervisorCode() == teacherId {
			return true
		}
	}

	return false
}
