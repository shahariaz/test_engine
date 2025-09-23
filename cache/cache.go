package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// CacheEntry represents a cache entry with expiration
type CacheEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
	HitCount  int         `json:"hit_count"`
}

// Cache provides thread-safe in-memory caching
type Cache struct {
	mu        sync.RWMutex
	entries   map[string]*CacheEntry
	ttl       time.Duration
	maxSize   int
	hitCount  int64
	missCount int64
}

// CacheStats provides cache statistics
type CacheStats struct {
	Entries   int     `json:"entries"`
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	HitRatio  float64 `json:"hit_ratio"`
	MaxSize   int     `json:"max_size"`
}

// NewCache creates a new cache instance
func NewCache(ttl time.Duration, maxSize int) *Cache {
	cache := &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
		HitCount:  0,
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.missCount++
		c.mu.Unlock()
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.missCount++
		c.mu.Unlock()
		return nil, false
	}

	// Update hit count
	c.mu.Lock()
	entry.HitCount++
	c.hitCount++
	c.mu.Unlock()

	return entry.Value, true
}

// GetOrSet retrieves a value or sets it if not found
func (c *Cache) GetOrSet(key string, provider func() interface{}) interface{} {
	if value, found := c.Get(key); found {
		return value
	}

	value := provider()
	c.Set(key, value)
	return value
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
	c.hitCount = 0
	c.missCount = 0
}

// GetStats returns cache statistics
func (c *Cache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hitCount + c.missCount
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(c.hitCount) / float64(total)
	}

	return CacheStats{
		Entries:   len(c.entries),
		HitCount:  c.hitCount,
		MissCount: c.missCount,
		HitRatio:  hitRatio,
		MaxSize:   c.maxSize,
	}
}

// GenerateKey creates a cache key from any object
func (c *Cache) GenerateKey(prefix string, obj interface{}) string {
	data, _ := json.Marshal(obj)
	hash := sha256.Sum256(data)
	return prefix + ":" + hex.EncodeToString(hash[:])
}

// evictLRU removes the least recently used entry
func (c *Cache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, entry := range c.entries {
		if entry.CreatedAt.Before(oldestTime) {
			oldestTime = entry.CreatedAt
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// cleanupExpired runs periodically to remove expired entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(time.Minute * 5) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// QueryCache provides specialized caching for DQL queries
type QueryCache struct {
	cache *Cache
}

// NewQueryCache creates a new query-specific cache
func NewQueryCache() *QueryCache {
	return &QueryCache{
		cache: NewCache(time.Minute*15, 1000), // 15 minute TTL, max 1000 entries
	}
}

// GetDQLQuery retrieves a cached DQL query
func (qc *QueryCache) GetDQLQuery(jsonQuery interface{}) (string, bool) {
	key := qc.cache.GenerateKey("dql", jsonQuery)
	if value, found := qc.cache.Get(key); found {
		return value.(string), true
	}
	return "", false
}

// SetDQLQuery caches a DQL query
func (qc *QueryCache) SetDQLQuery(jsonQuery interface{}, dqlQuery string) {
	key := qc.cache.GenerateKey("dql", jsonQuery)
	qc.cache.Set(key, dqlQuery)
}

// GetStats returns query cache statistics
func (qc *QueryCache) GetStats() CacheStats {
	return qc.cache.GetStats()
}

// Clear clears the query cache
func (qc *QueryCache) Clear() {
	qc.cache.Clear()
}
