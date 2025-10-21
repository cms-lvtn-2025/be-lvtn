package dataloader

import (
	"context"
	"fmt"
	"log"
	pb "thaily/proto/common"
	"thaily/src/graph/convert"

	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// createDefenceBatchFunc creates a batch function for loading councils
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
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful\", len(result), len(ids)")
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
