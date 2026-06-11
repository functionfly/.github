package statefabric

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	fabricCacheTTL     = 5 * time.Minute
	metricsCacheTTL   = 1 * time.Minute
	listCacheTTL      = 2 * time.Minute
	keyPatternCacheTTL = 10 * time.Minute
)

type CacheConfig struct {
	Enabled bool
	TTL     time.Duration
}

type StateFabricCache struct {
	redis        *redis.Client
	enabled     bool
	fabricTTL    time.Duration
	metricsTTL   time.Duration
	listTTL      time.Duration
	keyPatternTTL time.Duration
}

func NewStateFabricCache(redisClient *redis.Client) *StateFabricCache {
	return &StateFabricCache{
		redis:        redisClient,
		enabled:      redisClient != nil,
		fabricTTL:    fabricCacheTTL,
		metricsTTL:   metricsCacheTTL,
		listTTL:      listCacheTTL,
		keyPatternTTL: keyPatternCacheTTL,
	}
}

func (c *StateFabricCache) IsEnabled() bool {
	return c.enabled && c.redis != nil
}

func (c *StateFabricCache) fabricKey(tenantID, fabricID uuid.UUID) string {
	return fmt.Sprintf("sf:fabric:%s:%s", tenantID.String(), fabricID.String())
}

func (c *StateFabricCache) fabricListKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("sf:fabric:list:%s", tenantID.String())
}

func (c *StateFabricCache) metricsKey(fabricID uuid.UUID) string {
	return fmt.Sprintf("sf:metrics:%s", fabricID.String())
}

func (c *StateFabricCache) keyPatternKey(fabricID uuid.UUID, pattern string) string {
	return fmt.Sprintf("sf:keys:%s:%s", fabricID.String(), pattern)
}

// GetFabric retrieves a fabric from cache
func (c *StateFabricCache) GetFabric(ctx context.Context, tenantID, fabricID uuid.UUID) (*Fabric, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	key := c.fabricKey(tenantID, fabricID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var fabric Fabric
	if err := json.Unmarshal(data, &fabric); err != nil {
		return nil, err
	}

	return &fabric, nil
}

// SetFabric stores a fabric in cache
func (c *StateFabricCache) SetFabric(ctx context.Context, tenantID, fabricID uuid.UUID, fabric *Fabric) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.fabricKey(tenantID, fabricID)
	data, err := json.Marshal(fabric)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.fabricTTL).Err()
}

// InvalidateFabric removes a fabric from cache
func (c *StateFabricCache) InvalidateFabric(ctx context.Context, tenantID, fabricID uuid.UUID) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.fabricKey(tenantID, fabricID)
	return c.redis.Del(ctx, key).Err()
}

// InvalidateFabricList removes the fabric list from cache
func (c *StateFabricCache) InvalidateFabricList(ctx context.Context, tenantID uuid.UUID) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.fabricListKey(tenantID)
	return c.redis.Del(ctx, key).Err()
}

// GetFabricList retrieves a fabric list from cache
func (c *StateFabricCache) GetFabricList(ctx context.Context, tenantID uuid.UUID) ([]Fabric, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	key := c.fabricListKey(tenantID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var fabrics []Fabric
	if err := json.Unmarshal(data, &fabrics); err != nil {
		return nil, err
	}

	return fabrics, nil
}

// SetFabricList stores a fabric list in cache
func (c *StateFabricCache) SetFabricList(ctx context.Context, tenantID uuid.UUID, fabrics []Fabric) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.fabricListKey(tenantID)
	data, err := json.Marshal(fabrics)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.listTTL).Err()
}

// GetMetrics retrieves metrics from cache
func (c *StateFabricCache) GetMetrics(ctx context.Context, fabricID uuid.UUID) (*FabricMetrics, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	key := c.metricsKey(fabricID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var metrics FabricMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// SetMetrics stores metrics in cache
func (c *StateFabricCache) SetMetrics(ctx context.Context, fabricID uuid.UUID, metrics *FabricMetrics) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.metricsKey(fabricID)
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.metricsTTL).Err()
}

// InvalidateMetrics removes metrics from cache
func (c *StateFabricCache) InvalidateMetrics(ctx context.Context, fabricID uuid.UUID) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.metricsKey(fabricID)
	return c.redis.Del(ctx, key).Err()
}

// GetKeyPattern retrieves key pattern from cache
func (c *StateFabricCache) GetKeyPattern(ctx context.Context, fabricID uuid.UUID, pattern string) ([]FabricKeyEntry, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	key := c.keyPatternKey(fabricID, pattern)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var keys []FabricKeyEntry
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// SetKeyPattern stores key pattern in cache
func (c *StateFabricCache) SetKeyPattern(ctx context.Context, fabricID uuid.UUID, pattern string, keys []FabricKeyEntry) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.keyPatternKey(fabricID, pattern)
	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.keyPatternTTL).Err()
}

// InvalidateKeyPattern removes key pattern from cache
func (c *StateFabricCache) InvalidateKeyPattern(ctx context.Context, fabricID uuid.UUID, pattern string) error {
	if !c.IsEnabled() {
		return nil
	}

	key := c.keyPatternKey(fabricID, pattern)
	return c.redis.Del(ctx, key).Err()
}

// InvalidateAllFabricData removes all fabric-related data from cache
func (c *StateFabricCache) InvalidateAllFabricData(ctx context.Context, tenantID, fabricID uuid.UUID) error {
	if !c.IsEnabled() {
		return nil
	}

	keys := []string{
		c.fabricKey(tenantID, fabricID),
		c.metricsKey(fabricID),
	}

	pattern := fmt.Sprintf("sf:keys:%s:*", fabricID.String())
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		logrus.WithError(err).Warn("failed to scan key pattern keys during invalidation")
	}

	if len(keys) > 0 {
		return c.redis.Del(ctx, keys...).Err()
	}

	return nil
}

// RecordCacheHit records a cache hit metric
func (c *StateFabricCache) RecordCacheHit(tenantID, fabricID, cacheType string) {
	if !c.IsEnabled() {
		return
	}
}

// RecordCacheMiss records a cache miss metric
func (c *StateFabricCache) RecordCacheMiss(tenantID, fabricID, cacheType string) {
	if !c.IsEnabled() {
		return
	}
}