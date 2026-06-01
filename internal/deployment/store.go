package deployment

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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
	clusterClient  *redis.ClusterClient // Redis Cluster client (if using cluster mode)
	ttl            time.Duration        // Optional TTL for stored artifacts
	archiver       *R2ArtifactStore    // Optional R2 archiver for durable backup
	archiveAsync   bool                // Whether to archive asynchronously (non-blocking)
	archiveEnabled bool                // Whether archiving is enabled
	isCluster      bool                // Whether using Redis Cluster
}

// RedisStoreOption configures a RedisArtifactStore
type RedisStoreOption func(*RedisArtifactStore)

// WithR2Archiver configures the Redis store to archive artifacts to R2
func WithR2Archiver(archiver *R2ArtifactStore, async bool) RedisStoreOption {
	return func(s *RedisArtifactStore) {
		s.archiver = archiver
		s.archiveEnabled = true
		s.archiveAsync = async
	}
}

// NewRedisArtifactStore creates a new Redis-based artifact store
func NewRedisArtifactStore(addr, password string, db int, ttl time.Duration, opts ...RedisStoreOption) *RedisArtifactStore {
	// Check if Redis Cluster mode is enabled via environment variable
	if os.Getenv("REDIS_CLUSTER_ENABLED") == "true" {
		return newRedisClusterArtifactStore(ttl, opts...)
	}

	var tlsCfg *tls.Config
	if strings.Contains(addr, "upstash.io") {
		tlsCfg = &tls.Config{}
	}

	client := redis.NewClient(&redis.Options{
		Addr:      addr,
		Password:  password,
		DB:        db,
		TLSConfig: tlsCfg,
	})

	store := &RedisArtifactStore{
		client:         client,
		ttl:            ttl,
		archiveEnabled: false,
		archiveAsync:   true, // Default to async for performance
	}

	// Apply options
	for _, opt := range opts {
		opt(store)
	}

	// Auto-enable R2 archiver if environment variables are set and not already configured
	if !store.archiveEnabled && os.Getenv("R2_AUTO_ARCHIVE_ENABLED") == "true" {
		if archiver, err := NewR2ArtifactStore(); err == nil {
			store.archiver = archiver
			store.archiveEnabled = true
			store.archiveAsync = os.Getenv("R2_ARCHIVE_SYNC") != "true" // Async unless explicitly sync
			logrus.Info("R2 auto-archiving enabled for RedisArtifactStore")
		} else {
			logrus.Warnf("R2 auto-archive configured but failed to initialize: %v", err)
		}
	}

	return store
}

// newRedisClusterArtifactStore creates a Redis Cluster-based artifact store for horizontal scaling
func newRedisClusterArtifactStore(ttl time.Duration, opts ...RedisStoreOption) *RedisArtifactStore {
	clusterNodes := os.Getenv("REDIS_CLUSTER_NODES")
	if clusterNodes == "" {
		logrus.Warn("REDIS_CLUSTER_ENABLED=true but REDIS_CLUSTER_NODES is not set, falling back to single node")
		return newRedisArtifactStoreFromEnv(ttl, opts...)
	}

	var tlsCfg *tls.Config
	// Check if any node uses TLS
	if strings.Contains(clusterNodes, "upstash.io") {
		tlsCfg =&tls.Config{}
	}

	// Parse cluster nodes (format: node1:port,node2:port,node3:port)
	nodes := strings.Split(clusterNodes, ",")
	clusterOptions := &redis.ClusterOptions{
		Addrs:     nodes,
		TLSConfig: tlsCfg,
	}

	// Optional password for cluster auth
	if password := os.Getenv("REDIS_CLUSTER_PASSWORD"); password != "" {
		clusterOptions.Password = password
	}

	clusterClient := redis.NewClusterClient(clusterOptions)

	store :=&RedisArtifactStore{
		clusterClient:  clusterClient,
		ttl:            ttl,
		archiveEnabled: false,
		archiveAsync:   true,
		isCluster:      true,
	}

	// Apply options
	for _, opt := range opts {
		opt(store)
	}

	// Auto-enable R2 archiver if environment variables are set
	if !store.archiveEnabled && os.Getenv("R2_AUTO_ARCHIVE_ENABLED") == "true" {
		if archiver, err := NewR2ArtifactStore(); err == nil {
			store.archiver = archiver
			store.archiveEnabled = true
			store.archiveAsync = os.Getenv("R2_ARCHIVE_SYNC") != "true"
			logrus.Info("R2 auto-archiving enabled for RedisArtifactStore (Cluster mode)")
		} else {
			logrus.Warnf("R2 auto-archive configured but failed to initialize: %v", err)
		}
	}

	logrus.Infof("RedisArtifactStore initialized in Cluster mode with %d nodes", len(nodes))
	return store
}

