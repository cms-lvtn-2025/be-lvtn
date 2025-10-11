package controller

import (
	"context"
	pb "thaily/proto/role"
	"thaily/src/graph/model"
	"time"
)

func (c *Controller) pbRoleToModel(ctx context.Context, role *pb.ListRoleSystemsResponse) []*model.RoleSystem {
	if role == nil || role.GetRoleSystems() == nil {
		return nil
	}

	roleSystems := role.GetRoleSystems()
	result := make([]*model.RoleSystem, 0, len(roleSystems))

	for _, r := range roleSystems {
		// Convert protobuf RoleType to GraphQL RoleSystemRole
		var roleType model.RoleSystemRole
		switch r.GetRole() {
		case pb.RoleType_ACADEMIC_AFFAIRS_STAFF:
			roleType = model.RoleSystemRoleAcademicAffairsStaff
		case pb.RoleType_SUPERVISOR_LECTURER:
			roleType = model.RoleSystemRoleSupervisorLecturer
		case pb.RoleType_DEPARTMENT_LECTURER:
			roleType = model.RoleSystemRoleDepartmentLecturer
		case pb.RoleType_REVIEWER_LECTURER:
			roleType = model.RoleSystemRoleReviewerLecturer
		default:
			roleType = model.RoleSystemRoleAcademicAffairsStaff
		}

		// Convert timestamps
		var createdAt, updatedAt *time.Time
		if r.GetCreatedAt() != nil {
			t := r.GetCreatedAt().AsTime()
			createdAt = &t
		}
		if r.GetUpdatedAt() != nil {
			t := r.GetUpdatedAt().AsTime()
			updatedAt = &t
		}

		// Handle optional fields
		var teacherCode, createdBy, updatedBy *string
		if r.GetTeacherCode() != "" {
			tc := r.GetTeacherCode()
			teacherCode = &tc
		}
		if r.GetCreatedBy() != "" {
			cb := r.GetCreatedBy()
			createdBy = &cb
		}
		if r.GetUpdatedBy() != "" {
			ub := r.GetUpdatedBy()
			updatedBy = &ub
		}

		result = append(result, &model.RoleSystem{
			ID:           r.GetId(),
			Title:        r.GetTitle(),
			TeacherCode:  teacherCode,
			Role:         roleType,
			SemesterCode: r.GetSemesterCode(),
			Activate:     r.GetActivate(),
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			CreatedBy:    createdBy,
			UpdatedBy:    updatedBy,
		})
	}

	return result
}

func (c *Controller) GetRole(ctx context.Context, teacherId string) ([]*model.RoleSystem, error) {
	if teacherId == "" {
		return nil, nil
	}

	role, err := c.role.GetAllRoleByTeacherId(ctx, teacherId)
	if err != nil {
		return nil, err
	}

	return c.pbRoleToModel(ctx, role), nil
}
