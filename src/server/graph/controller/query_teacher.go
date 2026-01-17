package controller

import (
	"context"

	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// TEACHER QUERY CONTROLLER
// Handles: TeacherQuery, SupervisorQuery, CouncilMemberQuery, ReviewerQuery
// ============================================

// GetTeacherByRequest returns current logged-in teacher profile
func (c *Controller) GetTeacherByRequest(ctx context.Context) (*model.Teacher, error) {
	myId, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if myId == nil || role == nil || *role != "teacher" {
		return nil, nil
	}

	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}
	return convert.PbTeacherToModel(teacher.GetTeacher()), nil
}

// ============================================
// SUPERVISOR QUERY METHODS
// ============================================

// GetSupervisedTopicCouncils returns topic councils supervised by current teacher
func (c *Controller) GetSupervisedTopicCouncils(ctx context.Context, search *model.SearchRequestInput) (*model.TopicCouncilListResponse, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	// First get topic council IDs where teacher is supervisor
	supervisorSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "teacher_supervisor_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{*myId},
				},
			},
		},
	}

	supervisors, err := c.thesis.GetTopicCouncilSupervisorBySearch(ctx, convert.ConvertSearchRequestToPB(supervisorSearch))
	if err != nil {
		return nil, err
	}

	// Extract topic council IDs
	topicCouncilIds := make([]string, 0)
	for _, s := range supervisors.GetTopicCouncilSupervisors() {
		topicCouncilIds = append(topicCouncilIds, s.GetTopicCouncilCode())
	}

	if len(topicCouncilIds) == 0 {
		return &model.TopicCouncilListResponse{Total: 0, Data: []*model.TopicCouncil{}}, nil
	}

	// Get topic councils by IDs
	var newSearch model.SearchRequestInput
	tcFilter := &model.FilterCriteriaInput{
		Condition: &model.FilterConditionInput{
			Field:    "id",
			Operator: model.FilterOperatorIn,
			Values:   topicCouncilIds,
		},
	}

	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters:    append([]*model.FilterCriteriaInput{tcFilter}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters:    []*model.FilterCriteriaInput{tcFilter},
		}
	}

	topicCouncils, err := c.thesis.GetTopicCouncilBySearch(ctx, convert.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.TopicCouncilListResponse{
		Total: topicCouncils.GetTotal(),
		Data:  convert.PbTopicCouncilsToModel(topicCouncils.GetTopicCouncils()),
	}, nil
}

// ============================================
// COUNCIL MEMBER QUERY METHODS
// ============================================

// GetMyDefences returns defences where current teacher is council member
func (c *Controller) GetMyDefences(ctx context.Context, search *model.SearchRequestInput) (*model.DefenceListResponse, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	var newSearch model.SearchRequestInput
	teacherFilter := &model.FilterCriteriaInput{
		Condition: &model.FilterConditionInput{
			Field:    "teacher_code",
			Operator: model.FilterOperatorEqual,
			Values:   []string{*myId},
		},
	}

	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters:    append([]*model.FilterCriteriaInput{teacherFilter}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters:    []*model.FilterCriteriaInput{teacherFilter},
		}
	}

	defences, err := c.council.GetDefencesBySearch(ctx, convert.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.DefenceListResponse{
		Total: defences.GetTotal(),
		Data:  convert.PbDefencesToModel(defences.GetDefences()),
	}, nil
}
