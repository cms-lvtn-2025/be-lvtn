package controller

import (
	"context"
	pb "thaily/proto/common"
	"thaily/src/graph/convert"
	"thaily/src/graph/model"
)

// Teacher Profile

func (c *Controller) GetTeacherByRequest(ctx context.Context) (*model.Teacher, error) {
	myId, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if myId == nil || role == nil {
		return nil, nil
	}
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}
	return convert.PbTeacherToModel(teacher.GetTeacher()), nil
}

// GetTeacherById, GetStudentById - moved to resolvers_helper.go

// Supervisor Methods

func (c *Controller) GetSupervisedTopicCouncils(ctx context.Context, search *model.SearchRequestInput) (*model.SupervisorTopicCouncilAssignmentListResponse, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	var newSearch model.SearchRequestInput
	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters: append([]*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "teacher_supervisor_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{*myId},
					},
				},
			}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
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
	}

	assignments, err := c.thesis.GetTopicCouncilSupervisorBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.SupervisorTopicCouncilAssignmentListResponse{
		Total: assignments.GetTotal(),
		Data:  convert.PbTopicCouncilSupervisorsToSupervisorTopicCouncilAssignments(assignments.GetTopicCouncilSupervisors()),
	}, nil
}

func (c *Controller) GetSupervisedTopicCouncilDetail(ctx context.Context, id string) (*model.SupervisorTopicCouncilAssignment, error) {
	_, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	assignment, err := c.thesis.GetTopicCouncilSupervisorById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicCouncilSupervisorToSupervisorTopicCouncilAssignment(assignment.GetTopicCouncilSupervisor()), nil
}

func (c *Controller) GetSupervisorTopicCouncilById(ctx context.Context, id string) (*model.SupervisorTopicCouncil, error) {
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTopicCouncilToSupervisorTopicCouncil(topicCouncil.GetTopicCouncil()), nil
}

func (c *Controller) GetSupervisorTopicById(ctx context.Context, id string) (*model.SupervisorTopic, error) {
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTopicToSupervisorTopic(topic.GetTopic()), nil
}

func (c *Controller) GetSupervisorEnrollmentsByTopicCouncilId(ctx context.Context, topicCouncilId string) ([]*model.SupervisorEnrollment, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
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

	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbEnrollmentsToSupervisorEnrollments(enrollments.GetEnrollments()), nil
}

func (c *Controller) GetSupervisorTopicCouncilsByTopicId(ctx context.Context, topicId string) ([]*model.SupervisorTopicCouncil, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "topic_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{topicId},
					},
				},
			},
		},
	}

	topicCouncils, err := c.thesis.GetTopicCouncilBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicCouncilsToSupervisorTopicCouncils(topicCouncils.GetTopicCouncils()), nil
}

// Council Member Methods

func (c *Controller) GetMyDefences(ctx context.Context, search *model.SearchRequestInput) (*model.CouncilDefenceListResponse, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	var newSearch model.SearchRequestInput
	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters: append([]*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "teacher_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{*myId},
					},
				},
			}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters: []*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "teacher_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{*myId},
					},
				},
			},
		}
	}

	defences, err := c.council.GetDefencesBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.CouncilDefenceListResponse{
		Total: defences.GetTotal(),
		Data:  convert.PbDefencesToCouncilDefences(defences.GetDefences()),
	}, nil
}

func (c *Controller) GetMyDefenceDetail(ctx context.Context, id string) (*model.CouncilDefence, error) {
	_, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	defence, err := c.council.GetDefenceById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert.PbDefenceToCouncilDefence(defence.GetDefence()), nil
}

func (c *Controller) GetCouncilMemberById(ctx context.Context, id string) (*model.CouncilMemberCouncil, error) {
	council, err := c.council.GetCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbCouncilToCouncilMemberCouncil(council.GetCouncil()), nil
}

func (c *Controller) GetCouncilTopicCouncilById(ctx context.Context, id string) (*model.CouncilTopicCouncil, error) {
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTopicCouncilToCouncilTopicCouncil(topicCouncil.GetTopicCouncil()), nil
}

// GetTopicById, GetCouncilById - moved to resolvers_helper.go

