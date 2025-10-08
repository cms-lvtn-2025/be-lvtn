# Redis Integration Guide

## Tổng quan

Hệ thống sử dụng **2 Redis instances** với mục đích khác nhau:

1. **Redis Cache** (Port 6379) - Caching tại Backend Gateway
2. **Redis Queue** (Port 6380) - BullMQ job processing tại Microservices

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Backend Gateway                        │
│  ┌──────────────┐         ┌──────────────────────┐     │
│  │  REST API    │────────▶│   Redis Cache        │     │
│  │  GraphQL     │         │   (Port 6379)        │     │
│  └──────────────┘         └──────────────────────┘     │
│         │                                                │
│         │ gRPC                                           │
│         ▼                                                │
└─────────────────────────────────────────────────────────┘
          │
          │
┌─────────▼────────────────────────────────────────────────┐
│                  Microservices                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │  User    │  │ Thesis   │  │  File    │               │
│  │ Service  │  │ Service  │  │ Service  │               │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘               │
│       │             │             │                      │
│       └─────────────┴─────────────┘                      │
│                     │                                    │
│         ┌───────────▼───────────────────────┐            │
│         │      Redis Queue (BullMQ)         │            │
│         │        (Port 6380)                │            │
│         │                                   │            │
│         │  ┌─────────────────────────────┐ │            │
│         │  │  Job Queue:                 │ │            │
│         │  │  - email:send               │ │            │
│         │  │  - thesis:deadline-reminder │ │            │
│         │  │  - report:generate          │ │            │
│         │  └─────────────────────────────┘ │            │
│         └───────────────────────────────────┘            │
└──────────────────────────────────────────────────────────┘
```

---

## Part 1: Redis Cache (Backend Gateway)

### Purpose

- Cache GraphQL query results
- Cache user sessions
- Cache frequently accessed data
- Reduce database queries
- Improve response time

### Setup

#### 1. Cài đặt Redis Client cho Go

```bash
go get github.com/redis/go-redis/v9
```

#### 2. Tạo Cache Package

**File**: `src/pkg/cache/redis.go`

```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

type CacheConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewRedisCache(cfg CacheConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{
		client: client,
	}, nil
}

// Set stores a value with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a key
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if key exists
func (r *RedisCache) Exists(ctx context.Context, key string) bool {
	count, _ := r.client.Exists(ctx, key).Result()
	return count > 0
}

// SetNX sets only if key doesn't exist (for locks)
func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return r.client.SetNX(ctx, key, data, expiration).Result()
}

// Increment increments a counter
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// SetWithTags stores value with tags for bulk invalidation
func (r *RedisCache) SetWithTags(ctx context.Context, key string, value interface{}, tags []string, expiration time.Duration) error {
	// Store the actual data
	if err := r.Set(ctx, key, value, expiration); err != nil {
		return err
	}

	// Store tags mapping
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		r.client.SAdd(ctx, tagKey, key)
		r.client.Expire(ctx, tagKey, expiration)
	}

	return nil
}

// InvalidateByTag removes all keys with a specific tag
func (r *RedisCache) InvalidateByTag(ctx context.Context, tag string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)
	keys, err := r.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		r.client.Del(ctx, keys...)
	}

	return r.client.Del(ctx, tagKey).Err()
}
```

#### 3. Thêm vào Container

**File**: `src/pkg/container/container.go`

```go
type Container struct {
	Config   *config.Config
	Clients  *Clients
	Storage  *storage.MinioClient
	Cache    *cache.RedisCache  // Add cache
}

