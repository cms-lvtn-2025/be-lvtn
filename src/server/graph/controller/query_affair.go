package controller

import (
	"context"

	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// AFFAIR QUERY CONTROLLER
// Handles: AffairQuery resolvers (Academic Affairs Staff)
// Full access to all data
// ============================================

// ============================================
// USER MANAGEMENT QUERIES
// ============================================

// GetListTeachers returns all teachers with pagination
func (c *Controller) GetListTeachers(ctx context.Context, search model.SearchRequestInput) (*model.TeacherListResponse, error) {
	teachers, err := c.user.GetTeachersBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.TeacherListResponse{
		Total: teachers.GetTotal(),
		Data:  convert.PbTeachersToModel(teachers.GetTeachers()),
	}, nil
}

// GetListStudents returns all students with pagination
func (c *Controller) GetListStudents(ctx context.Context, search model.SearchRequestInput) (*model.StudentListResponse, error) {
	students, err := c.user.GetStudentsBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.StudentListResponse{
		Total: students.GetTotal(),
		Data:  convert.PbStudentsToModel(students.GetStudents()),
	}, nil
}

// ============================================
// ACADEMIC QUERIES
// ============================================

// GetAllSemesters returns all semesters
func (c *Controller) GetAllSemesters(ctx context.Context, search model.SearchRequestInput) (*model.SemesterListResponse, error) {
	semesters, err := c.academic.GetSemestersBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.SemesterListResponse{
		Total: semesters.GetTotal(),
		Data:  convert.PbSemestersToModel(semesters.Semesters),
	}, nil
}

// GetAllMajors returns all majors
func (c *Controller) GetAllMajors(ctx context.Context, search model.SearchRequestInput) (*model.MajorListResponse, error) {
	majors, err := c.academic.GetMajorsBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.MajorListResponse{
		Total: majors.GetTotal(),
		Data:  convert.PbMajorsToModel(majors.GetMajors()),
	}, nil
}

// GetAllFaculties returns all faculties
func (c *Controller) GetAllFaculties(ctx context.Context, search model.SearchRequestInput) (*model.FacultyListResponse, error) {
	faculties, err := c.academic.GetFacultiesBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.FacultyListResponse{
		Total: faculties.GetTotal(),
		Data:  convert.PbFacultiesToModel(faculties.GetFaculties()),
	}, nil
}

// ============================================
// THESIS QUERIES
// ============================================

// GetAllTopics returns all topics
func (c *Controller) GetAllTopics(ctx context.Context, search model.SearchRequestInput) (*model.TopicListResponse, error) {
	topics, err := c.thesis.GetTopicBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.TopicListResponse{
		Total: topics.GetTotal(),
		Data:  convert.PbTopicsToModel(topics.GetTopics()),
	}, nil
}

// GetAllEnrollments returns all enrollments
func (c *Controller) GetAllEnrollments(ctx context.Context, search model.SearchRequestInput) (*model.EnrollmentListResponse, error) {
	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.EnrollmentListResponse{
		Total: enrollments.GetTotal(),
		Data:  convert.PbEnrollmentsToModel(enrollments.GetEnrollments()),
	}, nil
}

// ============================================
// COUNCIL QUERIES
// ============================================

// GetAllCouncils returns all councils
func (c *Controller) GetAllCouncils(ctx context.Context, search model.SearchRequestInput) (*model.CouncilListResponse, error) {
	councils, err := c.council.GetCouncilBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.CouncilListResponse{
		Total: councils.GetTotal(),
		Data:  convert.PbCouncilsToModel(councils.GetCouncils()),
	}, nil
}

// GetAllGradeDefences returns all grade defences
func (c *Controller) GetAllGradeDefences(ctx context.Context, search model.SearchRequestInput) (*model.GradeDefenceListResponse, error) {
	gradeDefences, err := c.council.GetGradeDefenceBySearch(ctx, convert.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.GradeDefenceListResponse{
		Total: gradeDefences.GetTotal(),
		Data:  convert.PbGradeDefencesToModel(gradeDefences.GetGradeDefences()),
	}, nil
}
