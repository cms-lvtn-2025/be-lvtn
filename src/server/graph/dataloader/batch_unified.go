package dataloader

import (
	"context"
	"log"
	pb "thaily/proto/common"
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
	return func(ctx context.Context, ids []string) (map[string]*model.MajorInfo, error) {
		result := make(map[string]*model.MajorInfo)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetMajorsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				major, err := client.GetMajorById(ctx, id)
				if err != nil {
					continue
				}
				if major != nil && major.Major != nil {
					result[id] = convert.PbMajorToMajorInfo(major.Major)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Majors != nil {
			for _, pbMajor := range resp.Majors {
				if pbMajor != nil {
					result[pbMajor.Id] = convert.PbMajorToMajorInfo(pbMajor)
				}
			}
		}

		return result, nil
	}
}

// createSemesterInfoBatchFunc creates a batch function for loading SemesterInfo
func createSemesterInfoBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, *model.SemesterInfo] {
	return func(ctx context.Context, ids []string) (map[string]*model.SemesterInfo, error) {
		result := make(map[string]*model.SemesterInfo)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetSemestersByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				semester, err := client.GetSemesterById(ctx, id)
				if err != nil {
					continue
				}
				if semester != nil && semester.Semester != nil {
					result[id] = convert.PbSemesterToSemesterInfo(semester.Semester)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Semesters != nil {
			for _, pbSemester := range resp.Semesters {
				if pbSemester != nil {
					result[pbSemester.Id] = convert.PbSemesterToSemesterInfo(pbSemester)
				}
			}
		}

		return result, nil
	}
}

// createTeacherInfoBatchFunc creates a batch function for loading TeacherInfo
func createTeacherInfoBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.TeacherInfo] {
	return func(ctx context.Context, ids []string) (map[string]*model.TeacherInfo, error) {
		result := make(map[string]*model.TeacherInfo)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetTeachersByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				teacher, err := client.GetTeacherById(ctx, id)
				if err != nil {
					continue
				}
				if teacher != nil && teacher.Teacher != nil {
					result[id] = convert.PbTeacherToTeacherInfo(teacher.Teacher)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Teachers != nil {
			for _, pbTeacher := range resp.Teachers {
				if pbTeacher != nil {
					result[pbTeacher.Id] = convert.PbTeacherToTeacherInfo(pbTeacher)
				}
			}
		}

		return result, nil
	}
}

// createTeacherBatchFunc creates a batch function for loading Teacher
func createTeacherBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.Teacher] {
	return func(ctx context.Context, ids []string) (map[string]*model.Teacher, error) {
		result := make(map[string]*model.Teacher)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetTeachersByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				teacher, err := client.GetTeacherById(ctx, id)
				if err != nil {
					continue
				}
				if teacher != nil && teacher.Teacher != nil {
					result[id] = convert.PbTeacherToModel(teacher.Teacher)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Teachers != nil {
			for _, pbTeacher := range resp.Teachers {
				if pbTeacher != nil {
					result[pbTeacher.Id] = convert.PbTeacherToModel(pbTeacher)
				}
			}
		}

		return result, nil
	}
}

// createGradeDefenceByDefenceIdBatchFunc creates a batch function for loading grade defences by defence ID
func createGradeDefenceByDefenceIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.GradeDefence] {
	return func(ctx context.Context, ids []string) (map[string][]*model.GradeDefence, error) {
		result := make(map[string][]*model.GradeDefence)
		if len(ids) == 0 {
			return result, nil
		}

		newSearch := pb.SearchRequest{
			Pagination: &pb.Pagination{
				Page:       1,
				PageSize:   int32(len(ids) * 10),
				Descending: true,
				SortBy:     "created_at",
			},
			Filters: []*pb.FilterCriteria{
				{
					Criteria: &pb.FilterCriteria_Condition{
						Condition: &pb.FilterCondition{
							Field:    "defence_code",
							Operator: pb.FilterOperator_IN,
							Values:   ids,
						},
					},
				},
			},
		}

		gradeDefences, err := client.GetGradeDefenceBySearch(ctx, &newSearch)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				search := pb.SearchRequest{
					Pagination: &pb.Pagination{
						Page:       1,
						PageSize:   10,
						Descending: true,
						SortBy:     "created_at",
					},
					Filters: []*pb.FilterCriteria{
						{
							Criteria: &pb.FilterCriteria_Condition{
								Condition: &pb.FilterCondition{
									Field:    "defence_code",
									Operator: pb.FilterOperator_EQUAL,
									Values:   []string{id},
								},
							},
						},
					},
				}
				gradeDefence, err := client.GetGradeDefenceBySearch(ctx, &search)
				if err != nil {
					continue
				}
				if gradeDefence != nil {
					result[id] = convert.PbGradeDefencesToModel(gradeDefence.GetGradeDefences())
				}
			}
			return result, nil
		}

		if gradeDefences != nil {
			for _, gradeDefence := range gradeDefences.GetGradeDefences() {
				if gradeDefence != nil {
					result[gradeDefence.DefenceCode] = append(result[gradeDefence.DefenceCode], convert.PbGradeDefenceToModel(gradeDefence))
				}
			}
		}

		return result, nil
	}
}

