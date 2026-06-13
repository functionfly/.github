package vault

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// newTestRedis returns a redis client. If a local Redis isn't
// available, the test is skipped.
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	c := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		t.Skipf("no redis at 127.0.0.1:6379: %v", err)
	}
	return c
}

func TestSecretCache_GetMiss(t *testing.T) {
	rc := newTestRedis(t)
	cache := NewSecretCache(rc, CacheConfig{TTL: 5 * time.Second})
	defer cache.InvalidateTenant(context.Background(), uuid.New().String())

	tenantID := uuid.New().String()
	secretID := uuid.New().String()
	got, hit, err := cache.GetSecretMeta(context.Background(), tenantID, secretID)
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("expected miss")
	}
	if got != nil {
		t.Fatal("expected nil on miss")
	}
}

func TestSecretCache_SetGetRoundTrip(t *testing.T) {
	rc := newTestRedis(t)
	cache := NewSecretCache(rc, CacheConfig{TTL: 5 * time.Second})
	tenantID := uuid.New().String()
	defer cache.InvalidateTenant(context.Background(), tenantID)

	s := &Secret{
		ID:         uuid.New(),
		Name:       "API_KEY",
		SecretType: SecretTypeAPIKey,
	}
	if err := cache.SetSecretMeta(context.Background(), tenantID, s); err != nil {
		t.Fatal(err)
	}
	got, hit, err := cache.GetSecretMeta(context.Background(), tenantID, s.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("expected hit after set")
	}
	if got.ID != s.ID {
		t.Fatalf("id mismatch: got %s want %s", got.ID, s.ID)
	}
	if got.Name != s.Name {
		t.Fatalf("name mismatch: got %q want %q", got.Name, s.Name)
	}
}

func TestSecretCache_InvalidateSecret(t *testing.T) {
	rc := newTestRedis(t)
	cache := NewSecretCache(rc, CacheConfig{TTL: 5 * time.Second})
	tenantID := uuid.New().String()
	defer cache.InvalidateTenant(context.Background(), tenantID)
	secretID := uuid.New().String()

	if err := cache.SetSecretMeta(context.Background(), tenantID, &Secret{ID: uuid.New(), Name: "X"}); err != nil {
		t.Fatal(err)
	}
	// We can't easily set a specific secret_id (Secret.ID is the key),
	// so we test the invalidate flow with a known key directly.
	if err := cache.redis.Set(context.Background(), cache.metaKey(tenantID, secretID), "raw", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if err := cache.InvalidateSecret(context.Background(), tenantID, secretID); err != nil {
		t.Fatal(err)
	}
	_, hit, _ := cache.GetSecretMeta(context.Background(), tenantID, secretID)
	if hit {
		t.Fatal("expected miss after invalidate")
	}
}

func TestSecretCache_InvalidateTenant(t *testing.T) {
	rc := newTestRedis(t)
	cache := NewSecretCache(rc, CacheConfig{TTL: 5 * time.Second})
	tenantID := uuid.New().String()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := cache.SetSecretMeta(ctx, tenantID, &Secret{ID: uuid.New(), Name: "X"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := cache.InvalidateTenant(ctx, tenantID); err != nil {
		t.Fatal(err)
	}
	// Subsequent scan should find nothing.
	keys, _, err := cache.redis.Scan(ctx, 0, "vault:meta:"+tenantID+":*", 100).Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys after invalidate, got %d", len(keys))
	}
}

func TestSecretCache_TokenRoundTrip(t *testing.T) {
	rc := newTestRedis(t)
	cache := NewSecretCache(rc, CacheConfig{TTL: 5 * time.Second})
	tok := &AccessToken{ID: uuid.New(), Name: "k8s-pod"}
	if err := cache.SetToken(context.Background(), tok); err != nil {
		t.Fatal(err)
	}
	defer cache.InvalidateToken(context.Background(), tok.ID.String())
	got, hit, err := cache.GetToken(context.Background(), tok.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	if !hit || got.ID != tok.ID {
		t.Fatalf("miss: hit=%v id=%v", hit, got)
	}
}

func TestSecretCache_DisabledWhenRedisNil(t *testing.T) {
	cache := NewSecretCache(nil, CacheConfig{})
	if cache.Enabled() {
		t.Fatal("expected disabled when redis is nil")
	}
	_, hit, _ := cache.GetSecretMeta(context.Background(), "t", "s")
	if hit {
		t.Fatal("expected miss when disabled")
	}
}

func TestSecretCache_StatsIsZeroWhenDisabled(t *testing.T) {
	cache := NewSecretCache(nil, CacheConfig{})
	stats := cache.Stats(context.Background())
	if stats.MetaKeys != 0 || stats.TokenKeys != 0 {
		t.Fatalf("expected zero stats, got %+v", stats)
	}
}

func TestHashTokenSHA256(t *testing.T) {
	// Deterministic output: SHA-256("test") is a well-known value.
	got := HashTokenSHA256("test")
	if len(got) != 64 {
		t.Fatalf("hex len=%d, want 64", len(got))
	}
	if HashTokenSHA256("test") != got {
		t.Fatal("not deterministic")
	}
}

// TestSecretCache_EncodedJSON verifies the wire format.
func TestSecretCache_EncodedJSON(t *testing.T) {
	rc := newTestRedis(t)
	cache := NewSecretCache(rc, CacheConfig{TTL: 5 * time.Second})
	tenantID := uuid.New().String()
	defer cache.InvalidateTenant(context.Background(), tenantID)

	s := &Secret{ID: uuid.New(), Name: "KEY", SecretType: SecretTypeAPIKey}
	if err := cache.SetSecretMeta(context.Background(), tenantID, s); err != nil {
		t.Fatal(err)
	}
	raw, err := cache.redis.Get(context.Background(), cache.metaKey(tenantID, s.ID.String())).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	var decoded Secret
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Name != "KEY" {
		t.Fatalf("name=%q", decoded.Name)
	}
}
