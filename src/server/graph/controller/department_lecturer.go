package controller

import (
	"context"
	"thaily/src/server/graph/convert"
	convert2 "thaily/src/server/graph/convert"
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

// GetDepartmentTeachers returns all teachers in the department
func (c *Controller) GetDepartmentTeachers(ctx context.Context, search model.SearchRequestInput) ([]*model.Teacher, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get teacher's major to filter by department
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	majorCode := teacher.GetTeacher().GetMajorCode()

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{majorCode},
				},
			},
		}, search.Filters...),
	}

	teachers, err := c.user.GetTeachersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbTeachersToModel(teachers.GetTeachers()), nil
}

// GetDepartmentStudents returns all students in the department
func (c *Controller) GetDepartmentStudents(ctx context.Context, search model.SearchRequestInput) ([]*model.Student, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get teacher's major to filter by department
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	majorCode := teacher.GetTeacher().GetMajorCode()

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{majorCode},
				},
			},
		}, search.Filters...),
	}

	students, err := c.user.GetStudentsBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbStudentsToModel(students.GetStudents()), nil
}

// ============================================
// QUERY METHODS - Academic Management
// ============================================

// GetDepartmentSemesters returns all semesters
func (c *Controller) GetDepartmentSemesters(ctx context.Context, search model.SearchRequestInput) ([]*model.Semester, error) {
	semesters, err := c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return convert.PbSemestersToModel(semesters.GetSemesters()), nil
}

// GetDepartmentMajors returns all majors in the department
func (c *Controller) GetDepartmentMajors(ctx context.Context, search model.SearchRequestInput) ([]*model.Major, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get teacher's major to filter by department
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	majorCode := teacher.GetTeacher().GetMajorCode()

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "id",
					Operator: model.FilterOperatorEqual,
					Values:   []string{majorCode},
				},
			},
		}, search.Filters...),
	}

	majors, err := c.academic.GetMajorsBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbMajorsToModel(majors.GetMajors()), nil
}

// GetDepartmentFaculties returns all faculties
func (c *Controller) GetDepartmentFaculties(ctx context.Context, search model.SearchRequestInput) ([]*model.Faculty, error) {
	// TODO: Implement when GetFacultiesBySearch is available in gRPC client
	return []*model.Faculty{}, nil
}

// ============================================
// QUERY METHODS - Topic Management
// ============================================

// GetDepartmentTopics returns all topics in the department
func (c *Controller) GetDepartmentTopics(ctx context.Context, search model.SearchRequestInput) ([]*model.Topic, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get teacher's major to filter by department
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	majorCode := teacher.GetTeacher().GetMajorCode()

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{majorCode},
				},
			},
		}, search.Filters...),
	}

	topics, err := c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicsToModel(topics.GetTopics()), nil
}

// GetDepartmentTopicDetail returns topic detail by ID
func (c *Controller) GetDepartmentTopicDetail(ctx context.Context, id string) (*model.Topic, error) {
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicToModel(topic.GetTopic()), nil
}

// ============================================
// QUERY METHODS - Enrollment Management
// ============================================

// GetDepartmentEnrollments returns all enrollments in the department
func (c *Controller) GetDepartmentEnrollments(ctx context.Context, search model.SearchRequestInput) ([]*model.Enrollment, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get teacher's major to filter by department
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	majorCode := teacher.GetTeacher().GetMajorCode()

	// Get all topics in department first, then filter enrollments by topic
	topicSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{majorCode},
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
		return []*model.Enrollment{}, nil
	}

	// Filter enrollments by topic IDs through topic_council
	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return convert2.PbEnrollmentsToModel(enrollments.GetEnrollments()), nil
}

// GetDepartmentEnrollmentDetail returns enrollment detail by ID
func (c *Controller) GetDepartmentEnrollmentDetail(ctx context.Context, id string) (*model.Enrollment, error) {
	enrollment, err := c.thesis.GetEnrollmentById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert2.PbEnrollmentToModel(enrollment.GetEnrollment()), nil
}

// ============================================
// QUERY METHODS - Council Management
// ============================================

// GetDepartmentCouncils returns all councils in the department
func (c *Controller) GetDepartmentCouncils(ctx context.Context, search model.SearchRequestInput) ([]*model.Council, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get teacher's major to filter by department
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	majorCode := teacher.GetTeacher().GetMajorCode()

	// Add major filter to search
	newSearch := model.SearchRequestInput{
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "major_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{majorCode},
				},
			},
		}, search.Filters...),
	}

	councils, err := c.council.GetCouncilBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbCouncilsToModel(councils.GetCouncils()), nil
}

// GetDepartmentCouncilDetail returns council detail by ID
func (c *Controller) GetDepartmentCouncilDetail(ctx context.Context, id string) (*model.Council, error) {
	council, err := c.council.GetCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert.PbCouncilToModel(council.GetCouncil()), nil
}

// GetDepartmentDefences returns all defences of a council
func (c *Controller) GetDepartmentDefences(ctx context.Context, councilID string) ([]*model.Defence, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "council_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{councilID},
				},
			},
		},
	}

	defences, err := c.council.GetDefencesBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbDefencesToModel(defences.GetDefences()), nil
}

// GetDepartmentGradeDefences returns all grade defences
func (c *Controller) GetDepartmentGradeDefences(ctx context.Context, search model.SearchRequestInput) ([]*model.GradeDefence, error) {
	gradeDefences, err := c.council.GetGradeDefenceBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return convert.PbGradeDefencesToModel(gradeDefences.GetGradeDefences()), nil
}

// ============================================
// MUTATION METHODS - Council Management
// ============================================

// CreateCouncil creates a new council
func (c *Controller) CreateCouncil(ctx context.Context, input model.CreateCouncilInput) (*model.Council, error) {
	// TODO: Implement when CreateCouncil is available in gRPC client
	return nil, nil
}

// UpdateDepartmentCouncil updates a council
func (c *Controller) UpdateDepartmentCouncil(ctx context.Context, id string, input model.UpdateCouncilInput) (*model.Council, error) {
	// TODO: Implement when UpdateCouncil method signature matches
	return nil, nil
}

// AddDefenceToCouncil adds a defence member to council
func (c *Controller) AddDefenceToCouncil(ctx context.Context, input model.CreateDefenceInput) (*model.Defence, error) {
	// TODO: Implement when CreateDefence is available in gRPC client
	return nil, nil
}

// RemoveDefenceFromCouncil removes a defence member from council
func (c *Controller) RemoveDefenceFromCouncil(ctx context.Context, id string) (bool, error) {
	_, err := c.council.DeleteDefence(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ============================================
// MUTATION METHODS - Topic Management
// ============================================

// ApproveTopicStage1 approves topic stage 1
func (c *Controller) ApproveTopicStage1(ctx context.Context, id string) (*model.Topic, error) {
	// TODO: Implement when ApproveTopicStage1 is available in gRPC client
	return nil, nil
}

// RejectTopicStage1 rejects topic stage 1
func (c *Controller) RejectTopicStage1(ctx context.Context, id string, reason *string) (*model.Topic, error) {
	// TODO: Implement when RejectTopicStage1 is available in gRPC client
	return nil, nil
}

// AssignTopicToCouncil assigns a topic to council
func (c *Controller) AssignTopicToCouncil(ctx context.Context, topicCouncilID string, councilID string) (*model.TopicCouncil, error) {
	// TODO: Implement when AssignTopicToCouncil is available in gRPC client
	return nil, nil
}