func New(cfg *config.Config) (*Container, error) {
	// ... existing code

	// Initialize Redis Cache
	redisCache, err := cache.NewRedisCache(cache.CacheConfig{
		Host:     cfg.Redis.Cache.Host,
		Port:     cfg.Redis.Cache.Port,
		Password: cfg.Redis.Cache.Password,
		DB:       cfg.Redis.Cache.DB,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init redis cache: %w", err)
	}

	return &Container{
		Config:  cfg,
		Clients: clients,
		Storage: minioClient,
		Cache:   redisCache,  // Add to container
	}, nil
}
```

#### 4. Sử dụng Cache trong API

**Example 1: Cache user data**

```go
func (h *APIHandler) GetCurrentUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("user_id")

	cacheKey := fmt.Sprintf("user:%s", userID)

	// Try cache first
	var user User
	err := h.Cache.Get(ctx, cacheKey, &user)
	if err == nil {
		// Cache hit
		response.Success(c, user)
		return
	}

	// Cache miss - fetch from service
	userData, err := h.UserClient.GetUser(ctx, &pb.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		response.InternalError(c, "Failed to get user")
		return
	}

	// Store in cache (10 minutes)
	h.Cache.Set(ctx, cacheKey, userData, 10*time.Minute)

	response.Success(c, userData)
}
```

**Example 2: Cache with tags (for invalidation)**

```go
func (h *APIHandler) GetThesisList(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("user_id")

	cacheKey := fmt.Sprintf("thesis:list:%s", userID)

	// Try cache
	var theses []Thesis
	err := h.Cache.Get(ctx, cacheKey, &theses)
	if err == nil {
		response.Success(c, theses)
		return
	}

	// Fetch from service
	result, err := h.ThesisClient.GetThesisList(ctx, &pb.GetThesisListRequest{
		UserId: userID,
	})
	if err != nil {
		response.InternalError(c, "Failed to get thesis list")
		return
	}

	// Cache with tags (for easy invalidation)
	tags := []string{
		fmt.Sprintf("user:%s", userID),
		"thesis",
	}
	h.Cache.SetWithTags(ctx, cacheKey, result.Theses, tags, 5*time.Minute)

	response.Success(c, result.Theses)
}

// When user creates a new thesis, invalidate cache
func (h *APIHandler) CreateThesis(c *gin.Context) {
	// ... create thesis logic

	// Invalidate user's thesis list cache
	userID := c.GetString("user_id")
	h.Cache.InvalidateByTag(ctx, fmt.Sprintf("user:%s", userID))

	response.Success(c, thesis)
}
```

**Example 3: Rate limiting**

```go
func RateLimitMiddleware(cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userID := c.GetString("user_id")
		key := fmt.Sprintf("rate_limit:%s", userID)

		// Increment counter
		count, err := cache.Increment(ctx, key)
		if err != nil {
			count = 1
			cache.Set(ctx, key, 1, 1*time.Minute)
		}

		// Check limit (100 requests per minute)
		if count > 100 {
			response.Error(c, 429, "Too many requests")
			c.Abort()
			return
		}

		c.Next()
	}
}
```

---

## Part 2: Redis Queue (BullMQ - Microservices)

### Purpose

- Background job processing
- Send email notifications
- Generate reports
- Scheduled tasks
- Batch processing

### Architecture

```
Service → Add Job → Redis Queue → Worker → Execute Job
```

### Setup (Node.js Service Example)

#### 1. Cài đặt BullMQ

```bash
npm install bullmq ioredis
```

#### 2. Tạo Queue Manager

**File**: `src/service/user/queue/queue.js`

```javascript
const { Queue, Worker } = require('bullmq');

// Redis connection config
const redisConnection = {
  host: process.env.REDIS_QUEUE_HOST || 'localhost',
  port: process.env.REDIS_QUEUE_PORT || 6380,
  password: process.env.REDIS_QUEUE_PASSWORD || '',
};

// Create queues
const emailQueue = new Queue('email', { connection: redisConnection });
const reportQueue = new Queue('report', { connection: redisConnection });
const reminderQueue = new Queue('reminder', { connection: redisConnection });

module.exports = {
  emailQueue,
  reportQueue,
  reminderQueue,
};
```

#### 3. Tạo Workers

**File**: `src/service/user/queue/workers/email.worker.js`

```javascript
const { Worker } = require('bullmq');
const nodemailer = require('nodemailer');

const redisConnection = {
  host: process.env.REDIS_QUEUE_HOST || 'localhost',
  port: process.env.REDIS_QUEUE_PORT || 6380,
};

// Email transporter
const transporter = nodemailer.createTransport({
  host: process.env.SMTP_HOST,
  port: process.env.SMTP_PORT,
  auth: {
    user: process.env.SMTP_USER,
    pass: process.env.SMTP_PASS,
  },
});

