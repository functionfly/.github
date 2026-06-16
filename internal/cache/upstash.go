package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/claywarren/upstash-go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// UpstashRedisClient is a wrapper around upstash-go that provides
// a compatible interface with go-redis/v9 for use throughout the codebase.
type UpstashRedisClient struct {
	client      *upstash.Upstash
	stdRedis    *redis.Client // Standard go-redis client (used when Upstash isn't configured)
	isNil       bool
	useStdRedis bool // Flag to indicate we're using standard Redis
}

// UpstashConfig holds configuration for Upstash Redis connection
type UpstashConfig struct {
	URL       string // Upstash REST API URL
	Token     string // Upstash REST API Token
	DB        int    // Database number (kept for compatibility)
	IsUpstash bool
}

// NewUpstashConfig creates Upstash config from environment variables
func NewUpstashConfig() *UpstashConfig {
	config := &UpstashConfig{}

	// Check if using Upstash
	upstashURL := os.Getenv("UPSTASH_REDIS_REST_URL")
	upstashToken := os.Getenv("UPSTASH_REDIS_REST_TOKEN")

	if upstashURL != "" && upstashToken != "" {
		config.URL = upstashURL
		config.Token = upstashToken
		config.IsUpstash = true
		logrus.Info("Using Upstash Redis")
	} else {
		// Fall back to standard Redis
		config.IsUpstash = false
		logrus.Info("Using standard Redis (set UPSTASH_REDIS_REST_URL and UPSTASH_REDIS_REST_TOKEN for Upstash)")
	}

	// DB number (for compatibility, not used by Upstash)
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			config.DB = db
		} else {
			logrus.WithField("REDIS_DB", dbStr).Warn("Invalid REDIS_DB value, using 0")
		}
	}

	return config
}

// NewUpstashRedisClient creates a new Upstash Redis client
func NewUpstashRedisClient(config *UpstashConfig) *UpstashRedisClient {
	if !config.IsUpstash {
		return &UpstashRedisClient{isNil: true}
	}

	client, err := upstash.New(upstash.Options{
		Url:   config.URL,
		Token: config.Token,
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to create Upstash Redis client")
		return &UpstashRedisClient{isNil: true}
	}

	return &UpstashRedisClient{
		client: &client,
		isNil:  false,
	}
}

// NewUpstashRedisClientFromEnv creates a new Upstash Redis client from environment variables
func NewUpstashRedisClientFromEnv() *UpstashRedisClient {
	config := NewUpstashConfig()
	return NewUpstashRedisClient(config)
}

// NewUpstashRedisClientFromStandardRedis creates an UpstashRedisClient wrapper
// around a standard go-redis client. This is used for local development when
// Upstash is not configured but we still need the Upstash-compatible interface.
func NewUpstashRedisClientFromStandardRedis(rdb *redis.Client) *UpstashRedisClient {
	if rdb == nil {
		return &UpstashRedisClient{isNil: true}
	}
	return &UpstashRedisClient{
		stdRedis:    rdb,
		isNil:       false,
		useStdRedis: true,
	}
}

// isInitialized checks if the client is properly initialized
func (c *UpstashRedisClient) isInitialized() bool {
	if c == nil || c.isNil {
		return false
	}
	return c.client != nil || c.useStdRedis
}

// Get retrieves a value by key
func (c *UpstashRedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	if !c.isInitialized() {
		return nil, redis.Nil
	}

	// Use standard Redis client if configured
	if c.useStdRedis && c.stdRedis != nil {
		val, err := c.stdRedis.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				return nil, redis.Nil
			}
			return nil, err
		}
		return []byte(val), nil
	}

	val, err := c.client.Get(ctx, key)
	if err != nil {
		if err.Error() == "key not found" || err.Error() == "redis: nil" {
			return nil, redis.Nil
		}
		return nil, err
	}
	return []byte(val), nil
}

// Set stores a value with optional TTL
func (c *UpstashRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !c.isInitialized() {
		return fmt.Errorf("upstash client not initialized")
	}

	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data = string(jsonData)
	}

	// Use standard Redis client if configured
	if c.useStdRedis && c.stdRedis != nil {
		return c.stdRedis.Set(ctx, key, data, expiration).Err()
	}

	if expiration > 0 {
		return c.client.SetEX(ctx, key, int(expiration.Seconds()), data)
	}
	return c.client.Set(ctx, key, data)
}

// SetJSON marshals the value as JSON and stores it
func (c *UpstashRedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, expiration)
}

