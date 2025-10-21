package dataloader

import (
	"context"
	"log"
	pbCommon "thaily/proto/common"
	"thaily/src/graph/convert"

	pb "thaily/proto/thesis"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

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
					// Skip failed items - don't fail entire batch
					log.Printf("[DataLoader] Failed to fetch midterm %s: %v", id, err)
					continue
				}

				if topicCouncil != nil && topicCouncil.GetTopicCouncil() != nil {
					result[id] = convert.PbTopicCouncilToStudentTopicCouncil(topicCouncil.GetTopicCouncil())
				}
			}
		}
		// Map batch results
		if resp != nil && resp.TopicCouncils != nil {
			for _, pbTopicCouncil := range resp.TopicCouncils {
				if pbTopicCouncil != nil {
					result[pbTopicCouncil.Id] = convert.PbTopicCouncilToStudentTopicCouncil(pbTopicCouncil)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d topic successfully", len(result), len(ids))
		return result, nil
	}
}

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
					log.Printf("[DataLoader] Failed to fetch midterm %s: %v", id, err)
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
		log.Printf("[DataLoader] Batch loaded %d/%d topic successfully", len(result), len(ids))
		return result, nil
	}
}

// createMidtermBatchFunc creates a batch function for loading midterms
func createMidtermBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Midterm] {
	return func(ctx context.Context, ids []string) (map[string]*model.Midterm, error) {
		result := make(map[string]*model.Midterm)

		if len(ids) == 0 {
			return result, nil
		}

		// Use batch fetching method
		resp, err := client.GetMidtermsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			// Fallback to individual fetching if batch fails
			for _, id := range ids {
				midterm, err := client.GetMidtermById(ctx, id)
				if err != nil {
					// Skip failed items - don't fail entire batch
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

		// Map batch results
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

// createFinalBatchFunc creates a batch function for loading finals
func createFinalBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Final] {
	return func(ctx context.Context, ids []string) (map[string]*model.Final, error) {
		result := make(map[string]*model.Final)

		if len(ids) == 0 {
			return result, nil
		}

		// Use batch fetching method
		resp, err := client.GetFinalsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			// Fallback to individual fetching if batch fails
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

		// Map batch results
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

func createTopicForStudentBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.StudentTopic] {
	return func(ctx context.Context, ids []string) (map[string]*model.StudentTopic, error) {
		result := make(map[string]*model.StudentTopic)

		if len(ids) == 0 {
			return result, nil
		}

		// Use batch fetching method
		resp, err := client.GetTopicsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			// Fallback to individual fetching if batch fails
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

		// Map batch results
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

func createSupervisorForStudentBatchFunc(client *client.GRPCthesis) BatchFunc[string, []*model.StudentTopicSupervisor] {
	return func(ctx context.Context, ids []string) (map[string][]*model.StudentTopicSupervisor, error) {
		result := make(map[string][]*model.StudentTopicSupervisor)
		if len(ids) == 0 {
			return result, nil
		}
		newSearch := pbCommon.SearchRequest{
			Pagination: &pbCommon.Pagination{
				Page:       1,
				PageSize:   int32(len(ids) * 5),
				SortBy:     "created_at",
				Descending: true,
			},
			Filters: []*pbCommon.FilterCriteria{
				{
					Criteria: &pbCommon.FilterCriteria_Condition{
						Condition: &pbCommon.FilterCondition{
							Field:    "topic_council_code",
							Operator: pbCommon.FilterOperator_IN,
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
				newSearch = pbCommon.SearchRequest{
					Pagination: &pbCommon.Pagination{
						Page:       1,
						PageSize:   int32(len(ids) * 5),
						SortBy:     "created_at",
						Descending: true,
					},
					Filters: []*pbCommon.FilterCriteria{
						{
							Criteria: &pbCommon.FilterCriteria_Condition{
								Condition: &pbCommon.FilterCondition{
									Field:    "topic_council_code",
									Operator: pbCommon.FilterOperator_EQUAL,
									Values:   []string{id},
								},
							},
						},
					},
				}
				supervisor, err := client.GetTopicCouncilSupervisorBySearch(ctx, &newSearch)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch topic %s: %v", id, err)
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

// createTopicBatchFunc creates a batch function for loading topics
func createTopicBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Topic] {
	return func(ctx context.Context, ids []string) (map[string]*model.Topic, error) {
		result := make(map[string]*model.Topic)

		if len(ids) == 0 {
			return result, nil
		}

		// Use batch fetching method
		resp, err := client.GetTopicsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			// Fallback to individual fetching if batch fails
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

		// Map batch results
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

// convertPbMidtermToModel converts protobuf Midterm to GraphQL model
func convertPbMidtermToModel(pbMidterm *pb.Midterm) *model.Midterm {
	if pbMidterm == nil {
		return nil
	}

	// Convert protobuf enum (int32) to GraphQL enum (string)
	var status model.MidtermStatus
	switch pbMidterm.Status {
	case pb.MidtermStatus_NOT_SUBMITTED:
		status = model.MidtermStatusNotSubmitted
	case pb.MidtermStatus_SUBMITTED:
		status = model.MidtermStatusSubmitted
	case pb.MidtermStatus_PASS:
		status = model.MidtermStatusPass
	case pb.MidtermStatus_FAIL:
		status = model.MidtermStatusFail
	default:
		status = model.MidtermStatusNotSubmitted
	}

	modelMidterm := &model.Midterm{
		ID:     pbMidterm.Id,
		Title:  pbMidterm.Title,
		Status: status,
	}

	// Handle optional primitive fields (need pointers)
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

	// Handle timestamps
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
func convertPbFinalToModel(pbFinal *pb.Final) *model.Final {
	if pbFinal == nil {
		return nil
	}

	// Convert protobuf enum to GraphQL enum
	var status model.FinalStatus
	switch pbFinal.Status {
	case pb.FinalStatus_PENDING:
		status = model.FinalStatusPending
	case pb.FinalStatus_COMPLETED:
		status = model.FinalStatusCompleted
	case pb.FinalStatus_FAILED:
		status = model.FinalStatusFailed
	case pb.FinalStatus_PASSED:
		status = model.FinalStatusPassed
	default:
		status = model.FinalStatusPending
	}

	modelFinal := &model.Final{
		ID:     pbFinal.Id,
		Title:  pbFinal.Title,
		Status: status,
	}

	// Handle optional fields
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

	// Handle timestamps
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
func convertPbTopicToModel(pbTopic *pb.Topic) *model.Topic {
	if pbTopic == nil {
		return nil
	}

	// Convert protobuf enum to GraphQL enum
	var status model.TopicStatus
	switch pbTopic.Status {
	case pb.TopicStatus_IN_PROGRESS:
		status = model.TopicStatusInProgress
	case pb.TopicStatus_REJECTED:
		status = model.TopicStatusRejected
	case pb.TopicStatus_TOPIC_COMPLETED:
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

	// Handle optional fields
	if pbTopic.CreatedBy != "" {
		createdBy := pbTopic.CreatedBy
		modelTopic.CreatedBy = &createdBy
	}
	if pbTopic.UpdatedBy != "" {
		updatedBy := pbTopic.UpdatedBy
		modelTopic.UpdatedBy = &updatedBy
	}

	// Handle timestamps
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
