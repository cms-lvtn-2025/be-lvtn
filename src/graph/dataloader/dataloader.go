package dataloader

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BatchFunc defines the function signature for batch loading
// Input: slice of keys, Output: map of key->value and error
type BatchFunc[K comparable, V any] func(ctx context.Context, keys []K) (map[K]V, error)

// DataLoader provides batching and caching for data fetching
type DataLoader[K comparable, V any] struct {
	// batchFn is the function that fetches multiple items at once
	batchFn BatchFunc[K, V]

	// L2 cache - persistent across requests (with TTL)
	l2Cache   map[K]*cacheEntry[V]
	l2Mutex   sync.RWMutex
	l2TTL     time.Duration
	l2Enabled bool

	// Batching configuration
	batchWindow  time.Duration // time to wait for collecting keys (default 2ms)
	maxBatchSize int           // maximum batch size (0 = unlimited)

	// Current batch state
	batch      map[K][]chan<- *Result[V] // pending requests grouped by key
	batchMutex sync.Mutex
	batchTimer *time.Timer
}

// Result represents the result of a load operation
type Result[V any] struct {
	Value V
	Error error
}

// cacheEntry stores cached value with expiration
type cacheEntry[V any] struct {
	value     V
	expiresAt time.Time
}

// Config for DataLoader
type Config struct {
	BatchWindow  time.Duration // default: 2ms
	MaxBatchSize int           // default: 0 (unlimited)
	L2TTL        time.Duration // default: 5 minutes, set to 0 to disable L2
}

// NewDataLoader creates a new DataLoader instance
func NewDataLoader[K comparable, V any](batchFn BatchFunc[K, V], cfg *Config) *DataLoader[K, V] {
	if cfg == nil {
		cfg = &Config{
			BatchWindow:  2 * time.Millisecond,
			MaxBatchSize: 0,
			L2TTL:        5 * time.Minute,
		}
	}

	if cfg.BatchWindow == 0 {
		cfg.BatchWindow = 2 * time.Millisecond
	}

	dl := &DataLoader[K, V]{
		batchFn:      batchFn,
		batch:        make(map[K][]chan<- *Result[V]),
		batchWindow:  cfg.BatchWindow,
		maxBatchSize: cfg.MaxBatchSize,
		l2Enabled:    cfg.L2TTL > 0,
		l2TTL:        cfg.L2TTL,
	}

	if dl.l2Enabled {
		dl.l2Cache = make(map[K]*cacheEntry[V])
		// Start cleanup goroutine for L2 cache
		go dl.cleanupL2Cache()
	}

	return dl
}

// Load loads a single key
func (dl *DataLoader[K, V]) Load(ctx context.Context, key K) (V, error) {
	// Check L2 cache first
	if dl.l2Enabled {
		if value, found := dl.getFromL2(key); found {
			fmt.Printf("[DataLoader] Cache HIT for key: %v\n", key)
			return value, nil
		}
	}

	fmt.Printf("[DataLoader] Cache MISS for key: %v (will batch)\n", key)

	// Create result channel
	resultCh := make(chan *Result[V], 1)

	// Add to batch
	dl.batchMutex.Lock()
	dl.batch[key] = append(dl.batch[key], resultCh)

	// Schedule batch execution if not already scheduled
	if dl.batchTimer == nil {
		dl.batchTimer = time.AfterFunc(dl.batchWindow, func() {
			dl.executeBatch(ctx)
		})
	}

	dl.batchMutex.Unlock()

	// Wait for result
	result := <-resultCh
	return result.Value, result.Error
}

// LoadMany loads multiple keys and returns results in the same order
func (dl *DataLoader[K, V]) LoadMany(ctx context.Context, keys []K) ([]V, []error) {
	// L1 deduplication: track unique keys and their positions
	type keyInfo struct {
		key       K
		positions []int
	}

	uniqueKeys := make(map[K]*keyInfo)
	keyOrder := []K{}

	for i, key := range keys {
		if info, exists := uniqueKeys[key]; exists {
			info.positions = append(info.positions, i)
		} else {
			info := &keyInfo{
				key:       key,
				positions: []int{i},
			}
			uniqueKeys[key] = info
			keyOrder = append(keyOrder, key)
		}
	}

	// Check L2 cache and separate cached vs uncached keys
	uncachedKeys := []K{}
	cachedResults := make(map[K]V)

	if dl.l2Enabled {
		for _, key := range keyOrder {
			if value, found := dl.getFromL2(key); found {
				cachedResults[key] = value
			} else {
				uncachedKeys = append(uncachedKeys, key)
			}
		}
	} else {
		uncachedKeys = keyOrder
	}

	// Fetch uncached keys if any
	fetchedResults := make(map[K]*Result[V])
	if len(uncachedKeys) > 0 {
		// Create result channels for uncached keys
		resultChannels := make(map[K]chan *Result[V])
		for _, key := range uncachedKeys {
			resultChannels[key] = make(chan *Result[V], 1)
		}

		// Add to batch
		dl.batchMutex.Lock()
		for key, ch := range resultChannels {
			dl.batch[key] = append(dl.batch[key], ch)
		}

		// Schedule batch execution if not already scheduled
		if dl.batchTimer == nil {
			dl.batchTimer = time.AfterFunc(dl.batchWindow, func() {
				dl.executeBatch(ctx)
			})
		}
		dl.batchMutex.Unlock()

		// Collect results
		for key, ch := range resultChannels {
			result := <-ch
			fetchedResults[key] = result
		}
	}

	// Reconstruct results in original order
	values := make([]V, len(keys))
	errors := make([]error, len(keys))

	for _, key := range keyOrder {
		var value V
		var err error

		// Get from cached or fetched results
		if cachedValue, isCached := cachedResults[key]; isCached {
			value = cachedValue
		} else if result, isFetched := fetchedResults[key]; isFetched {
			value = result.Value
			err = result.Error
		}

		// Fill all positions for this key
		info := uniqueKeys[key]
		for _, pos := range info.positions {
			values[pos] = value
			errors[pos] = err
		}
	}

	return values, errors
}

