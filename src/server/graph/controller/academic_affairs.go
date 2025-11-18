package controller

import (
	"context"
	"fmt"
	"thaily/src/server/graph/convert"
	convert2 "thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
	"time"
)

// ============================================
// ACADEMIC AFFAIRS CONTROLLER
// Giáo vụ có quyền đọc và chỉnh sửa hầu hết các thông tin trong hệ thống
// ============================================

// ============================================
// QUERY METHODS - User Management
// ============================================

// GetListTeachers returns all teachers with pagination
func (c *Controller) GetListTeachers(ctx context.Context, search model.SearchRequestInput) (*model.TeacherListResponse, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	teachers, err := c.user.GetTeachersBySearch(ctx, c.ConvertSearchRequestToPB(search))
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
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	students, err := c.user.GetStudentsBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.StudentListResponse{
		Total: students.GetTotal(),
		Data:  convert.PbStudentsToModel(students.GetStudents()),
	}, nil
}

// GetStudentDetail returns student detail by ID
func (c *Controller) GetStudentDetail(ctx context.Context, id string) (*model.Student, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	student, err := c.user.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert.PbStudentToModel(student.GetStudent()), nil
}

// GetTeacherDetail returns teacher detail by ID
func (c *Controller) GetTeacherDetail(ctx context.Context, id string) (*model.Teacher, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	teacher, err := c.user.GetTeacherById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert.PbTeacherToModel(teacher.GetTeacher()), nil
}

// ============================================
// QUERY METHODS - Academic Management
// ============================================

// GetAllSemesters returns all semesters with pagination
func (c *Controller) GetAllSemesters(ctx context.Context, search model.SearchRequestInput) (*model.SemesterListResponse, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	semesters, err := c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.SemesterListResponse{
		Total: semesters.GetTotal(),
		Data:  convert.PbSemestersToModel(semesters.GetSemesters()),
	}, nil
}

// GetAllMajors returns all majors with pagination
func (c *Controller) GetAllMajors(ctx context.Context, search model.SearchRequestInput) (*model.MajorListResponse, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	majors, err := c.academic.GetMajorsBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.MajorListResponse{
		Total: majors.GetTotal(),
		Data:  convert.PbMajorsToModel(majors.GetMajors()),
	}, nil
}

// GetAllFaculties returns all faculties with pagination
func (c *Controller) GetAllFaculties(ctx context.Context, search model.SearchRequestInput) (*model.FacultyListResponse, error) {
	// TODO: Implement when GetFacultiesBySearch is available in gRPC client
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	faculties, err := c.academic.GetFacultiesBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}
	return &model.FacultyListResponse{
		Total: faculties.GetTotal(),
		Data:  convert.PbFacultiesToModel(faculties.GetFaculties()),
	}, nil
}

// ============================================
// QUERY METHODS - Topic Management
// ============================================

// GetAllTopics returns all topics with pagination
func (c *Controller) GetAllTopics(ctx context.Context, search model.SearchRequestInput) (*model.TopicListResponse, error) {
	_, check, err := c.RbacInfo(ctx, model.RoleSystemRoleAcademicAffairsStaff)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("No academic affairs staff role")
	}
	_, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if role == nil || (*role) != "teacher" {
		return nil, fmt.Errorf("No teacher role %s", *role)
	}

	topics, err := c.thesis.GetTopicBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.TopicListResponse{
		Total: topics.GetTotal(),
		Data:  convert2.PbTopicsToModel(topics.GetTopics()),
	}, nil
}

// GetTopicDetail returns topic detail by ID
func (c *Controller) GetTopicDetail(ctx context.Context, id string) (*model.Topic, error) {
	_, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if role == nil || (*role) == "teacher" {
		return nil, fmt.Errorf("No teacher")
	}
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert2.PbTopicToModel(topic.GetTopic()), nil
}

// ============================================
// QUERY METHODS - Enrollment Management
// ============================================

