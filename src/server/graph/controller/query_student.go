package controller

import (
	"context"

	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// STUDENT QUERY CONTROLLER
// Handles: StudentQuery resolvers
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

// GetEnrollmentsForStudent returns all enrollments for current student
func (c *Controller) GetEnrollmentsForStudent(ctx context.Context, search *model.SearchRequestInput) (*model.EnrollmentListResponse, error) {
	myId, _, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, err
	}

	var newSearch model.SearchRequestInput
	studentFilter := &model.FilterCriteriaInput{
		Condition: &model.FilterConditionInput{
			Field:    "student_code",
			Operator: model.FilterOperatorEqual,
			Values:   []string{*myId},
		},
	}

	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters:    append([]*model.FilterCriteriaInput{studentFilter}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters:    []*model.FilterCriteriaInput{studentFilter},
		}
	}

	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, convert.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.EnrollmentListResponse{
		Total: enrollments.GetTotal(),
		Data:  convert.PbEnrollmentsToModel(enrollments.GetEnrollments()),
	}, nil
}

// GetMySemesters returns all semesters for current student
func (c *Controller) GetMySemesters(ctx context.Context, search *model.SearchRequestInput) (*model.SemesterListResponse, error) {
	_, mySemesters, _, err := c.GetInfoAllRequest(ctx)
	if err != nil {
		return nil, err
	}

	var newSearch model.SearchRequestInput
	semesterFilter := &model.FilterCriteriaInput{
		Condition: &model.FilterConditionInput{
			Field:    "id",
			Operator: model.FilterOperatorIn,
			Values:   *mySemesters,
		},
	}

	if search != nil {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters:    append([]*model.FilterCriteriaInput{semesterFilter}, search.Filters...),
		}
	} else {
		newSearch = model.SearchRequestInput{
			Pagination: c.DefaultPagination(),
			Filters:    []*model.FilterCriteriaInput{semesterFilter},
		}
	}

	semesters, err := c.academic.GetSemestersBySearch(ctx, convert.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return &model.SemesterListResponse{
		Total: semesters.GetTotal(),
		Data:  convert.PbSemestersToModel(semesters.Semesters),
	}, nil
}

// GetGradeDefenceCriterionByDefenceCode returns grade defence criteria by defence code
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

	criteria, err := c.council.GetGradeDefenceCriteriaBySearch(ctx, convert.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}
	return convert.PbGradeDefenceCriteriaToModel(criteria.GradeDefenceCriteria), nil
}
