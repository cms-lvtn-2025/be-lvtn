package dataloader

import (
	"context"
	pbThesis "thaily/proto/thesis"
	"thaily/src/server/client"
	"thaily/src/server/graph/convert"
	"thaily/src/server/graph/model"
)

// ============================================
// UNIFIED BATCH FUNCTIONS FOR SCHEMA 2
// ============================================

// createMajorInfoBatchFunc creates a batch function for loading MajorInfo
func createMajorInfoBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, *model.MajorInfo] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.MajorInfo, error) {
		result := make(map[string]*model.MajorInfo)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		majors, _ := client.GetMajorsBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbMajor := range majors.GetMajors() {
			majorId := pbMajor.GetId()
			result[majorId] = convert.PbMajorToMajorInfo(pbMajor)
		}

		return result, nil
	}
}

// createSemesterInfoBatchFunc creates a batch function for loading SemesterInfo
func createSemesterInfoBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, *model.SemesterInfo] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.SemesterInfo, error) {
		result := make(map[string]*model.SemesterInfo)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		majors, _ := client.GetSemestersBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbSemester := range majors.GetSemesters() {
			semesterId := pbSemester.GetId()
			result[semesterId] = convert.PbSemesterToSemesterInfo(pbSemester)
		}

		return result, nil
	}
}

// createTeacherInfoBatchFunc creates a batch function for loading TeacherInfo
func createTeacherInfoBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.TeacherInfo] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.TeacherInfo, error) {
		result := make(map[string]*model.TeacherInfo)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		teachers, _ := client.GetTeachersBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbTeacher := range teachers.GetTeachers() {
			teacherId := pbTeacher.GetId()
			result[teacherId] = convert.PbTeacherToTeacherInfo(pbTeacher)
		}

		return result, nil
	}
}

// createTeacherBatchFunc creates a batch function for loading Teacher
func createTeacherBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.Teacher] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Teacher, error) {
		result := make(map[string]*model.Teacher)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{

				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		teachers, _ := client.GetTeachersBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbTeacher := range teachers.GetTeachers() {
			teacherId := pbTeacher.GetId()
			result[teacherId] = convert.PbTeacherToModel(pbTeacher)
		}
		return result, nil
	}
}

// createGradeDefenceByDefenceIdBatchFunc creates a batch function for loading grade defences by defence ID
func createGradeDefenceByDefenceIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.GradeDefence] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string][]*model.GradeDefence, error) {
		result := make(map[string][]*model.GradeDefence)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "defence_code",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		gradeDefences, _ := client.GetGradeDefenceBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbGradeDefence := range gradeDefences.GetGradeDefences() {
			defenceCode := pbGradeDefence.GetDefenceCode()
			result[defenceCode] = append(result[defenceCode], convert.PbGradeDefenceToModel(pbGradeDefence))
		}

		return result, nil
	}
}

// createGradeDefenceCriteriaByDefenceIdBatchFunc creates a batch function for loading GradeDefenceCriterion
func createGradeDefenceCriteriaByDefenceIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.GradeDefenceCriterion] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string][]*model.GradeDefenceCriterion, error) {
		result := make(map[string][]*model.GradeDefenceCriterion)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "grade_defence_code",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{

			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		gradeDefenceCriteria, _ := client.GetGradeDefenceCriteriaBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbCriterion := range gradeDefenceCriteria.GetGradeDefenceCriteria() {
			defenceCode := pbCriterion.GetGradeDefenceCode()
			result[defenceCode] = append(result[defenceCode], convert.PbGradeDefenceCriterionToModel(pbCriterion))
		}
		return result, nil
	}
}

// createTopicBatchFunc creates a batch function for loading Topic
func createTopicBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Topic] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Topic, error) {
		result := make(map[string]*model.Topic)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "id",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		topics, _ := client.GetTopicBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbTopic := range topics.GetTopics() {
			topicId := pbTopic.GetId()
			result[topicId] = convert.PbTopicToModel(pbTopic)
		}

		return result, nil
	}
}

// createMidtermBatchFunc creates a batch function for loading Midterm
func createMidtermBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Midterm] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Midterm, error) {
		result := make(map[string]*model.Midterm)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "enrollment_code",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		midterms, _ := client.GetMidtermBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))

		for _, pbMidterm := range midterms.GetMidterms() {
			midtermId := pbMidterm.GetId()
			result[midtermId] = convert.PbMidtermToModel(pbMidterm)
		}

		return result, nil
	}
}