// Del deletes keys
func (c *UpstashRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	// Use standard Redis client if configured
	if c.useStdRedis && c.stdRedis != nil {
		return c.stdRedis.Del(ctx, keys...).Result()
	}

	count, err := c.client.Del(ctx, keys...)
	return int64(count), err
}

// ZAdd adds member(s) to a sorted set
func (c *UpstashRedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	for _, m := range members {
		memberStr, ok := m.Member.(string)
		if !ok {
			continue
		}
		_, err := c.client.ZAdd(ctx, key, m.Score, memberStr)
		if err != nil {
			return 0, err
		}
	}
	return int64(len(members)), nil
}

// ZCard returns the cardinality of a sorted set
func (c *UpstashRedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	count, err := c.client.ZCard(ctx, key)
	return int64(count), err
}

// ZRange returns a range of members in a sorted set
func (c *UpstashRedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if !c.isInitialized() {
		return nil, nil
	}

	return c.client.ZRange(ctx, key, int(start), int(stop))
}

// ZRemRangeByScore removes members by score range
func (c *UpstashRedisClient) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	count, err := c.client.ZRemRangeByScore(ctx, key, min, max)
	return int64(count), err
}

// Expire sets expiration on a key
func (c *UpstashRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if !c.isInitialized() {
		return false, nil
	}

	// Use standard Redis client if configured
	if c.useStdRedis && c.stdRedis != nil {
		return c.stdRedis.Expire(ctx, key, expiration).Result()
	}

	result, err := c.client.Expire(ctx, key, int(expiration.Seconds()))
	return result > 0, err
}

// TTL returns the TTL of a key
func (c *UpstashRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	if !c.isInitialized() {
		return -1, nil
	}

	ttl, err := c.client.Ttl(ctx, key)
	if err != nil {
		return -1, err
	}
	if ttl < 0 {
		return -1, nil
	}
	return time.Duration(ttl) * time.Second, nil
}

// Keys returns all keys matching a pattern
func (c *UpstashRedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	if !c.isInitialized() {
		return nil, nil
	}

	return c.client.Keys(ctx, pattern)
}

// Ping checks if the client is connected
func (c *UpstashRedisClient) Ping(ctx context.Context) error {
	if !c.isInitialized() {
		return fmt.Errorf("upstash client not initialized")
	}

	_, err := c.client.Ping(ctx)
	return err
}

// Close closes the client connection
func (c *UpstashRedisClient) Close() error {
	// Upstash HTTP client doesn't need explicit close
	return nil
}

// Publish publishes a message to a channel
func (c *UpstashRedisClient) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	var data string
	switch v := message.(type) {
	case string:
		data = v
	default:
		jsonData, _ := json.Marshal(v)
		data = string(jsonData)
	}

	count, err := c.client.Publish(ctx, channel, data)
	return int64(count), err
}

// Subscribe subscribes to a channel
func (c *UpstashRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	// Upstash SSE subscription requires different handling
	// Return nil as this is a simplified implementation
	return nil
}

// MGet returns values for multiple keys
func (c *UpstashRedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if !c.isInitialized() {
		return nil, nil
	}

	vals, err := c.client.MGet(ctx, keys)
	if err != nil {
		return nil, err
	}

	results := make([]interface{}, len(vals))
	for i, v := range vals {
		if v == "" {
			results[i] = nil
		} else {
			results[i] = v
		}
	}
	return results, nil
}

// ZScore returns the score of a member
func (c *UpstashRedisClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	if !c.isInitialized() {
		return 0, redis.Nil
	}

	return c.client.ZScore(ctx, key, member)
}

// GetBytes retrieves a value as bytes
func (c *UpstashRedisClient) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return c.Get(ctx, key)
}

// GetString retrieves a value as string
func (c *UpstashRedisClient) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// SetWithTTL is a convenience method for Set with expiration
func (c *UpstashRedisClient) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}

// Delete removes keys
func (c *UpstashRedisClient) Delete(ctx context.Context, keys ...string) error {
	_, err := c.Del(ctx, keys...)
	return err
}

// Increment increments a counter
func (c *UpstashRedisClient) Increment(ctx context.Context, key string) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	val, err := c.client.Incr(ctx, key)
	return int64(val), err
}

// IncrementBy increments by a specific amount
func (c *UpstashRedisClient) IncrementBy(ctx context.Context, key string, increment int64) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	val, err := c.client.IncrBy(ctx, key, int(increment))
	return int64(val), err
}

// Decrement decrements a counter
func (c *UpstashRedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	if !c.isInitialized() {
		return 0, nil
	}

	val, err := c.client.Decr(ctx, key)
	return int64(val), err
}

// GetJSON retrieves and unmarshals JSON data
func (c *UpstashRedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