// GetAllEnrollments returns all enrollments with pagination
func (c *Controller) GetAllEnrollments(ctx context.Context, search model.SearchRequestInput) (*model.EnrollmentListResponse, error) {
	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.EnrollmentListResponse{
		Total: enrollments.GetTotal(),
		Data:  convert2.PbEnrollmentsToModel(enrollments.GetEnrollments()),
	}, nil
}

// GetEnrollmentDetail returns enrollment detail by ID
func (c *Controller) GetEnrollmentDetail(ctx context.Context, id string) (*model.Enrollment, error) {
	enrollment, err := c.thesis.GetEnrollmentById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert2.PbEnrollmentToModel(enrollment.GetEnrollment()), nil
}

// ============================================
// QUERY METHODS - Council Management
// ============================================

// GetAllCouncils returns all councils with pagination
func (c *Controller) GetAllCouncils(ctx context.Context, search model.SearchRequestInput) (*model.CouncilListResponse, error) {
	councils, err := c.council.GetCouncilBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.CouncilListResponse{
		Total: councils.GetTotal(),
		Data:  convert.PbCouncilsToModel(councils.GetCouncils()),
	}, nil
}

// GetCouncilDetail returns council detail by ID
func (c *Controller) GetCouncilDetail(ctx context.Context, id string) (*model.Council, error) {
	council, err := c.council.GetCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}

	return convert.PbCouncilToModel(council.GetCouncil()), nil
}

// GetDefencesByCouncil returns all defences of a council
func (c *Controller) GetDefencesByCouncil(ctx context.Context, councilID string) (*model.DefenceListResponse, error) {
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

	return &model.DefenceListResponse{
		Total: defences.GetTotal(),
		Data:  convert.PbDefencesToModel(defences.GetDefences()),
	}, nil
}

// GetAllGradeDefences returns all grade defences with pagination
func (c *Controller) GetAllGradeDefences(ctx context.Context, search model.SearchRequestInput) (*model.GradeDefenceListResponse, error) {
	gradeDefences, err := c.council.GetGradeDefenceBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.GradeDefenceListResponse{
		Total: gradeDefences.GetTotal(),
		Data:  convert.PbGradeDefencesToModel(gradeDefences.GetGradeDefences()),
	}, nil
}

// ============================================
// MUTATION METHODS - User Management
// ============================================

// CreateTeacher creates a new teacher
func (c *Controller) CreateTeacher(ctx context.Context, input model.CreateTeacherInput) (*model.Teacher, error) {
	// TODO: Implement when CreateTeacher is available in gRPC client
	return nil, nil
}

// UpdateTeacher updates a teacher
func (c *Controller) UpdateTeacher(ctx context.Context, id string, input model.UpdateTeacherInput) (*model.Teacher, error) {
	// TODO: Implement when UpdateTeacher is available in gRPC client
	return nil, nil
}

// DeleteTeacher deletes a teacher
func (c *Controller) DeleteTeacher(ctx context.Context, id string) (bool, error) {
	// TODO: Implement when DeleteTeacher is available in gRPC client
	return false, nil
}

// CreateStudent creates a new student
func (c *Controller) CreateStudent(ctx context.Context, input model.CreateStudentInput) (*model.Student, error) {
	// TODO: Implement when CreateStudent is available in gRPC client
	return nil, nil
}

// UpdateStudent updates a student
func (c *Controller) UpdateStudent(ctx context.Context, id string, input model.UpdateStudentInput) (*model.Student, error) {
	// TODO: Implement when UpdateStudent is available in gRPC client
	return nil, nil
}

// DeleteStudent deletes a student
func (c *Controller) DeleteStudent(ctx context.Context, id string) (bool, error) {
	// TODO: Implement when DeleteStudent is available in gRPC client
	return false, nil
}

// ============================================
// MUTATION METHODS - Academic Management
// ============================================

