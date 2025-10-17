package controller

//
//import (
//	"context"
//	"fmt"
//	"strings"
//	pb "thaily/proto/council"
//	pbRole "thaily/proto/role"
//	pbUser "thaily/proto/user"
//	"thaily/src/graph/helper"
//	"thaily/src/graph/model"
//	"time"
//
//	"github.com/golang-jwt/jwt/v5"
//)
//
//func (c *Controller) pbCouncilsToModel(resp *pb.ListCouncilsResponse) []*model.Council {
//	if resp == nil {
//		return nil
//	}
//	councils := resp.GetCouncils()
//	var total int32 = 0
//	if resp.Total != 0 {
//		total = resp.Total
//	}
//	var result []*model.Council
//	for _, council := range councils {
//		var createdAt, updatedAt *time.Time
//		var createdBy, updatedBy *string
//		if council.CreatedAt != nil {
//			t := council.CreatedAt.AsTime()
//			createdAt = &t
//		}
//		if council.UpdatedAt != nil {
//			t := council.UpdatedAt.AsTime()
//			createdAt = &t
//		}
//		if council.CreatedBy != "" {
//			createdBy = &council.CreatedBy
//		}
//		if council.UpdatedBy != "" {
//			updatedBy = &council.UpdatedBy
//		}
//		result = append(result, &model.Council{
//			ID:           council.Id,
//			Title:        council.Title,
//			Total:        &total,
//			MajorCode:    council.MajorCode,
//			SemesterCode: council.SemesterCode,
//			CreatedAt:    createdAt,
//			UpdatedAt:    updatedAt,
//			CreatedBy:    createdBy,
//			UpdatedBy:    updatedBy,
//		})
//
//	}
//	return result
//}
//
//func (c *Controller) pbCouncilToModel(resp *pb.GetCouncilResponse) *model.Council {
//	if resp == nil {
//		return nil
//	}
//	council := resp.GetCouncil()
//	var createdAt, updatedAt *time.Time
//	var createdBy, updatedBy *string
//	if council.CreatedAt != nil {
//		t := council.CreatedAt.AsTime()
//		createdAt = &t
//	}
//	if council.UpdatedAt != nil {
//		t := council.UpdatedAt.AsTime()
//		createdAt = &t
//	}
//	if council.CreatedBy != "" {
//		createdBy = &council.CreatedBy
//
//	}
//	if council.UpdatedBy != "" {
//		updatedBy = &council.UpdatedBy
//	}
//	return &model.Council{
//		ID:           council.Id,
//		Title:        council.Title,
//		MajorCode:    council.MajorCode,
//		SemesterCode: council.SemesterCode,
//		CreatedAt:    createdAt,
//		UpdatedAt:    updatedAt,
//		CreatedBy:    createdBy,
//		UpdatedBy:    updatedBy,
//	}
//}
//
//func (c *Controller) pbDefencesToModel(resp *pb.ListDefencesResponse) []*model.Defence {
//	if resp == nil {
//		return nil
//	}
//	defences := resp.GetDefences()
//	var result []*model.Defence
//	for _, defence := range defences {
//		var posision model.DefencePosition
//		switch defence.GetPosition() {
//		case pb.DefencePosition_PRESIDENT:
//			posision = model.DefencePositionPresident
//		case pb.DefencePosition_SECRETARY:
//			posision = model.DefencePositionSecretary
//		case pb.DefencePosition_REVIEWER:
//			posision = model.DefencePositionReviewer
//		case pb.DefencePosition_MEMBER:
//			posision = model.DefencePositionMember
//		default:
//			posision = model.DefencePositionMember
//
//		}
//		result = append(result, &model.Defence{
//			ID:          defence.Id,
//			Title:       defence.Title,
//			CouncilCode: defence.CouncilCode,
//			TeacherCode: defence.TeacherCode,
//			Position:    posision,
//		})
//	}
//	return result
//}
//
//func (c *Controller) pbSchedulesToModel(resp *pb.ListCouncilSchedulesResponse) []*model.CouncilSchedule {
//	if resp == nil {
//		return nil
//	}
//	schedules := resp.GetCouncilSchedules()
//	var result []*model.CouncilSchedule
//	for _, schedule := range schedules {
//		var createdAt, updatedAt, timeEnd, timeStart *time.Time
//		var councilCode, topicCode *string
//		if schedule.CreatedAt != nil {
//			t := schedule.CreatedAt.AsTime()
//			createdAt = &t
//		}
//		if schedule.UpdatedAt != nil {
//			t := schedule.UpdatedAt.AsTime()
//			updatedAt = &t
//		}
//		if schedule.TimeStart != nil {
//			t := schedule.TimeStart.AsTime()
//			timeStart = &t
//		}
//		if schedule.TimeEnd != nil {
//			t := schedule.TimeEnd.AsTime()
//			timeEnd = &t
//		}
//		if schedule.CouncilsCode != "" {
//			councilCode = &schedule.CouncilsCode
//		}
//		if schedule.TopicCode != "" {
//			topicCode = &schedule.TopicCode
//		}
//		result = append(result, &model.CouncilSchedule{
//			ID:           schedule.Id,
//			Status:       schedule.Status,
//			TimeStart:    timeStart,
//			TimeEnd:      timeEnd,
//			TopicCode:    topicCode,
//			CouncilsCode: councilCode,
//			CreatedAt:    createdAt,
//			UpdatedAt:    updatedAt,
//		})
//	}
//	return result
//}
//
//func (c *Controller) pbScheduleToModel(schedule *pb.CouncilSchedule) *model.CouncilSchedule {
//	var result *model.CouncilSchedule
//	var createdAt, updatedAt, timeEnd, timeStart *time.Time
//	var councilCode, topicCode *string
//	if schedule.CreatedAt != nil {
//		t := schedule.CreatedAt.AsTime()
//		createdAt = &t
//	}
//	if schedule.UpdatedAt != nil {
//		t := schedule.UpdatedAt.AsTime()
//		updatedAt = &t
//	}
//	if schedule.TimeStart != nil {
//		t := schedule.TimeStart.AsTime()
//		timeStart = &t
//	}
//	if schedule.TimeEnd != nil {
//		t := schedule.TimeEnd.AsTime()
//		timeEnd = &t
//	}
//	if schedule.CouncilsCode != "" {
//		councilCode = &schedule.CouncilsCode
//	}
//	if schedule.TopicCode != "" {
//		topicCode = &schedule.TopicCode
//	}
//	result = &model.CouncilSchedule{
//		ID:           schedule.Id,
//		Status:       schedule.Status,
//		TimeStart:    timeStart,
//		TimeEnd:      timeEnd,
//		TopicCode:    topicCode,
//		CouncilsCode: councilCode,
//		CreatedAt:    createdAt,
//		UpdatedAt:    updatedAt,
//	}
//	return result
//}
//
//func (c *Controller) pbGradeToModel(resp *pb.GetGradeDefenceResponse) *model.GradeDefence {
//	if resp == nil {
//		return nil
//	}
//	grade := resp.GetGradeDefence()
//	var createdAt, updatedAt *time.Time
//	var councilGrade, secretaryGrade *int32
//	if grade.CreatedAt != nil {
//		t := grade.CreatedAt.AsTime()
//		createdAt = &t
//	}
//	if grade.UpdatedAt != nil {
//		t := grade.UpdatedAt.AsTime()
//		updatedAt = &t
//	}
//	if grade.GetCouncil() != -1 {
//		councilGrade = &grade.Council
//	}
//	if grade.GetSecretary() != -1 {
//		secretaryGrade = &grade.Secretary
//	}
//	return &model.GradeDefence{
//		ID:        grade.Id,
//		Council:   councilGrade,
//		Secretary: secretaryGrade,
//		CreatedAt: createdAt,
//		UpdatedAt: updatedAt,
//	}
//}
//
//func (c *Controller) GetCouncils(ctx context.Context, search model.SearchRequestInput) ([]*model.Council, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	semester, ok := ctx.Value("semester").(string)
//	idsArr := strings.Split(claims["ids"].(string), ",")
//	myId := ""
//	if semester == "" {
//		myId = strings.Split(idsArr[0], "-")[1]
//	} else {
//		for _, id := range idsArr {
//			if strings.HasPrefix(id, semester+"-") {
//				myId = strings.Split(id, "-")[1]
//			}
//		}
//	}
//	if myId == "" {
//		return nil, fmt.Errorf("no teacher found for semester %s", semester)
//	}
//	var councils *pb.ListCouncilsResponse
//	var err error
//	if role == "student" {
//		return nil, fmt.Errorf("student role not allowed")
//	} else if role == "teacher" {
//		var newSearch model.SearchRequestInput
//		var permissions *pbRole.ListRoleSystemsResponse
//		permissions, err = c.role.GetAllRoleByTeacherId(ctx, myId)
//
//		if err != nil || permissions == nil || len(permissions.GetRoleSystems()) == 0 {
//			return nil, err
//		}
//		permissionMap := make(map[pbRole.RoleType]bool)
//		for _, permission := range permissions.GetRoleSystems() {
//			permissionMap[permission.Role] = true
//		}
//		if permissionMap[pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF] {
//			newSearch = model.SearchRequestInput{
//				Pagination: search.Pagination,
//				Filters: append([]*model.FilterCriteriaInput{
//					&model.FilterCriteriaInput{
//						Condition: &model.FilterConditionInput{
//							Field:    "semester_code",
//							Operator: model.FilterOperatorEqual,
//							Values:   []string{semester},
//						},
//					},
//				}, search.Filters...),
//			}
//			councils, err = c.council.GetCouncilBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
//		} else if permissionMap[pbRole.RoleType_DEPARTMENT_LECTURER] {
//			var user *pbUser.GetStudentResponse
//			user, err = c.user.GetUserById(ctx, myId)
//			if err != nil {
//				return nil, err
//			}
//			newSearch = model.SearchRequestInput{
//				Pagination: search.Pagination,
//				Filters: append([]*model.FilterCriteriaInput{
//					&model.FilterCriteriaInput{
//						Condition: &model.FilterConditionInput{
//							Field:    "major_code",
//							Operator: model.FilterOperatorEqual,
//							Values:   []string{user.GetStudent().GetMajorCode()},
//						},
//					},
//				}, search.Filters...),
//			}
//			councils, err = c.council.GetCouncilBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
//		} else {
//			return nil, fmt.Errorf("unknown role %s", role)
//		}
//	}
//	if err != nil {
//		return nil, err
//	}
//	return c.pbCouncilsToModel(councils), nil
//
//}
//
//func (c *Controller) GetCouncilByIdForDefences(ctx context.Context, id string) (*model.Council, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	if role != "teacher" {
//		return nil, fmt.Errorf("teacher role not allowed")
//	}
//	council, err := c.council.GetCouncilById(ctx, id)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbCouncilToModel(council), nil
//}
//
//func (c *Controller) GetCouncilByIdForSchedule(ctx context.Context, id string) (*model.Council, error) {
//	_, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	council, err := c.council.GetCouncilById(ctx, id)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbCouncilToModel(council), nil
//}
//
//func (c *Controller) GetDefencesByCouncilCode(ctx context.Context, councilCode string) ([]*model.Defence, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	if role != "teacher" {
//		return nil, fmt.Errorf("teacher role not allowed")
//	}
//	defences, err := c.council.GetDefencesByCouncilCode(ctx, councilCode)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbDefencesToModel(defences), nil
//}
//
//func (c *Controller) GetDefences(ctx context.Context, search model.SearchRequestInput) ([]*model.Defence, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	if role != "teacher" {
//		return nil, fmt.Errorf("student role not allowed")
//	}
//	semester, ok := ctx.Value("semester").(string)
//	idsArr := strings.Split(claims["ids"].(string), ",")
//	myId := ""
//	if semester == "" {
//		myId = strings.Split(idsArr[0], "-")[1]
//	} else {
//		for _, id := range idsArr {
//			if strings.HasPrefix(id, semester+"-") {
//				myId = strings.Split(id, "-")[1]
//			}
//		}
//	}
//	newSearch := model.SearchRequestInput{
//		Pagination: search.Pagination,
//		Filters: append([]*model.FilterCriteriaInput{
//			&model.FilterCriteriaInput{
//				Condition: &model.FilterConditionInput{
//					Field:    "teacher_code",
//					Operator: model.FilterOperatorEqual,
//					Values:   []string{myId},
//				},
//			},
//		}, search.Filters...),
//	}
//	defences, err := c.council.GetDefencesBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
//	if err != nil {
//		return nil, err
//	}
//	return c.pbDefencesToModel(defences), nil
//}
//
//func (c *Controller) GetSchedulesByCouncilCode(ctx context.Context, councilCode string) ([]*model.CouncilSchedule, error) {
//	_, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	schedules, err := c.council.GetSchedulesByCouncilCode(ctx, councilCode)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbSchedulesToModel(schedules), nil
//}
//
//func (c *Controller) GetScheduleByTopicCodeForTopic(ctx context.Context, topicCode string) (*model.CouncilSchedule, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	if role != "student" {
//		return nil, fmt.Errorf("student role not allowed")
//	}
//
//	schedule, err := c.council.GetScheduleByTopicCode(ctx, topicCode)
//	if err != nil {
//		return nil, err
//	}
//	if len(schedule.GetCouncilSchedules()) == 0 {
//		return nil, nil
//	}
//	return c.pbScheduleToModel(schedule.GetCouncilSchedules()[0]), nil
//}
//
//func (c *Controller) GetGradeByIdForEnrollment(ctx context.Context, id *string) (*model.GradeDefence, error) {
//	_, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	if id == nil {
//		return nil, nil
//	}
//	grade, err := c.council.GetGradeById(ctx, *id)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbGradeToModel(grade), nil
//}
