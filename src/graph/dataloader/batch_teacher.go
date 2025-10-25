package dataloader

import (
	"context"
	"log"
	pb "thaily/proto/common"
	"thaily/src/graph/convert"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// Teacher-specific batch functions

// createCouncilMemberCouncilBatchFunc creates a batch function for loading CouncilMemberCouncil
func createCouncilMemberCouncilBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.CouncilMemberCouncil] {
	return func(ctx context.Context, ids []string) (map[string]*model.CouncilMemberCouncil, error) {
		result := make(map[string]*model.CouncilMemberCouncil)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetCouncilsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				council, err := client.GetCouncilById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch council %s: %v", id, err)
					continue
				}
				if council != nil && council.Council != nil {
					result[id] = convert.PbCouncilToCouncilMemberCouncil(council.Council)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Councils != nil {
			for _, pbCouncil := range resp.Councils {
				if pbCouncil != nil {
					result[pbCouncil.Id] = convert.PbCouncilToCouncilMemberCouncil(pbCouncil)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d councils successfully", len(result), len(ids))
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
									Field:    "defence_code",
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

		log.Printf("[DataLoader] Batch loaded grade defences for %d/%d defences", len(result), len(ids))
		return result, nil
	}
}

// createCouncilTopicCouncilBatchFunc creates a batch function for loading CouncilTopicCouncil
func createCouncilTopicCouncilBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.CouncilTopicCouncil] {
	return func(ctx context.Context, ids []string) (map[string]*model.CouncilTopicCouncil, error) {
		result := make(map[string]*model.CouncilTopicCouncil)
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
				if topicCouncil != nil && topicCouncil.TopicCouncil != nil {
					result[id] = convert.PbTopicCouncilToCouncilTopicCouncil(topicCouncil.TopicCouncil)
				}
			}
			return result, nil
		}

		if resp != nil && resp.TopicCouncils != nil {
			for _, pbTopicCouncil := range resp.TopicCouncils {
				if pbTopicCouncil != nil {
					result[pbTopicCouncil.Id] = convert.PbTopicCouncilToCouncilTopicCouncil(pbTopicCouncil)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d topic councils successfully", len(result), len(ids))
		return result, nil
	}
}

// createReviewerTopicCouncilBatchFunc creates a batch function for loading ReviewerTopicCouncil
func createReviewerTopicCouncilBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.ReviewerTopicCouncil] {
	return func(ctx context.Context, ids []string) (map[string]*model.ReviewerTopicCouncil, error) {
		result := make(map[string]*model.ReviewerTopicCouncil)
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
				if topicCouncil != nil && topicCouncil.TopicCouncil != nil {
					result[id] = convert.PbTopicCouncilToReviewerTopicCouncil(topicCouncil.TopicCouncil)
				}
			}
			return result, nil
		}

		if resp != nil && resp.TopicCouncils != nil {
			for _, pbTopicCouncil := range resp.TopicCouncils {
				if pbTopicCouncil != nil {
					result[pbTopicCouncil.Id] = convert.PbTopicCouncilToReviewerTopicCouncil(pbTopicCouncil)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d reviewer topic councils successfully", len(result), len(ids))
		return result, nil
	}
}

// createSupervisorTopicCouncilBatchFunc creates a batch function for loading SupervisorTopicCouncil
func createSupervisorTopicCouncilBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.SupervisorTopicCouncil] {
	return func(ctx context.Context, ids []string) (map[string]*model.SupervisorTopicCouncil, error) {
		result := make(map[string]*model.SupervisorTopicCouncil)
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
				if topicCouncil != nil && topicCouncil.TopicCouncil != nil {
					result[id] = convert.PbTopicCouncilToSupervisorTopicCouncil(topicCouncil.TopicCouncil)
				}
			}
			return result, nil
		}

		if resp != nil && resp.TopicCouncils != nil {
			for _, pbTopicCouncil := range resp.TopicCouncils {
				if pbTopicCouncil != nil {
					result[pbTopicCouncil.Id] = convert.PbTopicCouncilToSupervisorTopicCouncil(pbTopicCouncil)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d supervisor topic councils successfully", len(result), len(ids))
		return result, nil
	}
}

// createCouncilDefenceBatchFunc creates a batch function for loading CouncilDefence
func createCouncilDefenceBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.CouncilDefence] {
	return func(ctx context.Context, ids []string) (map[string]*model.CouncilDefence, error) {
		result := make(map[string]*model.CouncilDefence)
		if len(ids) == 0 {
			return result, nil
		}

		resp, err := client.GetDefencesByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				defence, err := client.GetDefenceById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch defence %s: %v", id, err)
					continue
				}
				if defence != nil && defence.Defence != nil {
					result[id] = convert.PbDefenceToCouncilDefence(defence.Defence)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Defences != nil {
			for _, pbDefence := range resp.Defences {
				if pbDefence != nil {
					result[pbDefence.Id] = convert.PbDefenceToCouncilDefence(pbDefence)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d council defences successfully", len(result), len(ids))
		return result, nil
	}
}

// createReviewerTopicBatchFunc creates a batch function for loading ReviewerTopic
func createReviewerTopicBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.ReviewerTopic] {
	return func(ctx context.Context, ids []string) (map[string]*model.ReviewerTopic, error) {
		result := make(map[string]*model.ReviewerTopic)
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
					result[id] = convert.PbTopicToReviewerTopic(topic.Topic)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Topics != nil {
			for _, pbTopic := range resp.Topics {
				if pbTopic != nil {
					result[pbTopic.Id] = convert.PbTopicToReviewerTopic(pbTopic)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d reviewer topics successfully", len(result), len(ids))
		return result, nil
	}
}

// createSupervisorTopicBatchFunc creates a batch function for loading SupervisorTopic
func createSupervisorTopicBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.SupervisorTopic] {
	return func(ctx context.Context, ids []string) (map[string]*model.SupervisorTopic, error) {
		result := make(map[string]*model.SupervisorTopic)
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
					result[id] = convert.PbTopicToSupervisorTopic(topic.Topic)
				}
			}
			return result, nil
		}

		if resp != nil && resp.Topics != nil {
			for _, pbTopic := range resp.Topics {
				if pbTopic != nil {
					result[pbTopic.Id] = convert.PbTopicToSupervisorTopic(pbTopic)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d supervisor topics successfully", len(result), len(ids))
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
					log.Printf("[DataLoader] Failed to fetch teacher %s: %v", id, err)
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

		log.Printf("[DataLoader] Batch loaded %d/%d teachers successfully", len(result), len(ids))
		return result, nil
	}
}
