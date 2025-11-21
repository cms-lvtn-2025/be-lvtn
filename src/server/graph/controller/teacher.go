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
	convert2 "thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized")
	}
	if myId == nil {
		return nil, fmt.Errorf("not authorized")
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
		Data:  convert2.PbTopicCouncilSupervisorsToSupervisorTopicCouncilAssignments(assignments.GetTopicCouncilSupervisors()),
	}, nil
}

func (c *Controller) GetSupervisedTopicCouncilDetail(ctx context.Context, id string) (*model.SupervisorTopicCouncilAssignment, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized")
	}
	if myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	assignment, err := c.thesis.GetTopicCouncilSupervisorById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicCouncilSupervisorToSupervisorTopicCouncilAssignment(assignment.GetTopicCouncilSupervisor()), nil
}

func (c *Controller) GetSupervisorTopicCouncilById(ctx context.Context, id string) (*model.SupervisorTopicCouncil, error) {
	_, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicCouncilToSupervisorTopicCouncil(topicCouncil.GetTopicCouncil()), nil
}

func (c *Controller) GetSupervisorTopicById(ctx context.Context, id string) (*model.SupervisorTopic, error) {
	// role teacher
	_, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if role == nil || (*role) != "teacher" {
		return nil, fmt.Errorf("Not teacher")
	}
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicToSupervisorTopic(topic.GetTopic()), nil
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

	return convert2.PbEnrollmentsToSupervisorEnrollments(enrollments.GetEnrollments()), nil
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

	return convert2.PbTopicCouncilsToSupervisorTopicCouncils(topicCouncils.GetTopicCouncils()), nil
}

// Council Member Methods

func (c *Controller) GetMyDefences(ctx context.Context, search *model.SearchRequestInput) (*model.CouncilDefenceListResponse, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized")
	}
	if myId == nil {
		return nil, fmt.Errorf("not authorized")
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
	return convert2.PbTopicCouncilToCouncilTopicCouncil(topicCouncil.GetTopicCouncil()), nil
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

	return convert2.PbEnrollmentsToCouncilEnrollments(enrollments.GetEnrollments()), nil
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

	return convert2.PbTopicCouncilsToCouncilTopicCouncils(topicCouncils.GetTopicCouncils()), nil
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

	return convert2.PbTopicCouncilSupervisorsToModel(supervisors.GetTopicCouncilSupervisors()), nil
}

// Reviewer Methods

func (c *Controller) GetMyGradeReviews(ctx context.Context, search *model.SearchRequestInput) (*model.ReviewerGradeReviewListResponse, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not authorized")
	}
	if myId == nil {
		return nil, fmt.Errorf("not authorized")
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
		Data:  convert2.PbGradeReviewsToReviewerGradeReviews(gradeReviews.GetGradeReviews()),
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

	return convert2.PbGradeReviewToReviewerGradeReview(gradeReview.GetGradeReview()), nil
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

	return convert2.PbEnrollmentToReviewerEnrollment(enrollments.GetEnrollments()[0]), nil
}

func (c *Controller) GetReviewerTopicCouncilById(ctx context.Context, id string) (*model.ReviewerTopicCouncil, error) {
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicCouncilToReviewerTopicCouncil(topicCouncil.GetTopicCouncil()), nil
}

func (c *Controller) GetReviewerTopicById(ctx context.Context, id string) (*model.ReviewerTopic, error) {
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicToReviewerTopic(topic.GetTopic()), nil
}

// GetFilesByTopicId - moved to resolvers_helper.go

