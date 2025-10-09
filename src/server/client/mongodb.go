package client

import (
	"context"
	"fmt"
	"log"
	"thaily/src/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.MongoConfig
}

// GetClient returns the underlying MongoDB client
func (m *MongoClient) GetClient() *mongo.Client {
	return m.client
}

// GetDatabase returns the default database
func (m *MongoClient) GetDatabase() *mongo.Database {
	return m.database
}

// GetCollection returns a collection from the default database
func (m *MongoClient) GetCollection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// Ping checks if MongoDB connection is alive
func (m *MongoClient) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return m.client.Ping(ctx, readpref.Primary())
}

// Close disconnects from MongoDB
func (m *MongoClient) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(cfg config.MongoConfig) (*MongoClient, error) {
	// Set client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(uint64(cfg.MaxPoolSize)).
		SetMinPoolSize(uint64(cfg.MinPoolSize)).
		SetMaxConnIdleTime(time.Duration(cfg.MaxConnIdleTime) * time.Second).
		SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Second).
		SetServerSelectionTimeout(time.Duration(cfg.ServerSelectionTimeout) * time.Second)

	// Set auth if credentials are provided
	if cfg.Username != "" && cfg.Password != "" {
		credential := options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		}
		if cfg.AuthSource != "" {
			credential.AuthSource = cfg.AuthSource
		}
		clientOptions.SetAuth(credential)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Printf("Connected to MongoDB at %s (Database: %s)", cfg.URI, cfg.Database)

	// Get database
	database := client.Database(cfg.Database)

	return &MongoClient{
		client:   client,
		database: database,
		config:   &cfg,
	}, nil
}