// createFinalBatchFunc creates a batch function for loading Final
func createFinalBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Final] {
	return func(ctx context.Context, ids []string, filters []*model.FilterCriteriaInput) (map[string]*model.Final, error) {
		result := make(map[string]*model.Final)
		if len(ids) == 0 {
			return result, nil
		}

		filters = append(filters, &model.FilterCriteriaInput{
			Condition: &model.FilterConditionInput{
				Field:    "enrollment_code",
				Operator: model.FilterOperatorIn,
				Values:   ids,
			},
		})

		page := int32(1)
		pageSize := int32(len(ids) * 10)

		searchReq := &model.SearchRequestInput{
			Pagination: &model.PaginationInput{
				Page:     &page,
				PageSize: &pageSize,
			},
			Filters: filters,
		}

		finals, _ := client.GetFinalBySearch(ctx, convert.ConvertSearchRequestToPB(*searchReq))
		for _, pbFinal := range finals.GetFinals() {
			finalId := pbFinal.GetId()
			result[finalId] = convert.PbFinalToModel(pbFinal)
		}

		return result, nil
	}
}

// ============================================
// HELPER CONVERTER FUNCTIONS
// ============================================

// convertPbMidtermToModel converts protobuf Midterm to GraphQL model
func convertPbMidtermToModel(pbMidterm *pbThesis.Midterm) *model.Midterm {
	if pbMidterm == nil {
		return nil
	}

	var status model.MidtermStatus
	switch pbMidterm.Status {
	case pbThesis.MidtermStatus_NOT_SUBMITTED:
		status = model.MidtermStatusNotSubmitted
	case pbThesis.MidtermStatus_SUBMITTED:
		status = model.MidtermStatusSubmitted
	case pbThesis.MidtermStatus_PASS:
		status = model.MidtermStatusPass
	case pbThesis.MidtermStatus_FAIL:
		status = model.MidtermStatusFail
	default:
		status = model.MidtermStatusNotSubmitted
	}

	modelMidterm := &model.Midterm{
		ID:     pbMidterm.Id,
		Title:  pbMidterm.Title,
		Status: status,
	}

	if pbMidterm.Grade != 0 {
		grade := pbMidterm.Grade
		modelMidterm.Grade = &grade
	}
	if pbMidterm.Feedback != "" {
		feedback := pbMidterm.Feedback
		modelMidterm.Feedback = &feedback
	}
	if pbMidterm.CreatedBy != "" {
		createdBy := pbMidterm.CreatedBy
		modelMidterm.CreatedBy = &createdBy
	}
	if pbMidterm.UpdatedBy != "" {
		updatedBy := pbMidterm.UpdatedBy
		modelMidterm.UpdatedBy = &updatedBy
	}
	if pbMidterm.CreatedAt != nil {
		createdAt := pbMidterm.CreatedAt.AsTime()
		modelMidterm.CreatedAt = &createdAt
	}
	if pbMidterm.UpdatedAt != nil {
		updatedAt := pbMidterm.UpdatedAt.AsTime()
		modelMidterm.UpdatedAt = &updatedAt
	}

	return modelMidterm
}

// convertPbFinalToModel converts protobuf Final to GraphQL model
func convertPbFinalToModel(pbFinal *pbThesis.Final) *model.Final {
	if pbFinal == nil {
		return nil
	}

	var status model.FinalStatus
	switch pbFinal.Status {
	case pbThesis.FinalStatus_PENDING:
		status = model.FinalStatusPending
	case pbThesis.FinalStatus_COMPLETED:
		status = model.FinalStatusCompleted
	case pbThesis.FinalStatus_FAILED:
		status = model.FinalStatusFailed
	case pbThesis.FinalStatus_PASSED:
		status = model.FinalStatusPassed
	default:
		status = model.FinalStatusPending
	}

	modelFinal := &model.Final{
		ID:     pbFinal.Id,
		Title:  pbFinal.Title,
		Status: status,
	}

	if pbFinal.SupervisorGrade != 0 {
		grade := pbFinal.SupervisorGrade
		modelFinal.SupervisorGrade = &grade
	}
	if pbFinal.FinalGrade != 0 {
		grade := pbFinal.FinalGrade
		modelFinal.FinalGrade = &grade
	}
	if pbFinal.Notes != "" {
		notes := pbFinal.Notes
		modelFinal.Notes = &notes
	}
	if pbFinal.CreatedBy != "" {
		createdBy := pbFinal.CreatedBy
		modelFinal.CreatedBy = &createdBy
	}
	if pbFinal.UpdatedBy != "" {
		updatedBy := pbFinal.UpdatedBy
		modelFinal.UpdatedBy = &updatedBy
	}
	if pbFinal.CreatedAt != nil {
		createdAt := pbFinal.CreatedAt.AsTime()
		modelFinal.CreatedAt = &createdAt
	}
	if pbFinal.UpdatedAt != nil {
		updatedAt := pbFinal.UpdatedAt.AsTime()
		modelFinal.UpdatedAt = &updatedAt
	}

	return modelFinal
}