// newRedisArtifactStoreFromEnv creates a single-node Redis store from environment variables (fallback)
func newRedisArtifactStoreFromEnv(ttl time.Duration, opts ...RedisStoreOption) *RedisArtifactStore {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		fmt.Sscanf(redisDB, "%d",&db)
	}

	return NewRedisArtifactStore(addr, password, db, ttl, opts...)
}

// NewRedisArtifactStoreFromClient creates a Redis artifact store using an existing Redis client
func NewRedisArtifactStoreFromClient(client *redis.Client, ttl time.Duration, opts ...RedisStoreOption) *RedisArtifactStore {
	store := &RedisArtifactStore{
		client:         client,
		ttl:            ttl,
		archiveEnabled: false,
		archiveAsync:   true,
	}

	// Apply options
	for _, opt := range opts {
		opt(store)
	}

	// Auto-enable R2 archiver if environment variables are set
	if !store.archiveEnabled && os.Getenv("R2_AUTO_ARCHIVE_ENABLED") == "true" {
		if archiver, err := NewR2ArtifactStore(); err == nil {
			store.archiver = archiver
			store.archiveEnabled = true
			store.archiveAsync = os.Getenv("R2_ARCHIVE_SYNC") != "true"
			logrus.Info("R2 auto-archiving enabled for RedisArtifactStore")
		} else {
			logrus.Warnf("R2 auto-archive configured but failed to initialize: %v", err)
		}
	}

	return store
}

// Store stores an artifact with the given key in Redis and optionally archives to R2
func (s *RedisArtifactStore) Store(ctx context.Context, key string, data []byte) error {
	// Store in Redis (primary, fast access)
	var err error
	if s.isCluster {
		err = s.clusterClient.Set(ctx, key, data, s.ttl).Err()
	} else if s.ttl > 0 {
		err = s.client.Set(ctx, key, data, s.ttl).Err()
	} else {
		err = s.client.Set(ctx, key, data, 0).Err()
	}
	if err != nil {
		return fmt.Errorf("failed to store artifact in Redis: %w", err)
	}

	// Archive to R2 if enabled (for durable, cross-region backup)
	if s.archiveEnabled && s.archiver != nil {
		if s.archiveAsync {
			// Non-blocking async archival
			go s.archiveToR2(ctx, key, data)
		} else {
			// Blocking sync archival
			if archiveErr := s.archiveToR2(ctx, key, data); archiveErr != nil {
				// Log but don't fail the Redis store on archive failure
				logrus.Warnf("Failed to archive artifact %s to R2: %v", key, archiveErr)
			}
		}
	}

	return nil
}

// archiveToR2 archives an artifact to R2 for durable backup
func (s *RedisArtifactStore) archiveToR2(ctx context.Context, key string, data []byte) error {
	// Use a background context with timeout to avoid blocking indefinitely
	archiveCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := s.archiver.Store(archiveCtx, key, data); err != nil {
		logrus.Warnf("R2 archive failed for %s: %v", key, err)
		return err
	}

	logrus.Debugf("R2 archive successful for %s (size: %d bytes)", key, len(data))
	return nil
}

// ArchiveToR2Sync performs synchronous archiving to R2 (useful for critical deployments)
func (s *RedisArtifactStore) ArchiveToR2Sync(ctx context.Context, key string, data []byte) error {
	if s.archiver == nil {
		return fmt.Errorf("R2 archiver not configured")
	}
	return s.archiver.Store(ctx, key, data)
}

// RetrieveFromR2 retrieves an artifact directly from R2 (bypassing Redis, useful for disaster recovery)
func (s *RedisArtifactStore) RetrieveFromR2(ctx context.Context, key string) ([]byte, error) {
	if s.archiver == nil {
		return nil, fmt.Errorf("R2 archiver not configured")
	}
	return s.archiver.Retrieve(ctx, key)
}

// GetR2Archiver returns the underlying R2 archiver for direct access
func (s *RedisArtifactStore) GetR2Archiver() *R2ArtifactStore {
	return s.archiver
}

// IsArchiveEnabled returns whether R2 archiving is enabled
func (s *RedisArtifactStore) IsArchiveEnabled() bool {
	return s.archiveEnabled
}

// Retrieve retrieves an artifact by key from Redis
func (s *RedisArtifactStore) Retrieve(ctx context.Context, key string) ([]byte, error) {
	var result string
	var err error

	if s.isCluster {
		result, err = s.clusterClient.Get(ctx, key).Result()
	} else {
		result, err = s.client.Get(ctx, key).Result()
	}

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
	var err error
	if s.isCluster {
		err = s.clusterClient.Del(ctx, key).Err()
	} else {
		err = s.client.Del(ctx, key).Err()
	}
	return err
}

// Ping tests the Redis connection
func (s *RedisArtifactStore) Ping(ctx context.Context) error {
	if s.isCluster {
		// For cluster, ping the first node
		return s.clusterClient.Ping(ctx).Err()
	}
	return s.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (s *RedisArtifactStore) Close() error {
	if s.isCluster {
		return s.clusterClient.Close()
	}
	return s.client.Close()
}
