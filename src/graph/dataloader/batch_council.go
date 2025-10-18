package dataloader

import (
	"context"
	"log"

	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// createCouncilBatchFunc creates a batch function for loading councils
func createCouncilBatchFunc(client *client.GRPCCouncil) BatchFunc[string, *model.Council] {
	return func(ctx context.Context, ids []string) (map[string]*model.Council, error) {
		result := make(map[string]*model.Council)

		// TODO: Implement batch fetching from gRPC client
		// When GetCouncilsByIds is implemented, replace this with:
		// resp, err := client.GetCouncilsByIds(ctx, ids)

		for _, id := range ids {
			council, err := client.GetCouncilById(ctx, id)
			if err != nil {
				log.Printf("[DataLoader] Failed to fetch council %s: %v", id, err)
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

		log.Printf("[DataLoader] Loaded %d/%d councils successfully", len(result), len(ids))
		return result, nil
	}
}
