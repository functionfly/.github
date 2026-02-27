package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// StateCache provides Redis-based caching for state values
type StateCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewStateCache creates a new state cache
func NewStateCache(client *redis.Client, defaultTTL time.Duration) *StateCache {
	if defaultTTL == 0 {
		defaultTTL = 5 * time.Minute // Default 5 minute TTL
	}
	return &StateCache{
		client: client,
		ttl:    defaultTTL,
	}
}

// cacheKey generates a cache key for a state value
func (c *StateCache) cacheKey(stateID uuid.UUID, key string) string {
	return fmt.Sprintf("state:%s:%s", stateID.String(), key)
}

// Get retrieves a value from cache
func (c *StateCache) Get(ctx context.Context, stateID uuid.UUID, key string) (map[string]interface{}, error) {
	cacheKey := c.cacheKey(stateID, key)

	data, err := c.client.Get(ctx, cacheKey).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var value map[string]interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return value, nil
}

// Set stores a value in cache
func (c *StateCache) Set(ctx context.Context, stateID uuid.UUID, key string, value map[string]interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}

	cacheKey := c.cacheKey(stateID, key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for cache: %w", err)
	}

	if err := c.client.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (c *StateCache) Delete(ctx context.Context, stateID uuid.UUID, key string) error {
	cacheKey := c.cacheKey(stateID, key)

	if err := c.client.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// InvalidateState invalidates all cached values for a state
func (c *StateCache) InvalidateState(ctx context.Context, stateID uuid.UUID) error {
	pattern := fmt.Sprintf("state:%s:*", stateID.String())

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	return nil
}

// GetAll retrieves all cached values for a state
func (c *StateCache) GetAll(ctx context.Context, stateID uuid.UUID) (map[string]map[string]interface{}, error) {
	pattern := fmt.Sprintf("state:%s:*", stateID.String())

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	result := make(map[string]map[string]interface{})
	for i, val := range values {
		if val == nil {
			continue
		}

		data, ok := val.(string)
		if !ok {
			continue
		}

		var value map[string]interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			logrus.Warnf("failed to unmarshal cached value for key %s: %v", keys[i], err)
			continue
		}

		// Extract key from cache key
		var key string
		fmt.Sscanf(keys[i], "state:"+stateID.String()+":%s", &key)
		result[key] = value
	}

	return result, nil
}

// StatePubSub provides Redis pub/sub for state change notifications
type StatePubSub struct {
	client *redis.Client
}

// NewStatePubSub creates a new state pub/sub handler
func NewStatePubSub(client *redis.Client) *StatePubSub {
	return &StatePubSub{
		client: client,
	}
}

// Publish publishes a state change event
func (p *StatePubSub) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := p.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Subscribe subscribes to state change events
func (p *StatePubSub) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return p.client.Subscribe(ctx, channels...)
}

// StateChangeEvent represents a state change notification
type StateChangeEvent struct {
	StateID    uuid.UUID `json:"state_id"`
	Key        string    `json:"key"`
	EventType  string    `json:"event_type"` // "set", "delete", "restore"
	Timestamp  time.Time `json:"timestamp"`
	SourceType string    `json:"source_type"`
	SourceID   string    `json:"source_id"`
}

// PublishStateChange publishes a state change event
func (p *StatePubSub) PublishStateChange(ctx context.Context, event *StateChangeEvent) error {
	channel := fmt.Sprintf("state_changes:%s", event.StateID.String())
	return p.Publish(ctx, channel, event)
}

// SubscribeToStateChanges subscribes to state changes for a specific state
func (p *StatePubSub) SubscribeToStateChanges(ctx context.Context, stateID uuid.UUID) *redis.PubSub {
	channel := fmt.Sprintf("state_changes:%s", stateID.String())
	return p.Subscribe(ctx, channel)
}

// SubscribeToAllStateChanges subscribes to all state changes for a tenant
func (p *StatePubSub) SubscribeToAllStateChanges(ctx context.Context, tenantID uuid.UUID) *redis.PubSub {
	channel := fmt.Sprintf("tenant:%s:state_changes", tenantID.String())
	return p.Subscribe(ctx, channel)
}

// TriggerEvent represents a trigger invocation event
type TriggerEvent struct {
	TriggerID     uuid.UUID   `json:"trigger_id"`
	StateID       uuid.UUID   `json:"state_id"`
	Key           string      `json:"key"`
	EventType     string      `json:"event_type"`
	PreviousValue interface{} `json:"previous_value,omitempty"`
	NewValue      interface{} `json:"new_value,omitempty"`
	Timestamp     time.Time   `json:"timestamp"`
}

// PublishTrigger publishes a trigger invocation event
func (p *StatePubSub) PublishTrigger(ctx context.Context, event *TriggerEvent) error {
	channel := "state_triggers"
	return p.Publish(ctx, channel, event)
}