// CreateSemester creates a new semester
func (c *Controller) CreateSemester(ctx context.Context, input model.CreateSemesterInput) (*model.Semester, error) {
	// TODO: Implement when CreateSemester is available in gRPC client
	return nil, nil
}

// UpdateSemester updates a semester
func (c *Controller) UpdateSemester(ctx context.Context, id string, input model.UpdateSemesterInput) (*model.Semester, error) {
	// TODO: Implement when UpdateSemester is available in gRPC client
	return nil, nil
}

// DeleteSemester deletes a semester
func (c *Controller) DeleteSemester(ctx context.Context, id string) (bool, error) {
	// TODO: Implement when DeleteSemester is available in gRPC client
	return false, nil
}

// CreateMajor creates a new major
func (c *Controller) CreateMajor(ctx context.Context, input model.CreateMajorInput) (*model.Major, error) {
	// TODO: Implement when CreateMajor is available in gRPC client
	return nil, nil
}

// UpdateMajor updates a major
func (c *Controller) UpdateMajor(ctx context.Context, id string, input model.UpdateMajorInput) (*model.Major, error) {
	// TODO: Implement when UpdateMajor is available in gRPC client
	return nil, nil
}

// DeleteMajor deletes a major
func (c *Controller) DeleteMajor(ctx context.Context, id string) (bool, error) {
	// TODO: Implement when DeleteMajor is available in gRPC client
	return false, nil
}

// CreateFaculty creates a new faculty
func (c *Controller) CreateFaculty(ctx context.Context, input model.CreateFacultyInput) (*model.Faculty, error) {
	// TODO: Implement when CreateFaculty is available in gRPC client
	return nil, nil
}

// UpdateFaculty updates a faculty
func (c *Controller) UpdateFaculty(ctx context.Context, id string, input model.UpdateFacultyInput) (*model.Faculty, error) {
	// TODO: Implement when UpdateFaculty is available in gRPC client
	return nil, nil
}

// DeleteFaculty deletes a faculty
func (c *Controller) DeleteFaculty(ctx context.Context, id string) (bool, error) {
	// TODO: Implement when DeleteFaculty is available in gRPC client
	return false, nil
}

// ============================================
// MUTATION METHODS - Council Management
// ============================================

// ApproveCouncil approves a council and sets time start
func (c *Controller) ApproveCouncil(ctx context.Context, id string, timeStart time.Time) (*model.Council, error) {
	// TODO: Implement when ApproveCouncil is available in gRPC client
	return nil, nil
}

// UpdateCouncil updates a council
func (c *Controller) UpdateCouncil(ctx context.Context, id string, input model.UpdateCouncilInput) (*model.Council, error) {
	// TODO: Implement when UpdateCouncil method signature matches
	return nil, nil
}

// DeleteCouncil deletes a council
func (c *Controller) DeleteCouncil(ctx context.Context, id string) (bool, error) {
	_, err := c.council.DeleteCouncil(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ============================================
// MUTATION METHODS - Topic Management
// ============================================

// ApproveTopic approves topic stage 2 (final approval)
func (c *Controller) ApproveTopic(ctx context.Context, id string) (*model.Topic, error) {
	// TODO: Implement when ApproveTopic is available in gRPC client
	return nil, nil
}

// RejectTopic rejects a topic
func (c *Controller) RejectTopic(ctx context.Context, id string, reason *string) (*model.Topic, error) {
	// TODO: Implement when RejectTopic is available in gRPC client
	return nil, nil
}

// UpdateTopic updates a topic
func (c *Controller) UpdateTopic(ctx context.Context, id string, input model.UpdateTopicInput) (*model.Topic, error) {
	// TODO: Implement when UpdateTopic is available in gRPC client
	return nil, nil
}

// DeleteTopic deletes a topic
func (c *Controller) DeleteTopic(ctx context.Context, id string) (bool, error) {
	// TODO: Implement when DeleteTopic is available in gRPC client
	return false, nil
}
