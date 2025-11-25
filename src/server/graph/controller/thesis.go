package controller

import (
	"context"
	"thaily/src/server/graph/convert"
	convert2 "thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// Student-related methods (GetEnrollmentsForStudent, GetTopicForStudent, GetMidtermById, etc.)
// have been moved to controller/student.go
//
//func (c *Controller) pbTopicsToModel(resp *pb.ListTopicsResponse) []*model.Topic {
//	if resp == nil {
//		return nil
//	}
//	topics := resp.GetTopics()
//	var total *int32
//	if resp.Total != 0 {
//		total = &resp.Total
//	}
//	result := make([]*model.Topic, 0, len(topics))
//	for _, topic := range topics {
//		var status model.TopicStatus
//		switch topic.GetStatus() {
//		case pb.TopicStatus_IN_PROGRESS:
//			status = model.TopicStatusInProgress
//		case pb.TopicStatus_REJECTED:
//			status = model.TopicStatusRejected
//		case pb.TopicStatus_TOPIC_COMPLETED:
//
//		}
//		var createdAt, updatedAt *time.Time
//		if topic.GetCreatedAt() != nil {
//			t := topic.GetCreatedAt().AsTime()
//			createdAt = &t
//		}
//		if topic.GetUpdatedAt() != nil {
//			t := topic.GetUpdatedAt().AsTime()
//			updatedAt = &t
//		}
//
//		// handle optional fields
//		var createdBy, updatedBy *string
//		if topic.CreatedBy != "" {
//			createdBy = &topic.CreatedBy
//		}
//		if topic.UpdatedBy != "" {
//			updatedBy = &topic.UpdatedBy
//		}
//
//		result = append(result, &model.Topic{
//			ID:           topic.GetId(),
//			Total:        total,
//			Title:        topic.GetTitle(),
//			MajorCode:    topic.GetMajorCode(),
//			Status:       status,
//			CreatedAt:    createdAt,
//			UpdatedAt:    updatedAt,
//			SemesterCode: topic.GetSemesterCode(),
//			CreatedBy:    createdBy,
//			UpdatedBy:    updatedBy,
//		})
//	}
//	return result
//}
//
//func (c *Controller) pbTopicToModel(resp *pb.GetTopicResponse) *model.Topic {
//	if resp == nil {
//		return nil
//	}
//	topic := resp.GetTopic()
//	var status model.TopicStatus
//	switch topic.GetStatus() {
//	case pb.TopicStatus_IN_PROGRESS:
//		status = model.TopicStatusInProgress
//	case pb.TopicStatus_REJECTED:
//		status = model.TopicStatusRejected
//	case pb.TopicStatus_TOPIC_COMPLETED:
//
//	}
//	var createdAt, updatedAt *time.Time
//	if topic.GetCreatedAt() != nil {
//		t := topic.GetCreatedAt().AsTime()
//		createdAt = &t
//	}
//	if topic.GetUpdatedAt() != nil {
//		t := topic.GetUpdatedAt().AsTime()
//		updatedAt = &t
//	}
//
//	// handle optional fields
//	var createdBy, updatedBy *string
//	if topic.CreatedBy != "" {
//		createdBy = &topic.CreatedBy
//	}
//	if topic.UpdatedBy != "" {
//		updatedBy = &topic.UpdatedBy
//	}
//	return &model.Topic{
//		ID:           topic.GetId(),
//		Title:        topic.GetTitle(),
//		MajorCode:    topic.GetMajorCode(),
//		Status:       status,
//		CreatedAt:    createdAt,
//		UpdatedAt:    updatedAt,
//		SemesterCode: topic.GetSemesterCode(),
//		CreatedBy:    createdBy,
//		UpdatedBy:    updatedBy,
//	}
//}
//
//func (c *Controller) pbEnrollmentsToModel(resp *pb.ListEnrollmentsResponse) []*model.Enrollment {
//	if resp == nil {
//		return nil
//	}
//	enrollments := resp.GetEnrollments()
//	result := make([]*model.Enrollment, 0, len(enrollments))
//	for _, enrollment := range enrollments {
//		var createdAt, updatedAt *time.Time
//		if enrollment.GetCreatedAt() != nil {
//			t := enrollment.GetCreatedAt().AsTime()
//			createdAt = &t
//		}
//		if enrollment.GetUpdatedAt() != nil {
//			t := enrollment.GetUpdatedAt().AsTime()
//			updatedAt = &t
//		}
//		var createdBy, updatedBy, finalCode *string
//		if enrollment.CreatedBy != "" {
//			createdBy = &enrollment.CreatedBy
//		}
//		if enrollment.UpdatedBy != "" {
//			updatedBy = &enrollment.UpdatedBy
//		}
//		result = append(result, &model.Enrollment{
//			ID:          enrollment.GetId(),
//			Title:       enrollment.GetTitle(),
//			StudentCode: enrollment.GetStudentCode(),
//			MidtermCode: enrollment.MidtermCode,
//			FinalCode:   finalCode,
//			CreatedAt:   createdAt,
//			UpdatedAt:   updatedAt,
//			CreatedBy:   createdBy,
//			UpdatedBy:   updatedBy,
//		})
//
//	}
//	return result
//}
//
//func (c *Controller) pbMidtermToModel(resp *pb.GetMidtermResponse) *model.Midterm {
//	if resp == nil {
//		return nil
//	}
//	midterm := resp.GetMidterm()
//
//	var status model.MidtermStatus
//	switch midterm.GetStatus() {
//	case pb.MidtermStatus_NOT_SUBMITTED:
//		status = model.MidtermStatusNotSubmitted
//	case pb.MidtermStatus_SUBMITTED:
//		status = model.MidtermStatusSubmitted
//	case pb.MidtermStatus_PASS:
//		status = model.MidtermStatusPass
//	case pb.MidtermStatus_FAIL:
//		status = model.MidtermStatusFail
//	default:
//		status = model.MidtermStatusNotSubmitted
//	}
//	// field optional
//	var gradeInt *int32
//	var feedBack, createdBy, updatedBy *string
//	var createdAt, updatedAt *time.Time
//	if midterm.GetGrade() != -1 {
//		gradeInt = &midterm.Grade
//	}
//	if midterm.GetFeedback() != "" {
//		feedBack = &midterm.Feedback
//	}
//	if midterm.GetCreatedBy() != "" {
//		createdBy = &midterm.CreatedBy
//	}
//	if midterm.GetUpdatedBy() != "" {
//		updatedBy = &midterm.UpdatedBy
//	}
//	if midterm.GetCreatedAt() != nil {
//		t := midterm.GetCreatedAt().AsTime()
//		createdAt = &t
//	}
//	if midterm.GetUpdatedAt() != nil {
//		t := midterm.GetUpdatedAt().AsTime()
//		updatedAt = &t
//	}
//
//	result := &model.Midterm{
//		ID:        midterm.GetId(),
//		Title:     midterm.GetTitle(),
//		Status:    status,
//		Grade:     gradeInt,
//		Feedback:  feedBack,
//		CreatedBy: createdBy,
//		UpdatedBy: updatedBy,
//		CreatedAt: createdAt,
//		UpdatedAt: updatedAt,
//	}
//	return result
//}
//
//func (c *Controller) pbFinalToModel(resp *pb.GetFinalResponse) *model.Final {
//	if resp == nil {
//		return nil
//	}
//	final := resp.GetFinal()
//	var supervisorGrade, finalGrade *int32
//	var notes, createdBy, updatedBy *string
//	var status model.FinalStatus
//	var createdAt, updatedAt *time.Time
//	switch final.GetStatus() {
//	case pb.FinalStatus_PENDING:
//		status = model.FinalStatusPending
//	case pb.FinalStatus_COMPLETED:
//		status = model.FinalStatusCompleted
//	case pb.FinalStatus_FAILED:
//		status = model.FinalStatusFailed
//	case pb.FinalStatus_PASSED:
//		status = model.FinalStatusPassed
//	default:
//		status = model.FinalStatusPending
//
//	}
//	if final.GetSupervisorGrade() != -1 {
//		supervisorGrade = &final.SupervisorGrade
//	}
//	if final.GetFinalGrade() != -1 {
//		finalGrade = &final.FinalGrade
//	}
//	if final.GetNotes() != "" {
//		notes = &final.Notes
//	}
//	if final.GetCreatedBy() != "" {
//		createdBy = &final.CreatedBy
//	}
//	if final.GetUpdatedBy() != "" {
//		updatedBy = &final.UpdatedBy
//	}
//	if final.GetCreatedAt() != nil {
//		t := final.GetCreatedAt().AsTime()
//		createdAt = &t
//	}
//	if final.GetUpdatedAt() != nil {
//		t := final.GetUpdatedAt().AsTime()
//		updatedAt = &t
//	}
//
//	return &model.Final{
//		ID:              final.GetId(),
//		Title:           final.GetTitle(),
//		Status:          status,
//		SupervisorGrade: supervisorGrade,
//		FinalGrade:      finalGrade,
//		Notes:           notes,
//		CreatedBy:       createdBy,
//		UpdatedBy:       updatedBy,
//		CreatedAt:       createdAt,
//		UpdatedAt:       updatedAt,
//	}
//
//}
//
//func (c *Controller) GetTopics(ctx context.Context, search model.SearchRequestInput) ([]*model.Topic, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	semester, ok := ctx.Value("semester").(string)
//
//	idsArr := strings.Split(claims["ids"].(string), ",")
//	myId := ""
//	if semester == "" {
//		myId = strings.Split(idsArr[0], ":")[1]
//	} else {
//		for _, id := range idsArr {
//			if strings.HasPrefix(id, semester+":") {
//				myId = strings.Split(id, ":")[1]
//			}
//		}
//	}
//	if myId == "" {
//		return nil, fmt.Errorf("no teacher found for semester %s", semester)
//	}
//	var topics *pb.ListTopicsResponse
//
//	if role == "student" {
//		return nil, fmt.Errorf("student role not allowed")
//	} else if role == "teacher" {
//		var err error
//		var newSearch model.SearchRequestInput
//
//		permissions, err := c.role.GetAllRoleByTeacherId(ctx, myId)
//
//		if err != nil || permissions == nil || len(permissions.GetRoleSystems()) == 0 {
//			return nil, err
//		}
//		permissionMap := make(map[pbRole.RoleType]bool)
//		for _, permission := range permissions.GetRoleSystems() {
//			permissionMap[permission.Role] = true
//		}
//		if permissionMap[pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF] {
//			topics, err = c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(search))
//		} else if permissionMap[pbRole.RoleType_DEPARTMENT_LECTURER] {
//			user, err := c.user.GetUserById(ctx, myId)
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
//							Values:   []string{user.GetStudent().GetSemesterCode()},
//						},
//					},
//				}, search.Filters...),
//			}
//			topics, err = c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
//
//		} else if permissionMap[pbRole.RoleType_TEACHER] {
//			newSearch = model.SearchRequestInput{
//				Pagination: search.Pagination,
//				Filters: append([]*model.FilterCriteriaInput{
//					&model.FilterCriteriaInput{
//						Condition: &model.FilterConditionInput{
//							Field:    "teacher_supervisor_code",
//							Operator: model.FilterOperatorEqual,
//							Values:   []string{myId},
//						},
//					},
//				}, search.Filters...),
//			}
//			topics, err = c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
//		} else {
//			return nil, err
//		}
//		if err != nil {
//			return nil, err
//		}
//
//	}
//	return c.pbTopicsToModel(topics), nil
//}
//
//func (c *Controller) GetTopicByIdForSchedule(ctx context.Context, id *string) (*model.Topic, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	if id == nil {
//		return nil, nil
//	}
//	if role != "teacher" {
//		return nil, fmt.Errorf("no teacher found for student role %s", role)
//	}
//	topic, err := c.thesis.GetTopicById(ctx, *id)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbTopicToModel(topic), nil
//}
//
//func (c *Controller) GetEnrollmentsChild(ctx context.Context, topicCode string) ([]*model.Enrollment, error) {
//	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	role, ok := claims["role"].(string)
//	if !ok {
//		return nil, fmt.Errorf("not authorized")
//	}
//	semester, ok := ctx.Value("semester").(string)
//
//	idsArr := strings.Split(claims["ids"].(string), ",")
//	myId := ""
//	if semester == "" {
//		myId = strings.Split(idsArr[0], ":")[1]
//	} else {
//		for _, id := range idsArr {
//			if strings.HasPrefix(id, semester+":") {
//				myId = strings.Split(id, ":")[1]
//			}
//		}
//	}
//	if myId == "" {
//		return nil, fmt.Errorf("no teacher found for semester %s", semester)
//	}
//	var enrolls *pb.ListEnrollmentsResponse
//	if role == "student" {
//		return nil, fmt.Errorf("student role not allowed")
//	} else if role == "teacher" {
//		//enrolls, err = c.thesis.GetEnrollmentByTopicCode(ctx, topicCode)
//		//if err != nil {
//		//	return nil, err
//		//}
//		return nil, nil
//	}
//
//	return c.pbEnrollmentsToModel(enrolls), nil
//}
//
//func (c *Controller) GetEnrollments(ctx context.Context, search model.SearchRequestInput) ([]*model.Enrollment, error) {
//	//claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
//	//if !ok {
//	//	return nil, fmt.Errorf("not authorized")
//	//}
//	//role, ok := claims["role"].(string)
//	//if !ok {
//	//	return nil, fmt.Errorf("not authorized")
//	//}
//	//semester, ok := ctx.Value("semester").(string)
//	//
//	//idsArr := strings.Split(claims["ids"].(string), ",")
//	//myId := ""
//	//if semester == "" {
//	//	myId = strings.Split(idsArr[0], ":")[1]
//	//} else {
//	//	for _, id := range idsArr {
//	//		if strings.HasPrefix(id, semester+":") {
//	//			myId = strings.Split(id, ":")[1]
//	//		}
//	//	}
//	//}
//	//if myId == "" {
//	//	return nil, fmt.Errorf("no teacher found for semester %s", semester)
//	//}
//	var err error
//	var enrolls *pb.ListEnrollmentsResponse
//	enrolls, err = c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(search))
//
//	//if role == "student" {
//	//	//var err error
//	//	//var newSearch model.SearchRequestInput
//	//	//newSearch = model.SearchRequestInput{
//	//	//	Pagination: search.Pagination,
//	//	//	Filters: append([]*model.FilterCriteriaInput{
//	//	//		&model.FilterCriteriaInput{
//	//	//			Condition: &model.FilterConditionInput{
//	//	//				Field:    "student_code",
//	//	//				Operator: model.FilterOperatorEqual,
//	//	//				Values:   []string{myId},
//	//	//			},
//	//	//		},
//	//	//	}, search.Filters...),
//	//	//}
//	//	enrolls, err = c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(search))
//	//	if err != nil {
//	//		return nil, err
//	//	}
//	//} else if role == "teacher" {
//	//	return nil, fmt.Errorf("teacher role not allowed")
//	//}
//	if err != nil {
//		return nil, err
//	}
//
//	return c.pbEnrollmentsToModel(enrolls), nil
//}
//
//func (c *Controller) GetMidterm(ctx context.Context, midtermCode *string) (*model.Midterm, error) {
//	if midtermCode == nil {
//		return nil, fmt.Errorf("no teacher found for midterm")
//	}
//	midterm, err := c.thesis.GetMidtermById(ctx, *midtermCode)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbMidtermToModel(midterm), nil
//}
//
//func (c *Controller) GetFinal(ctx context.Context, finalCode *string) (*model.Final, error) {
//	if finalCode == nil {
//		return nil, fmt.Errorf("no teacher found for final")
//	}
//	final, err := c.thesis.GetFinalById(ctx, *finalCode)
//	if err != nil {
//		return nil, err
//	}
//	return c.pbFinalToModel(final), err
//}

// ============================================
// RESOLVER HELPER METHODS
// ============================================

// GetTopicsByMajorId returns all topics for a given major
func (c *Controller) GetTopicsByMajorId(ctx context.Context, majorId string) ([]*model.Topic, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{majorId},
				},
			},
		},
	}

	topics, err := c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicsToModel(topics.GetTopics()), nil
}

