package dataloader

import (
	"context"
	"fmt"
	"log"
	pb "thaily/proto/common"
	pbThesis "thaily/proto/thesis"
	"thaily/src/graph/convert"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// ============================================
// STUDENT DATALOADER BATCH FUNCTIONS
// All student-related batch loading functions
// ============================================

// ============================================
// ACADEMIC BATCH FUNCTIONS (Major, Semester)
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
				Major, err := client.GetMajorById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Major %s: %v", id, err)
					continue
				}

				if Major != nil && Major.Major != nil {
					result[id] = convert.PbMajorToMajorInfo(Major.Major)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Majors != nil {
			for _, pbMajor := range resp.Majors {
				if pbMajor != nil {
					result[pbMajor.Id] = convert.PbMajorToMajorInfo(pbMajor)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d Majors successfully", len(result), len(ids))
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
				Semester, err := client.GetSemesterById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Semester %s: %v", id, err)
					continue
				}

				if Semester != nil && Semester.Semester != nil {
					result[id] = convert.PbSemesterToSemesterInfo(Semester.Semester)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Semesters != nil {
			for _, pbSemester := range resp.Semesters {
				if pbSemester != nil {
					result[pbSemester.Id] = convert.PbSemesterToSemesterInfo(pbSemester)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d Semesters successfully", len(result), len(ids))
		return result, nil
	}
}

// ============================================
// COUNCIL BATCH FUNCTIONS (Defence, Council, GradeDefence)
// ============================================

// createDefenceByCouncilIDBatchFunc creates a batch function for loading Defences by Council ID
func createDefenceByCouncilIDBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.StudentDefenceInfo] {
	return func(ctx context.Context, ids []string) (map[string][]*model.StudentDefenceInfo, error) {
		result := make(map[string][]*model.StudentDefenceInfo)
		if len(ids) == 0 {
			return result, nil
		}
		var search pb.SearchRequest
		search = pb.SearchRequest{
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
							Field:    "council_code",
							Operator: pb.FilterOperator_IN,
							Values:   ids,
						},
					},
				},
			},
		}

		resp, err := client.GetDefencesBySearch(ctx, &search)

		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				search = pb.SearchRequest{
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
									Field:    "council_code",
									Operator: pb.FilterOperator_IN,
									Values:   []string{id},
								},
							},
						},
					},
				}
				Defences, err := client.GetDefencesBySearch(ctx, &search)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Defences %s: %v", id, err)
					continue
				}

				if Defences != nil && Defences.Defences != nil {
					result[id] = convert.PbDefencesToStudentDefencesInfo(Defences.GetDefences())
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Defences != nil {
			for _, pbDefences := range resp.Defences {
				if pbDefences != nil {
					fmt.Println("key", pbDefences.CouncilCode, ids)
					result[pbDefences.CouncilCode] = append(result[pbDefences.CouncilCode], convert.PbDefenceToStudentDefenceInfo(pbDefences))
				}
			}
		}
		log.Printf("[DataLoader] Batch loaded %d/%d Defences successfully", len(result), len(ids))
		return result, nil
	}
}

// createDefenceInfoByIdBatchFunc creates a batch function for loading Defence by ID
func createDefenceInfoByIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.StudentDefenceInfo] {
	return func(ctx context.Context, ids []string) (map[string]*model.StudentDefenceInfo, error) {
		result := make(map[string]*model.StudentDefenceInfo)
		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetDefencesByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				defence, err := client.GetDefenceById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Defence %s: %v", id, err)
					continue
				}
				if defence != nil {
					result[id] = convert.PbDefenceToStudentDefenceInfo(defence.GetDefence())
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}
		if resp != nil && resp.Defences != nil {
			for _, pbDefences := range resp.Defences {
				if pbDefences != nil {
					result[pbDefences.Id] = convert.PbDefenceToStudentDefenceInfo(pbDefences)
				}
			}
		}
		return result, nil
	}
}

