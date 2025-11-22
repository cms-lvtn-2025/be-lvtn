package controller

import (
	"context"
	"fmt"
	pbCouncil "thaily/proto/council"
	pbThesis "thaily/proto/thesis"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// DEPARTMENT LECTURER CONTROLLER
// Giáo viên bộ môn có quyền xem tất cả topic, enrollment trong bộ môn
// và tạo council, phê duyệt topic lần 1
// ============================================

// ============================================
// QUERY METHODS - User Management
// ============================================

func (c *Controller) GetAllMajorByMajorCode(ctx context.Context) ([]string, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("no department lecturer role")
	}
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}
	majorCode := teacher.GetTeacher().GetMajorCode()
	major, err := c.academic.GetMajorById(ctx, majorCode)
	if err != nil {
		return nil, err
	}
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "faculty_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{major.GetMajor().GetFacultyCode()},
				},
			},
		},
	}
	majors, err := c.academic.GetMajorsBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}
	majorCodes := make([]string, 0)
	for _, major := range majors.GetMajors() {
		majorCodes = append(majorCodes, major.GetId())
	}
	return majorCodes, nil
}

// GetDepartmentTeachers returns all teachers in the department
func (c *Controller) GetDepartmentTeachers(ctx context.Context, search model.SearchRequestInput) (*model.TeacherListResponse, error) {
	majorCodes, err := c.GetAllMajorByMajorCode(ctx)
	if err != nil {
		return nil, err
	}
	semester, ok := ctx.Value("semester").(string)
	if !ok {
		return nil, fmt.Errorf("no semester")
	}
	fmt.Println("semester", semester)
	fmt.Println("majorCodes", majorCodes)

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorIn,
					Values:   majorCodes,
				},
			},
			{
				Condition: &model.FilterConditionInput{
					Field:    "semester_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{semester},
				},
			},
		}, search.Filters...),
	}

	teachers, err := c.user.GetTeachersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.TeacherListResponse{
		Total: teachers.GetTotal(),
		Data:  convert.PbTeachersToModel(teachers.GetTeachers()),
	}, nil
}

// GetDepartmentStudents returns all students in the department
func (c *Controller) GetDepartmentStudents(ctx context.Context, search model.SearchRequestInput) (*model.StudentListResponse, error) {
	majorCodes, err := c.GetAllMajorByMajorCode(ctx)
	if err != nil {
		return nil, err
	}
	semester, ok := ctx.Value("semester").(string)
	if !ok {
		return nil, fmt.Errorf("no semester")
	}

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorIn,
					Values:   majorCodes,
				},
			},
			{
				Condition: &model.FilterConditionInput{
					Field:    "semester_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{semester},
				},
			},
		}, search.Filters...),
	}

	students, err := c.user.GetStudentsBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.StudentListResponse{
		Total: students.GetTotal(),
		Data:  convert.PbStudentsToModel(students.GetStudents()),
	}, nil
}

// ============================================
// QUERY METHODS - Academic Management
// ============================================

// GetDepartmentSemesters returns all semesters
func (c *Controller) GetDepartmentSemesters(ctx context.Context, search model.SearchRequestInput) (*model.SemesterListResponse, error) {
	return nil, nil
	// _, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	// if err != nil {
	// 	return nil, err
	// }
	// if !check {
	// 	return nil, fmt.Errorf("no department lecturer role")
	// }
	// semesters, err := c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(search))
	// if err != nil {
	// 	return nil, err
	// }

	// return &model.SemesterListResponse{
	// 	Total: semesters.GetTotal(),
	// 	Data:  convert.PbSemestersToModel(semesters.GetSemesters()),
	// }, nil
}

