package client

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"thaily/src/config"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type RedisClient struct {
	client *redis.Client
	config *config.RedisConfig
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Ping checks if Redis connection is alive
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set sets a key-value pair with optional expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del deletes one or more keys
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration time for a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	// Create Redis client options
	opts := &redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	}

	// Create Redis client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis at %s (DB: %d)", cfg.Address, cfg.DB)

	return &RedisClient{
		client: client,
		config: &cfg,
	}, nil
}

// generateCacheKey creates a unique cache key from prefix and parameters
func GenerateCacheKey(prefix string, params interface{}) string {
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s%x", prefix, hash[:16])
}

// GetCachedProto retrieves cached protobuf message from Redis
func GetCachedProto(ctx context.Context, redisClient *redis.Client, key string, dest proto.Message) (bool, error) {
	if redisClient == nil {
		return false, nil
	}

	data, err := redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil // Cache miss
	}
	if err != nil {
		log.Printf("Redis get error for key %s: %v", key, err)
		return false, nil // Don't fail on cache errors
	}

	if err := proto.Unmarshal(data, dest); err != nil {
		log.Printf("Proto unmarshal error for key %s: %v", key, err)
		return false, nil
	}

	return true, nil
}

// SetCachedProto stores protobuf message in Redis
func SetCachedProto(ctx context.Context, redisClient *redis.Client, key string, msg proto.Message, ttl time.Duration) {
	if redisClient == nil {
		return
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Proto marshal error for key %s: %v", key, err)
		return
	}

	if err := redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("Redis set error for key %s: %v", key, err)
	}
}

// InvalidateCacheByPattern invalidates cache keys matching a pattern
func InvalidateCacheByPattern(ctx context.Context, redisClient *redis.Client, pattern string) error {
	if redisClient == nil {
		return nil
	}

	iter := redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("Failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	return iter.Err()
}

// InvalidateCacheByKey invalidates a specific cache key
func InvalidateCacheByKey(ctx context.Context, redisClient *redis.Client, key string) error {
	if redisClient == nil {
		return nil
	}

	if err := redisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("Failed to delete cache key %s: %v", key, err)
		return err
	}
	return nil
}