// GetTopicsBySemesterId returns all topics for a given semester
func (c *Controller) GetTopicsBySemesterId(ctx context.Context, semesterId string) ([]*model.Topic, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "semester_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{semesterId},
				},
			},
		},
	}

	topics, err := c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicsToModel(topics.GetTopics()), nil
}

// GetStudentById returns a student by ID
func (c *Controller) GetStudentById(ctx context.Context, id string) (*model.Student, error) {
	student, err := c.user.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbStudentToModel(student.GetStudent()), nil
}

// GetTeacherById returns a teacher by ID
func (c *Controller) GetTeacherById(ctx context.Context, id string) (*model.Teacher, error) {
	teacher, err := c.user.GetTeacherById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTeacherToModel(teacher.GetTeacher()), nil
}

// GetTopicById returns a topic by ID
func (c *Controller) GetTopicById(ctx context.Context, id string) (*model.Topic, error) {
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicToModel(topic.GetTopic()), nil
}

// GetMidtermById returns a midterm by ID
func (c *Controller) GetMidtermById(ctx context.Context, id string) (*model.Midterm, error) {
	midterm, err := c.thesis.GetMidtermById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbMidtermToModel(midterm.GetMidterm()), nil
}

// GetFinalById returns a final by ID
func (c *Controller) GetFinalById(ctx context.Context, id string) (*model.Final, error) {
	final, err := c.thesis.GetFinalById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbFinalToModel(final.GetFinal()), nil
}

