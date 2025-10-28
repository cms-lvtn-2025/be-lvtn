# DataLoader Architecture

## Overview
This package implements Facebook's DataLoader pattern for batching and caching database queries in GraphQL resolvers.

## File Structure

```
src/graph/dataloader/
├── dataloader.go          # Core DataLoader implementation (generic, reusable)
├── loaders.go             # Loaders registry and initialization
├── batch_thesis.go        # Batch functions for thesis-related entities
├── batch_council.go       # Batch functions for council-related entities
├── batch_user.go          # Batch functions for user-related entities (future)
├── context.go             # Context helpers for middleware
└── ARCHITECTURE.md        # This file
```

## How It Works

### 1. File Organization

#### `dataloader.go` - Core Implementation
- Generic DataLoader implementation using Go generics
- Handles batching, caching (L1/L2), and chunking
- **DO NOT MODIFY** unless changing core DataLoader behavior
- Reusable across all entity types

#### `loaders.go` - Registry
- Defines `Loaders` struct with all DataLoader instances
- `NewLoaders()` function creates configured instances
- **ADD NEW LOADERS HERE** when adding new entities
- Example:
```go
type Loaders struct {
    CouncilByID *DataLoader[string, *model.Council]
    MidtermByID *DataLoader[string, *model.Midterm]
    FinalByID   *DataLoader[string, *model.Final]
    TopicByID   *DataLoader[string, *model.Topic]
    // Add your new loader here
}
```

#### `batch_*.go` - Entity-Specific Batch Functions
- One file per domain/service (e.g., `batch_thesis.go`, `batch_council.go`)
- Contains batch functions and model converters for that domain
- **ADD NEW FILES** for new domains (e.g., `batch_user.go`, `batch_academic.go`)

### 2. Adding a New DataLoader

#### Step 1: Create batch function in appropriate file

```go
// In batch_thesis.go (or create new file like batch_enrollment.go)

func createEnrollmentBatchFunc(client *client.GRPCthesis) BatchFunc[string, *model.Enrollment] {
    return func(ctx context.Context, ids []string) (map[string]*model.Enrollment, error) {
        result := make(map[string]*model.Enrollment)

        if len(ids) == 0 {
            return result, nil
        }

        // Try batch method first
        resp, err := client.GetEnrollmentsByIds(ctx, ids)
        if err != nil {
            log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
            // Fallback to individual fetching
            for _, id := range ids {
                enrollment, err := client.GetEnrollmentById(ctx, id)
                if err != nil {
                    log.Printf("[DataLoader] Failed to fetch enrollment %s: %v", id, err)
                    continue
                }
                if enrollment != nil && enrollment.Enrollment != nil {
                    result[id] = convertPbEnrollmentToModel(enrollment.Enrollment)
                }
            }
            return result, nil
        }

        // Map batch results
        if resp != nil && resp.Enrollments != nil {
            for _, pbEnrollment := range resp.Enrollments {
                if pbEnrollment != nil {
                    result[pbEnrollment.Id] = convertPbEnrollmentToModel(pbEnrollment)
                }
            }
        }

        log.Printf("[DataLoader] Batch loaded %d/%d enrollments successfully", len(result), len(ids))
        return result, nil
    }
}

func convertPbEnrollmentToModel(pb *pb.Enrollment) *model.Enrollment {
    // Convert protobuf to GraphQL model
    // ...
}
```

#### Step 2: Register in `loaders.go`

```go
type Loaders struct {
    // ... existing loaders
    EnrollmentByID *DataLoader[string, *model.Enrollment]  // ADD THIS
}

func NewLoaders(...) *Loaders {
    defaultConfig := &Config{
        BatchWindow:  2 * time.Millisecond,
        MaxBatchSize: 300,
        L2TTL:        5 * time.Minute,
    }

    return &Loaders{
        // ... existing loaders
        EnrollmentByID: NewDataLoader(
            createEnrollmentBatchFunc(thesisClient),  // ADD THIS
            defaultConfig,
        ),
    }
}
```

#### Step 3: Use in resolver

```go
func (r *enrollmentResolver) Midterm(ctx context.Context, obj *model.Enrollment) (*model.Midterm, error) {
    loaders := dataloader.GetLoaders(ctx)
    return loaders.MidtermByID.Load(ctx, obj.MidtermCode)
}
```

### 3. Best Practices

#### Batch Function Pattern
1. **Always check for empty input**: `if len(ids) == 0 { return result, nil }`
2. **Try batch method first**: Use `GetXByIds()` if available
3. **Fallback to individual**: If batch fails, fetch individually
4. **Never fail entire batch**: Skip failed items, continue with others
5. **Log everything**: Use `log.Printf()` for debugging