// ============================================
// MUTATION METHODS
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
// SUPERVISOR MUTATIONS - Midterm & Final Grading
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

	// Verify teacher is supervisor of this enrollment's topic
	enrollment, err := c.thesis.GetEnrollmentById(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}

	// Check if teacher is supervisor of the topic_council
	topicCouncilId := enrollment.GetEnrollment().GetTopicCouncilCode()
	if !c.verifySupervisor(ctx, *myId, topicCouncilId) {
		return nil, fmt.Errorf("not authorized: not supervisor of this topic")
	}

	// Get existing midterm from enrollment
	midtermCode := enrollment.GetEnrollment().GetMidtermCode()
	if midtermCode == "" {
		return nil, fmt.Errorf("no midterm found for this enrollment")
	}

	// Convert status
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

	return convert2.PbMidtermToModel(resp.GetMidterm()), nil
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

	// Get midterm to verify ownership
	midterm, err := c.thesis.GetMidtermById(ctx, midtermID)
	if err != nil {
		return nil, err
	}

	// Verify supervisor - search for enrollment by midterm_code
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

	return convert2.PbMidtermToModel(resp.GetMidterm()), nil
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

	// Verify teacher is supervisor
	enrollment, err := c.thesis.GetEnrollmentById(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}

	if !c.verifySupervisor(ctx, *myId, enrollment.GetEnrollment().GetTopicCouncilCode()) {
		return nil, fmt.Errorf("not authorized")
	}

	// Get existing final from enrollment
	finalCode := enrollment.GetEnrollment().GetFinalCode()
	if finalCode == "" {
		return nil, fmt.Errorf("no final found for this enrollment")
	}

	// Convert status
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

	return convert2.PbFinalToModel(resp.GetFinal()), nil
}

// FeedbackFinal adds feedback to final
func (c *Controller) FeedbackFinal(ctx context.Context, finalID string, notes string) (*model.Final, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	// Get final to verify ownership
	final, err := c.thesis.GetFinalById(ctx, finalID)
	if err != nil {
		return nil, err
	}

	// Verify supervisor - search for enrollment by final_code
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

	return convert2.PbFinalToModel(resp.GetFinal()), nil
}

