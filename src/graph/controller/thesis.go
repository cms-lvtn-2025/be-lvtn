package controller

import (
	"context"
	"fmt"
	"strings"
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
	result := make([]*model.Topic, len(topics))
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
	result := make([]*model.Enrollment, len(enrollments))
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

func (c *Controller) GetTopics(ctx context.Context, pag model.Pagination) ([]*model.Topic, error) {
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
	var page int32 = 1
	var pageSize int32 = 10
	order := true
	sort := "created_at"
	if pag.Page != nil {
		page = *(pag.Page)
	}
	if pag.PageSize != nil {
		pageSize = *(pag.PageSize)
	}
	if pag.Order != nil {
		order = *pag.Order
	}
	if pag.Sort != nil {
		sort = *pag.Sort
	}

	if role == "student" {
		topics, err = c.thesis.GetTopicByStudentCode(ctx, myId, page, pageSize, sort, order)
		if err != nil {
			return nil, err
		}
	} else if role == "teacher" {
		topics, err = c.thesis.GetTopicByTeacherCode(ctx, myId, page, pageSize, sort, order)
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
