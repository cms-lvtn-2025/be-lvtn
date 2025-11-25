package convert

import (
	"thaily/proto/council"
	"thaily/src/server/graph/model"
)

// PbCouncilToModel converts protobuf Council to GraphQL Council
func PbCouncilToModel(pb *council.Council) *model.Council {
	if pb == nil {
		return nil
	}

	result := &model.Council{
		ID:           pb.Id,
		Title:        pb.Title,
		MajorCode:    pb.MajorCode,
		SemesterCode: pb.SemesterCode,
	}

	// Handle optional TimeStart
	if pb.TimeStart != nil {
		t := pb.TimeStart.AsTime()
		result.TimeStart = &t
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

// PbDefenceToModel converts protobuf Defence to GraphQL Defence
func PbDefenceToModel(pb *council.Defence) *model.Defence {
	if pb == nil {
		return nil
	}

	result := &model.Defence{
		ID:          pb.Id,
		Title:       pb.Title,
		CouncilCode: pb.CouncilCode,
		TeacherCode: pb.TeacherCode,
		Position:    PbDefencePositionToModel(pb.Position),
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

// PbGradeDefenceToModel converts protobuf GradeDefence to GraphQL GradeDefence
func PbGradeDefenceToModel(pb *council.GradeDefence) *model.GradeDefence {
	if pb == nil {
		return nil
	}

	result := &model.GradeDefence{
		ID:             pb.Id,
		DefenceCode:    pb.DefenceCode,
		EnrollmentCode: pb.EnrollmentCode,
	}

	// Handle optional Note
	if pb.Note != "" {
		result.Note = &pb.Note
	}

	// Handle optional TotalScore
	if pb.TotalScore != 0 {
		result.TotalScore = &pb.TotalScore
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

// PbGradeDefenceCriterionToModel converts protobuf GradeDefenceCriterion to GraphQL GradeDefenceCriterion
func PbGradeDefenceCriterionToModel(pb *council.GradeDefenceCriterion) *model.GradeDefenceCriterion {
	if pb == nil {
		return nil
	}

	result := &model.GradeDefenceCriterion{
		ID:               pb.Id,
		GradeDefenceCode: pb.GradeDefenceCode,
	}

	// Handle optional Name
	if pb.Name != "" {
		result.Name = &pb.Name
	}

	// Handle optional Score
	if pb.Score != "" {
		result.Score = &pb.Score
	}

	// Handle optional MaxScore
	if pb.MaxScore != "" {
		result.MaxScore = &pb.MaxScore
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

// PbDefencePositionToModel converts protobuf DefencePosition enum to GraphQL DefencePosition enum
func PbDefencePositionToModel(pb council.DefencePosition) model.DefencePosition {
	switch pb {
	case council.DefencePosition_PRESIDENT:
		return model.DefencePositionPresident
	case council.DefencePosition_SECRETARY:
		return model.DefencePositionSecretary
	case council.DefencePosition_REVIEWER:
		return model.DefencePositionReviewer
	case council.DefencePosition_MEMBER:
		return model.DefencePositionMember
	default:
		return model.DefencePositionPresident
	}
}

// PbCouncilsToModel converts array of protobuf Councils to GraphQL Councils
func PbCouncilsToModel(pbs []*council.Council) []*model.Council {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Council, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbCouncilToModel(pb))
		}
	}
	return result
}

// PbDefencesToModel converts array of protobuf Defences to GraphQL Defences
func PbDefencesToModel(pbs []*council.Defence) []*model.Defence {
	if pbs == nil {
		return nil
	}

	result := make([]*model.Defence, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbDefenceToModel(pb))
		}
	}
	return result
}

// PbGradeDefencesToModel converts array of protobuf GradeDefences to GraphQL GradeDefences
func PbGradeDefencesToModel(pbs []*council.GradeDefence) []*model.GradeDefence {
	if pbs == nil {
		return nil
	}

	result := make([]*model.GradeDefence, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbGradeDefenceToModel(pb))
		}
	}
	return result
}

// PbGradeDefenceCriteriaToModel converts array of protobuf GradeDefenceCriteria to GraphQL GradeDefenceCriteria
func PbGradeDefenceCriteriaToModel(pbs []*council.GradeDefenceCriterion) []*model.GradeDefenceCriterion {
	if pbs == nil {
		return nil
	}

	result := make([]*model.GradeDefenceCriterion, 0, len(pbs))
	for _, pb := range pbs {
		if pb != nil {
			result = append(result, PbGradeDefenceCriterionToModel(pb))
		}
	}
	return result
}


// ============================================
// LIST RESPONSE FACTORY FUNCTIONS
// ============================================

// CreateCouncilListResponse creates a CouncilListResponse
func CreateCouncilListResponse(councils []*model.Council, total int32) *model.CouncilListResponse {
	return &model.CouncilListResponse{
		Data:  councils,
		Total: total,
	}
}

// CreateDefenceListResponse creates a DefenceListResponse
func CreateDefenceListResponse(defences []*model.Defence, total int32) *model.DefenceListResponse {
	return &model.DefenceListResponse{
		Data:  defences,
		Total: total,
	}
}

// CreateGradeDefenceListResponse creates a GradeDefenceListResponse
func CreateGradeDefenceListResponse(gradeDefences []*model.GradeDefence, total int32) *model.GradeDefenceListResponse {
	return &model.GradeDefenceListResponse{
		Data:  gradeDefences,
		Total: total,
	}
}

// CreateGradeDefenceCriterionListResponse creates a GradeDefenceCriterionListResponse
func CreateGradeDefenceCriterionListResponse(criteria []*model.GradeDefenceCriterion, total int32) *model.GradeDefenceCriterionListResponse {
	return &model.GradeDefenceCriterionListResponse{
		Data:  criteria,
		Total: total,
	}
}

