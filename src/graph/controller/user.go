package controller

import (
	"context"
	"fmt"
	"strings"
	pb "thaily/proto/user"
	"thaily/src/graph/helper"
	"thaily/src/graph/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (c *Controller) pbStudentToModel(resp *pb.GetStudentResponse) *model.Student {
	if resp == nil {
		return nil
	}

	s := resp.GetStudent()

	// convert enum Gender
	var gender model.Gender
	switch s.Gender {
	case pb.Gender_MALE:
		gender = model.GenderMale
	case pb.Gender_FEMALE:
		gender = model.GenderFemale
	case pb.Gender_OTHER:
		gender = model.GenderOther
	}

	// convert timestamp
	var createdAt, updatedAt *time.Time
	if s.CreatedAt != nil {
		t := s.CreatedAt.AsTime()
		createdAt = &t
	}
	if s.UpdatedAt != nil {
		t := s.UpdatedAt.AsTime()
		updatedAt = &t
	}

	return &model.Student{
		ID:           s.Id,
		Email:        s.Email,
		Phone:        s.Phone,
		Username:     s.Username,
		Gender:       &gender,
		MajorCode:    s.MajorCode,
		ClassCode:    &s.ClassCode,
		SemesterCode: s.SemesterCode,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		CreatedBy:    &s.CreatedBy,
		UpdatedBy:    &s.UpdatedBy,
	}
}

func (c *Controller) pbTeacherToModel(resp *pb.GetTeacherResponse) *model.Teacher {
	if resp == nil {
		return nil
	}

	t := resp.GetTeacher()

	// convert enum
	var gender model.Gender
	switch t.Gender {
	case pb.Gender_MALE:
		gender = model.GenderMale
	case pb.Gender_FEMALE:
		gender = model.GenderFemale
	case pb.Gender_OTHER:
		gender = model.GenderOther
	}
	// convert timestamp
	var createdAt, updatedAt *time.Time
	if t.CreatedAt != nil {
		t := t.CreatedAt.AsTime()
		createdAt = &t
	}
	if t.UpdatedAt != nil {
		t := t.UpdatedAt.AsTime()
		updatedAt = &t
	}

	return &model.Teacher{
		ID:           t.Id,
		Email:        t.Email,
		Username:     t.Username,
		MajorCode:    t.MajorCode,
		SemesterCode: t.SemesterCode,
		Gender:       &gender,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		CreatedBy:    &t.CreatedBy,
		UpdatedBy:    &t.UpdatedBy,
	}
}

func (c *Controller) GetInfoStudent(ctx context.Context) (*model.Student, error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("claims not found in context")
	}
	fmt.Print(claims)
	semester, _ := ctx.Value("semester").(string)

	// fmt.Print(ids)
	idsArr := strings.Split(claims["ids"].(string), ",")
	myId := ""
	if (semester == "") && len(idsArr) > 0 {
		myId = strings.Split(idsArr[0], "-")[1]
	} else {
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+"-") {
				myId = strings.Split(id, "-")[1]
			}
		}
	}
	fmt.Println("myId:", myId, "semester:", semester)
	if myId == "" {
		return nil, fmt.Errorf("no student found for semester %s", semester)
	}

	user, err := c.user.GetUserById(ctx, myId)

	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return c.pbStudentToModel(user), nil
}

func (c *Controller) GetInfoTeacher(ctx context.Context) (*model.Teacher, error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	fmt.Print(claims)
	if !ok {
		return nil, fmt.Errorf("claims not found in context")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("email not found in claims")
	}

	semester, _ := ctx.Value("semester").(string)
	idsArr := strings.Split(claims["ids"].(string), ",")
	myId := ""
	if semester == "" {
		myId = strings.Split(idsArr[0], "-")[1]
	} else {
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+"-") {
				myId = strings.Split(id, "-")[1]
			}
		}
	}
	if myId == "" {
		return nil, fmt.Errorf("no teacher found for semester %s", semester)
	}

	teacher, err := c.user.GetTeacherById(ctx, myId)
	if err != nil {
		return nil, err
	}
	if teacher == nil && teacher.GetTeacher().GetEmail() != email {
		return nil, fmt.Errorf("teacher not found or email mismatch")
	}

	// convert timestamp
	return c.pbTeacherToModel(teacher), nil
}

func (c *Controller) GetStudentById(ctx context.Context, id string) (*model.Student, error) {
	student, err := c.user.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.pbStudentToModel(student), nil
}
