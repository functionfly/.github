package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryArtifactStore(t *testing.T) {
	store := NewMemoryArtifactStore()
	ctx := context.Background()

	// Test Store and Retrieve
	data := []byte("test artifact data")
	err := store.Store(ctx, "test-key", data)
	require.NoError(t, err)

	retrieved, err := store.Retrieve(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)

	// Test Delete
	err = store.Delete(ctx, "test-key")
	require.NoError(t, err)

	_, err = store.Retrieve(ctx, "test-key")
	assert.Error(t, err)

	// Test non-existent key
	_, err = store.Retrieve(ctx, "non-existent")
	assert.Error(t, err)
}

func TestRedisArtifactStore(t *testing.T) {
	// Start a mini Redis server for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Test connection
	ctx := context.Background()
	err = rdb.Ping(ctx).Err()
	require.NoError(t, err)

	// Create Redis store
	store := NewRedisArtifactStoreFromClient(rdb, 1*time.Hour)

	// Test Store and Retrieve
	data := []byte("test Redis artifact data")
	err = store.Store(ctx, "redis-test-key", data)
	require.NoError(t, err)

	retrieved, err := store.Retrieve(ctx, "redis-test-key")
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)

	// Test Delete
	err = store.Delete(ctx, "redis-test-key")
	require.NoError(t, err)

	_, err = store.Retrieve(ctx, "redis-test-key")
	assert.Error(t, err)

	// Test non-existent key
	_, err = store.Retrieve(ctx, "non-existent-redis")
	assert.Error(t, err)

	// Test Ping
	err = store.Ping(ctx)
	assert.NoError(t, err)
}

func TestRedisArtifactStoreTTL(t *testing.T) {
	// Start a mini Redis server for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Enable key expiration in miniredis
	mr.SetTime(time.Now())

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create Redis store with short TTL
	store := NewRedisArtifactStoreFromClient(rdb, 100*time.Millisecond)

	ctx := context.Background()
	data := []byte("test TTL data")

	// Store data
	err = store.Store(ctx, "ttl-test-key", data)
	require.NoError(t, err)

	// Should be retrievable immediately
	retrieved, err := store.Retrieve(ctx, "ttl-test-key")
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)

	// Fast-forward time to expire the key
	mr.FastForward(200 * time.Millisecond)

	// Should no longer exist
	_, err = store.Retrieve(ctx, "ttl-test-key")
	assert.Error(t, err)
}