// Email worker
const emailWorker = new Worker(
  'email',
  async (job) => {
    const { to, subject, template, data } = job.data;

    console.log(`Processing email job: ${job.id}`);

    // Render template
    const html = renderTemplate(template, data);

    // Send email
    const info = await transporter.sendMail({
      from: process.env.SMTP_FROM,
      to,
      subject,
      html,
    });

    console.log(`Email sent: ${info.messageId}`);

    return { messageId: info.messageId };
  },
  {
    connection: redisConnection,
    concurrency: 5, // Process 5 jobs concurrently
  }
);

// Event handlers
emailWorker.on('completed', (job) => {
  console.log(`Job ${job.id} completed`);
});

emailWorker.on('failed', (job, err) => {
  console.error(`Job ${job.id} failed:`, err);
});

function renderTemplate(template, data) {
  // TODO: Implement template rendering
  return `<h1>${template}</h1>`;
}

module.exports = emailWorker;
```

**File**: `src/service/thesis/queue/workers/reminder.worker.js`

```javascript
const { Worker } = require('bullmq');
const { emailQueue } = require('../queue');

const reminderWorker = new Worker(
  'reminder',
  async (job) => {
    const { type, thesisId, userId } = job.data;

    console.log(`Processing reminder: ${type} for thesis ${thesisId}`);

    // Get thesis and user data
    const thesis = await getThesis(thesisId);
    const user = await getUser(userId);

    // Add email job
    await emailQueue.add('send', {
      to: user.email,
      subject: 'Thesis Deadline Reminder',
      template: 'thesis_reminder',
      data: {
        userName: user.name,
        thesisTitle: thesis.title,
        deadline: thesis.deadline,
      },
    });

    return { success: true };
  },
  {
    connection: {
      host: process.env.REDIS_QUEUE_HOST || 'localhost',
      port: process.env.REDIS_QUEUE_PORT || 6380,
    },
  }
);

module.exports = reminderWorker;
```

#### 4. Add Jobs từ Service

**Example: Thesis Service gọi queue**

```javascript
const { emailQueue, reminderQueue } = require('./queue/queue');

class ThesisService {
  async submitThesis(thesisId, userId) {
    // Save thesis to database
    await this.thesisRepo.updateStatus(thesisId, 'SUBMITTED');

    // Add email notification job
    await emailQueue.add('send', {
      to: 'advisor@university.edu',
      subject: 'New Thesis Submitted',
      template: 'thesis_submitted',
      data: {
        thesisId,
        userId,
      },
    });

    console.log('Email job added to queue');
  }

  async scheduleDeadlineReminder(thesisId, userId, deadline) {
    // Schedule reminder 7 days before deadline
    const reminderDate = new Date(deadline);
    reminderDate.setDate(reminderDate.getDate() - 7);

    const delay = reminderDate.getTime() - Date.now();

    if (delay > 0) {
      await reminderQueue.add(
        'deadline',
        {
          type: 'deadline',
          thesisId,
          userId,
        },
        {
          delay, // Delay in milliseconds
        }
      );

      console.log(`Reminder scheduled for ${reminderDate}`);
    }
  }
}
```

#### 5. Start Workers

**File**: `src/service/user/index.js`

```javascript
const emailWorker = require('./queue/workers/email.worker');
const reminderWorker = require('./queue/workers/reminder.worker');

// Start gRPC server
startGRPCServer();

// Workers are already running after import
console.log('Workers started');