// createCouncilForStudentBatchFunc creates a batch function for loading StudentCouncil
func createCouncilForStudentBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.StudentCouncil] {
	return func(ctx context.Context, ids []string) (map[string]*model.StudentCouncil, error) {
		result := make(map[string]*model.StudentCouncil)
		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetCouncilsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				council, err := client.GetCouncilById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Council %s: %v", id, err)
				}
				if council != nil && council.GetCouncil() != nil {
					result[id] = convert.PbCouncilToStudentCouncil(council.GetCouncil())
				}
			}
		}
		if resp != nil && resp.Councils != nil {
			for _, council := range resp.Councils {
				if council != nil {
					result[council.Id] = convert.PbCouncilToStudentCouncil(council)
				}
			}
		}
		return result, nil
	}
}

// createGradeDefencesForStudentBatchFunc creates a batch function for loading StudentGradeDefence
func createGradeDefencesForStudentBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.StudentGradeDefence] {
	return func(ctx context.Context, ids []string) (map[string][]*model.StudentGradeDefence, error) {
		result := make(map[string][]*model.StudentGradeDefence)
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
							Field:    "enrollment_code",
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
				newSearch = pb.SearchRequest{
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
									Field:    "enrollment_code",
									Operator: pb.FilterOperator_EQUAL,
									Values:   []string{id},
								},
							},
						},
					},
				}
				gradeDefence, err := client.GetGradeDefenceBySearch(ctx, &newSearch)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch GradeDefence %s: %v", id, err)
					continue
				}
				if gradeDefence != nil {
					result[id] = convert.PbGradeDefencesToStudentGradeDefences(gradeDefence.GetGradeDefences())
				}

			}
		}
		if gradeDefences != nil {
			for _, gradeDefence := range gradeDefences.GetGradeDefences() {
				result[gradeDefence.EnrollmentCode] = append(result[gradeDefence.EnrollmentCode], convert.PbGradeDefenceToStudentGradeDefence(gradeDefence))
			}
		}
		return result, nil
	}
}

// createGradeDefenceCriteriaForByDefenceIdBatchFunc creates a batch function for loading GradeDefenceCriterion
func createGradeDefenceCriteriaForByDefenceIdBatchFunc(client *client.GRPCCouncil) BatchFunc[string, []*model.GradeDefenceCriterion] {
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
				newSearch = pb.SearchRequest{
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
									Field:    "council_code",
									Operator: pb.FilterOperator_EQUAL,
									Values:   []string{id},
								},
							},
						},
					},
				}
				gradeDefence, err := client.GetGradeDefenceCriteriaBySearch(ctx, &newSearch)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch GradeDefence %s: %v", id, err)
					continue
				}
				if gradeDefence != nil {
					result[id] = convert.PbGradeDefenceCriteriaToModel(gradeDefence.GetGradeDefenceCriteria())
				}

			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}
		if resp != nil {
			for _, gradeDefence := range resp.GetGradeDefenceCriteria() {
				if gradeDefence != nil {
					result[gradeDefence.GetGradeDefenceCode()] = append(result[gradeDefence.GradeDefenceCode], convert.PbGradeDefenceCriterionToModel(gradeDefence))
				}
			}
		}
		return result, nil
	}
}

// ============================================
// THESIS BATCH FUNCTIONS (Topic, TopicCouncil, Midterm, Final, GradeReview, Supervisor)
// ============================================

