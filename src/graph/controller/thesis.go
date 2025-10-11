package controller

import (
	"context"
	"fmt"
	"strings"
	pbRole "thaily/proto/role"
	pb "thaily/proto/thesis"
	"thaily/src/graph/helper"
	"thaily/src/graph/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (c *Controller) pbTopicsToModel(resp *pb.ListTopicsResponse) []*model.Topic {
	if resp == nil {
		return nil
	}
	topics := resp.GetTopics()
	var total *int32
	if resp.Total != 0 {
		total = &resp.Total
	}
	result := make([]*model.Topic, 0, len(topics))
	for _, topic := range topics {
		var status model.TopicStatus
		switch topic.GetStatus() {
		case pb.TopicStatus_TOPIC_PENDING:
			status = model.TopicStatusPending
		case pb.TopicStatus_APPROVED:
			status = model.TopicStatusApproved
		case pb.TopicStatus_IN_PROGRESS:
			status = model.TopicStatusInProgress
		case pb.TopicStatus_REJECTED:
			status = model.TopicStatusRejected
		case pb.TopicStatus_TOPIC_COMPLETED:
			status = model.TopicStatusCompleted
		default:
			status = model.TopicStatusPending

		}
		var createdAt, updatedAt *time.Time
		if topic.GetCreatedAt() != nil {
			t := topic.GetCreatedAt().AsTime()
			createdAt = &t
		}
		if topic.GetUpdatedAt() != nil {
			t := topic.GetUpdatedAt().AsTime()
			updatedAt = &t
		}

		// handle optional fields
		var createdBy, updatedBy *string
		if topic.CreatedBy != "" {
			createdBy = &topic.CreatedBy
		}
		if topic.UpdatedBy != "" {
			updatedBy = &topic.UpdatedBy
		}

		result = append(result, &model.Topic{
			ID:                    topic.GetId(),
			Total:                 total,
			Title:                 topic.GetTitle(),
			MajorCode:             topic.GetMajorCode(),
			Status:                status,
			CreatedAt:             createdAt,
			UpdatedAt:             updatedAt,
			SemesterCode:          topic.GetSemesterCode(),
			TeacherSupervisorCode: topic.GetTeacherSupervisorCode(),
			TimeEnd:               topic.TimeEnd.AsTime(),
			TimeStart:             topic.TimeStart.AsTime(),
			CreatedBy:             createdBy,
			UpdatedBy:             updatedBy,
		})
	}
	return result
}

func (c *Controller) pbEnrollmentsToModel(resp *pb.ListEnrollmentsResponse) []*model.Enrollment {
	if resp == nil {
		return nil
	}
	enrollments := resp.GetEnrollments()
	result := make([]*model.Enrollment, 0, len(enrollments))
	for _, enrollment := range enrollments {
		var createdAt, updatedAt *time.Time
		if enrollment.GetCreatedAt() != nil {
			t := enrollment.GetCreatedAt().AsTime()
			createdAt = &t
		}
		if enrollment.GetUpdatedAt() != nil {
			t := enrollment.GetUpdatedAt().AsTime()
			updatedAt = &t
		}
		var createdBy, updatedBy, midTermCode, topicCode, finalCode, gradeCode *string
		if enrollment.CreatedBy != "" {
			createdBy = &enrollment.CreatedBy
		}
		if enrollment.UpdatedBy != "" {
			updatedBy = &enrollment.UpdatedBy
		}
		if enrollment.TopicCode != "" {
			topicCode = &enrollment.TopicCode
		}
		if enrollment.FinalCode != "" {
			finalCode = &enrollment.FinalCode
		}
		if enrollment.GradeCode != "" {
			gradeCode = &enrollment.GradeCode
		}
		if enrollment.MidtermCode != "" {
			midTermCode = &enrollment.MidtermCode
		}
		result = append(result, &model.Enrollment{
			ID:          enrollment.GetId(),
			Title:       enrollment.GetTitle(),
			StudentCode: enrollment.GetStudentCode(),
			MidtermCode: midTermCode,
			FinalCode:   finalCode,
			GradeCode:   gradeCode,
			TopicCode:   topicCode,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			CreatedBy:   createdBy,
			UpdatedBy:   updatedBy,
		})

	}
	return result
}