// ApproveMidtermFile approves a midterm file
func (c *Controller) ApproveMidtermFile(ctx context.Context, fileID string) (*model.File, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	// TODO: Verify supervisor owns the file's topic
	// For now, just update the file status

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

	// TODO: Store reason somewhere if needed

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
// COUNCIL MEMBER MUTATIONS - Grade Defence
// ============================================

// CreateGradeDefence creates a grade defence (council member only)
func (c *Controller) CreateGradeDefence(ctx context.Context, input model.CreateGradeDefenceInput) (*model.GradeDefence, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	// Verify teacher is member of the defence
	defence, err := c.council.GetDefenceById(ctx, input.DefenceCode)
	if err != nil {
		return nil, err
	}

	if defence.GetDefence().GetTeacherCode() != *myId {
		return nil, fmt.Errorf("not authorized: not member of this defence")
	}

	req := &pbCouncil.CreateGradeDefenceRequest{
		DefenceCode:    input.DefenceCode,
		EnrollmentCode: input.EnrollmentCode,
		CreatedBy:      *myId,
	}

	if input.TotalScore != nil {
		score := int32(*input.TotalScore)
		req.TotalScore = &score
	}

	if input.Note != nil {
		req.Note = input.Note
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

	// Verify ownership through defence
	gradeDefence, err := c.council.GetGradeById(ctx, id)
	if err != nil {
		return nil, err
	}

	defence, err := c.council.GetDefenceById(ctx, gradeDefence.GetGradeDefence().GetDefenceCode())
	if err != nil {
		return nil, err
	}

	if defence.GetDefence().GetTeacherCode() != *myId {
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
		score := int32(*input.TotalScore)
		req.TotalScore = &score
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

	// Verify ownership
	gradeDefence, err := c.council.GetGradeById(ctx, input.GradeDefenceCode)
	if err != nil {
		return nil, err
	}

	defence, err := c.council.GetDefenceById(ctx, gradeDefence.GetGradeDefence().GetDefenceCode())
	if err != nil {
		return nil, err
	}

	if defence.GetDefence().GetTeacherCode() != *myId {
		return nil, fmt.Errorf("not authorized")
	}

	req := &pbCouncil.CreateGradeDefenceCriterionRequest{
		GradeDefenceCode: input.GradeDefenceCode,
		Name:             &input.Name,
		Score:            &input.Score,
		MaxScore:         &input.MaxScore,
		CreatedBy:        myId,
	}

	resp, err := c.council.CreateGradeDefenceCriterion(ctx, req)
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

	// Verify ownership
	criterion, err := c.council.GetGradeCriterionById(ctx, id)
	if err != nil {
		return nil, err
	}

	gradeDefence, err := c.council.GetGradeById(ctx, criterion.GetGradeDefenceCriterion().GetGradeDefenceCode())
	if err != nil {
		return nil, err
	}

	defence, err := c.council.GetDefenceById(ctx, gradeDefence.GetGradeDefence().GetDefenceCode())
	if err != nil {
		return nil, err
	}

	if defence.GetDefence().GetTeacherCode() != *myId {
		return nil, fmt.Errorf("not authorized")
	}

	req := &pbCouncil.UpdateGradeDefenceCriterionRequest{
		Id:        id,
		UpdatedBy: myId,
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

	// Verify ownership
	criterion, err := c.council.GetGradeCriterionById(ctx, id)
	if err != nil {
		return false, err
	}

	gradeDefence, err := c.council.GetGradeById(ctx, criterion.GetGradeDefenceCriterion().GetGradeDefenceCode())
	if err != nil {
		return false, err
	}

	defence, err := c.council.GetDefenceById(ctx, gradeDefence.GetGradeDefence().GetDefenceCode())
	if err != nil {
		return false, err
	}

	if defence.GetDefence().GetTeacherCode() != *myId {
		return false, fmt.Errorf("not authorized")
	}

	_, err = c.council.DeleteGradeDefenceCriterion(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ============================================
// REVIEWER MUTATIONS - Grade Review
// ============================================

// UpdateGradeReview updates a grade review (reviewer only)
func (c *Controller) UpdateGradeReview(ctx context.Context, id string, input model.UpdateGradeReviewInput) (*model.ReviewerGradeReview, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	// Verify teacher is the reviewer
	gradeReview, err := c.thesis.GetGradeReviewById(ctx, id)
	if err != nil {
		return nil, err
	}

	if gradeReview.GetGradeReview().GetTeacherCode() != *myId {
		return nil, fmt.Errorf("not authorized: not reviewer of this grade review")
	}

	req := &pbThesis.UpdateGradeReviewRequest{
		Id:        id,
		UpdatedBy: *myId,
	}

	if input.ReviewGrade != nil {
		grade := int32(*input.ReviewGrade)
		req.ReviewGrade = &grade
	}
	if input.Status != nil {
		var status pbThesis.FinalStatus
		switch *input.Status {
		case model.FinalStatusPending:
			status = pbThesis.FinalStatus_PENDING
		case model.FinalStatusPassed:
			status = pbThesis.FinalStatus_PASSED
		case model.FinalStatusFailed:
			status = pbThesis.FinalStatus_FAILED
		case model.FinalStatusCompleted:
			status = pbThesis.FinalStatus_COMPLETED
		}
		req.Status = &status
	}
	if input.Notes != nil {
		req.Notes = input.Notes
	}

	resp, err := c.thesis.UpdateGradeReview(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert2.PbGradeReviewToReviewerGradeReview(resp.GetGradeReview()), nil
}

// CompleteGradeReview marks grade review as completed
func (c *Controller) CompleteGradeReview(ctx context.Context, id string) (*model.ReviewerGradeReview, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleTeacher)
	if err != nil {
		return nil, err
	}
	if !check || myId == nil {
		return nil, fmt.Errorf("not authorized")
	}

	// Verify teacher is the reviewer
	gradeReview, err := c.thesis.GetGradeReviewById(ctx, id)
	if err != nil {
		return nil, err
	}

	if gradeReview.GetGradeReview().GetTeacherCode() != *myId {
		return nil, fmt.Errorf("not authorized")
	}

	// Update status to completed and set completion date
	status := pbThesis.FinalStatus_COMPLETED
	now := timestamppb.New(time.Now())
	req := &pbThesis.UpdateGradeReviewRequest{
		Id:             id,
		Status:         &status,
		CompletionDate: now,
		UpdatedBy:      *myId,
	}

	resp, err := c.thesis.UpdateGradeReview(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert2.PbGradeReviewToReviewerGradeReview(resp.GetGradeReview()), nil
}

// ============================================
// HELPER METHODS
// ============================================

// verifySupervisor checks if teacher is supervisor of a topic council
func (c *Controller) verifySupervisor(ctx context.Context, teacherId string, topicCouncilId string) bool {
	newSearch := pb.SearchRequest{
		Pagination: &pb.Pagination{
			Page:     1,
			PageSize: 1,
		},
		Filters: []*pb.FilterCriteria{
			{
				Criteria: &pb.FilterCriteria_Condition{
					Condition: &pb.FilterCondition{
						Field:    "teacher_supervisor_code",
						Operator: pb.FilterOperator_EQUAL,
						Values:   []string{teacherId},
					},
				},
			},
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
		return false
	}

	return supervisors.GetTotal() > 0
}
