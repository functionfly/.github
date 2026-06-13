package quota

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// stubRedis is a hand-rolled Redis stub for sliding-window tests.
// It implements the quota.Redis subset and tracks the bucket in
// an in-memory map.
type stubRedis struct {
	buckets map[string][]int64
}

func newStubRedis() *stubRedis {
	return &stubRedis{buckets: map[string][]int64{}}
}

func (s *stubRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringResult("", nil)
	return cmd
}
func (s *stubRedis) Set(ctx context.Context, key string, value interface{}, exp time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusResult("", nil)
	return cmd
}
func (s *stubRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return redis.NewIntResult(0, nil)
}
func (s *stubRedis) Incr(ctx context.Context, key string) *redis.IntCmd {
	return redis.NewIntResult(0, nil)
}
func (s *stubRedis) Expire(ctx context.Context, key string, exp time.Duration) *redis.BoolCmd {
	return redis.NewBoolResult(true, nil)
}
func (s *stubRedis) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	// The limiter's pipeline issues ZRemRangeByScore, ZCard, ZAdd,
	// Expire in that order. We need to return matching commands so
	// the limiter can read the ZCARD result at index 1.
	card := redis.NewIntCmd(ctx, "zcard")
	card.SetVal(int64(len(s.buckets)))
	return []redis.Cmder{
		redis.NewIntCmd(ctx, "zremrangebyscore"),
		card,
		redis.NewIntCmd(ctx, "zadd"),
		redis.NewBoolCmd(ctx, "expire"),
	}, nil
}

func TestSlidingWindowLimiter_DisabledWhenRedisNil(t *testing.T) {
	l := NewSlidingWindowLimiter(nil, "test", time.Minute, 5)
	d, err := l.Allow(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed {
		t.Fatal("nil redis must allow (fail-open)")
	}
}

func TestSlidingWindowLimiter_LimitIsSetOnDecision(t *testing.T) {
	l := NewSlidingWindowLimiter(newStubRedis(), "test", time.Minute, 5)
	d, _ := l.Allow(context.Background(), uuid.New())
	if d.Limit != 5 {
		t.Fatalf("limit=%d, want 5", d.Limit)
	}
	if d.Headers["X-RateLimit-Limit"] != "5" {
		t.Fatalf("X-RateLimit-Limit=%q", d.Headers["X-RateLimit-Limit"])
	}
}

func TestSlidingWindowLimiter_LoggerOverride(t *testing.T) {
	l := NewSlidingWindowLimiter(newStubRedis(), "test", time.Minute, 5)
	// Just exercise the setter; we don't assert on logger output.
	l.SetLogger(nil)
}
