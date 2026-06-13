package vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// SecretCache caches the *metadata-only* projection of a secret.
// The zero-knowledge model means we never cache plaintext — but
// caching the metadata is still a big win because it lets the
// list endpoint skip the DB and the audit pipeline skip the
// ciphertext re-serialization on hot paths.
//
// Cache keys:
//
//	vault:meta:{tenant_id}:{secret_id}            -> SecretMeta JSON
//	vault:list:{tenant_id}:{namespace}:{limit}:{off} -> {id, ...} JSON
//	vault:token:{token_hash}                     -> AccessToken JSON
type SecretCache struct {
	redis  *redis.Client
	logger *logrus.Logger
	ttl    time.Duration
	// enabled flips off the cache entirely (useful for tests).
	enabled bool
}

// CacheConfig configures SecretCache.
type CacheConfig struct {
	// TTL is how long a metadata entry stays in the cache. Default 60s.
	TTL time.Duration
	// Logger is the structured logger; nil means logrus.New().
	Logger *logrus.Logger
}

// NewSecretCache constructs a SecretCache. If redis is nil, the
// cache is a no-op (each method silently returns cache-miss).
func NewSecretCache(redisClient *redis.Client, cfg CacheConfig) *SecretCache {
	if cfg.TTL == 0 {
		cfg.TTL = 60 * time.Second
	}
	logger := cfg.Logger
	if logger == nil {
		logger = logrus.New()
	}
	return &SecretCache{
		redis:   redisClient,
		logger:  logger,
		ttl:     cfg.TTL,
		enabled: redisClient != nil,
	}
}

// Enabled reports whether the cache is connected to Redis.
func (c *SecretCache) Enabled() bool { return c != nil && c.enabled }

// GetSecretMeta returns the cached metadata for a secret, or
// (nil, false, nil) on miss.
func (c *SecretCache) GetSecretMeta(ctx context.Context, tenantID, secretID string) (*Secret, bool, error) {
	if !c.Enabled() {
		return nil, false, nil
	}
	key := c.metaKey(tenantID, secretID)
	raw, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var s Secret
	if err := json.Unmarshal(raw, &s); err != nil {
		// Corrupt entry — treat as miss and let the caller refresh.
		return nil, false, nil
	}
	return &s, true, nil
}

// SetSecretMeta writes a secret's metadata to the cache.
func (c *SecretCache) SetSecretMeta(ctx context.Context, tenantID string, s *Secret) error {
	if !c.Enabled() || s == nil {
		return nil
	}
	key := c.metaKey(tenantID, s.ID.String())
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, data, c.ttl).Err()
}

// InvalidateSecret removes the metadata entry for a secret.
func (c *SecretCache) InvalidateSecret(ctx context.Context, tenantID, secretID string) error {
	if !c.Enabled() {
		return nil
	}
	return c.redis.Del(ctx, c.metaKey(tenantID, secretID)).Err()
}

// InvalidateTenant drops every cached metadata entry for a tenant.
// We do this with a SCAN, not KEYS, to keep Redis happy in
// production-sized keyspaces.
func (c *SecretCache) InvalidateTenant(ctx context.Context, tenantID string) error {
	if !c.Enabled() {
		return nil
	}
	pattern := fmt.Sprintf("vault:meta:%s:*", tenantID)
	var cursor uint64
	for {
		keys, next, err := c.redis.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.redis.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}

// GetToken returns the cached access token by its public ID.
func (c *SecretCache) GetToken(ctx context.Context, tokenID string) (*AccessToken, bool, error) {
	if !c.Enabled() {
		return nil, false, nil
	}
	raw, err := c.redis.Get(ctx, c.tokenKey(tokenID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var t AccessToken
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, false, nil
	}
	return &t, true, nil
}

// SetToken writes a token to the cache.
func (c *SecretCache) SetToken(ctx context.Context, t *AccessToken) error {
	if !c.Enabled() || t == nil {
		return nil
	}
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, c.tokenKey(t.ID.String()), data, c.ttl).Err()
}

// InvalidateToken removes a token entry.
func (c *SecretCache) InvalidateToken(ctx context.Context, tokenID string) error {
	if !c.Enabled() {
		return nil
	}
	return c.redis.Del(ctx, c.tokenKey(tokenID)).Err()
}

// Stats returns lightweight cache counters. Used by the /health
// surface and the operator dashboard.
func (c *SecretCache) Stats(ctx context.Context) CacheStats {
	stats := CacheStats{}
	if !c.Enabled() {
		return stats
	}
	// Estimate the cache size with a SCAN. We cap at 10k keys to
	// avoid scanning enormous keyspaces from a health check.
	const cap = 10000
	var cursor uint64
	patterns := []string{"vault:meta:*", "vault:token:*"}
	for _, p := range patterns {
		var count int64
		for {
			keys, next, err := c.redis.Scan(ctx, cursor, p, 200).Result()
			if err != nil {
				c.logger.WithError(err).Debug("cache stats scan failed")
				break
			}
			count += int64(len(keys))
			if next == 0 {
				break
			}
			cursor = next
			if count > cap {
				break
			}
		}
		if p == "vault:meta:*" {
			stats.MetaKeys = count
		} else {
			stats.TokenKeys = count
		}
	}
	return stats
}

// CacheStats is a snapshot of the cache contents.
type CacheStats struct {
	MetaKeys  int64 `json:"meta_keys"`
	TokenKeys int64 `json:"token_keys"`
}

func (c *SecretCache) metaKey(tenantID, secretID string) string {
	return "vault:meta:" + tenantID + ":" + secretID
}

func (c *SecretCache) tokenKey(tokenID string) string {
	return "vault:token:" + tokenID
}

// HashTokenSHA256 computes a SHA-256 hex digest. Exposed here so
// callers that want to key the cache on the token's hash (rather
// than its UUID) don't have to import crypto/sha256 directly.
func HashTokenSHA256(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ErrCacheMiss is returned by cache lookups that miss when the
// caller wants a sentinel. Currently unused (we use (nil, false, nil))
// but exported for future Go-cache style use.
var ErrCacheMiss = errors.New("vault: cache miss")
