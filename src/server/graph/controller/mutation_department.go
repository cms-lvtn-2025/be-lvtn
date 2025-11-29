package controller

import (
	"context"
	"fmt"
	"slices"

	pbCouncil "thaily/proto/council"
	pbThesis "thaily/proto/thesis"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// DEPARTMENT MUTATION CONTROLLER
// Handles: DepartmentMutation resolvers (Department Lecturer)
// Council management and topic approval stage 1
// ============================================

// ============================================
// COUNCIL MANAGEMENT MUTATIONS
// ============================================

// CreateCouncil creates a new council in department
func (c *Controller) CreateCouncil(ctx context.Context, input model.CreateCouncilInput) (*model.Council, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires department lecturer role")
	}
	teacher, err := c.user.GetTeacherById(ctx, *userInfo)
	if err != nil {
		return nil, err
	}
	if teacher == nil || teacher.GetTeacher() == nil {
		return nil, fmt.Errorf("teacher not found")
	}

	resp, err := c.council.CreateCouncil(ctx, &pbCouncil.CreateCouncilRequest{
		Title:        input.Title,
		MajorCode:    input.MajorCode,
		SemesterCode: teacher.GetTeacher().SemesterCode,
		CreatedBy:    *userInfo,
	})
	if err != nil {
		return nil, err
	}

	return convert.PbCouncilToModel(resp.GetCouncil()), nil
}

// UpdateDepartmentCouncil updates a council (only if not approved yet)
func (c *Controller) UpdateDepartmentCouncil(ctx context.Context, id string, input model.UpdateCouncilInput) (*model.Council, error) {
	majorCodes, semesterCode, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}
	if semesterCode == nil {
		return nil, fmt.Errorf("no semester code found")
	}

	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check || err != nil {
		return nil, fmt.Errorf("not authorized: requires department lecturer role")
	}
	// get one and check
	council, err := c.council.GetCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return nil, fmt.Errorf("cannot modify: council already approved")
	}
	if slices.Contains(majorCodes, council.GetCouncil().MajorCode) {
		return nil, fmt.Errorf("cannot modify: council is not in the same major")
	}
	if council.GetCouncil().SemesterCode != *semesterCode {
		return nil, fmt.Errorf("cannot modify: council is not in the same semester")
	}

	updatedBy := *myId
	// update council
	req := &pbCouncil.UpdateCouncilRequest{
		Id:        id,
		Title:     input.Title,
		UpdatedBy: updatedBy,
	}

	updateCouncil, err := c.council.UpdateCouncil(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbCouncilToModel(updateCouncil.GetCouncil()), nil
}

// AddDefenceToCouncil adds a defence member to council
// check member is major in semester
// check defence is not already in council
func (c *Controller) AddDefenceToCouncil(ctx context.Context, input model.CreateDefenceInput) (*model.Defence, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires department lecturer role")
	}
	majorCodes, semesterCode, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}
	if semesterCode == nil {
		return nil, fmt.Errorf("no semester code found")
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}

	// Check if council is already approved
	council, err := c.council.GetCouncilById(ctx, input.CouncilCode)
	if err != nil {
		return nil, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return nil, fmt.Errorf("cannot modify: council already approved")
	}

	// Convert DefencePosition enum
	var position pbCouncil.DefencePosition
	switch input.Position {
	case model.DefencePositionPresident:
		position = pbCouncil.DefencePosition_PRESIDENT
	case model.DefencePositionSecretary:
		position = pbCouncil.DefencePosition_SECRETARY
	case model.DefencePositionReviewer:
		position = pbCouncil.DefencePosition_REVIEWER
	case model.DefencePositionMember:
		position = pbCouncil.DefencePosition_MEMBER
	default:
		position = pbCouncil.DefencePosition_MEMBER
	}
	// get teacher by teacher code
	teacher, err := c.user.GetTeacherById(ctx, input.TeacherCode)
	if err != nil {
		return nil, err
	}
	// check semester and major
	if teacher.GetTeacher().SemesterCode != *semesterCode {
		return nil, fmt.Errorf("teacher is not in the same semester")
	}
	if !slices.Contains(majorCodes, teacher.GetTeacher().MajorCode) {
		return nil, fmt.Errorf("teacher is not in the same major")
	}

	// check semester and major
	if council.GetCouncil().SemesterCode != *semesterCode {
		return nil, fmt.Errorf("council is not in the same semester")
	}
	if !slices.Contains(majorCodes, council.GetCouncil().MajorCode) {
		return nil, fmt.Errorf("council is not in the same major")
	}

	resp, err := c.council.CreateDefence(ctx, &pbCouncil.CreateDefenceRequest{
		Title:       input.Title,
		CouncilCode: input.CouncilCode,
		TeacherCode: input.TeacherCode,
		Position:    position,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, err
	}

	return convert.PbDefenceToModel(resp.GetDefence()), nil
}

