package convert

import (
	"thaily/proto/academic"
	"thaily/src/graph/model"
)

// PbSemesterToModel converts protobuf Semester to GraphQL Semester
func PbSemesterToModel(pb *academic.Semester) *model.Semester {
	if pb == nil {
		return nil
	}

	result := &model.Semester{
		ID:    pb.Id,
		Title: pb.Title,
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

// PbSemesterToSemesterInfo converts protobuf Semester to GraphQL SemesterInfo
func PbSemesterToSemesterInfo(pb *academic.Semester) *model.SemesterInfo {
	if pb == nil {
		return nil
	}

	return &model.SemesterInfo{
		ID:    pb.Id,
		Title: pb.Title,
	}
}

// PbMajorToModel converts protobuf Major to GraphQL Major
func PbMajorToModel(pb *academic.Major) *model.Major {
	if pb == nil {
		return nil
	}

	result := &model.Major{
		ID:          pb.Id,
		Title:       pb.Title,
		FacultyCode: pb.FacultyCode,
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

// PbMajorToMajorInfo converts protobuf Major to GraphQL MajorInfo
func PbMajorToMajorInfo(pb *academic.Major) *model.MajorInfo {
	if pb == nil {
		return nil
	}

	return &model.MajorInfo{
		ID:          pb.Id,
		Title:       pb.Title,
		FacultyCode: pb.FacultyCode,
	}
}

// PbSemestersToModel converts array of protobuf Semesters to GraphQL Semesters
func PbSemestersToModel(pbs []*academic.Semester) []*model.Semester {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Semester, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbSemesterToModel(pb))
		}
	}
	return result
}

// PbMajorsToModel converts array of protobuf Majors to GraphQL Majors
func PbMajorsToModel(pbs []*academic.Major) []*model.Major {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Major, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbMajorToModel(pb))
		}
	}
	return result
}

// PbFacultyToModel converts protobuf Faculty to GraphQL Faculty
func PbFacultyToModel(pb *academic.Faculty) *model.Faculty {
	if pb == nil {
		return nil
	}

	result := &model.Faculty{
		ID:    pb.Id,
		Title: pb.Title,
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

// PbFacultiesToModel converts array of protobuf Faculties to GraphQL Faculties
func PbFacultiesToModel(pbs []*academic.Faculty) []*model.Faculty {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Faculty, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbFacultyToModel(pb))
		}
	}
	return result
}

// ============================================
// LIST RESPONSE FACTORY FUNCTIONS
// ============================================

// CreateSemesterListResponse creates a SemesterListResponse
func CreateSemesterListResponse(semesters []*model.Semester, total int32) *model.SemesterListResponse {
	return &model.SemesterListResponse{
		Data:  semesters,
		Total: total,
	}
}

// CreateMajorListResponse creates a MajorListResponse
func CreateMajorListResponse(majors []*model.Major, total int32) *model.MajorListResponse {
	return &model.MajorListResponse{
		Data:  majors,
		Total: total,
	}
}

// CreateFacultyListResponse creates a FacultyListResponse
func CreateFacultyListResponse(faculties []*model.Faculty, total int32) *model.FacultyListResponse {
	return &model.FacultyListResponse{
		Data:  faculties,
		Total: total,
	}
}
