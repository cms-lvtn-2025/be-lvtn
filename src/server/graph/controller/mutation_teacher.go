package controller

import (
	"context"
	"fmt"

	pbCouncil "thaily/proto/council"
	pbFile "thaily/proto/file"
	pbThesis "thaily/proto/thesis"
	pbUser "thaily/proto/user"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/dataloader"
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

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateTeacher(*myId)
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
	// check midterm is not pass and not fail
	// get enrollment
	enrollment, err := c.thesis.GetEnrollmentById(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}
	midterm, err := c.thesis.GetMidtermById(ctx, enrollment.GetEnrollment().GetMidtermCode())
	if err != nil {
		return nil, err
	}
	if midterm.GetMidterm().GetStatus() == pbThesis.MidtermStatus_PASS || midterm.GetMidterm().GetStatus() == pbThesis.MidtermStatus_FAIL {
		return nil, fmt.Errorf("midterm is already graded")
	}

	if midterm.GetMidterm().GetCreatedBy() != *myId {
		return nil, fmt.Errorf("you are not the creator of this midterm")
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
		Id:        midterm.GetMidterm().GetId(),
		Grade:     &grade,
		Status:    &status,
		UpdatedBy: *myId,
	}

	if input.Feedback != nil {
		req.Feedback = input.Feedback
	}

	resp, err := c.thesis.UpdateMidterm(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error")
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && resp.GetMidterm() != nil {
		loaders.InvalidateMidterm(resp.GetMidterm().GetId())
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
	// status is not submitted or not graded
	if midterm.GetMidterm().GetStatus() != pbThesis.MidtermStatus_NOT_SUBMITTED && midterm.GetMidterm().GetStatus() != pbThesis.MidtermStatus_SUBMITTED {
		return nil, fmt.Errorf("midterm is not submitted or not graded")
	}

	if midterm.GetMidterm().GetCreatedBy() != *myId {
		return nil, fmt.Errorf("you are not the creator of this midterm")
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

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateMidterm(midtermID)
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

	finalCode := enrollment.GetEnrollment().GetFinalCode()
	if finalCode == "" {
		return nil, fmt.Errorf("no final found for this enrollment")
	}
	// status is not pending or not completed
	final, err := c.thesis.GetFinalById(ctx, finalCode)
	if err != nil {
		return nil, err
	}
	if final.GetFinal().GetStatus() != pbThesis.FinalStatus_PENDING {
		return nil, fmt.Errorf("final is not pending")
	}

	if final.GetFinal().GetCreatedBy() != *myId {
		return nil, fmt.Errorf("you are not the creator of this final")
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

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && resp.GetFinal() != nil {
		loaders.InvalidateFinal(resp.GetFinal().GetId())
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

	if final.GetFinal().GetStatus() != pbThesis.FinalStatus_PENDING {
		return nil, fmt.Errorf("final is not pending")
	}
	if final.GetFinal().GetCreatedBy() != *myId {
		return nil, fmt.Errorf("you are not the creator of this final")
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

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateFinal(finalID)
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

	// Note: File cache invalidation happens via FilesByMidtermId/FilesByFinalId
	// Since we don't have the midterm/final ID here, cache will expire naturally

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

	if len(input.Students) == 0 {
		return nil, fmt.Errorf("students are required")
	}

	// check teacher ID is member of topic council

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

		midtermResp, err := c.thesis.CreateMidterm(ctx, &pbThesis.CreateMidtermRequest{
			Title:     fmt.Sprintf("Midterm for %s", student),
			Status:    pbThesis.MidtermStatus_NOT_SUBMITTED,
			CreatedBy: *myId,
		})
		if err != nil {
			return nil, err
		}
		finalResp, err := c.thesis.CreateFinal(ctx, &pbThesis.CreateFinalRequest{
			Title:     fmt.Sprintf("Final for %s", student),
			Status:    pbThesis.FinalStatus_PENDING,
			CreatedBy: *myId,
		})
		if err != nil {
			return nil, err
		}
		enrollmentResp, err := c.thesis.CreateEnrollment(ctx, &pbThesis.CreateEnrollmentRequest{
			TopicCouncilCode: resources.topicCouncilId,
			StudentCode:      idStudent,
			CreatedBy:        *myId,
			Title:            fmt.Sprintf("Enrollment for %s", student),
			MidtermCode:      &midtermResp.Midterm.Id,
			FinalCode:        &finalResp.Final.Id,
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
	// check defence have member

	defence, err := c.council.GetDefenceById(ctx, input.DefenceCode)
	if err != nil {
		return nil, err
	}
	if defence.GetDefence().GetTeacherCode() != *myId {
		return nil, fmt.Errorf("you are not a member of this defence")
	}

	// check enrollment is of defence topic council
	enrollment, err := c.thesis.GetEnrollmentById(ctx, input.EnrollmentCode)
	if err != nil {
		return nil, err
	}
	// get topic council
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, enrollment.GetEnrollment().GetTopicCouncilCode())
	if err != nil {
		return nil, err
	}
	// check defence is council code
	if defence.GetDefence().GetCouncilCode() != topicCouncil.GetTopicCouncil().GetCouncilCode() {
		return nil, fmt.Errorf("defence is not of this topic council")
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

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateGradeDefence(input.DefenceCode, input.EnrollmentCode)
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
	// check member
	gradeDefence, err := c.council.GetGradeById(ctx, id)
	if err != nil {
		return nil, err
	}
	if gradeDefence.GetGradeDefence().GetCreatedBy() != *myId {
		return nil, fmt.Errorf("you are not a member of this defence")
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

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && resp.GetGradeDefence() != nil {
		gd := resp.GetGradeDefence()
		loaders.InvalidateGradeDefence(gd.GetDefenceCode(), gd.GetEnrollmentCode())
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

	// check teacher is member of defence
	defenceGrade, err := c.council.GetGradeById(ctx, input.GradeDefenceCode)
	if err != nil {
		return nil, err
	}
	// check defence is member of defence
	if defenceGrade.GetGradeDefence().GetCreatedBy() != *myId {
		return nil, fmt.Errorf("you are not a member of this defence")
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

	// Invalidate cache - criteria are cached by defence ID
	if loaders := dataloader.GetLoaders(ctx); loaders != nil {
		loaders.InvalidateGradeDefence(input.GradeDefenceCode, "")
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
	// check teacher is member of defence by created by
	gradeCriterion, err := c.council.GetGradeCriterionById(ctx, id)
	if err != nil {
		return nil, err
	}
	if gradeCriterion.GetGradeDefenceCriterion().GetCreatedBy() != *myId {
		return nil, fmt.Errorf("you are not the creator of this criterion")
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

	// Invalidate cache - criteria are cached by defence ID
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && resp.GetGradeDefenceCriterion() != nil {
		loaders.InvalidateGradeDefence(resp.GetGradeDefenceCriterion().GetGradeDefenceCode(), "")
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

	// Get criterion before delete to know the grade defence ID
	criterion, _ := c.council.GetGradeCriterionById(ctx, id)

	_, err = c.council.DeleteGradeDefenceCriterion(ctx, id)
	if err != nil {
		return false, err
	}

	// Invalidate cache
	if loaders := dataloader.GetLoaders(ctx); loaders != nil && criterion != nil && criterion.GetGradeDefenceCriterion() != nil {
		loaders.InvalidateGradeDefence(criterion.GetGradeDefenceCriterion().GetGradeDefenceCode(), "")
	}

	return true, nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// verifySupervisor checks if teacher is supervisor of a topic council
// func (c *Controller) verifySupervisor(ctx context.Context, teacherId string, topicCouncilId string) bool {
// 	search := &pb.SearchRequest{
// 		Pagination: &pb.Pagination{Page: 1, PageSize: 10},
// 		Filters: []*pb.FilterCriteria{
// 			{
// 				Criteria: &pb.FilterCriteria_Condition{
// 					Condition: &pb.FilterCondition{
// 						Field:    "topic_council_code",
// 						Operator: pb.FilterOperator_EQUAL,
// 						Values:   []string{topicCouncilId},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	supervisors, err := c.thesis.GetTopicCouncilSupervisorBySearch(ctx, search)
// 	if err != nil {
// 		return false
// 	}

// 	for _, s := range supervisors.GetTopicCouncilSupervisors() {
// 		if s.GetTeacherSupervisorCode() == teacherId {
// 			return true
// 		}
// 	}

// 	return false
// }

// func (c *Controller) verifyTopic(ctx context.Context, topicId string) bool {
// 	topic, err := c.thesis.GetTopicById(ctx, topicId)
// 	if err != nil {
// 		return false
// 	}
// 	if topic.GetTopic().GetStatus() != pbThesis.TopicStatus_IN_PROGRESS {
// 		return false
// 	}
// 	return true
// }

// func (c *Controller) verifyDefenceMember(ctx context.Context, teacherId string, defenceId string) bool {
// 	search := &pb.SearchRequest{
// 		Pagination: &pb.Pagination{Page: 1, PageSize: 10},
// 		Filters: []*pb.FilterCriteria{
// 			{
// 				Criteria: &pb.FilterCriteria_Condition{
// 					Condition: &pb.FilterCondition{
// 						Field:    "defence_code",
// 						Operator: pb.FilterOperator_EQUAL,
// 						Values:   []string{defenceId},
// 					},
// 				},
// 			},
// 			{
// 				Criteria: &pb.FilterCriteria_Condition{
// 					Condition: &pb.FilterCondition{
// 						Field:    "teacher_code",
// 						Operator: pb.FilterOperator_EQUAL,
// 						Values:   []string{teacherId},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	defences, err := c.council.GetDefencesBySearch(ctx, search)
// 	if err != nil {
// 		return false
// 	}
// 	if len(defences.GetDefences()) == 0 {
// 		return false
// 	}
// 	return true
// }
