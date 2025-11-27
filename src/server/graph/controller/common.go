package controller

import (
	"context"
	"fmt"

	pbRole "thaily/proto/role"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// COMMON GETTERS - Used by multiple resolvers
// ============================================

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

// GetSemesterById returns a semester by ID
func (c *Controller) GetSemesterById(ctx context.Context, id string) (*model.Semester, error) {
	semester, err := c.academic.GetSemesterById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbSemesterToModel(semester.GetSemester()), nil
}

// GetMajorById returns a major by ID
func (c *Controller) GetMajorById(ctx context.Context, id string) (*model.Major, error) {
	major, err := c.academic.GetMajorById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbMajorToModel(major.GetMajor()), nil
}

// GetFacultyById returns a faculty by ID
func (c *Controller) GetFacultyById(ctx context.Context, id string) (*model.Faculty, error) {
	faculty, err := c.academic.GetFacultyById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbFacultyToModel(faculty.GetFaculty()), nil
}

// GetTopicById returns a topic by ID
func (c *Controller) GetTopicById(ctx context.Context, id string) (*model.Topic, error) {
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTopicToModel(topic.GetTopic()), nil
}

// GetEnrollmentById returns an enrollment by ID
func (c *Controller) GetEnrollmentById(ctx context.Context, id string) (*model.Enrollment, error) {
	enrollment, err := c.thesis.GetEnrollmentById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbEnrollmentToModel(enrollment.GetEnrollment()), nil
}

// GetCouncilById returns a council by ID
func (c *Controller) GetCouncilById(ctx context.Context, id string) (*model.Council, error) {
	council, err := c.council.GetCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbCouncilToModel(council.GetCouncil()), nil
}

// GetDefenceById returns a defence by ID
func (c *Controller) GetDefenceById(ctx context.Context, id string) (*model.Defence, error) {
	defence, err := c.council.GetDefenceById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbDefenceToModel(defence.GetDefence()), nil
}

// GetTopicCouncilById returns a topic council by ID
func (c *Controller) GetTopicCouncilById(ctx context.Context, id string) (*model.TopicCouncil, error) {
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	fmt.Print("xxxxxxxxxxxxxxxxxx", topicCouncil)
	if err != nil {
		return nil, err
	}
	return convert.PbTopicCouncilToModel(topicCouncil.GetTopicCouncil()), nil
}

// GetMidtermById returns a midterm by ID
func (c *Controller) GetMidtermById(ctx context.Context, id string) (*model.Midterm, error) {
	midterm, err := c.thesis.GetMidtermById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbMidtermToModel(midterm.GetMidterm()), nil
}

// GetFinalById returns a final by ID
func (c *Controller) GetFinalById(ctx context.Context, id string) (*model.Final, error) {
	final, err := c.thesis.GetFinalById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbFinalToModel(final.GetFinal()), nil
}

// GetGradeReviewById returns a grade review by ID

// GetDefencesByCouncil returns all defences for a council
func (c *Controller) GetDefencesByCouncil(ctx context.Context, councilID string) (*model.DefenceListResponse, error) {
	search := model.SearchRequestInput{
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

	defences, err := c.council.GetDefencesBySearch(ctx, c.ConvertSearchRequestToPB(search))
	if err != nil {
		return nil, err
	}

	return &model.DefenceListResponse{
		Total: defences.GetTotal(),
		Data:  convert.PbDefencesToModel(defences.GetDefences()),
	}, nil
}

// ============================================
// ROLE HELPERS
// ============================================

// GetRole returns all roles for a teacher
func (c *Controller) GetRole(ctx context.Context, teacherId string) ([]*model.RoleSystem, error) {
	roles, err := c.role.GetAllRoleByTeacherId(ctx, teacherId)
	if err != nil {
		return nil, err
	}
	return c.pbRoleToModel(roles), nil
}

// GetRolesByTeacherId returns all roles for a teacher (alias)
func (c *Controller) GetRolesByTeacherId(ctx context.Context, teacherId string) ([]*model.RoleSystem, error) {
	return c.GetRole(ctx, teacherId)
}

func (c *Controller) pbRoleToModel(resp *pbRole.ListRoleSystemsResponse) []*model.RoleSystem {
	if resp == nil {
		return nil
	}

	roles := resp.GetRoleSystems()
	result := make([]*model.RoleSystem, 0, len(roles))

	for _, r := range roles {
		var role model.RoleSystemRole
		switch r.Role {
		case pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF:
			role = model.RoleSystemRoleAcademicAffairsStaff
		case pbRole.RoleType_DEPARTMENT_LECTURER:
			role = model.RoleSystemRoleDepartmentLecturer
		case pbRole.RoleType_TEACHER:
			role = model.RoleSystemRoleTeacher
		default:
			role = model.RoleSystemRoleTeacher
		}

		teacherCode := r.TeacherCode
		result = append(result, &model.RoleSystem{
			ID:          r.Id,
			TeacherCode: &teacherCode,
			Role:        role,
		})
	}

	return result
}

// ============================================
// INFO GETTERS (for DataLoader batch functions)
// ============================================

// GetMajorInfo returns major info by ID
func (c *Controller) GetMajorInfo(ctx context.Context, id string) (*model.MajorInfo, error) {
	major, err := c.academic.GetMajorById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbMajorToMajorInfo(major.GetMajor()), nil
}

// GetSemesterInfo returns semester info by ID
func (c *Controller) GetSemesterInfo(ctx context.Context, id string) (*model.SemesterInfo, error) {
	semester, err := c.academic.GetSemesterById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbSemesterToSemesterInfo(semester.GetSemester()), nil
}
