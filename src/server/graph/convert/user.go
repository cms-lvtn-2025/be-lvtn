package convert

import (
	"thaily/proto/user"
	"thaily/src/server/graph/model"
)

// PbStudentToModel converts protobuf Student to GraphQL Student
func PbStudentToModel(pb *user.Student) *model.Student {
	if pb == nil {
		return nil
	}

	result := &model.Student{
		ID:           pb.Id,
		Email:        pb.Email,
		Phone:        pb.Phone,
		Username:     pb.Username,
		MajorCode:    pb.MajorCode,
		SemesterCode: pb.SemesterCode,
		Mssv:         pb.Mssv,
	}

	// Handle optional ClassCode
	if pb.ClassCode != "" {
		result.ClassCode = &pb.ClassCode
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

	// Handle Gender enum
	gender := PbGenderToModel(pb.Gender)
	result.Gender = &gender

	return result
}

// PbTeacherToModel converts protobuf Teacher to GraphQL Teacher
func PbTeacherToModel(pb *user.Teacher) *model.Teacher {
	if pb == nil {
		return nil
	}

	result := &model.Teacher{
		ID:           pb.Id,
		Email:        pb.Email,
		Username:     pb.Username,
		MajorCode:    pb.MajorCode,
		SemesterCode: pb.SemesterCode,
		Msgv:         pb.Msgv,
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

	// Handle Gender enum
	gender := PbGenderToModel(pb.Gender)
	result.Gender = &gender

	return result
}

// PbGenderToModel converts protobuf Gender enum to GraphQL Gender enum
func PbGenderToModel(pb user.Gender) model.Gender {
	switch pb {
	case user.Gender_MALE:
		return model.GenderMale
	case user.Gender_FEMALE:
		return model.GenderFemale
	case user.Gender_OTHER:
		return model.GenderOther
	default:
		return model.GenderMale
	}
}

// PbStudentsToModel converts array of protobuf Students to GraphQL Students
func PbStudentsToModel(pbs []*user.Student) []*model.Student {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Student, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbStudentToModel(pb))
		}
	}
	return result
}

// PbTeachersToModel converts array of protobuf Teachers to GraphQL Teachers
func PbTeachersToModel(pbs []*user.Teacher) []*model.Teacher {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Teacher, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbTeacherToModel(pb))
		}
	}
	return result
}

// ============================================
// INFO TYPE CONVERTERS
// ============================================

// PbTeacherToTeacherInfo converts protobuf Teacher to GraphQL TeacherInfo
func PbTeacherToTeacherInfo(pb *user.Teacher) *model.TeacherInfo {
	if pb == nil {
		return nil
	}

	result := &model.TeacherInfo{
		ID:        pb.Id,
		Email:     pb.Email,
		Username:  pb.Username,
		MajorCode: pb.MajorCode,
	}

	// Handle Gender enum
	gender := PbGenderToModel(pb.Gender)
	result.Gender = &gender

	return result
}

// ============================================
// LIST RESPONSE FACTORY FUNCTIONS
// ============================================

// CreateStudentListResponse creates a StudentListResponse
func CreateStudentListResponse(students []*model.Student, total int32) *model.StudentListResponse {
	return &model.StudentListResponse{
		Data:  students,
		Total: total,
	}
}

// CreateTeacherListResponse creates a TeacherListResponse
func CreateTeacherListResponse(teachers []*model.Teacher, total int32) *model.TeacherListResponse {
	return &model.TeacherListResponse{
		Data:  teachers,
		Total: total,
	}
}