// Graceful shutdown
process.on('SIGTERM', async () => {
  await emailWorker.close();
  await reminderWorker.close();
  process.exit(0);
});
```

### Job Patterns

#### 1. Simple Job

```javascript
await emailQueue.add('send', {
  to: 'user@example.com',
  subject: 'Hello',
  message: 'World',
});
```

#### 2. Delayed Job

```javascript
await emailQueue.add(
  'send',
  { to: 'user@example.com', subject: 'Reminder' },
  { delay: 3600000 } // 1 hour delay
);
```

#### 3. Repeatable Job (Cron)

```javascript
await reportQueue.add(
  'daily-report',
  {},
  {
    repeat: {
      pattern: '0 0 * * *', // Daily at midnight
    },
  }
);
```

#### 4. Priority Job

```javascript
await emailQueue.add(
  'urgent-email',
  { to: 'admin@example.com', subject: 'URGENT' },
  { priority: 1 } // Higher priority
);
```

#### 5. Job with Retry

```javascript
await emailQueue.add(
  'send',
  { to: 'user@example.com' },
  {
    attempts: 3, // Retry 3 times
    backoff: {
      type: 'exponential',
      delay: 1000, // Start with 1 second
    },
  }
);
```

---

## Configuration

### Config structure

**File**: `src/config/config.go`

```go
type Config struct {
	Server   ServerConfig
	Services ServiceConfig
	Redis    RedisConfig  // Add Redis config
}

type RedisConfig struct {
	Cache CacheRedisConfig
	Queue QueueRedisConfig
}

type CacheRedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type QueueRedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() (*Config, error) {
	// ...

	cfg := &Config{
		// ... existing configs

		Redis: RedisConfig{
			Cache: CacheRedisConfig{
				Host:     getEnv("REDIS_CACHE_HOST", "localhost"),
				Port:     getEnv("REDIS_CACHE_PORT", "6379"),
				Password: os.Getenv("REDIS_CACHE_PASSWORD"),
				DB:       getEnvInt("REDIS_CACHE_DB", 0),
			},
			Queue: QueueRedisConfig{
				Host:     getEnv("REDIS_QUEUE_HOST", "localhost"),
				Port:     getEnv("REDIS_QUEUE_PORT", "6380"),
				Password: os.Getenv("REDIS_QUEUE_PASSWORD"),
				DB:       getEnvInt("REDIS_QUEUE_DB", 0),
			},
		},
	}

	return cfg, nil
}
```

---

## Monitoring

### BullMQ UI (Optional)

```bash
npm install -g bull-board

# Run UI
bull-board --redis redis://localhost:6380
```

Visit: http://localhost:3000

### Redis CLI Monitoring

```bash
# Connect to cache
redis-cli -p 6379

# Connect to queue
redis-cli -p 6380

# Monitor commands
MONITOR

# Check queue stats
LLEN bull:email:wait
LLEN bull:email:active
LLEN bull:email:failed
```

---

## Best Practices

### 1. Cache Keys Naming

```
<entity>:<id>:<field>

Examples:
user:123:profile
thesis:456:metadata
user:123:thesis:list
```

### 2. Cache Expiration Strategy

```go
// Short-lived (1-5 minutes) - frequently changing data
cache.Set(ctx, key, data, 2*time.Minute)

// Medium-lived (10-30 minutes) - semi-static data
cache.Set(ctx, key, data, 15*time.Minute)

// Long-lived (1-24 hours) - static data
cache.Set(ctx, key, data, 6*time.Hour)
```

### 3. Job Naming Convention

```
<service>:<action>

Examples:
email:send
thesis:deadline-reminder
report:generate-pdf
data:sync-users
```

### 4. Error Handling

```javascript
worker.on('failed', async (job, err) => {
  console.error(`Job ${job.id} failed:`, err);

  // Log to monitoring system
  await logError({
    jobId: job.id,
    queue: job.queueName,
    error: err.message,
  });

  // Alert if critical
  if (job.data.critical) {
    await sendAlert({
      message: `Critical job failed: ${job.id}`,
    });
  }
});
```

---

## Troubleshooting

### Cache Issues

```bash
# Check Redis connection
redis-cli -p 6379 ping

# View all keys
redis-cli -p 6379 KEYS "*"

# Check memory usage
redis-cli -p 6379 INFO memory

# Clear all cache (use carefully!)
redis-cli -p 6379 FLUSHDB
```

### Queue Issues

```bash
# Check queue connection
redis-cli -p 6380 ping

# View pending jobs
redis-cli -p 6380 LLEN bull:email:wait

# View failed jobs
redis-cli -p 6380 LLEN bull:email:failed

# Clear failed jobs
redis-cli -p 6380 DEL bull:email:failed
```
