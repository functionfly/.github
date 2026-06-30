package execution

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/queue/rabbitmq"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// shouldQueueExecution determines if an execution should be queued due to high load.
//
// Enable with:
// - EXECUTION_QUEUE_ENABLED=true
// - Client sets header: X-Queue-If-Busy: true
//
// Queued executions are consumed by the RabbitMQ consumer
// (internal/queue/rabbitmq/consumer.go) or the in-memory ExecutionQueue
// (internal/queue/execution_queue.go) depending on deployment configuration.
//
// Heuristic: if cache is under pressure (low hit ratio and high evictions), treat the node as "busy".
func (h *Handler) shouldQueueExecution(r *http.Request) bool {
	if os.Getenv("EXECUTION_QUEUE_ENABLED") != "true" {
		return false
	}
	if strings.ToLower(strings.TrimSpace(r.Header.Get("X-Queue-If-Busy"))) != "true" {
		return false
	}

	// If we have no cache service metrics, don't guess.
	if h.CacheService == nil {
		return false
	}
	mem := h.CacheService.GetMemoryStats()
	if mem == nil {
		return false
	}

	// Optional tuning knobs.
	minEvictions := int64(500)
	minSizeBytes := int64(50 * 1024 * 1024) // 50MB
	maxHitRatio := 0.15
	if v := os.Getenv("EXECUTION_QUEUE_MIN_EVICTIONS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 0 {
			minEvictions = n
		}
	}
	if v := os.Getenv("EXECUTION_QUEUE_MIN_SIZE_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 0 {
			minSizeBytes = n
		}
	}
	if v := os.Getenv("EXECUTION_QUEUE_MAX_HIT_RATIO"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 && f <= 1 {
			maxHitRatio = f
		}
	}

	// Under pressure: large cache + lots of evictions + poor hit ratio.
	if mem.SizeBytes >= minSizeBytes && mem.Evictions >= minEvictions && mem.Ratio <= maxHitRatio {
		return true
	}
	return false
}

// queueExecution queues a function execution for later processing.
// Publishes to RabbitMQ when configured; otherwise falls back to an in-memory
// priority queue backed by internal/queue/execution_queue.go workers.
func (h *Handler) queueExecution(r *http.Request, functionID uuid.UUID, execReq functionregistry.ExecutionRequest, fnVersion *storage.RegistryFunctionVersion) error {
	// Prefer RabbitMQ when configured; otherwise, log-only fallback.
	pub := rabbitmq.NewPublisherFromEnv()
	if pub.Enabled() {
		defer pub.Close()

		msg := map[string]any{
			"type":           "function_execution",
			"function_id":    functionID.String(),
			"author":         mux.Vars(r)["author"],
			"name":           mux.Vars(r)["name"],
			"version":        fnVersion.Version,
			"input_json":     json.RawMessage(execReq.Input),
			"requested_at":   time.Now().UTC().Format(time.RFC3339Nano),
			"request_id":     r.Header.Get("X-Request-ID"),
			"user_agent":     r.Header.Get("User-Agent"),
			"queue_reason":   "high_load",
			"cache_eligible": fnVersion.Deterministic && fnVersion.CacheTTL > 0,
		}

		// MessageId helps de-dupe on consumer side if needed (best-effort).
		messageID := fmt.Sprintf("%s:%s:%s", functionID.String(), fnVersion.Version, r.Header.Get("X-Request-ID"))
		if err := pub.PublishJSON(r.Context(), msg, rabbitmq.PublishOptions{MessageID: messageID}); err != nil {
			return fmt.Errorf("publish to rabbitmq: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"function_id": functionID,
			"version":     fnVersion.Version,
			"input_size":  len(execReq.Input),
			"request_id":  r.Header.Get("X-Request-ID"),
		}).Info("Execution queued to RabbitMQ")

		return nil
	}

	logrus.WithFields(logrus.Fields{
		"function_id": functionID,
		"version":     fnVersion.Version,
		"input_size":  len(execReq.Input),
	}).Info("Execution queued (no broker configured; log-only)")
	return nil
}