// GetTopicCouncilById returns a topic council by ID
func (c *Controller) GetTopicCouncilById(ctx context.Context, id string) (*model.TopicCouncil, error) {
	access, err := c.GetRbacDynamicRole(ctx, []string{"council", "department", "affair"})
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, nil
	}
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicCouncilToModel(topicCouncil.GetTopicCouncil()), nil
}

// GetGradeReviewById returns a grade review by ID
func (c *Controller) GetGradeReviewById(ctx context.Context, id string) (*model.GradeReview, error) {
	gradeReview, err := c.thesis.GetGradeReviewById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbGradeReviewToModel(gradeReview.GetGradeReview()), nil
}

// GetGradeDefencesByEnrollmentId returns all grade defences for a given enrollment
func (c *Controller) GetGradeDefencesByEnrollmentId(ctx context.Context, enrollmentId string) ([]*model.GradeDefence, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "enrollment_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{enrollmentId},
				},
			},
		},
	}

	gradeDefences, err := c.council.GetGradeDefenceBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefencesToModel(gradeDefences.GetGradeDefences()), nil
}

// GetFilesByTopicId returns all files for a given topic
func (c *Controller) GetFilesByTopicId(ctx context.Context, topicId string) ([]*model.File, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "topic_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{topicId},
				},
			},
		},
	}

	files, err := c.file.GetFileBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbFilesToModel(files.GetFiles()), nil
}

