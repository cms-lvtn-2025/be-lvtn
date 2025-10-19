package controller

import (
	"context"
	"thaily/src/graph/convert"
	"thaily/src/graph/model"
)

// Student

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

func (c *Controller) GetTeacherInfoById(ctx context.Context, id string) (*model.StudentTeacherInfo, error) {
	teacher, err := c.user.GetTeacherById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.PbTeacherToStudentTeacherInfo(teacher.GetTeacher()), nil
}