// createGradeDefenceCriteriaByDefenceIdBatchFunc creates a batch function for loading GradeDefenceCriterion
func createGradeDefenceCriteriaByDefenceIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.GradeDefenceCriterion] {
	return func(ctx context.Context, ids []string) (map[string][]*model.GradeDefenceCriterion, error) {
		result := make(map[string][]*model.GradeDefenceCriterion)
		if len(ids) == 0 {
			return result, nil
		}

		newSearch := pb.SearchRequest{
			Pagination: &pb.Pagination{
				Page:       1,
				PageSize:   int32(len(ids) * 5),
				Descending: true,
				SortBy:     "created_at",
			},
			Filters: []*pb.FilterCriteria{
				{
					Criteria: &pb.FilterCriteria_Condition{
						Condition: &pb.FilterCondition{
							Field:    "grade_defence_code",
							Operator: pb.FilterOperator_IN,
							Values:   ids,
						},
					},
				},
			},
		}

		resp, err := client.GetGradeDefenceCriteriaBySearch(ctx, &newSearch)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				search := pb.SearchRequest{
					Pagination: &pb.Pagination{
						Page:       1,
						PageSize:   10,
						Descending: true,
						SortBy:     "created_at",
					},
					Filters: []*pb.FilterCriteria{
						{
							Criteria: &pb.FilterCriteria_Condition{
								Condition: &pb.FilterCondition{
									Field:    "grade_defence_code",
									Operator: pb.FilterOperator_EQUAL,
									Values:   []string{id},
								},
							},
						},
					},
				}
				criteria, err := client.GetGradeDefenceCriteriaBySearch(ctx, &search)
				if err != nil {
					continue
				}
				if criteria != nil {
					result[id] = convert.PbGradeDefenceCriteriaToModel(criteria.GetGradeDefenceCriteria())
				}
			}
			return result, nil
		}

		if resp != nil {
			for _, criterion := range resp.GetGradeDefenceCriteria() {
				if criterion != nil {
					result[criterion.GradeDefenceCode] = append(result[criterion.GradeDefenceCode], convert.PbGradeDefenceCriterionToModel(criterion))
				}
			}
		}

		return result, nil
	}
}

// createTopicBatchFunc creates a batch function for loading Topic
func createTopicBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Topic] {
	return func(ctx context.Context, ids []string) (map[string]*model.Topic, error) {
		result := make(map[string]*model.Topic)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetTopicsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				topic, err := client.GetTopicById(ctx, id)
				if err != nil {
					continue
				}
				if topic != nil && topic.Topic != nil {
					result[id] = convert.PbTopicToModel(topic.Topic)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Topics != nil {
			for _, pbTopic := range resp.Topics {
				if pbTopic != nil {
					result[pbTopic.Id] = convert.PbTopicToModel(pbTopic)
				}
			}
		}

		return result, nil
	}
}

// createMidtermBatchFunc creates a batch function for loading Midterm
func createMidtermBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Midterm] {
	return func(ctx context.Context, ids []string) (map[string]*model.Midterm, error) {
		result := make(map[string]*model.Midterm)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetMidtermsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				midterm, err := client.GetMidtermById(ctx, id)
				if err != nil {
					continue
				}
				if midterm != nil && midterm.Midterm != nil {
					result[id] = convertPbMidtermToModel(midterm.Midterm)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Midterms != nil {
			for _, pbMidterm := range resp.Midterms {
				if pbMidterm != nil {
					result[pbMidterm.Id] = convertPbMidtermToModel(pbMidterm)
				}
			}
		}

		return result, nil
	}
}

// createFinalBatchFunc creates a batch function for loading Final
func createFinalBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Final] {
	return func(ctx context.Context, ids []string) (map[string]*model.Final, error) {
		result := make(map[string]*model.Final)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetFinalsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				final, err := client.GetFinalById(ctx, id)
				if err != nil {
					continue
				}
				if final != nil && final.Final != nil {
					result[id] = convertPbFinalToModel(final.Final)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Finals != nil {
			for _, pbFinal := range resp.Finals {
				if pbFinal != nil {
					result[pbFinal.Id] = convertPbFinalToModel(pbFinal)
				}
			}
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
