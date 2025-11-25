package controller

import (
	"context"
	"fmt"

	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// DEPARTMENT QUERY CONTROLLER
// Handles: DepartmentQuery resolvers (Department Lecturer)
// Scoped to department's faculty majors
// ============================================

// GetDepartmentMajorCodes returns all major codes in current teacher's faculty
func (c *Controller) GetDepartmentMajorCodes(ctx context.Context) ([]string, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not a department lecturer")
	}

	// Get teacher's major
	teacher, err := c.user.GetTeacherById(ctx, *myId)
	if err != nil {
		return nil, err
	}

	majorCode := teacher.GetTeacher().GetMajorCode()
	major, err := c.academic.GetMajorById(ctx, majorCode)
	if err != nil {
		return nil, err
	}

	// Get all majors in same faculty
	search := model.SearchRequestInput{
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

	majors, err := c.academic.GetMajorsBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	majorCodes := make([]string, 0, len(majors.GetMajors()))
	for _, m := range majors.GetMajors() {
		majorCodes = append(majorCodes, m.GetId())
	}

	return majorCodes, nil
}

// ============================================
// USER QUERIES (scoped to department)
// ============================================

// GetDepartmentTeachers returns teachers in department's faculty
func (c *Controller) GetDepartmentTeachers(ctx context.Context, search model.SearchRequestInput) (*model.TeacherListResponse, error) {
	majorCodes, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}

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

// GetDepartmentStudents returns students in department's faculty
func (c *Controller) GetDepartmentStudents(ctx context.Context, search model.SearchRequestInput) (*model.StudentListResponse, error) {
	majorCodes, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}

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
// ACADEMIC QUERIES (scoped to department)
// ============================================

// GetDepartmentSemesters returns semesters (all - no department scope needed)
func (c *Controller) GetDepartmentSemesters(ctx context.Context, search model.SearchRequestInput) (*model.SemesterListResponse, error) {
	semesters, err := c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.SemesterListResponse{
		Total: semesters.GetTotal(),
		Data:  convert.PbSemestersToModel(semesters.Semesters),
	}, nil
}

// GetDepartmentMajors returns majors in department's faculty
func (c *Controller) GetDepartmentMajors(ctx context.Context, search model.SearchRequestInput) (*model.MajorListResponse, error) {
	majorCodes, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}

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

// GetDepartmentFaculties returns department's faculty
func (c *Controller) GetDepartmentFaculties(ctx context.Context, search model.SearchRequestInput) (*model.FacultyListResponse, error) {
	myId, check, err := c.RbacInfo(ctx, model.RoleSystemRoleDepartmentLecturer)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("not a department lecturer")
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
		Pagination: search.Pagination,
		Filters: append([]*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "id",
					Operator: model.FilterOperatorEqual,
					Values:   []string{major.GetMajor().GetFacultyCode()},
				},
			},
		}, search.Filters...),
	}

	faculties, err := c.academic.GetFacultiesBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.FacultyListResponse{
		Total: faculties.GetTotal(),
		Data:  convert.PbFacultiesToModel(faculties.GetFaculties()),
	}, nil
}

// ============================================
// THESIS QUERIES (scoped to department)
// ============================================

// GetDepartmentTopics returns topics in department's faculty
func (c *Controller) GetDepartmentTopics(ctx context.Context, search model.SearchRequestInput) (*model.TopicListResponse, error) {
	majorCodes, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}

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

// GetDepartmentEnrollments returns enrollments in department's faculty
func (c *Controller) GetDepartmentEnrollments(ctx context.Context, search model.SearchRequestInput) (*model.EnrollmentListResponse, error) {
	majorCodes, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}

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
		}, search.Filters...),
	}

	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.EnrollmentListResponse{
		Total: enrollments.GetTotal(),
		Data:  convert.PbEnrollmentsToModel(enrollments.GetEnrollments()),
	}, nil
}

// ============================================
// COUNCIL QUERIES (scoped to department)
// ============================================

// GetDepartmentCouncils returns councils in department's faculty
func (c *Controller) GetDepartmentCouncils(ctx context.Context, search model.SearchRequestInput) (*model.CouncilListResponse, error) {
	majorCodes, err := c.GetDepartmentMajorCodes(ctx)
	if err != nil {
		return nil, err
	}

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

// GetDepartmentGradeDefences returns grade defences in department (placeholder)
func (c *Controller) GetDepartmentGradeDefences(ctx context.Context, search model.SearchRequestInput) (*model.GradeDefenceListResponse, error) {
	// TODO: Implement proper filtering by department scope
	return &model.GradeDefenceListResponse{
		Total: 0,
		Data:  []*model.GradeDefence{},
	}, nil
}
