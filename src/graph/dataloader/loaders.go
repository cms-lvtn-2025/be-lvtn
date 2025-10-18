package dataloader

import (
	"time"

	"thaily/src/graph/model"
	"thaily/src/server/client"
)

// Loaders holds all dataloaders for the application
type Loaders struct {
	CouncilByID *DataLoader[string, *model.Council]
	MidtermByID *DataLoader[string, *model.Midterm]
	FinalByID   *DataLoader[string, *model.Final]
	TopicByID   *DataLoader[string, *model.Topic]
}

// NewLoaders creates a new Loaders instance with all dataloaders
func NewLoaders(
	userClient *client.GRPCUser,
	thesisClient *client.GRPCthesis,
	councilClient *client.GRPCCouncil,
) *Loaders {
	// Default configuration for all loaders
	defaultConfig := &Config{
		BatchWindow:  2 * time.Millisecond,
		MaxBatchSize: 300,
		L2TTL:        5 * time.Minute,
	}

	return &Loaders{
		CouncilByID: NewDataLoader(
			createCouncilBatchFunc(councilClient),
			defaultConfig,
		),
		MidtermByID: NewDataLoader(
			createMidtermBatchFunc(thesisClient),
			defaultConfig,
		),
		FinalByID: NewDataLoader(
			createFinalBatchFunc(thesisClient),
			defaultConfig,
		),
		TopicByID: NewDataLoader(
			createTopicBatchFunc(thesisClient),
			defaultConfig,
		),
	}
}