// createStudentTopicCouncilInfoBatchFunc creates a batch function for loading StudentTopicCouncil
func createStudentTopicCouncilInfoBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.StudentTopicCouncil] {
	return func(ctx context.Context, ids []string) (map[string]*model.StudentTopicCouncil, error) {
		result := make(map[string]*model.StudentTopicCouncil)
		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetTopicCouncilsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				topicCouncil, err := client.GetTopicCouncilById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch topic council %s: %v", id, err)
					continue
				}

				if topicCouncil != nil && topicCouncil.GetTopicCouncil() != nil {
					result[id] = convert.PbTopicCouncilToStudentTopicCouncil(topicCouncil.GetTopicCouncil())
				}
			}
		}
		if resp != nil && resp.TopicCouncils != nil {
			for _, pbTopicCouncil := range resp.TopicCouncils {
				if pbTopicCouncil != nil {
					result[pbTopicCouncil.Id] = convert.PbTopicCouncilToStudentTopicCouncil(pbTopicCouncil)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d topic councils successfully", len(result), len(ids))
		return result, nil
	}
}

// createTopicForStudentBatchFunc creates a batch function for loading StudentTopic
func createTopicForStudentBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.StudentTopic] {
	return func(ctx context.Context, ids []string) (map[string]*model.StudentTopic, error) {
		result := make(map[string]*model.StudentTopic)

		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetTopicsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				topic, err := client.GetTopicById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch topic %s: %v", id, err)
					continue
				}

				if topic != nil && topic.Topic != nil {
					result[id] = convert.PbTopicToStudentTopic(topic.Topic)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Topics != nil {
			for _, pbTopic := range resp.Topics {
				if pbTopic != nil {
					result[pbTopic.Id] = convert.PbTopicToStudentTopic(pbTopic)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d topics successfully", len(result), len(ids))
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
					log.Printf("[DataLoader] Failed to fetch midterm %s: %v", id, err)
					continue
				}

				if midterm != nil && midterm.Midterm != nil {
					result[id] = convertPbMidtermToModel(midterm.Midterm)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Midterms != nil {
			for _, pbMidterm := range resp.Midterms {
				if pbMidterm != nil {
					result[pbMidterm.Id] = convertPbMidtermToModel(pbMidterm)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d midterms successfully", len(result), len(ids))
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
					log.Printf("[DataLoader] Failed to fetch final %s: %v", id, err)
					continue
				}

				if final != nil && final.Final != nil {
					result[id] = convertPbFinalToModel(final.Final)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Finals != nil {
			for _, pbFinal := range resp.Finals {
				if pbFinal != nil {
					result[pbFinal.Id] = convertPbFinalToModel(pbFinal)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d finals successfully", len(result), len(ids))
		return result, nil
	}
}

// createGradeViewBatchFunc creates a batch function for loading GradeReview
func createGradeViewBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.GradeReview] {
	return func(ctx context.Context, ids []string) (map[string]*model.GradeReview, error) {
		result := make(map[string]*model.GradeReview)
		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetGradeReviewsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				gradeReview, err := client.GetGradeReviewById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch grade review %s: %v", id, err)
				}
				if gradeReview != nil && gradeReview.GetGradeReview() != nil {
					result[id] = convert.PbGradeReviewToModel(gradeReview.GetGradeReview())
				}
			}
		}
		if resp != nil && resp.GradeReviews != nil {
			for _, pbGradeReview := range resp.GradeReviews {
				if pbGradeReview != nil {
					result[pbGradeReview.Id] = convert.PbGradeReviewToModel(pbGradeReview)
				}
			}
		}
		log.Printf("[DataLoader] Batch loaded %d/%d grade reviews successfully", len(result), len(ids))
		return result, nil
	}
}

// createSupervisorForStudentBatchFunc creates a batch function for loading StudentTopicSupervisor
func createSupervisorForStudentBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.StudentTopicSupervisor] {
	return func(ctx context.Context, ids []string) (map[string][]*model.StudentTopicSupervisor, error) {
		result := make(map[string][]*model.StudentTopicSupervisor)
		if len(ids) == 0 {
			return result, nil
		}
		newSearch := pb.SearchRequest{
			Pagination: &pb.Pagination{
				Page:       1,
				PageSize:   int32(len(ids) * 5),
				SortBy:     "created_at",
				Descending: true,
			},
			Filters: []*pb.FilterCriteria{
				{
					Criteria: &pb.FilterCriteria_Condition{
						Condition: &pb.FilterCondition{
							Field:    "topic_council_code",
							Operator: pb.FilterOperator_IN,
							Values:   ids,
						},
					},
				},
			},
		}
		resp, err := client.GetTopicCouncilSupervisorBySearch(ctx, &newSearch)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				newSearch = pb.SearchRequest{
					Pagination: &pb.Pagination{
						Page:       1,
						PageSize:   int32(len(ids) * 5),
						SortBy:     "created_at",
						Descending: true,
					},
					Filters: []*pb.FilterCriteria{
						{
							Criteria: &pb.FilterCriteria_Condition{
								Condition: &pb.FilterCondition{
									Field:    "topic_council_code",
									Operator: pb.FilterOperator_EQUAL,
									Values:   []string{id},
								},
							},
						},
					},
				}
				supervisor, err := client.GetTopicCouncilSupervisorBySearch(ctx, &newSearch)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch supervisors %s: %v", id, err)
					continue
				}
				if supervisor != nil {
					result[id] = convert.PbTopicCouncilSupervisorsToStudentTopicSupervisors(supervisor.GetTopicCouncilSupervisors())
				}
			}
			return result, err
		}
		if resp != nil {
			for _, supervisor := range resp.GetTopicCouncilSupervisors() {
				if supervisor != nil {
					result[supervisor.TopicCouncilCode] = append(result[supervisor.TopicCouncilCode], convert.PbTopicCouncilSupervisorToStudentTopicSupervisor(supervisor))
				}
			}
		}
		return result, err
	}
}

// createTopicBatchFunc creates a batch function for loading Topic (general)
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
					log.Printf("[DataLoader] Failed to fetch topic %s: %v", id, err)
					continue
				}

				if topic != nil && topic.Topic != nil {
					result[id] = convertPbTopicToModel(topic.Topic)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Topics != nil {
			for _, pbTopic := range resp.Topics {
				if pbTopic != nil {
					result[pbTopic.Id] = convertPbTopicToModel(pbTopic)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d topics successfully", len(result), len(ids))
		return result, nil
	}
}

// ============================================
// USER BATCH FUNCTIONS (Teacher for student view)
// ============================================

// createTeacherInfoBatchFunc creates a batch function for loading StudentTeacherInfo
func createTeacherInfoBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.StudentTeacherInfo] {
	return func(ctx context.Context, ids []string) (map[string]*model.StudentTeacherInfo, error) {
		result := make(map[string]*model.StudentTeacherInfo)

		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetTeachersByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				Teacher, err := client.GetTeacherById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Teacher %s: %v", id, err)
					continue
				}

				if Teacher != nil && Teacher.Teacher != nil {
					result[id] = convert.PbTeacherToStudentTeacherInfo(Teacher.Teacher)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Teachers != nil {
			for _, pbTeacher := range resp.Teachers {
				if pbTeacher != nil {
					result[pbTeacher.Id] = convert.PbTeacherToStudentTeacherInfo(pbTeacher)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d Teachers successfully", len(result), len(ids))
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

// convertPbTopicToModel converts protobuf Topic to GraphQL model
func convertPbTopicToModel(pbTopic *pbThesis.Topic) *model.Topic {
	if pbTopic == nil {
		return nil
	}

	var status model.TopicStatus
	switch pbTopic.Status {
	case pbThesis.TopicStatus_IN_PROGRESS:
		status = model.TopicStatusInProgress
	case pbThesis.TopicStatus_REJECTED:
		status = model.TopicStatusRejected
	case pbThesis.TopicStatus_TOPIC_COMPLETED:
		status = model.TopicStatusTopicCompleted
	default:
		status = model.TopicStatusInProgress
	}

	modelTopic := &model.Topic{
		ID:           pbTopic.Id,
		Title:        pbTopic.Title,
		MajorCode:    pbTopic.MajorCode,
		SemesterCode: pbTopic.SemesterCode,
		Status:       status,
	}

	if pbTopic.CreatedBy != "" {
		createdBy := pbTopic.CreatedBy
		modelTopic.CreatedBy = &createdBy
	}
	if pbTopic.UpdatedBy != "" {
		updatedBy := pbTopic.UpdatedBy
		modelTopic.UpdatedBy = &updatedBy
	}

	if pbTopic.CreatedAt != nil {
		createdAt := pbTopic.CreatedAt.AsTime()
		modelTopic.CreatedAt = &createdAt
	}
	if pbTopic.UpdatedAt != nil {
		updatedAt := pbTopic.UpdatedAt.AsTime()
		modelTopic.UpdatedAt = &updatedAt
	}

	return modelTopic
}