// GetTopicCouncilsByTopicId returns all topic councils for a given topic
func (c *Controller) GetTopicCouncilsByTopicId(ctx context.Context, topicId string) ([]*model.TopicCouncil, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "topic_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{topicId},
				},
			},
		},
	}

	topicCouncils, err := c.thesis.GetTopicCouncilBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicCouncilsToModel(topicCouncils.GetTopicCouncils()), nil
}

// GetEnrollmentsByTopicCouncilId returns all enrollments for a given topic council
func (c *Controller) GetEnrollmentsByTopicCouncilId(ctx context.Context, topicCouncilId string) ([]*model.Enrollment, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "topic_council_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{topicCouncilId},
				},
			},
		},
	}

	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert2.PbEnrollmentsToModel(enrollments.GetEnrollments()), nil
}

// GetSupervisorsByTopicCouncilId returns all supervisors for a given topic council
func (c *Controller) GetSupervisorsByTopicCouncilId(ctx context.Context, topicCouncilId string) ([]*model.TopicCouncilSupervisor, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "topic_council_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{topicCouncilId},
				},
			},
		},
	}

	supervisors, err := c.thesis.GetTopicCouncilSupervisorBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicCouncilSupervisorsToModel(supervisors.GetTopicCouncilSupervisors()), nil
}

// GetEnrollmentById returns an enrollment by ID
func (c *Controller) GetEnrollmentById(ctx context.Context, id string) (*model.Enrollment, error) {
	enrollment, err := c.thesis.GetEnrollmentById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbEnrollmentToModel(enrollment.GetEnrollment()), nil
}
