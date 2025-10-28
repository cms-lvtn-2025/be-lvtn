package controller

import (
	"context"
	"thaily/src/server/graph/convert"
	convert2 "thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// STUDENT CONTROLLER - All student-related methods
// ============================================

// ============================================
// STUDENT PROFILE & AUTH
// ============================================

// GetStudentByRequest returns current logged-in student profile
func (c *Controller) GetStudentByRequest(ctx context.Context) (*model.Student, error) {
	myId, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	if myId == nil || role == nil {
		return nil, nil
	}
	student, err := c.user.GetUserById(ctx, *myId)
	if err != nil {
		return nil, err
	}
	return convert.PbStudentToModel(student.GetStudent()), nil
}

// ============================================
// STUDENT - ENROLLMENT METHODS
// ============================================

// GetEnrollmentsForStudent returns all enrollments for current student
func (c *Controller) GetEnrollmentsForStudent(ctx context.Context, search *model.SearchRequestInput) (*model.StudentEnrollmentListResponse, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	var newSearch model.SearchRequestInput
	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters: append([]*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "student_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{*myId},
					},
				},
			}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters: []*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "student_code",
						Operator: model.FilterOperatorEqual,
						Values:   []string{*myId},
					},
				},
			},
		}
	}
	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}
	return &model.StudentEnrollmentListResponse{
		Total: enrollments.GetTotal(),
		Data:  convert2.PbEnrollmentsToStudentEnrollment(enrollments.GetEnrollments()),
	}, nil
}

// GetEnrollmentForStudent returns enrollment detail for student
func (c *Controller) GetEnrollmentForStudent(ctx context.Context, id string) (*model.StudentEnrollment, error) {
	_, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	enrollment, err := c.thesis.GetEnrollmentById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbEnrollmentToStudentEnrollment(enrollment.GetEnrollment()), nil
}

// ============================================
// STUDENT - TOPIC & TOPIC COUNCIL METHODS
// ============================================

// GetTopicCouncilForStudent returns topic council for student
func (c *Controller) GetTopicCouncilForStudent(ctx context.Context, id string) (*model.StudentTopicCouncil, error) {
	_, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	topicCouncil, err := c.thesis.GetTopicCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicCouncilToStudentTopicCouncil(topicCouncil.GetTopicCouncil()), nil
}

// GetTopicForStudent returns topic for student
func (c *Controller) GetTopicForStudent(ctx context.Context, id string) (*model.StudentTopic, error) {
	_, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}
	topic, err := c.thesis.GetTopicById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicToStudentTopic(topic.GetTopic()), nil
}

// GetSupervisorByTopicId returns all supervisors for a topic council
func (c *Controller) GetSupervisorByTopicId(ctx context.Context, topicCouncilCode string) ([]*model.StudentTopicSupervisor, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "topic_council_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{topicCouncilCode},
				},
			},
		},
	}
	supervisor, err := c.thesis.GetTopicCouncilSupervisorBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}
	return convert2.PbTopicCouncilSupervisorsToStudentTopicSupervisors(supervisor.GetTopicCouncilSupervisors()), nil
}

// ============================================
// STUDENT - GRADING METHODS (Midterm, Final, GradeReview)
// ============================================

// GetMidtermById, GetFinalById - moved to resolvers_helper.go

// GetGradeViewById returns grade review by ID
func (c *Controller) GetGradeViewById(ctx context.Context, id string) (*model.GradeReview, error) {
	gradeReview, err := c.thesis.GetGradeReviewById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbGradeReviewToModel(gradeReview.GetGradeReview()), nil
}

// ============================================
// STUDENT - COUNCIL & DEFENCE METHODS
// ============================================

// GetCouncilByIdForStudent returns council for student
func (c *Controller) GetCouncilByIdForStudent(ctx context.Context, id string) (*model.StudentCouncil, error) {
	council, err := c.council.GetCouncilById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert2.PbCouncilToStudentCouncil(council.GetCouncil()), nil
}

// GetDefenceInfoByCouncilId returns defence info by council ID
func (c *Controller) GetDefenceInfoByCouncilId(ctx context.Context, id string) (*model.StudentDefenceInfo, error) {
	defence, err := c.council.GetDefenceById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbDefenceToStudentDefenceInfo(defence.GetDefence()), nil
}

// GetDefenceInfoById returns defence info by ID
func (c *Controller) GetDefenceInfoById(ctx context.Context, id string) (*model.StudentDefenceInfo, error) {
	defence, err := c.council.GetDefenceById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbDefenceToStudentDefenceInfo(defence.GetDefence()), nil
}

// GetGradeDefenceInfoByEnrollmentCode returns grade defence by enrollment code
func (c *Controller) GetGradeDefenceInfoByEnrollmentCode(ctx context.Context, enrollmentCode string) ([]*model.StudentGradeDefence, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "enrollment_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{enrollmentCode},
				},
			},
		},
	}
	gradeDefence, err := c.council.GetGradeDefenceBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}
	return convert.PbGradeDefencesToStudentGradeDefences(gradeDefence.GradeDefences), nil
}

// GetGradeDefenceCriterionByDefenceCode returns grade defence criterion by defence code
func (c *Controller) GetGradeDefenceCriterionByDefenceCode(ctx context.Context, defenceCode string) ([]*model.GradeDefenceCriterion, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "grade_defence_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{defenceCode},
				},
			},
		},
	}
	criterion, err := c.council.GetGradeDefenceCriteriaBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}
	return convert.PbGradeDefenceCriteriaToModel(criterion.GradeDefenceCriteria), nil
}

// ============================================
// STUDENT - ACADEMIC METHODS (Semester, Major)
// ============================================

// GetMySemesters returns all semesters for current student
func (c *Controller) GetMySemesters(ctx context.Context, search *model.SearchRequestInput) (*model.SemesterListResponse, error) {
	_, mySemesters, _, err := c.GetInfoAllRequest(ctx)
	if err != nil {
		return nil, err
	}
	var newSearch model.SearchRequestInput
	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters: append([]*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "id",
						Operator: model.FilterOperatorIn,
						Values:   *mySemesters,
					},
				},
			}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters: []*model.FilterCriteriaInput{
				{
					Condition: &model.FilterConditionInput{
						Field:    "id",
						Operator: model.FilterOperatorIn,
						Values:   *mySemesters,
					},
				},
			},
		}
	}
	semesters, err := c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}
	return &model.SemesterListResponse{
		Total: semesters.GetTotal(),
		Data:  convert.PbSemestersToModel(semesters.Semesters),
	}, nil
}

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

// ============================================
// STUDENT - TEACHER INFO (for student view)
// ============================================

// GetTeacherInfoById returns teacher info for student view
func (c *Controller) GetTeacherInfoById(ctx context.Context, id string) (*model.StudentTeacherInfo, error) {
	teacher, err := c.user.GetTeacherById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTeacherToStudentTeacherInfo(teacher.GetTeacher()), nil
}
