package billing

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueueFailedCharge(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()
	payload := AIFailedBillingPayload{
		TenantID:     "tenant_123",
		ExecutionID:  "exec_456",
		FunctionID:   "func_789",
		Provider:     "openai",
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 50,
		Attempts:     0,
	}

	err = EnqueueFailedCharge(ctx, rdb, payload)
	require.NoError(t, err)

	queueLen, err := rdb.LLen(ctx, aiBillingRetryQueueKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), queueLen)

	delayedLen, err := rdb.ZCard(ctx, aiBillingDelayedQueueKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), delayedLen)

	data, err := rdb.LPop(ctx, aiBillingRetryQueueKey).Result()
	require.NoError(t, err)

	var result AIFailedBillingPayload
	err = json.Unmarshal([]byte(data), &result)
	require.NoError(t, err)
	assert.Equal(t, "tenant_123", result.TenantID)
	assert.Equal(t, "exec_456", result.ExecutionID)
	assert.Equal(t, "openai", result.Provider)
	assert.Equal(t, 0, result.Attempts)
}

func TestEnqueueFailedCharge_ZeroAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()
	payload := AIFailedBillingPayload{
		TenantID: "tenant_123",
		Provider: "anthropic",
		Model:    "claude-3",
		Attempts: 5,
	}

	err = EnqueueFailedCharge(ctx, rdb, payload)
	require.NoError(t, err)

	data, err := rdb.LPop(ctx, aiBillingRetryQueueKey).Result()
	require.NoError(t, err)

	var result AIFailedBillingPayload
	err = json.Unmarshal([]byte(data), &result)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Attempts)
}

func TestQueueLength(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	length, err := QueueLength(ctx, rdb)
	require.NoError(t, err)
	assert.Equal(t, int64(0), length)

	payload := AIFailedBillingPayload{
		TenantID: "tenant_123",
		Provider: "openai",
		Model:    "gpt-4",
	}
	err = EnqueueFailedCharge(ctx, rdb, payload)
	require.NoError(t, err)

	length, err = QueueLength(ctx, rdb)
	require.NoError(t, err)
	assert.Equal(t, int64(1), length)
}

func TestAIBillingRetryWorker_StartStop(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	worker := NewAIBillingRetryWorker(rdb, "http://localhost:8080", nil)

	ctx, cancel := context.WithCancel(context.Background())
	go worker.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	worker.Stop()
	cancel()

	time.Sleep(50 * time.Millisecond)
}

func TestAIBillingRetryWorker_ProcessRetryQueue_Empty(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	suspendCalled := false
	suspendFn := func(ctx context.Context, tenantID string, reason string) error {
		suspendCalled = true
		return nil
	}

	worker := &AIBillingRetryWorker{
		redis:       rdb,
		billingURL:  "http://localhost:8080",
		stop:        make(chan struct{}),
		suspendFunc: suspendFn,
	}

	ctx := context.Background()
	worker.processRetryQueue(ctx)

	assert.False(t, suspendCalled)
}

func TestAIBillingRetryWorker_ProcessRetryQueue_RetrySuccess(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	now := time.Now()
	mr.SetTime(now)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	payload := AIFailedBillingPayload{
		TenantID:     "tenant_123",
		ExecutionID:  "exec_456",
		Provider:     "openai",
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 50,
		Attempts:     0,
	}

	data, _ := json.Marshal(payload)

	pipe := rdb.Pipeline()
	pipe.LPush(ctx, aiBillingRetryQueueKey, data)
	pipe.ZAdd(ctx, aiBillingDelayedQueueKey, redis.Z{
		Score:  float64(now.Unix() - 10),
		Member: string(data),
	})
	pipe.Exec(ctx)

	worker := &AIBillingRetryWorker{
		redis:      rdb,
		billingURL: "http://localhost:8080",
		stop:       make(chan struct{}),
	}

	worker.processRetryQueue(ctx)

	queueLen, _ := rdb.LLen(ctx, aiBillingRetryQueueKey).Result()
	assert.Equal(t, int64(0), queueLen)
}

