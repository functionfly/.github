package team_memory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// ============================================
// Team Context Cache for Repeated Agent Calls
// ============================================

// ContextCacheEntry represents a cached team context
type ContextCacheEntry struct {
	Key            string                `json:"key"`
	TenantID       uuid.UUID             `json:"tenant_id"`
	TeamID         uuid.UUID             `json:"team_id"`
	ContextString  string                `json:"context_string"`
	Sources        []*storage.TeamMemory `json:"sources"`
	QueryHash      string                `json:"query_hash"`   // Hash of the query for cache key
	MemoryTypes    []string              `json:"memory_types"` // Cached for these types
	Categories     []string              `json:"categories"`   // Cached for these categories
	CreatedAt      time.Time             `json:"created_at"`
	ExpiresAt      time.Time             `json:"expires_at"`
	AccessCount    int                   `json:"access_count"`
	LastAccessedAt time.Time             `json:"last_accessed_at"`
}

// IsExpired checks if the cache entry has expired
func (e *ContextCacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// ContextCache provides caching for team memory context
type ContextCache struct {
	entries   map[string]*ContextCacheEntry
	mu        sync.RWMutex
	ttl       time.Duration
	maxSize   int
	hitCount  int
	missCount int
}

// NewContextCache creates a new context cache
func NewContextCache(ttl time.Duration, maxSize int) *ContextCache {
	if ttl == 0 {
		ttl = 5 * time.Minute // Default: 5 minute TTL
	}
	if maxSize == 0 {
		maxSize = 1000 // Default: max 1000 entries
	}

	return &ContextCache{
		entries: make(map[string]*ContextCacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// CacheKey generates a cache key from request parameters
func (c *ContextCache) CacheKey(tenantID, teamID uuid.UUID, query string, memoryTypes, categories []string) string {
	// Create a consistent hash from the parameters
	h := sha256.New()
	h.Write([]byte(tenantID.String()))
	h.Write([]byte(teamID.String()))
	h.Write([]byte(query))
	for _, mt := range memoryTypes {
		h.Write([]byte(mt))
	}
	for _, cat := range categories {
		h.Write([]byte(cat))
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:32] // First 32 chars of hash
}

// Get retrieves a cached context
func (c *ContextCache) Get(key string) (*ContextCacheEntry, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		c.mu.Lock()
		c.missCount++
		c.mu.Unlock()
		return nil, false
	}

	if entry.IsExpired() {
		c.mu.Lock()
		delete(c.entries, key)
		c.missCount++
		c.mu.Unlock()
		return nil, false
	}

	// Update access stats
	c.mu.Lock()
	entry.AccessCount++
	entry.LastAccessedAt = time.Now()
	c.hitCount++
	c.mu.Unlock()

	return entry, true
}

// Set stores a context in the cache
func (c *ContextCache) Set(key string, tenantID, teamID uuid.UUID, contextString string, sources []*storage.TeamMemory, queryHash string, memoryTypes, categories []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity (simple LRU eviction)
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	entry := &ContextCacheEntry{
		Key:           key,
		TenantID:      tenantID,
		TeamID:        teamID,
		ContextString: contextString,
		Sources:       sources,
		QueryHash:     queryHash,
		MemoryTypes:   memoryTypes,
		Categories:    categories,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(c.ttl),
		AccessCount:   1,
	}

	c.entries[key] = entry
}

// evictOldest removes the oldest entry from the cache
// Must be called with c.mu held (write lock)
func (c *ContextCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestTime.IsZero() || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// Invalidate removes a specific entry from the cache
func (c *ContextCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// InvalidateTeam clears all cache entries for a team
func (c *ContextCache) InvalidateTeam(teamID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, entry := range c.entries {
		if entry.TeamID == teamID {
			delete(c.entries, key)
		}
	}
}

// InvalidateTenant clears all cache entries for a tenant
func (c *ContextCache) InvalidateTenant(tenantID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, entry := range c.entries {
		if entry.TenantID == tenantID {
			delete(c.entries, key)
		}
	}
}

// Clear removes all entries from the cache
func (c *ContextCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*ContextCacheEntry)
}

// Stats returns cache statistics
func (c *ContextCache) Stats() (hits, misses, size int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hitCount, c.missCount, len(c.entries)
}

// ============================================
// Cached Context Provider
// ============================================

// CachedContextProvider wraps a context provider with caching
type CachedContextProvider struct {
	inner *AgentContextProvider
	cache *ContextCache
	repo  storage.Repository
}

// NewCachedContextProvider creates a new cached context provider
func NewCachedContextProvider(inner *AgentContextProvider, cache *ContextCache, repo storage.Repository) *CachedContextProvider {
	return &CachedContextProvider{
		inner: inner,
		cache: cache,
		repo:  repo,
	}
}

// BuildContext builds context with caching
func (c *CachedContextProvider) BuildContext(ctx context.Context, req ContextRequest) (*ContextResponse, error) {
	start := time.Now()
	teamIDStr := req.TeamID.String()

	// Generate cache key
	cacheKey := c.cache.CacheKey(req.TenantID, req.TeamID, req.CurrentTask, req.MemoryTypes, req.RelevantCategories)

	// Try to get from cache
	if cached, ok := c.cache.Get(cacheKey); ok {
		monitoring.RecordTeamMemoryCacheHit(teamIDStr)
		monitoring.RecordTeamMemoryContextInjectionDuration(teamIDStr, time.Since(start))

		return &ContextResponse{
			Context: cached.ContextString,
			Sources: cached.Sources,
		}, nil
	}

	monitoring.RecordTeamMemoryCacheMiss(teamIDStr)

	// Build context from inner provider
	resp, err := c.inner.BuildContext(ctx, req)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, req.TenantID, req.TeamID, resp.Context, resp.Sources, "", req.MemoryTypes, req.RelevantCategories)

	// Record metric
	monitoring.RecordTeamMemoryContextInjectionDuration(teamIDStr, time.Since(start))

	return resp, nil
}

// InvalidateCache invalidates the cache for a specific team
func (c *CachedContextProvider) InvalidateCache(teamID uuid.UUID) {
	c.cache.InvalidateTeam(teamID)
}

// InvalidateMemory invalidates cache entries that include a specific memory.
// Iterates through all cache entries to find those referencing the memory.
// O(n*m) where n=entries, m=sources per entry. Could be optimized with reverse index.
func (c *CachedContextProvider) InvalidateMemory(memoryID uuid.UUID) {
	// Collect keys while holding read lock
	c.cache.mu.RLock()
	keysToInvalidate := make([]string, 0)
	for key, entry := range c.cache.entries {
		for _, source := range entry.Sources {
			if source.ID == memoryID {
				keysToInvalidate = append(keysToInvalidate, key)
				break
			}
		}
	}
	c.cache.mu.RUnlock()

	// Invalidate outside the lock to avoid deadlocks
	for _, key := range keysToInvalidate {
		c.cache.Invalidate(key)
	}
}

// CacheStats returns cache statistics
func (c *CachedContextProvider) CacheStats() (hits, misses, size int) {
	return c.cache.Stats()
}

// ============================================
// Redis-backed distributed cache
// ============================================

// DistributedContextCache provides Redis-backed caching for multi-instance deployments
type DistributedContextCache struct {
	localCache    *ContextCache
	redis         *redis.Client
	ttl           time.Duration
	keyPrefix     string
	pubSubEnabled bool
}

// NewDistributedContextCache creates a new distributed context cache with Redis backing
func NewDistributedContextCache(localCache *ContextCache, redisClient *redis.Client, ttl time.Duration) *DistributedContextCache {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &DistributedContextCache{
		localCache:    localCache,
		redis:         redisClient,
		ttl:           ttl,
		keyPrefix:     "team_memory:context:",
		pubSubEnabled: redisClient != nil,
	}
}

// cacheKey generates a Redis key for a cache entry
func (d *DistributedContextCache) cacheKey(key string) string {
	return d.keyPrefix + key
}

// pubSubChannel returns the pub/sub channel for invalidation messages
func (d *DistributedContextCache) pubSubChannel(teamID uuid.UUID) string {
	return fmt.Sprintf("team_memory:invalidate:%s", teamID.String())
}

// Get retrieves from local cache with Redis fallback
func (d *DistributedContextCache) Get(ctx context.Context, key string) (*ContextCacheEntry, bool) {
	// Try local cache first (fastest)
	if entry, ok := d.localCache.Get(key); ok {
		return entry, true
	}

	// Fall back to Redis if available
	if d.redis == nil {
		return nil, false
	}

	redisKey := d.cacheKey(key)
	data, err := d.redis.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		logrus.WithError(err).Warn("Failed to get team memory context from Redis")
		return nil, false
	}

	var entry ContextCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		logrus.WithError(err).Warn("Failed to unmarshal team memory context from Redis")
		return nil, false
	}

	// Check if expired (Redis TTL should handle this, but double-check)
	if entry.IsExpired() {
		_ = d.redis.Del(ctx, redisKey).Err()
		return nil, false
	}

	// Populate local cache for next time
	d.localCache.Set(key, entry.TenantID, entry.TeamID, entry.ContextString, entry.Sources, entry.QueryHash, entry.MemoryTypes, entry.Categories)

	return &entry, true
}

// Set stores in both local and Redis cache
func (d *DistributedContextCache) Set(ctx context.Context, key string, tenantID, teamID uuid.UUID, contextString string, sources []*storage.TeamMemory, queryHash string, memoryTypes, categories []string) {
	// Always update local cache
	d.localCache.Set(key, tenantID, teamID, contextString, sources, queryHash, memoryTypes, categories)

	// Also store in Redis if available
	if d.redis == nil {
		return
	}

	entry := &ContextCacheEntry{
		Key:           key,
		TenantID:      tenantID,
		TeamID:        teamID,
		ContextString: contextString,
		Sources:       sources,
		QueryHash:     queryHash,
		MemoryTypes:   memoryTypes,
		Categories:    categories,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(d.ttl),
		AccessCount:   1,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		logrus.WithError(err).Warn("Failed to marshal team memory context for Redis")
		return
	}

	redisKey := d.cacheKey(key)
	if err := d.redis.Set(ctx, redisKey, data, d.ttl).Err(); err != nil {
		logrus.WithError(err).Warn("Failed to store team memory context in Redis")
	}
}

// Invalidate removes a specific entry from both caches
func (d *DistributedContextCache) Invalidate(ctx context.Context, key string) {
	d.localCache.Invalidate(key)

	if d.redis != nil {
		redisKey := d.cacheKey(key)
		if err := d.redis.Del(ctx, redisKey).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to invalidate team memory context in Redis")
		}
	}
}

