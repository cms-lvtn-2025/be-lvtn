package dataloader

import (
	"context"
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
				PageSize:   10,
				Descending: true,
				SortBy:     "created_at",
			},
			Filters: []*pb.FilterCriteria{
				{
					Criteria: &pb.FilterCriteria_Condition{
						Condition: &pb.FilterCondition{
							Field:    "council_id",
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
									Field:    "council_id",
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
					result[pbDefences.Id] = append(result[pbDefences.Id], convert.PbDefenceToStudentDefenceInfo(pbDefences))
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d Defences successfully", len(result), len(ids))
		return result, nil
	}
}
