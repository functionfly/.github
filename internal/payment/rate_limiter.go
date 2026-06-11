package payment

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RateLimitConfig struct {
	MaxCalls  int
	WindowSec int
}

var RateLimitRules = map[string]RateLimitConfig{
	"POST:/v1/agent/*/credits/purchase":  {MaxCalls: 10, WindowSec: 60},
	"POST:/v1/agent/*/credits/checkout":  {MaxCalls: 10, WindowSec: 60},
	"POST:/admin/billing/credit":          {MaxCalls: 5, WindowSec: 60},
	"POST:/admin/billing/debit":           {MaxCalls: 5, WindowSec: 60},
	"POST:/v1/tenant/*/payment/create_payment_intent": {MaxCalls: 30, WindowSec: 60},
}

type SlidingWindowRateLimiter struct {
	redis *redis.Client
}

func NewSlidingWindowRateLimiter(redisClient *redis.Client) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		redis: redisClient,
	}
}

func (r *SlidingWindowRateLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if r.redis == nil {
		return true, nil
	}

	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	pipe := r.redis.Pipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	countCmd := pipe.ZCard(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		logrus.Warn("Rate limit check failed, allowing request", "key", key, "err", err)
		return true, nil
	}

	count := countCmd.Val()
	if count >= int64(limit) {
		return false, nil
	}

	r.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d:%d", now, count),
	})
	r.redis.Expire(ctx, key, window)

	return true, nil
}

func (r *SlidingWindowRateLimiter) CheckOperation(ctx context.Context, operation string) (bool, error) {
	config, ok := RateLimitRules[operation]
	if !ok {
		return true, nil
	}

	key := fmt.Sprintf("ratelimit:%s", operation)
	return r.Check(ctx, key, config.MaxCalls, time.Duration(config.WindowSec)*time.Second)
}

func (r *SlidingWindowRateLimiter) GetRemainingCalls(ctx context.Context, operation string) (int, error) {
	if r.redis == nil {
		return -1, nil
	}

	config, ok := RateLimitRules[operation]
	if !ok {
		return -1, nil
	}

	key := fmt.Sprintf("ratelimit:%s", operation)
	now := time.Now().Unix()
	windowStart := now - int64(config.WindowSec)

	count, err := r.redis.ZCount(ctx, key, strconv.FormatInt(windowStart, 10), strconv.FormatInt(now, 10)).Result()
	if err != nil {
		return 0, err
	}

	return config.MaxCalls - int(count), nil
}
