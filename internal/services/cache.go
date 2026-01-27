package services

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cache entry with expiration
type CacheEntry struct {
	Value      interface{}
	ExpiresAt  time.Time
	CreatedAt  time.Time
	AccessCount int
}

// Cache provides in-memory caching with expiration
type Cache struct {
	data  map[string]*CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
	maxSize int
}

// NewCache creates a new cache with default TTL
func NewCache(ttl time.Duration, maxSize int) *Cache {
	return &Cache{
		data:    make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Set stores a value in cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entry if at capacity
	if len(c.data) >= c.maxSize && c.data[key] == nil {
		c.evictOldest()
	}

	c.data[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
	}
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		delete(c.data, key)
		return nil, false
	}

	// Update access count
	entry.AccessCount++
	return entry.Value, true
}

// Delete removes a value from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*CacheEntry)
}

// evictOldest removes the least recently used entry
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.data {
		if oldestTime.IsZero() || entry.AccessCount < c.data[oldestKey].AccessCount {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

// Size returns the number of entries in cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// CacheKey generates a cache key from components
func CacheKey(parts ...string) string {
	key := ""
	for _, part := range parts {
		key += part + ":"
	}
	return key[:len(key)-1]
}

// HashCacheKey generates a hash-based cache key for large inputs
func HashCacheKey(data string) string {
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("hash:%x", hash)
}

// CacheManager manages multiple caches for different data types
type CacheManager struct {
	caches map[string]*Cache
	mu     sync.RWMutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]*Cache),
	}
}

// CreateCache creates a new cache
func (cm *CacheManager) CreateCache(name string, ttl time.Duration, maxSize int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.caches[name] = NewCache(ttl, maxSize)
}

// GetCache retrieves a cache
func (cm *CacheManager) GetCache(name string) *Cache {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.caches[name]
}

// Set stores a value in a named cache
func (cm *CacheManager) Set(cacheName string, key string, value interface{}) {
	cache := cm.GetCache(cacheName)
	if cache != nil {
		cache.Set(key, value)
	}
}

// Get retrieves a value from a named cache
func (cm *CacheManager) Get(cacheName string, key string) (interface{}, bool) {
	cache := cm.GetCache(cacheName)
	if cache != nil {
		return cache.Get(key)
	}
	return nil, false
}

// CacheStats returns cache statistics
type CacheStats struct {
	Name      string
	Size      int
	MaxSize   int
	HitRate   float64
}

// Stats returns statistics for a cache
func (cm *CacheManager) Stats(cacheName string) *CacheStats {
	cm.mu.RLock()
	cache, exists := cm.caches[cacheName]
	cm.mu.RUnlock()

	if !exists {
		return nil
	}

	return &CacheStats{
		Name:    cacheName,
		Size:    cache.Size(),
		MaxSize: cache.maxSize,
	}
}

// BatchProcessor processes items in batches for efficiency
type BatchProcessor struct {
	batchSize   int
	maxWaitTime time.Duration
	processFn   func([]interface{}) error
	queue       []interface{}
	mu          sync.Mutex
	ticker      *time.Ticker
	done        chan struct{}
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int, maxWaitTime time.Duration, processFn func([]interface{}) error) *BatchProcessor {
	bp := &BatchProcessor{
		batchSize:   batchSize,
		maxWaitTime: maxWaitTime,
		processFn:   processFn,
		queue:       make([]interface{}, 0, batchSize),
		ticker:      time.NewTicker(maxWaitTime),
		done:        make(chan struct{}),
	}

	go bp.processLoop()
	return bp
}

// Add adds an item to the batch queue
func (bp *BatchProcessor) Add(item interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.queue = append(bp.queue, item)

	// Process if batch is full
	if len(bp.queue) >= bp.batchSize {
		return bp.flush()
	}

	return nil
}

// processLoop processes batches on timer
func (bp *BatchProcessor) processLoop() {
	for {
		select {
		case <-bp.ticker.C:
			bp.mu.Lock()
			if len(bp.queue) > 0 {
				bp.flush()
			}
			bp.mu.Unlock()
		case <-bp.done:
			bp.ticker.Stop()
			return
		}
	}
}

// flush processes the current queue
func (bp *BatchProcessor) flush() error {
	if len(bp.queue) == 0 {
		return nil
	}

	batch := make([]interface{}, len(bp.queue))
	copy(batch, bp.queue)
	bp.queue = bp.queue[:0]

	return bp.processFn(batch)
}

// Close stops the batch processor
func (bp *BatchProcessor) Close() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	close(bp.done)

	// Process remaining items
	if len(bp.queue) > 0 {
		return bp.flush()
	}

	return nil
}

// Memoizer caches function results
type Memoizer struct {
	cache *Cache
}

// NewMemoizer creates a new memoizer
func NewMemoizer(ttl time.Duration, maxSize int) *Memoizer {
	return &Memoizer{
		cache: NewCache(ttl, maxSize),
	}
}

// Memoize wraps a function with memoization
func (m *Memoizer) Memoize(key string, fn func() (interface{}, error)) (interface{}, error) {
	// Check cache first
	if cached, found := m.cache.Get(key); found {
		return cached, nil
	}

	// Compute result
	result, err := fn()
	if err != nil {
		return nil, err
	}

	// Cache result
	m.cache.Set(key, result)
	return result, nil
}