// GetDepartmentMajors returns all majors in the department
func (c *Controller) GetDepartmentMajors(ctx context.Context, search model.SearchRequestInput) (*model.MajorListResponse, error) {
	majorCodes, err := c.GetAllMajorByMajorCode(ctx)
	if err != nil {
		return nil, err
	}
	semester, ok := ctx.Value("semester").(string)
	if !ok {
		return nil, fmt.Errorf("no semester")
	}
	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "id",
					Operator: model.FilterOperatorIn,
					Values:   majorCodes,
				},
			},
			{
				Condition: &model.FilterConditionInput{
					Field:    "semester_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{semester},
				},
			},
		}, search.Filters...),
	}

	majors, err := c.academic.GetMajorsBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.MajorListResponse{
		Total: majors.GetTotal(),
		Data:  convert.PbMajorsToModel(majors.GetMajors()),
	}, nil
}

// GetDepartmentFaculties returns all faculties
func (c *Controller) GetDepartmentFaculties(ctx context.Context, search model.SearchRequestInput) (*model.FacultyListResponse, error) {
	// TODO: Implement when GetFacultiesBySearch is available in gRPC client
	return nil, nil

	// majorCode := teacher.GetTeacher().GetMajorCode()
	// semesterCode := teacher.GetTeacher().GetSemesterCode()
	// newSearch := model.SearchRequestInput{
	// 	Pagination: search.Pagination,
	// 	Filters: []*model.FilterCriteriaInput{
	// 		{
	// 			Condition: &model.FilterConditionInput{
	// 				Field:    "major_code",
	// 				Operator: model.FilterOperatorEqual,
	// 				Values:   []string{majorCode},
	// 			},
	// 		},
	// 		{
	// 			Condition: &model.FilterConditionInput{
	// 				Field:    "semester_code",
	// 				Operator: model.FilterOperatorEqual,
	// 				Values:   []string{semesterCode},
	// 			},
	// 		},
	// 	},
	// }
	// faculties, err := c.academic.GetFacultiesBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	// if err != nil {
	// 	return nil, err
	// }
	// return &model.FacultyListResponse{
	// 	Total: faculties.GetTotal(),
	// 	Data:  convert.PbFacultiesToModel(faculties.GetFaculties()),
	// }, nil
}

// ============================================
// QUERY METHODS - Topic Management
// ============================================

// GetDepartmentTopics returns all topics in the department
func (c *Controller) GetDepartmentTopics(ctx context.Context, search model.SearchRequestInput) (*model.TopicListResponse, error) {
	majorCodes, err := c.GetAllMajorByMajorCode(ctx)
	if err != nil {
		return nil, err
	}
	semester, ok := ctx.Value("semester").(string)
	if !ok {
		return nil, fmt.Errorf("no semester")
	}
	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorIn,
					Values:   majorCodes,
				},
			},
			{
				Condition: &model.FilterConditionInput{
					Field:    "semester_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{semester},
				},
			},
		}, search.Filters...),
	}

	topics, err := c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.TopicListResponse{
		Total: topics.GetTotal(),
		Data:  convert.PbTopicsToModel(topics.GetTopics()),
	}, nil
}

// GetDepartmentTopicDetail returns topic detail by ID
func (c *Controller) GetDepartmentTopicDetail(ctx context.Context, id string) (*model.Topic, error) {
	return nil, nil
	// myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	// if err != nil {
	// 	return nil, err
	// }
	// if !check {
	// 	return nil, fmt.Errorf("no department lecturer role")
	// }
	// teacher, err := c.user.GetTeacherById(ctx, *myId)
	// if err != nil {
	// 	return nil, err
	// }
	// topic, err := c.thesis.GetTopicById(ctx, id)
	// if err != nil {
	// 	return nil, err
	// }
	// return convert.PbTopicToModel(topic.GetTopic()), nil
}

// ============================================
// QUERY METHODS - Enrollment Management
// ============================================