func (c *Controller) GetCouncilEnrollmentsByTopicCouncilId(ctx context.Context, topicCouncilId string) ([]*model.CouncilEnrollment, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
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

	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbEnrollmentsToCouncilEnrollments(enrollments.GetEnrollments()), nil
}

func (c *Controller) GetCouncilDefencesByCouncilId(ctx context.Context, councilId string) ([]*model.CouncilDefence, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "council_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{councilId},
					},
				},
			},
		},
	}

	defences, err := c.council.GetDefencesBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbDefencesToCouncilDefences(defences.GetDefences()), nil
}

func (c *Controller) GetCouncilTopicCouncilsByCouncilId(ctx context.Context, councilId string) ([]*model.CouncilTopicCouncil, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "council_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{councilId},
					},
				},
			},
		},
	}

	topicCouncils, err := c.thesis.GetTopicCouncilBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicCouncilsToCouncilTopicCouncils(topicCouncils.GetTopicCouncils()), nil
}

func (c *Controller) GetCouncilGradeDefencesByDefenceId(ctx context.Context, defenceId string) ([]*model.GradeDefence, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "defence_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{defenceId},
					},
				},
			},
		},
	}

	gradeDefences, err := c.council.GetGradeDefenceBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefencesToModel(gradeDefences.GetGradeDefences()), nil
}

func (c *Controller) GetCouncilGradeDefencesByEnrollmentId(ctx context.Context, enrollmentId string) ([]*model.GradeDefence, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "enrollment_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{enrollmentId},
					},
				},
			},
		},
	}

	gradeDefences, err := c.council.GetGradeDefenceBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefencesToModel(gradeDefences.GetGradeDefences()), nil
}

func (c *Controller) GetSupervisorTopicCouncilSupervisorsByTopicCouncilId(ctx context.Context, topicCouncilId string) ([]*model.TopicCouncilSupervisor, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   100,
			Descending: true,
			SortBy:     "created_at",
		},
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

	supervisors, err := c.thesis.GetTopicCouncilSupervisorBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicCouncilSupervisorsToModel(supervisors.GetTopicCouncilSupervisors()), nil
}

// Reviewer Methods

func (c *Controller) GetMyGradeReviews(ctx context.Context, search *model.SearchRequestInput) (*model.ReviewerGradeReviewListResponse, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	var newSearch model.SearchRequestInput
	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters: append([]*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "teacher_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{*myId},
					},
				},
			}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters: []*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "teacher_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{*myId},
					},
				},
			},
		}
	}

	gradeReviews, err := c.thesis.GetGradeReviewBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.ReviewerGradeReviewListResponse{
		Total: gradeReviews.GetTotal(),
		Data:  convert.PbGradeReviewsToReviewerGradeReviews(gradeReviews.GetGradeReviews()),
	}, nil
}

func (c *Controller) GetMyGradeReviewDetail(ctx context.Context, id string) (*model.ReviewerGradeReview, error) {
	_, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	gradeReview, err := c.thesis.GetGradeReviewById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert.PbGradeReviewToReviewerGradeReview(gradeReview.GetGradeReview()), nil
}

func (c *Controller) GetReviewerEnrollmentByGradeReviewId(ctx context.Context, gradeReviewId string) (*model.ReviewerEnrollment, error) {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:       1,
			PageSize:   1,
			Descending: true,
			SortBy:     "created_at",
		},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "grade_review_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{gradeReviewId},
					},
				},
			},
		},
	}

	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, &newSearch)
	if err != nil {
		return nil, err
	}

	if len(enrollments.GetEnrollments()) == 0 {
		return nil, nil
	}

	return convert.PbEnrollmentToReviewerEnrollment(enrollments.GetEnrollments()[0]), nil
}

func (c *Controller) GetReviewerTopicCouncilById(ctx context.Context, id string) (*model.ReviewerTopicCouncil, error) {
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTopicCouncilToReviewerTopicCouncil(topicCouncil.GetTopicCouncil()), nil
}

func (c *Controller) GetReviewerTopicById(ctx context.Context, id string) (*model.ReviewerTopic, error) {
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTopicToReviewerTopic(topic.GetTopic()), nil
}

// GetFilesByTopicId - moved to resolvers_helper.go