// executeBatch executes the current batch
func (dl *DataLoader[K, V]) executeBatch(ctx context.Context) {
	dl.batchMutex.Lock()

	// Take current batch
	currentBatch := dl.batch
	dl.batch = make(map[K][]chan<- *Result[V])
	dl.batchTimer = nil

	dl.batchMutex.Unlock()

	if len(currentBatch) == 0 {
		return
	}

	// Extract unique keys (L1 deduplication happens here)
	keys := make([]K, 0, len(currentBatch))
	for key := range currentBatch {
		keys = append(keys, key)
	}

	fmt.Printf("[DataLoader] Executing batch with %d unique keys: %v\n", len(keys), keys)

	// Execute batch function
	results, err := dl.batchFn(ctx, keys)

	if err != nil {
		fmt.Printf("[DataLoader] Batch execution failed: %v\n", err)
	} else {
		fmt.Printf("[DataLoader] Batch execution returned %d results\n", len(results))
	}

	// Distribute results to waiting channels
	successKeys := []K{}
	failedKeys := []K{}

	for key, channels := range currentBatch {
		var result *Result[V]

		if err != nil {
			result = &Result[V]{Error: err}
			failedKeys = append(failedKeys, key)
		} else if value, ok := results[key]; ok {
			result = &Result[V]{Value: value}

			// Store in L2 cache
			if dl.l2Enabled {
				dl.setToL2(key, value)
				successKeys = append(successKeys, key)
			}
		} else {
			var zero V
			result = &Result[V]{
				Value: zero,
				Error: fmt.Errorf("key not found in batch results: %v", key),
			}
			failedKeys = append(failedKeys, key)
		}

		// Send result to all waiting channels
		for _, ch := range channels {
			ch <- result
		}
	}

	if len(successKeys) > 0 {
		fmt.Printf("[DataLoader] Cached %d keys to L2: %v\n", len(successKeys), successKeys)
	}
	if len(failedKeys) > 0 {
		fmt.Printf("[DataLoader] Failed to load %d keys: %v\n", len(failedKeys), failedKeys)
	}
}

// L2 cache operations
func (dl *DataLoader[K, V]) getFromL2(key K) (V, bool) {
	dl.l2Mutex.RLock()
	defer dl.l2Mutex.RUnlock()

	entry, exists := dl.l2Cache[key]
	if !exists {
		var zero V
		return zero, false
	}

	// Check expiration
	if time.Now().After(entry.expiresAt) {
		var zero V
		return zero, false
	}

	return entry.value, true
}

func (dl *DataLoader[K, V]) setToL2(key K, value V) {
	dl.l2Mutex.Lock()
	defer dl.l2Mutex.Unlock()

	dl.l2Cache[key] = &cacheEntry[V]{
		value:     value,
		expiresAt: time.Now().Add(dl.l2TTL),
	}
}

// ClearL2 clears the L2 cache
func (dl *DataLoader[K, V]) ClearL2() {
	if !dl.l2Enabled {
		return
	}

	dl.l2Mutex.Lock()
	defer dl.l2Mutex.Unlock()

	dl.l2Cache = make(map[K]*cacheEntry[V])
}

// ClearL2Key removes a specific key from L2 cache
func (dl *DataLoader[K, V]) ClearL2Key(key K) {
	if !dl.l2Enabled {
		return
	}

	dl.l2Mutex.Lock()
	defer dl.l2Mutex.Unlock()

	delete(dl.l2Cache, key)
}

// cleanupL2Cache periodically removes expired entries
func (dl *DataLoader[K, V]) cleanupL2Cache() {
	if !dl.l2Enabled {
		return
	}

	ticker := time.NewTicker(dl.l2TTL / 2)
	defer ticker.Stop()

	for range ticker.C {
		dl.l2Mutex.Lock()
		now := time.Now()
		for key, entry := range dl.l2Cache {
			if now.After(entry.expiresAt) {
				delete(dl.l2Cache, key)
			}
		}
		dl.l2Mutex.Unlock()
	}
}

// Prime pre-loads a value into L2 cache
func (dl *DataLoader[K, V]) Prime(key K, value V) {
	if !dl.l2Enabled {
		return
	}
	dl.setToL2(key, value)
}