func TestAIBillingRetryWorker_MaxRetriesExceeded(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	now := time.Now()
	mr.SetTime(now)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	suspendCalled := false
	suspendFn := func(ctx context.Context, tenantID string, reason string) error {
		suspendCalled = true
		return nil
	}

	payload := AIFailedBillingPayload{
		TenantID:     "tenant_123",
		ExecutionID:  "exec_456",
		Provider:     "openai",
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 50,
		Attempts:     maxAIRetries,
	}

	data, _ := json.Marshal(payload)

	pipe := rdb.Pipeline()
	pipe.LPush(ctx, aiBillingRetryQueueKey, data)
	pipe.ZAdd(ctx, aiBillingDelayedQueueKey, redis.Z{
		Score:  float64(now.Unix() - 10),
		Member: string(data),
	})
	pipe.Exec(ctx)

	worker := &AIBillingRetryWorker{
		redis:       rdb,
		billingURL:  "http://localhost:8080",
		stop:        make(chan struct{}),
		suspendFunc: suspendFn,
	}

	worker.processRetryQueue(ctx)

	assert.True(t, suspendCalled)
}

func TestAIBillingRetryWorker_ExponentialBackoff(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	now := time.Now()
	mr.SetTime(now)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	attempts := []int{0, 1, 2, 3}
	expectedDelays := []int{30, 120, 480, 1920}

	for i, attempt := range attempts {
		payload := AIFailedBillingPayload{
			TenantID: "tenant_123",
			Provider: "openai",
			Model:    "gpt-4",
			Attempts: attempt,
		}

		data, _ := json.Marshal(payload)

		pipe := rdb.Pipeline()
		pipe.LPush(ctx, aiBillingRetryQueueKey, data)
		pipe.ZAdd(ctx, aiBillingDelayedQueueKey, redis.Z{
			Score:  float64(now.Unix() - 10),
			Member: string(data),
		})
		pipe.Exec(ctx)

		delays := []int{30, 120, 480, 1920}
		if i < len(delays) {
			assert.Equal(t, expectedDelays[i], delays[attempt])
		}
	}
}

func TestAIFailedBillingPayload_JSON(t *testing.T) {
	payload := AIFailedBillingPayload{
		TenantID:     "tenant_123",
		ExecutionID:  "exec_456",
		FunctionID:   "func_789",
		Provider:     "openai",
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 50,
		Attempts:     3,
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var result AIFailedBillingPayload
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, payload.TenantID, result.TenantID)
	assert.Equal(t, payload.ExecutionID, result.ExecutionID)
	assert.Equal(t, payload.FunctionID, result.FunctionID)
	assert.Equal(t, payload.Provider, result.Provider)
	assert.Equal(t, payload.Model, result.Model)
	assert.Equal(t, payload.InputTokens, result.InputTokens)
	assert.Equal(t, payload.OutputTokens, result.OutputTokens)
	assert.Equal(t, payload.Attempts, result.Attempts)
}

func TestNewAIBillingRetryWorker(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	suspendFn := func(ctx context.Context, tenantID string, reason string) error {
		return nil
	}

	worker := NewAIBillingRetryWorker(rdb, "http://localhost:8080", suspendFn)

	assert.NotNil(t, worker)
	assert.Equal(t, rdb, worker.redis)
	assert.Equal(t, "http://localhost:8080", worker.billingURL)
	assert.NotNil(t, worker.suspendFunc)
	assert.NotNil(t, worker.client)
	assert.Equal(t, 10*time.Second, worker.client.Timeout)
}

func TestAIBillingRetryWorker_AttemptCharge(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	worker := &AIBillingRetryWorker{
		redis:      rdb,
		billingURL: "http://localhost:8080",
		stop:       make(chan struct{}),
	}

	payload := AIFailedBillingPayload{
		TenantID:     "tenant_123",
		ExecutionID:  "exec_456",
		FunctionID:   "func_789",
		Provider:     "openai",
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 50,
		Attempts:     1,
	}

	ctx := context.Background()
	err = worker.attemptCharge(ctx, payload)
	assert.Error(t, err)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "ai_billing:failures", aiBillingRetryQueueKey)
	assert.Equal(t, "ai_billing:delayed", aiBillingDelayedQueueKey)
	assert.Equal(t, 4, maxAIRetries)
}
