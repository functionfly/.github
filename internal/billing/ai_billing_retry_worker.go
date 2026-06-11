package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// Redis key for the retry queue (pending items)
	aiBillingRetryQueueKey = "ai_billing:failures"
	// Redis key for delayed retry entries (score = Unix timestamp)
	aiBillingDelayedQueueKey = "ai_billing:delayed"
	// Max retry attempts before giving up and alerting
	maxAIRetries = 4
)

// AIFailedBillingPayload holds all data needed to retry a failed AI billing charge.
type AIFailedBillingPayload struct {
	TenantID     string `json:"tenant_id"`
	ExecutionID  string `json:"execution_id,omitempty"`
	FunctionID   string `json:"function_id,omitempty"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Attempts     int    `json:"attempts"`
}

// EnqueueFailedCharge adds a failed AI billing charge to the retry queue with
// exponential backoff. Called immediately after a wallet debit failure.
func EnqueueFailedCharge(ctx context.Context, rdb *redis.Client, payload AIFailedBillingPayload) error {
	payload.Attempts = 0
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ai_billing: failed to marshal payload: %w", err)
	}

	// Exponential backoff: 30s, 2m, 8m, 32m
	delays := []int{30, 120, 480, 1920}
	delay := 30
	if len(delays) > 0 {
		delay = delays[0]
	}

	// Store in the pending queue with score = now + delay
	// We use a delayed sorted set to implement the retry timing
	pipe := rdb.Pipeline()
	pipe.LPush(ctx, aiBillingRetryQueueKey, data)
	pipe.ZAdd(ctx, aiBillingDelayedQueueKey, redis.Z{
		Score:  float64(time.Now().Unix() + int64(delay)),
		Member: string(data),
	})
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("ai_billing: failed to enqueue retry: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": payload.TenantID,
		"delay_sec": delay,
	}).Warn("ai_billing: failed charge enqueued for retry")
	return nil
}

// AIBillingRetryWorker is a background worker that processes failed AI billing charges
// from the Redis retry queue with exponential backoff.
type AIBillingRetryWorker struct {
	redis      *redis.Client
	billingURL string
	client     *http.Client
	stop       chan struct{}
	// suspendFunc is called when retries are exhausted to suspend AI access for a tenant
	// The function signature matches DunningManager.SuspendService pattern
	suspendFunc func(ctx context.Context, tenantID string, reason string) error
}

// NewAIBillingRetryWorker creates a new retry worker.
// suspendFn is called when retries are exhausted - it should suspend AI access for the tenant.
func NewAIBillingRetryWorker(redisClient *redis.Client, billingURL string, suspendFn func(ctx context.Context, tenantID string, reason string) error) *AIBillingRetryWorker {
	return &AIBillingRetryWorker{
		redis:      redisClient,
		billingURL: billingURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stop:        make(chan struct{}),
		suspendFunc: suspendFn,
	}
}

// Start begins the background retry loop.
func (w *AIBillingRetryWorker) Start(ctx context.Context) {
	logrus.Info("ai_billing: retry worker starting")
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("ai_billing: retry worker stopping (context cancelled)")
			return
		case <-w.stop:
			logrus.Info("ai_billing: retry worker stopping")
			return
		case <-ticker.C:
			w.processRetryQueue(ctx)
		}
	}
}

// Stop signals the worker to stop.
func (w *AIBillingRetryWorker) Stop() {
	close(w.stop)
}

// processRetryQueue checks for due retry items and attempts to charge.
func (w *AIBillingRetryWorker) processRetryQueue(ctx context.Context) {
	now := time.Now().Unix()

	// Atomically move due items from delayed set to pending list
	// ZRANGEBYSCORE returns items with score <= now
	cmds, err := w.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// Get items whose delay has expired
		pipe.ZRangeByScore(ctx, aiBillingDelayedQueueKey, &redis.ZRangeBy{
			Min:   "-inf",
			Max:   fmt.Sprintf("%d", now),
			Count: 10,
		})
		return nil
	})
	if err != nil {
		logrus.WithError(err).Warn("ai_billing: retry pipeline failed")
		return
	}

	rangeCmd, ok := cmds[0].(*redis.StringSliceCmd)
	if !ok || rangeCmd.Err() != nil {
		return
	}
	items, err := rangeCmd.Result()
	if err != nil || len(items) == 0 {
		return
	}

	// Remove processed items from delayed queue
	pipe := w.redis.Pipeline()
	for _, item := range items {
		pipe.ZRem(ctx, aiBillingDelayedQueueKey, item)
	}
	pipe.Exec(ctx)

	delays := []int{30, 120, 480, 1920}

	for _, data := range items {
		var payload AIFailedBillingPayload
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			logrus.WithError(err).Warn("ai_billing: failed to unmarshal retry payload")
			continue
		}

		payload.Attempts++
		attempts := payload.Attempts

		// Attempt to charge
		err := w.attemptCharge(ctx, payload)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"tenant_id":    payload.TenantID,
				"execution_id": payload.ExecutionID,
			}).Info("ai_billing: retry succeeded")
			// Remove from pending queue (already removed from delayed)
			w.redis.LRem(ctx, aiBillingRetryQueueKey, 1, data)
			continue
		}

		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id": payload.TenantID,
			"attempt":   attempts,
			"max":       maxAIRetries,
		}).Warn("ai_billing: retry charge failed")

		if attempts >= maxAIRetries {
			// Give up: remove from pending and alert
			w.redis.LRem(ctx, aiBillingRetryQueueKey, 1, data)
			w.sendSuspensionAlert(ctx, payload)
			continue
		}

		// Re-enqueue with next backoff delay
		payload.Attempts = attempts
		newData, _ := json.Marshal(payload)
		delay := delays[0]
		if attempts-1 < len(delays) {
			delay = delays[attempts-1]
		}
		pipe := w.redis.Pipeline()
		pipe.ZAdd(ctx, aiBillingDelayedQueueKey, redis.Z{
			Score:  float64(time.Now().Unix() + int64(delay)),
			Member: string(newData),
		})
		pipe.Exec(ctx)
	}
}

// attemptCharge POSTs the billing charge to the orchestrator endpoint.
func (w *AIBillingRetryWorker) attemptCharge(ctx context.Context, payload AIFailedBillingPayload) error {
	type chargeReq struct {
		TenantID     string `json:"tenant_id"`
		ExecutionID  string `json:"execution_id,omitempty"`
		FunctionID   string `json:"function_id,omitempty"`
		Provider     string `json:"provider"`
		Model        string `json:"model"`
		InputTokens  int    `json:"input_tokens"`
		OutputTokens int    `json:"output_tokens"`
	}

	reqBody := chargeReq{
		TenantID:     payload.TenantID,
		ExecutionID:  payload.ExecutionID,
		FunctionID:   payload.FunctionID,
		Provider:     payload.Provider,
		Model:        payload.Model,
		InputTokens:  payload.InputTokens,
		OutputTokens: payload.OutputTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	url := w.billingURL + "/v1/billing/ai/charge"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = nil

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != 402 && resp.StatusCode != 409 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil // 2xx, 402 (retry later), or 409 (already charged) are all considered success
}

// sendSuspensionAlert suspends AI access for a tenant that has exhausted retries.
// It calls the configured suspend function to block further AI executions.
func (w *AIBillingRetryWorker) sendSuspensionAlert(ctx context.Context, payload AIFailedBillingPayload) {
	logrus.WithFields(logrus.Fields{
		"tenant_id":    payload.TenantID,
		"execution_id": payload.ExecutionID,
		"provider":     payload.Provider,
		"model":        payload.Model,
	}).Error("ai_billing: MAX RETRIES EXHAUSTED — suspending AI access for tenant")

	// Actually suspend AI access for the tenant
	if w.suspendFunc != nil {
		if err := w.suspendFunc(ctx, payload.TenantID, "ai_billing_retries_exhausted"); err != nil {
			logrus.WithError(err).WithField("tenant_id", payload.TenantID).
				Error("ai_billing: failed to suspend AI access for tenant")
		} else {
			logrus.WithField("tenant_id", payload.TenantID).
				Info("ai_billing: AI access suspended for tenant after exhausted retries")
		}
	}
}

// QueueLength returns the current retry queue size (for monitoring).
func QueueLength(ctx context.Context, rdb *redis.Client) (int64, error) {
	return rdb.LLen(ctx, aiBillingRetryQueueKey).Result()
}
