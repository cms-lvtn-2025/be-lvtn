package controller

import (
	"context"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// All student-related methods have been moved to controller/student.go

// ============================================
// RESOLVER HELPER METHODS
// ============================================

// GetEnrollmentsByStudentId returns all enrollments for a given student
func (c *Controller) GetEnrollmentsByStudentId(ctx context.Context, studentId string) ([]*model.Enrollment, error) {

	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "student_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{studentId},
				},
			},
		},
	}

	enrollments, err := c.thesis.GetEnrollmentBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbEnrollmentsToModel(enrollments.GetEnrollments()), nil
}