func (c *Controller) pbMidtermToModel(resp *pb.GetMidtermResponse) *model.Midterm {
	if resp == nil {
		return nil
	}
	midterm := resp.GetMidterm()

	var status model.MidtermStatus
	switch midterm.GetStatus() {
	case pb.MidtermStatus_GRADED:
		status = model.MidtermStatusGraded
	case pb.MidtermStatus_NOT_SUBMITTED:
		status = model.MidtermStatusNotSubmitted
	case pb.MidtermStatus_SUBMITTED:
		status = model.MidtermStatusSubmitted
	default:
		status = model.MidtermStatusNotSubmitted
	}
	// field optional
	var gradeInt *int32
	var feedBack, createdBy, updatedBy *string
	var createdAt, updatedAt *time.Time
	if midterm.GetGrade() != -1 {
		gradeInt = &midterm.Grade
	}
	if midterm.GetFeedback() != "" {
		feedBack = &midterm.Feedback
	}
	if midterm.GetCreatedBy() != "" {
		createdBy = &midterm.CreatedBy
	}
	if midterm.GetUpdatedBy() != "" {
		updatedBy = &midterm.UpdatedBy
	}
	if midterm.GetCreatedAt() != nil {
		t := midterm.GetCreatedAt().AsTime()
		createdAt = &t
	}
	if midterm.GetUpdatedAt() != nil {
		t := midterm.GetUpdatedAt().AsTime()
		updatedAt = &t
	}

	result := &model.Midterm{
		ID:        midterm.GetId(),
		Title:     midterm.GetTitle(),
		Status:    status,
		Grade:     gradeInt,
		Feedback:  feedBack,
		CreatedBy: createdBy,
		UpdatedBy: updatedBy,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return result
}

func (c *Controller) pbFinalToModel(resp *pb.GetFinalResponse) *model.Final {
	if resp == nil {
		return nil
	}
	final := resp.GetFinal()
	var supervisorGrade, reviewerGrade, finalGrade *int32
	var notes, createdBy, updatedBy *string
	var status model.FinalStatus
	var createdAt, updatedAt *time.Time
	switch final.GetStatus() {
	case pb.FinalStatus_PENDING:
		status = model.FinalStatusPending
	case pb.FinalStatus_COMPLETED:
		status = model.FinalStatusCompleted
	case pb.FinalStatus_FAILED:
		status = model.FinalStatusFailed
	case pb.FinalStatus_PASSED:
		status = model.FinalStatusPassed
	default:
		status = model.FinalStatusPending

	}
	if final.GetSupervisorGrade() != -1 {
		supervisorGrade = &final.SupervisorGrade
	}
	if final.GetReviewerGrade() != -1 {
		reviewerGrade = &final.ReviewerGrade
	}
	if final.GetFinalGrade() != -1 {
		finalGrade = &final.FinalGrade
	}
	if final.GetNotes() != "" {
		notes = &final.Notes
	}
	if final.GetCreatedBy() != "" {
		createdBy = &final.CreatedBy
	}
	if final.GetUpdatedBy() != "" {
		updatedBy = &final.UpdatedBy
	}
	if final.GetCreatedAt() != nil {
		t := final.GetCreatedAt().AsTime()
		createdAt = &t
	}
	if final.GetUpdatedAt() != nil {
		t := final.GetUpdatedAt().AsTime()
		updatedAt = &t
	}

	return &model.Final{
		ID:              final.GetId(),
		Title:           final.GetTitle(),
		Status:          status,
		SupervisorGrade: supervisorGrade,
		ReviewerGrade:   reviewerGrade,
		FinalGrade:      finalGrade,
		Notes:           notes,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

}

func (c *Controller) GetTopics(ctx context.Context, search model.SearchRequestInput) ([]*model.Topic, error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("not authorized")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("not authorized")
	}
	semester, ok := ctx.Value("semester").(string)

	idsArr := strings.Split(claims["ids"].(string), ",")
	myId := ""
	if semester == "" {
		myId = strings.Split(idsArr[0], "-")[1]
	} else {
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+"-") {
				myId = strings.Split(id, "-")[1]
			}
		}
	}
	if myId == "" {
		return nil, fmt.Errorf("no teacher found for semester %s", semester)
	}
	var topics *pb.ListTopicsResponse
	var err error
	var newSearch model.SearchRequestInput
	if role == "student" {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters: append([]*model.FilterCriteriaInput{
				&model.FilterCriteriaInput{
					Condition: &model.FilterConditionInput{
						Field:    "student_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{myId},
					},
				},
			}, search.Filters...),
		}

		topics, err = c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
		if err != nil {
			return nil, err
		}
	} else if role == "teacher" {
		permissions, err := c.role.GetAllRoleByTeacherId(ctx, myId)

		if err != nil || permissions == nil || len(permissions.GetRoleSystems()) == 0 {
			return nil, err
		}
		permissionMap := make(map[pbRole.RoleType]bool)
		for _, permission := range permissions.GetRoleSystems() {
			permissionMap[permission.Role] = true
		}
		if permissionMap[pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF] {
			topics, err = c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(search))
		} else if permissionMap[pbRole.RoleType_DEPARTMENT_LECTURER] {
			user, err := c.user.GetUserById(ctx, myId)
			if err != nil {
				return nil, err
			}
			newSearch = model.SearchRequestInput{
				Pagination: search.Pagination,
				Filters: append([]*model.FilterCriteriaInput{
					&model.FilterCriteriaInput{
						Condition: &model.FilterConditionInput{
							Field:    "major_code",
							Operator: model.FilterOperatorEqual,
							Values:   []string{user.GetStudent().GetSemesterCode()},
						},
					},
				}, search.Filters...),
			}
			topics, err = c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))

		} else if permissionMap[pbRole.RoleType_SUPERVISOR_LECTURER] {
			newSearch = model.SearchRequestInput{
				Pagination: search.Pagination,
				Filters: append([]*model.FilterCriteriaInput{
					&model.FilterCriteriaInput{
						Condition: &model.FilterConditionInput{
							Field:    "teacher_supervisor_code",
							Operator: model.FilterOperatorEqual,
							Values:   []string{myId},
						},
					},
				}, search.Filters...),
			}
			topics, err = c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
		} else {
			return nil, err
		}
		if err != nil {
			return nil, err
		}

	}
	return c.pbTopicsToModel(topics), nil
}

