// Package asynq provides a Redis-backed job queue for FunctionFly.
//
// This package is the Go equivalent of BullMQ (Node.js) — same Redis backend,
// same job semantics (delayed, scheduled, unique, priority, retries with
// backoff, dead-letter, rate-limited queues). BullMQ itself cannot be used
// here because the orchestrator is Go and FunctionFly has no long-lived
// Node.js worker process; the asynq library (github.com/hibiken/asynq)
// provides the same feature set for Go services and uses Redis as its
// storage layer.
//
// To swap in BullMQ later, replace the contents of this package with a thin
// HTTP/gRPC client that enqueues jobs into a Node.js BullMQ worker service
// and translates asynq job types to BullMQ job names. The public surface
// (EnqueueAIBillingRetry, EnqueueTrustWebhookDelivery, etc.) would not
// change.
package asynq

// Job type identifiers. These string constants are the names callers pass
// when enqueueing work, and the names the worker mux dispatches on. They
// also become the queue names when using per-type queues (see
// ManagerConfig.Queues).
const (
	// TypeAIBillingRetry retries a failed AI provider charge with
	// exponential backoff. Replaces the hand-rolled ZSET/ZRANGEBYSCORE
	// logic in internal/billing/ai_billing_retry_worker.go.
	TypeAIBillingRetry = "ai_billing:retry"

	// TypeTrustWebhookDelivery delivers a single Trust API webhook
	// attempt. Replaces the goroutine-per-delivery model and the
	// "ProcessRetries" / "ProcessPendingDeliveries" pollers in
	// internal/storage/trustapi/webhook_service.go.
	TypeTrustWebhookDelivery = "trust:webhook:delivery"
)

// Default queue names. Asynq supports multiple named queues with
// per-queue priority and concurrency; we start with one default queue
// for all job types, and per-type queues can be added later via
// ManagerConfig.Queues.
const (
	DefaultQueue = "default"

	// QueueCritical is reserved for work that must run before any
	// other job in the same worker (e.g. revocation webhooks).
	QueueCritical = "critical"
)
