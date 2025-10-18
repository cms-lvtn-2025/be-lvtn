package dataloader

import (
	"context"
	"log"
	"time"

	pb "thaily/proto/thesis"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// Loaders holds all dataloaders for the application
type Loaders struct {
	CouncilByID *DataLoader[string, *model.Council]
	MidtermByID *DataLoader[string, *model.Midterm]
}

// NewLoaders creates a new Loaders instance with all dataloaders
func NewLoaders(
	userClient *client.GRPCUser,
	thesisClient *client.GRPCthesis,
	councilClient *client.GRPCCouncil,
) *Loaders {
	return &Loaders{

		CouncilByID: NewDataLoader(
			createCouncilBatchFunc(councilClient),
			&Config{
				BatchWindow:  2 * time.Millisecond,
				MaxBatchSize: 300,
				L2TTL:        5 * time.Minute,
			},
		),
		MidtermByID: NewDataLoader(
			createMidtermBatchFunc(thesisClient),
			&Config{
				BatchWindow:  2 * time.Millisecond,
				MaxBatchSize: 300,
				L2TTL:        5 * time.Minute,
			},
		),
	}
}

// createCouncilBatchFunc creates a batch function for loading councils
func createCouncilBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.Council] {
	return func(ctx context.Context, ids []string) (map[string]*model.Council, error) {
		result := make(map[string]*model.Council)

		// TODO: Implement batch fetching from gRPC client

		for _, id := range ids {
			council, err := client.GetCouncilById(ctx, id)
			if err != nil {
				continue
			}

			if council != nil && council.Council != nil {
				result[id] = &model.Council{
					ID:    council.Council.Id,
					Title: council.Council.Title,
					// Map other fields as needed
				}
			}
		}

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
