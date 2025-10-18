package convert

import (
	"thaily/proto/role"
	"thaily/src/graph/model"
)

// PbRoleSystemToModel converts protobuf RoleSystem to GraphQL RoleSystem
func PbRoleSystemToModel(pb *role.RoleSystem) *model.RoleSystem {
	if pb == nil {
		return nil
	}

	result := &model.RoleSystem{
		ID:           pb.Id,
		Title:        pb.Title,
		Role:         PbRoleTypeToModel(pb.Role),
		SemesterCode: pb.SemesterCode,
		Activate:     pb.Activate,
	}

	// Handle optional TeacherCode
	if pb.TeacherCode != "" {
		result.TeacherCode = &pb.TeacherCode
	}

	// Handle timestamps
	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		result.CreatedAt = &t
	}
	if pb.UpdatedAt != nil {
		t := pb.UpdatedAt.AsTime()
		result.UpdatedAt = &t
	}

	// Handle CreatedBy/UpdatedBy
	if pb.CreatedBy != "" {
		result.CreatedBy = &pb.CreatedBy
	}
	if pb.UpdatedBy != "" {
		result.UpdatedBy = &pb.UpdatedBy
	}

	return result
}

// PbRoleTypeToModel converts protobuf RoleType enum to GraphQL RoleSystemRole enum
func PbRoleTypeToModel(pb role.RoleType) model.RoleSystemRole {
	switch pb {
	case role.RoleType_ACADEMIC_AFFAIRS_STAFF:
		return model.RoleSystemRoleAcademicAffairsStaff
	case role.RoleType_DEPARTMENT_LECTURER:
		return model.RoleSystemRoleDepartmentLecturer
	case role.RoleType_TEACHER:
		return model.RoleSystemRoleTeacher
	default:
		return model.RoleSystemRoleTeacher
	}
}

// PbRoleSystemsToModel converts array of protobuf RoleSystems to GraphQL RoleSystems
func PbRoleSystemsToModel(pbs []*role.RoleSystem) []*model.RoleSystem {
	if pbs == nil {
		return nil
	}

	result := make([]*model.RoleSystem, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbRoleSystemToModel(pb))
		}
	}
	return result
}

// ============================================
// LIST RESPONSE FACTORY FUNCTIONS
// ============================================

// CreateRoleSystemListResponse creates a RoleSystemListResponse
func CreateRoleSystemListResponse(roleSystems []*model.RoleSystem, total int32) *model.RoleSystemListResponse {
	return &model.RoleSystemListResponse{
		Data:  roleSystems,
		Total: total,
	}
}