#### Model Conversion
- Create separate `convertPbXToModel()` functions
- Handle nil checks
- Handle optional fields (use pointers)
- Convert enums correctly
- Convert timestamps using `AsTime()`

#### File Organization
- Group related entities in same file (e.g., Midterm, Final, Topic in `batch_thesis.go`)
- Keep files under 500 lines - split if larger
- Name files `batch_<domain>.go` (e.g., `batch_user.go`, `batch_academic.go`)

### 4. Configuration

All loaders use the same default configuration (defined in `loaders.go`):

```go
defaultConfig := &Config{
    BatchWindow:  2 * time.Millisecond,  // Time to collect IDs before executing
    MaxBatchSize: 300,                   // Max IDs per database query
    L2TTL:        5 * time.Minute,       // Cache TTL (per-request only)
}
```

To use custom config for specific loader:
```go
EnrollmentByID: NewDataLoader(
    createEnrollmentBatchFunc(thesisClient),
    &Config{
        BatchWindow:  5 * time.Millisecond,  // Custom: longer window
        MaxBatchSize: 100,                   // Custom: smaller batches
        L2TTL:        10 * time.Minute,      // Custom: longer cache
    },
),
```

### 5. Multi-layer Caching

#### L1 Cache (Deduplication)
- Within single batch window (2ms)
- Example: `[1,2,3,1,2,3]` → deduplicated to `[1,2,3]`
- Automatic, no configuration needed

#### L2 Cache (In-Memory, Per-Request)
- Stored in RAM
- Shared across all batches **within same request**
- Expires after 5 minutes OR when request ends (whichever is first)
- Prevents duplicate queries in same request
- **NOT shared across requests** (by design, prevents stale data)

#### Redis Cache (Global, Cross-Request)
- Implemented in gRPC client layer (`src/server/client/thesis.go`)
- Shared across all requests
- TTL: 10 minutes (configurable)
- Check before database query

### 6. Example: Complete Flow

Request: Load topics for 100 students

```
1. GraphQL Resolver calls: loaders.TopicByID.Load(ctx, topicID)
2. DataLoader collects IDs for 2ms (BatchWindow)
3. After 2ms: 100 IDs collected
4. L1 Deduplication: 100 → 80 unique IDs
5. L2 Cache check: 80 → 60 not cached (20 cache hits)
6. MaxBatchSize chunking: 60 IDs → 1 chunk (under 300 limit)
7. Execute batch function: createTopicBatchFunc()
   - Check Redis: 60 → 40 not in Redis (20 Redis hits)
   - Database query: SELECT * FROM topics WHERE id IN (40 IDs)
   - Store 40 results in Redis
8. Store 60 results in L2 cache
9. Return all 100 results (80 unique + 20 duplicates)
```

### 7. Troubleshooting

#### Issue: "Batch loaded 0/100 successfully"
- Check gRPC client has `GetXByIds()` method
- Check batch function returns `map[string]*model.X` correctly
- Check model conversion handles nil values

#### Issue: "MaxBatchSize not working"
- Ensure `MaxBatchSize` is set in Config
- Check logs for chunking: `[DataLoader] Splitting into chunks`
- Verify chunks are executed sequentially

#### Issue: "Cache not working"
- L2 cache is per-request only (intentional)
- Check `L2TTL > 0` in Config
- Use Redis cache for cross-request caching

### 8. Migration Guide

#### Before (old structure)
All batch functions in `loaders.go`:
```go
// loaders.go (500+ lines)
func NewLoaders(...) { ... }
func createCouncilBatchFunc(...) { ... }  // 50 lines
func createMidtermBatchFunc(...) { ... }  // 80 lines
func convertPbMidtermToModel(...) { ... } // 40 lines
// ... 10 more functions
```

#### After (new structure)
Separated by domain:
```
loaders.go (50 lines)          - Registry only
batch_council.go (40 lines)    - Council batch logic
batch_thesis.go (300 lines)    - Midterm, Final, Topic batch logic
```

Benefits:
- ✅ Easy to find code (by domain)
- ✅ Scales to 100+ tables
- ✅ Multiple developers can work simultaneously
- ✅ Clear separation of concerns

## Summary

1. **Core logic** → `dataloader.go` (don't modify)
2. **Registry** → `loaders.go` (add new loaders here)
3. **Batch functions** → `batch_*.go` (create one file per domain)
4. **One file per domain** (thesis, council, user, academic, etc.)
5. **Follow the pattern** in existing files when adding new loaders
