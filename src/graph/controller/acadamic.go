package controller

import (
	"context"
	pb "thaily/proto/academic"
	"thaily/src/graph/model"
	"time"
)

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
