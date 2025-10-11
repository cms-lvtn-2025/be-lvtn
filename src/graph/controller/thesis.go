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

func (c *Controller) pbTopicsToModel(resp *pb.ListTopicsResponse) *[]model.Topic {
	if resp == nil {
		return nil
	}
	topics := resp.GetTopics()
	result := make([]model.Topic, len(topics))
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

		result = append(result, model.Topic{
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
	return &result
}

func (c *Controller) GetTopic(ctx context.Context) (*[]model.Topic, error) {
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
	if role == "student" {
		topics, err = c.thesis.GetTopicByStudentCode(ctx, myId, 1, 20, "created_at")
		if err != nil {
			return nil, err
		}
	} else if role == "teacher" {
		topics, err = c.thesis.GetTopicByTeacherCode(ctx, myId, 1, 20, "created_at")
		if err != nil {
			return nil, err
		}
	}
	return c.pbTopicsToModel(topics), nil
}