// GetDepartmentEnrollments returns all enrollments in the department
func (c *Controller) GetDepartmentEnrollments(ctx context.Context, search model.SearchRequestInput) (*model.EnrollmentListResponse, error) {
	majorCodes, err := c.GetAllMajorByMajorCode(ctx)
	if err != nil {
		return nil, err
	}
	semester, ok := ctx.Value("semester").(string)
	if !ok {
		return nil, fmt.Errorf("no semester")
	}
	// Get all topics in department first, then filter enrollments by topic
	topicSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorIn,
					Values:   majorCodes,
				},
			},
			{
				Condition: &model.FilterConditionInput{
					Field:    "semester_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{semester},
				},
			},
		},
	}

	topics, err := c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(topicSearch))
	if err != nil {
		return nil, err
	}

	// Extract topic IDs to filter enrollments
	topicIDs := make([]string, 0)
	for _, topic := range topics.GetTopics() {
		topicIDs = append(topicIDs, topic.GetId())
	}

	if len(topicIDs) == 0 {
		return &model.EnrollmentListResponse{
			Total: 0,
			Data:  []*model.Enrollment{},
		}, nil
	}

	// Filter enrollments by topic IDs through topic_council
	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.EnrollmentListResponse{
		Total: enrollments.GetTotal(),
		Data:  convert.PbEnrollmentsToModel(enrollments.GetEnrollments()),
	}, nil
}

// GetDepartmentEnrollmentDetail returns enrollment detail by ID
func (c *Controller) GetDepartmentEnrollmentDetail(ctx context.Context, id string) (*model.Enrollment, error) {
	return nil, nil
	// enrollment, err := c.thesis.GetEnrollmentById(ctx, id)
	// if err != nil {
	// 	return nil, err
	// }

	// return convert.PbEnrollmentToModel(enrollment.GetEnrollment()), nil
}

// ============================================
// QUERY METHODS - Council Management
// ============================================

// GetDepartmentCouncils returns all councils in the department
func (c *Controller) GetDepartmentCouncils(ctx context.Context, search model.SearchRequestInput) (*model.CouncilListResponse, error) {
	majorCodes, err := c.GetAllMajorByMajorCode(ctx)
	if err != nil {
		return nil, err
	}
	semester, ok := ctx.Value("semester").(string)
	if !ok {
		return nil, fmt.Errorf("no semester")
	}

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorIn,
					Values:   majorCodes,
				},
			},
			{
				Condition: &model.FilterConditionInput{
					Field:    "semester_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{semester},
				},
			},
		}, search.Filters...),
	}

	councils, err := c.council.GetCouncilBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.CouncilListResponse{
		Total: councils.GetTotal(),
		Data:  convert.PbCouncilsToModel(councils.GetCouncils()),
	}, nil
}

// GetDepartmentCouncilDetail returns council detail by ID
func (c *Controller) GetDepartmentCouncilDetail(ctx context.Context, id string) (*model.Council, error) {
	return nil, nil
	// council, err := c.council.GetCouncilById(ctx, id)
	// if err != nil {
	// 	return nil, err
	// }

	// return convert.PbCouncilToModel(council.GetCouncil()), nil
}

// GetDepartmentDefences returns all defences of a council
func (c *Controller) GetDepartmentDefences(ctx context.Context, councilID string) (*model.DefenceListResponse, error) {
	return nil, nil
	// newSearch := model.SearchRequestInput{
	// 	Pagination: c.DefaultPagination(),
	// 	Filters: []*model.FilterCriteriaInput{
	// 		{
	// 			Condition: &model.FilterConditionInput{
	// 				Field:    "council_code",
	// 				Operator: model.FilterOperatorEqual,
	// 				Values:   []string{councilID},
	// 			},
	// 		},
	// 	},
	// }

	// defences, err := c.council.GetDefencesBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	// if err != nil {
	// 	return nil, err
	// }

	// return &model.DefenceListResponse{
	// 	Total: defences.GetTotal(),
	// 	Data:  convert.PbDefencesToModel(defences.GetDefences()),
	// }, nil
}

// GetDepartmentGradeDefences returns all grade defences
func (c *Controller) GetDepartmentGradeDefences(ctx context.Context, search model.SearchRequestInput) (*model.GradeDefenceListResponse, error) {

	return nil, nil
}

// ============================================
// MUTATION METHODS - Council Management
// ============================================

