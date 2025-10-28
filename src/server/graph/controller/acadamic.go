package controller

import (
	"context"
	"fmt"
	"strings"
	pb "thaily/proto/academic"
	pbRole "thaily/proto/role"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/helper"
	"thaily/src/server/graph/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Student-related methods (GetMySemesters, GetMajorInfo, GetSemesterInfo)
// have been moved to controller/student.go

func (c *Controller) pbMajorToModel(resp *pb.GetMajorResponse) *model.Major {
	if resp == nil {
		return nil
	}
	m := resp.GetMajor()
	var createdAt, updatedAt *time.Time
	if m.CreatedAt != nil {
		t := m.CreatedAt.AsTime()
		createdAt = &t
	}
	if m.UpdatedAt != nil {
		t := m.UpdatedAt.AsTime()
		updatedAt = &t
	}
	return &model.Major{
		ID:          m.Id,
		Title:       m.Title,
		FacultyCode: m.FacultyCode,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		CreatedBy:   &m.CreatedBy,
		UpdatedBy:   &m.UpdatedBy,
	}
}

func (c *Controller) pbSemesterToModel(resp *pb.GetSemesterResponse) *model.Semester {
	if resp == nil {
		return nil
	}
	s := resp.GetSemester()
	var createdAt, updatedAt *time.Time
	if s.CreatedAt != nil {
		t := s.CreatedAt.AsTime()
		createdAt = &t

	}
	if s.UpdatedAt != nil {
		t := s.UpdatedAt.AsTime()
		updatedAt = &t
	}
	return &model.Semester{
		ID:        s.Id,
		Title:     s.Title,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		CreatedBy: &s.CreatedBy,
		UpdatedBy: &s.UpdatedBy,
	}
}

func (c *Controller) pbSemestersToModel(resp *pb.ListSemestersResponse) []*model.Semester {
	if resp == nil {
		return nil
	}
	s := resp.GetSemesters()
	var createdAt, updatedAt *time.Time
	results := []*model.Semester{}
	for _, m := range s {
		if m.CreatedAt != nil {
			t := m.CreatedAt.AsTime()
			createdAt = &t
		}
		if m.UpdatedAt != nil {
			t := m.UpdatedAt.AsTime()
			updatedAt = &t
		}
		results = append(results, &model.Semester{
			ID:        m.Id,
			Title:     m.Title,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			CreatedBy: &m.CreatedBy,
			UpdatedBy: &m.UpdatedBy,
		})
	}
	return results
}

func (c *Controller) GetMajorByCode(ctx context.Context, code string) (*model.Major, error) {
	if code == "" {
		return nil, nil
	}

	res, err := c.academic.GetMajorById(ctx, code)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Major == nil {
		return nil, nil
	}
	return c.pbMajorToModel(res), nil
}

func (c *Controller) GetSemesterByCode(ctx context.Context, code string) (*model.Semester, error) {
	if code == "" {
		return nil, nil
	}
	res, err := c.academic.GetSemesterById(ctx, code)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Semester == nil {
		return nil, nil
	}
	return c.pbSemesterToModel(res), nil
}

func (c *Controller) GetSemesters(ctx context.Context, search model.SearchRequestInput) ([]*model.Semester, error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("not authorized")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("not authorized")
	}
	idsArr := strings.Split(claims["ids"].(string), ",")
	Values := []string{}
	semester, ok := ctx.Value("semester").(string)
	myId := ""
	for _, id := range idsArr {
		parts := strings.Split(id, "-")
		if len(parts) == 2 {
			if semester != "" && strings.HasPrefix(id, semester+"-") {
				myId = strings.Split(id, "-")[1]
			}
			Values = append(Values, parts[1])
		}

	}

	var semesters *pb.ListSemestersResponse
	var err error
	var newSearch model.SearchRequestInput
	if role == "student" {
		newSearch = model.SearchRequestInput{
			Pagination: search.Pagination,
			Filters: append([]*model.FilterCriteriaInput{
				&model.FilterCriteriaInput{
					Condition: &model.FilterConditionInput{
						Field:    "id",
						Operator: model.FilterOperatorIn,
						Values:   Values,
					},
				},
			}, search.Filters...),
		}
		semesters, err = c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	} else if role == "teacher" {
		var permissions *pbRole.ListRoleSystemsResponse
		permissions, err = c.role.GetAllRoleByTeacherId(ctx, myId)
		if err != nil || permissions == nil || len(permissions.GetRoleSystems()) == 0 {
			return nil, err
		}
		permissionMap := make(map[pbRole.RoleType]bool)
		for _, permission := range permissions.GetRoleSystems() {
			permissionMap[permission.Role] = true
		}
		if permissionMap[pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF] {
			semesters, err = c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
		} else if permissionMap[pbRole.RoleType_DEPARTMENT_LECTURER] {
			newSearch = model.SearchRequestInput{
				Pagination: search.Pagination,
				Filters: append([]*model.FilterCriteriaInput{
					&model.FilterCriteriaInput{
						Condition: &model.FilterConditionInput{
							Field:    "id",
							Operator: model.FilterOperatorIn,
							Values:   Values,
						},
					},
				}, search.Filters...),
			}
			semesters, err = c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
		} else if permissionMap[pbRole.RoleType_ACADEMIC_AFFAIRS_STAFF] {
			newSearch = model.SearchRequestInput{
				Pagination: search.Pagination,
				Filters: append([]*model.FilterCriteriaInput{
					&model.FilterCriteriaInput{
						Condition: &model.FilterConditionInput{
							Field:    "id",
							Operator: model.FilterOperatorIn,
							Values:   Values,
						},
					},
				}, search.Filters...),
			}
			semesters, err = c.academic.GetSemestersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
		} else {
			return nil, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return c.pbSemestersToModel(semesters), nil
}

// ============================================
// RESOLVER HELPER METHODS
// ============================================

// GetMajorsByFacultyId returns all majors for a given faculty
func (c *Controller) GetMajorsByFacultyId(ctx context.Context, facultyId string) ([]*model.Major, error) {
	newSearch := model.SearchRequestInput{
		Pagination: c.DefaultPagination(),
		Filters: []*model.FilterCriteriaInput{
			{
				Condition: &model.FilterConditionInput{
					Field:    "faculty_code",
					Operator: model.FilterOperatorEqual,
					Values:   []string{facultyId},
				},
			},
		},
	}

	majors, err := c.academic.GetMajorsBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbMajorsToModel(majors.GetMajors()), nil
}

// GetStudentsBySemesterId returns all students for a given semester
func (c *Controller) GetStudentsBySemesterId(ctx context.Context, semesterId string) ([]*model.Student, error) {
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

	students, err := c.user.GetStudentsBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbStudentsToModel(students.GetStudents()), nil
}

// GetTeachersBySemesterId returns all teachers for a given semester
func (c *Controller) GetTeachersBySemesterId(ctx context.Context, semesterId string) ([]*model.Teacher, error) {
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

	teachers, err := c.user.GetTeachersBySearch(ctx, c.ConvertSearchRequestToPB(newSearch))
	if err != nil {
		return nil, err
	}

	return convert.PbTeachersToModel(teachers.GetTeachers()), nil
}

// GetSemesterById returns a semester by ID
func (c *Controller) GetSemesterById(ctx context.Context, id string) (*model.Semester, error) {
	semester, err := c.academic.GetSemesterById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbSemesterToModel(semester.GetSemester()), nil
}