func (c *Controller) GetEnrollments(ctx context.Context, topicCode string) ([]*model.Enrollment, error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("not authorized")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("not authorized")
	}
	semester, ok := ctx.Value("semester").(string)

	idsArr := strings.Split(claims["ids"].(string), ",")
	myId := ""
	if semester == "" {
		myId = strings.Split(idsArr[0], "-")[1]
	} else {
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+"-") {
				myId = strings.Split(id, "-")[1]
			}
		}
	}
	if myId == "" {
		return nil, fmt.Errorf("no teacher found for semester %s", semester)
	}
	var enrolls *pb.ListEnrollmentsResponse
	var err error
	if role == "student" {
		enrolls, err = c.thesis.GetEnrolmentByTopicCodeAndStudentCode(ctx, topicCode, myId)
		if err != nil {
			return nil, err
		}
	} else if role == "teacher" {
		enrolls, err = c.thesis.GetEnrollmentByTopicCode(ctx, topicCode)
		if err != nil {
			return nil, err
		}
	}
	
	return c.pbEnrollmentsToModel(enrolls), nil
}

func (c *Controller) GetMidterm(ctx context.Context, midtermCode *string) (*model.Midterm, error) {
	if midtermCode == nil {
		return nil, fmt.Errorf("no teacher found for midterm")
	}
	midterm, err := c.thesis.GetMidtermById(ctx, *midtermCode)
	if err != nil {
		return nil, err
	}
	return c.pbMidtermToModel(midterm), nil
}

func (c *Controller) GetFinal(ctx context.Context, finalCode *string) (*model.Final, error) {
	if finalCode == nil {
		return nil, fmt.Errorf("no teacher found for final")
	}
	final, err := c.thesis.GetFinalById(ctx, *finalCode)
	if err != nil {
		return nil, err
	}
	return c.pbFinalToModel(final), err
}