// CreateCouncil creates a new council
func (c *Controller) CreateCouncil(ctx context.Context, input model.CreateCouncilInput) (*model.Council, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("no department lecturer role")
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}
	semester, ok := ctx.Value("semester").(string)
	if !ok {
		return nil, fmt.Errorf("no semester")
	}

	resp, err := c.council.CreateCouncil(ctx, &pbCouncil.CreateCouncilRequest{
		Title:        input.Title,
		MajorCode:    input.MajorCode,
		SemesterCode: semester,
		CreatedBy:    createdBy,
	})
	if err != nil {
		return nil, err
	}

	return convert.PbCouncilToModel(resp.GetCouncil()), nil
}

// UpdateDepartmentCouncil updates a council
func (c *Controller) UpdateDepartmentCouncil(ctx context.Context, id string, input model.UpdateCouncilInput) (*model.Council, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("no department lecturer role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	council, err := c.council.GetCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return nil, fmt.Errorf("council already approved")
	}

	req := &pbCouncil.UpdateCouncilRequest{
		Id:        id,
		UpdatedBy: updatedBy,
	}

	if input.Title != nil {
		req.Title = input.Title
	}

	resp, err := c.council.UpdateCouncil(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbCouncilToModel(resp.GetCouncil()), nil
}

// AddDefenceToCouncil adds a defence member to council
func (c *Controller) AddDefenceToCouncil(ctx context.Context, input model.CreateDefenceInput) (*model.Defence, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("no department lecturer role")
	}

	createdBy := ""
	if userInfo != nil {
		createdBy = *userInfo
	}
	council, err := c.council.GetCouncilById(ctx, input.CouncilCode)
	if err != nil {
		return nil, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return nil, fmt.Errorf("council already approved")
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
		return false, fmt.Errorf("no department lecturer role")
	}
	defence, err := c.council.GetDefenceById(ctx, id)
	if err != nil {
		return false, err
	}
	council, err := c.council.GetCouncilById(ctx, defence.GetDefence().GetCouncilCode())
	if err != nil {
		return false, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return false, fmt.Errorf("council already approved")
	}

	_, err = c.council.DeleteDefence(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ============================================
// MUTATION METHODS - Topic Management
// ============================================

// ApproveTopicStage1 approves topic stage 1 by updating status to APPROVED_1
func (c *Controller) ApproveTopicStage1(ctx context.Context, id string) (*model.Topic, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("no department lecturer role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	// Update topic status to APPROVED_1
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

// RejectTopicStage1 rejects topic stage 1 by updating status to REJECTED
func (c *Controller) RejectTopicStage1(ctx context.Context, id string, reason *string) (*model.Topic, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("no department lecturer role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}

	// Update topic status to REJECTED
	status := pbThesis.TopicStatus_REJECTED
	req := &pbThesis.UpdateTopicRequest{
		Id:        id,
		Status:    &status,
		UpdatedBy: updatedBy,
	}

	// TODO: Handle reason field if it exists in Topic model
	// The reason might need to be stored in a different field or table

	resp, err := c.thesis.UpdateTopic(ctx, req)
	if err != nil {
		return nil, err
	}

	return convert.PbTopicToModel(resp.GetTopic()), nil
}

// AssignTopicToCouncil assigns a topic to council
func (c *Controller) AssignTopicToCouncil(ctx context.Context, topicCouncilID string, councilID string) (*model.TopicCouncil, error) {
	userInfo, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("no department lecturer role")
	}

	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}
	council, err := c.council.GetCouncilById(ctx, councilID)
	if err != nil {
		return nil, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return nil, fmt.Errorf("council already approved")
	}

	// Use UpdateTopicCouncil to assign council_code
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
		return false, fmt.Errorf("no department lecturer role")
	}
	updatedBy := ""
	if userInfo != nil {
		updatedBy = *userInfo
	}
	council, err := c.council.GetCouncilById(ctx, councilID)
	if err != nil {
		return false, err
	}
	if council.GetCouncil().GetTimeStart() != nil {
		return false, fmt.Errorf("council already approved")
	}
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, topicCouncilID)
	if err != nil {
		return false, err
	}
	if topicCouncil.GetTopicCouncil().GetCouncilCode() != councilID {
		return false, fmt.Errorf("topic council not found in council")
	}
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