// RemoveDefenceFromCouncil removes a defence member from council
func (c *Controller) RemoveDefenceFromCouncil(ctx context.Context, id string) (bool, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires department lecturer role")
	}
	majorCodes, semesterCode, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return false, err
	}
	if semesterCode == nil {
		return false, fmt.Errorf("no semester code found")
	}

	// Get defence to find council
	defence, err := c.council.GetDefenceById(ctx, id)
	if err != nil {
		return false, err
	}

	// Check if council is already approved

	council, err := c.council.GetCouncilById(ctx, defence.GetDefence().GetCouncilCode())
	if err != nil {
		return false, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return false, fmt.Errorf("cannot modify: council already approved")
	}

	// check semester and major
	if council.GetCouncil().SemesterCode != *semesterCode {
		return false, fmt.Errorf("council is not in the same semester")
	}
	if !slices.Contains(majorCodes, council.GetCouncil().MajorCode) {
		return false, fmt.Errorf("council is not in the same major")
	}

	_, err = c.council.DeleteDefence(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ============================================
// TOPIC MANAGEMENT MUTATIONS
// ============================================

// ApproveTopicStage1 approves topic stage 1
func (c *Controller) ApproveTopicStage1(ctx context.Context, id string) (*model.Topic, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires department lecturer role")
	}
	majorCodes, semesterCode, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	if topic.GetTopic().SemesterCode != *semesterCode {
		return nil, fmt.Errorf("topic is not in the same semester")
	}
	if !slices.Contains(majorCodes, topic.GetTopic().MajorCode) {
		return nil, fmt.Errorf("topic is not in the same major")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	status := pbThesis.TopicStatus_APPROVED_1
	req := &pbThesis.UpdateTopicRequest{
		Id:        id,
		Status:    &status,
		UpdatedBy: updatedBy,
	}

	resp, err := c.thesis.UpdateTopic(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicToModel(resp.GetTopic()), nil
}

// RejectTopicStage1 rejects topic stage 1
func (c *Controller) RejectTopicStage1(ctx context.Context, id string, reason *string) (*model.Topic, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires department lecturer role")
	}
	majorCodes, semesterCode, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}
	topic, err := c.thesis.GetTopicById(ctx, id)

	if err != nil {
		return nil, err
	}
	if topic.GetTopic().SemesterCode != *semesterCode {
		return nil, fmt.Errorf("topic is not in the same semester")
	}

	if !slices.Contains(majorCodes, topic.GetTopic().MajorCode) {
		return nil, fmt.Errorf("topic is not in the same major")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	status := pbThesis.TopicStatus_REJECTED
	req := &pbThesis.UpdateTopicRequest{
		Id:        id,
		Status:    &status,
		UpdatedBy: updatedBy,
	}

	resp, err := c.thesis.UpdateTopic(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicToModel(resp.GetTopic()), nil
}

// AssignTopicToCouncil assigns a topic to council
func (c *Controller) AssignTopicToCouncil(ctx context.Context, topicCouncilID string, councilID string) (*model.TopicCouncil, error) {
	majorCodes, semesterCode, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}
	if semesterCode == nil {
		return nil, fmt.Errorf("no semester code found")
	}

	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized: requires department lecturer role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	// get topic council
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, topicCouncilID)
	if err != nil {
		return nil, err
	}
	// get topic by topic code
	topic, err := c.thesis.GetTopicById(ctx, topicCouncil.GetTopicCouncil().GetTopicCode())
	if err != nil {
		return nil, err
	}
	if topic.GetTopic().SemesterCode != *semesterCode {
		return nil, fmt.Errorf("topic is not in the same semester")
	}
	if !slices.Contains(majorCodes, topic.GetTopic().MajorCode) {
		return nil, fmt.Errorf("topic is not in the same major")
	}
	// get council and check

	// Check if council is already approved
	council, err := c.council.GetCouncilById(ctx, councilID)
	if err != nil {
		return nil, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return nil, fmt.Errorf("cannot assign: council already approved")
	}
	// check semester and major
	if council.GetCouncil().SemesterCode != *semesterCode {
		return nil, fmt.Errorf("council is not in the same semester")
	}
	if !slices.Contains(majorCodes, council.GetCouncil().MajorCode) {
		return nil, fmt.Errorf("council is not in the same major")
	}

	req := &pbThesis.UpdateTopicCouncilRequest{
		Id:          topicCouncilID,
		CouncilCode: &councilID,
		UpdatedBy:   updatedBy,
	}

	resp, err := c.thesis.UpdateTopicCouncil(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicCouncilToModel(resp.GetTopicCouncil()), nil
}

// RemoveTopicFromCouncil removes a topic from council
func (c *Controller) RemoveTopicFromCouncil(ctx context.Context, topicCouncilID string, councilID string) (bool, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return false, err
	}
	if !check {
		return false, fmt.Errorf("not authorized: requires department lecturer role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	// Check if council is already approved
	council, err := c.council.GetCouncilById(ctx, councilID)
	if err != nil {
		return false, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return false, fmt.Errorf("cannot remove: council already approved")
	}

	majorCodes, semesterCode, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return false, err
	}
	if semesterCode == nil {
		return false, fmt.Errorf("no semester code found")
	}
	if !slices.Contains(majorCodes, council.GetCouncil().MajorCode) {
		return false, fmt.Errorf("council is not in the same major")
	}
	// Verify topic council belongs to this council
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, topicCouncilID)
	if err != nil {
		return false, err
	}
	if topicCouncil.GetTopicCouncil().GetCouncilCode() != councilID {
		return false, fmt.Errorf("topic council does not belong to this council")
	}
	// get topic by topic code
	topic, err := c.thesis.GetTopicById(ctx, topicCouncil.GetTopicCouncil().GetTopicCode())
	if err != nil {
		return false, err
	}
	if topic.GetTopic().SemesterCode != *semesterCode {
		return false, fmt.Errorf("topic is not in the same semester")
	}
	if !slices.Contains(majorCodes, topic.GetTopic().MajorCode) {
		return false, fmt.Errorf("topic is not in the same major")
	}

	// Remove by setting council_code to "remove"
	removeString := "remove"
	req := &pbThesis.UpdateTopicCouncilRequest{
		Id:          topicCouncilID,
		CouncilCode: &removeString,
		UpdatedBy:   updatedBy,
	}

	_, err = c.thesis.UpdateTopicCouncil(ctx, req)
	if err != nil {
		return false, err
	}

	return true, nil
}
