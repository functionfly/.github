package deployment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MemoryArtifactStore is an in-memory implementation of ArtifactStore for development/testing
type MemoryArtifactStore struct {
	mu    sync.RWMutex
	store map[string][]byte
}

// NewMemoryArtifactStore creates a new in-memory artifact store
func NewMemoryArtifactStore() *MemoryArtifactStore {
	return &MemoryArtifactStore{
		store: make(map[string][]byte),
	}
}

// Store stores an artifact with the given key
func (s *MemoryArtifactStore) Store(ctx context.Context, key string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = make([]byte, len(data))
	copy(s.store[key], data)
	return nil
}

// Retrieve retrieves an artifact by key
func (s *MemoryArtifactStore) Retrieve(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.store[key]
	if !ok {
		return nil, fmt.Errorf("artifact not found: %s", key)
	}
	// Return a copy to prevent external modifications
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// Delete removes an artifact by key
func (s *MemoryArtifactStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
	return nil
}

// RedisArtifactStore is a Redis-based implementation of ArtifactStore for production use
type RedisArtifactStore struct {
	client *redis.Client
	ttl    time.Duration // Optional TTL for stored artifacts
}

// NewRedisArtifactStore creates a new Redis-based artifact store
func NewRedisArtifactStore(addr, password string, db int, ttl time.Duration) *RedisArtifactStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisArtifactStore{
		client: client,
		ttl:    ttl,
	}
}

// NewRedisArtifactStoreFromClient creates a Redis artifact store using an existing Redis client
func NewRedisArtifactStoreFromClient(client *redis.Client, ttl time.Duration) *RedisArtifactStore {
	return &RedisArtifactStore{
		client: client,
		ttl:    ttl,
	}
}

// Store stores an artifact with the given key in Redis
func (s *RedisArtifactStore) Store(ctx context.Context, key string, data []byte) error {
	if s.ttl > 0 {
		return s.client.Set(ctx, key, data, s.ttl).Err()
	}
	return s.client.Set(ctx, key, data, 0).Err()
}

// Retrieve retrieves an artifact by key from Redis
func (s *RedisArtifactStore) Retrieve(ctx context.Context, key string) ([]byte, error) {
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("artifact not found: %s", key)
		}
		return nil, fmt.Errorf("failed to retrieve artifact: %w", err)
	}
	return []byte(result), nil
}

// Delete removes an artifact by key from Redis
func (s *RedisArtifactStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// Ping tests the Redis connection
func (s *RedisArtifactStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (s *RedisArtifactStore) Close() error {
	return s.client.Close()
}