// InvalidateTeam clears cache for a team across all instances via pub/sub
func (d *DistributedContextCache) InvalidateTeam(ctx context.Context, teamID uuid.UUID) {
	d.localCache.InvalidateTeam(teamID)

	if d.redis == nil {
		return
	}

	// Delete all keys matching team pattern
	pattern := d.keyPrefix + "*"
	iter := d.redis.Scan(ctx, 0, pattern, 0).Iterator()
	var keysToDelete []string
	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}

	if err := iter.Err(); err != nil {
		logrus.WithError(err).Warn("Failed to scan team memory keys for invalidation")
	}

	if len(keysToDelete) > 0 {
		if err := d.redis.Del(ctx, keysToDelete...).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to delete team memory keys from Redis")
		}
	}

	// Publish invalidation message for other instances
	if d.pubSubEnabled {
		channel := d.pubSubChannel(teamID)
		message := fmt.Sprintf(`{"team_id":"%s","timestamp":"%s"}`, teamID.String(), time.Now().Format(time.RFC3339))
		if err := d.redis.Publish(ctx, channel, message).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to publish team memory invalidation")
		}
	}
}

// InvalidateTenant clears all cache entries for a tenant
func (d *DistributedContextCache) InvalidateTenant(ctx context.Context, tenantID uuid.UUID) {
	d.localCache.InvalidateTenant(tenantID)

	if d.redis == nil {
		return
	}

	// Scan and delete all keys (inefficient but correct for tenant isolation)
	pattern := d.keyPrefix + "*"
	iter := d.redis.Scan(ctx, 0, pattern, 0).Iterator()
	var keysToDelete []string
	for iter.Next(ctx) {
		key := iter.Val()
		// Get entry and check tenant
		data, err := d.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var entry ContextCacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}
		if entry.TenantID == tenantID {
			keysToDelete = append(keysToDelete, key)
		}
	}

	if len(keysToDelete) > 0 {
		if err := d.redis.Del(ctx, keysToDelete...).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to delete tenant memory keys from Redis")
		}
	}
}

// SubscribeToInvalidations starts listening for invalidation messages from other instances
func (d *DistributedContextCache) SubscribeToInvalidations(ctx context.Context, handler func(teamID uuid.UUID)) error {
	if d.redis == nil || !d.pubSubEnabled {
		return nil
	}

	// Subscribe to all invalidation channels
	pubsub := d.redis.PSubscribe(ctx, "team_memory:invalidate:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		// Extract team ID from channel or message
		teamIDStr := msg.Payload
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			d.localCache.InvalidateTeam(teamID)
			if handler != nil {
				handler(teamID)
			}
		}
	}

	return nil
}

// Stats returns combined cache statistics
func (d *DistributedContextCache) Stats(ctx context.Context) (localHits, localMisses, localSize int, redisKeys int64, err error) {
	localHits, localMisses, localSize = d.localCache.Stats()

	if d.redis == nil {
		return localHits, localMisses, localSize, 0, nil
	}

	// Count Redis keys
	iter := d.redis.Scan(ctx, 0, d.keyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		redisKeys++
	}

	return localHits, localMisses, localSize, redisKeys, iter.Err()
}
