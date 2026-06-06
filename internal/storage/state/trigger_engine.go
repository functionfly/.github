package state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ============================================
// Database Models for Trigger Queue and Observability
// ============================================

// QueuedTriggerEvent represents a trigger event waiting to be processed
type QueuedTriggerEvent struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TriggerID     uuid.UUID  `json:"trigger_id" gorm:"type:uuid;not null;index:idx_trigger_queue_status,idx_trigger_queue_created"`
	StateID       uuid.UUID  `json:"state_id" gorm:"type:uuid;not null;index"`
	TenantID      uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_trigger_queue_tenant,idx_trigger_queue_created"`
	Key           string     `json:"key" gorm:"size:500"`
	EventType     string     `json:"event_type" gorm:"size:50"` // "set", "delete", "patch"
	OldValue      *JSONMap   `json:"old_value,omitempty" gorm:"type:jsonb"`
	NewValue      *JSONMap   `json:"new_value,omitempty" gorm:"type:jsonb"`
	CorrelationID string     `json:"correlation_id" gorm:"size:255;index"`
	Payload       JSONMap    `json:"payload" gorm:"type:jsonb"`
	Status        string     `json:"status" gorm:"size:50;default:'pending';index:idx_trigger_queue_status,idx_trigger_queue_created"` // "pending", "processing", "completed", "failed", "dead_letter"
	Attempts      int        `json:"attempts" gorm:"default:0"`
	MaxAttempts   int        `json:"max_attempts" gorm:"default:3"`
	NextAttemptAt *time.Time `json:"next_attempt_at,omitempty;index:idx_trigger_queue_next_attempt"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	LastErrorAt   *time.Time `json:"last_error_at,omitempty"`
	ExecutedAt    *time.Time `json:"executed_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime;index:idx_trigger_queue_created"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (QueuedTriggerEvent) TableName() string {
	return "trigger_event_queue"
}

// TriggerExecutionLog records all trigger executions for observability
type TriggerExecutionLog struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TriggerID        uuid.UUID  `json:"trigger_id" gorm:"type:uuid;not null;index:idx_exec_log_trigger_created,idx_exec_log_tenant_created"`
	StateID          uuid.UUID  `json:"state_id" gorm:"type:uuid;not null;index"`
	TenantID         uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_exec_log_tenant_created"`
	QueuedEventID    *uuid.UUID `json:"queued_event_id,omitempty" gorm:"type:uuid;index"`
	EventType        string     `json:"event_type" gorm:"size:50"`
	Key              string     `json:"key" gorm:"size:500"`
	TargetType       string     `json:"target_type" gorm:"size:50"` // "function", "webhook"
	TargetURL        string     `json:"target_url" gorm:"size:1000"`
	PayloadSizeBytes int        `json:"payload_size_bytes"`
	Status           string     `json:"status" gorm:"size:50;index:idx_exec_log_status_created"` // "success", "error", "timeout", "rate_limited"
	HTTPStatusCode   *int       `json:"http_status_code,omitempty"`
	ResponseSize     *int       `json:"response_size,omitempty"`
	DurationMs       int64      `json:"duration_ms"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	RetryCount       int        `json:"retry_count" gorm:"default:0"`
	CorrelationID    string     `json:"correlation_id" gorm:"size:255;index"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime;index:idx_exec_log_trigger_created,idx_exec_log_tenant_created,idx_exec_log_status_created"`
}

func (TriggerExecutionLog) TableName() string {
	return "trigger_execution_logs"
}

// TriggerDeadLetter holds events that failed permanently
type TriggerDeadLetter struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OriginalEventID uuid.UUID  `json:"original_event_id" gorm:"type:uuid;not null;index"`
	TriggerID       uuid.UUID  `json:"trigger_id" gorm:"type:uuid;not null;index"`
	StateID         uuid.UUID  `json:"state_id" gorm:"type:uuid;not null;index"`
	TenantID        uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	EventType       string     `json:"event_type" gorm:"size:50"`
	Key             string     `json:"key" gorm:"size:500"`
	Payload         JSONMap    `json:"payload" gorm:"type:jsonb"`
	FinalError      string     `json:"final_error" gorm:"size:2000"`
	FailedAttempts  int        `json:"failed_attempts"`
	CorrelationID   string     `json:"correlation_id" gorm:"size:255"`
	CanRetry        bool       `json:"can_retry" gorm:"default:true"`
	RetriedAt       *time.Time `json:"retried_at,omitempty"`
	RetriedSuccess  bool       `json:"retried_success" gorm:"default:false"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (TriggerDeadLetter) TableName() string {
	return "trigger_dead_letter"
}

// ============================================
// Trigger Executor Interfaces
// ============================================

// TriggerExecutionResult represents the result of a trigger execution
type TriggerExecutionResult struct {
	TriggerID      uuid.UUID `json:"trigger_id"`
	Success        bool      `json:"success"`
	HTTPStatusCode int       `json:"http_status_code,omitempty"`
	Output         *JSONMap  `json:"output,omitempty"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	DurationMs     int64     `json:"duration_ms"`
	ExecutedAt     time.Time `json:"executed_at"`
	Retryable      bool      `json:"retryable"` // Whether this error is retryable
	ResponseSize   int       `json:"response_size,omitempty"`
}

// TriggerExecutor is an interface for executing trigger functions
type TriggerExecutor interface {
	ExecuteFunction(ctx context.Context, functionID uuid.UUID, payload JSONMap) (*TriggerExecutionResult, error)
	ExecuteTrigger(ctx context.Context, event QueuedTriggerEvent, trigger StateTrigger) (*TriggerExecutionResult, error)
	GetTargetURL(trigger StateTrigger) string
}

// FunctionTriggerExecutor executes triggers by calling internal functions via HTTP
type FunctionTriggerExecutor struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	logger     *logrus.Logger
}

// NewHTTPTriggerExecutor creates a new HTTP trigger executor (backward compatibility alias)
func NewHTTPTriggerExecutor(baseURL, apiKey string, logger *logrus.Logger) *FunctionTriggerExecutor {
	return NewFunctionTriggerExecutor(baseURL, apiKey, logger)
}

// NewFunctionTriggerExecutor creates a new function trigger executor
func NewFunctionTriggerExecutor(baseURL, apiKey string, logger *logrus.Logger) *FunctionTriggerExecutor {
	if logger == nil {
		logger = logrus.New()
	}
	return &FunctionTriggerExecutor{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
		logger: logger,
	}
}

// GetTargetURL returns the target URL for a function trigger
func (e *FunctionTriggerExecutor) GetTargetURL(trigger StateTrigger) string {
	if trigger.TargetFunctionID != nil {
		return fmt.Sprintf("%s/v1/functions/%s/execute", e.baseURL, trigger.TargetFunctionID.String())
	}
	return fmt.Sprintf("%s/v1/functions/by-name/%s/execute", e.baseURL, trigger.TargetFunction)
}

// ExecuteFunction executes a function via HTTP (backward compatibility)
func (e *FunctionTriggerExecutor) ExecuteFunction(ctx context.Context, functionID uuid.UUID, payload JSONMap) (*TriggerExecutionResult, error) {
	start := time.Now()

	url := fmt.Sprintf("%s/v1/functions/%s/execute", e.baseURL, functionID.String())

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    functionID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to marshal payload: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    false,
		}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    functionID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    false,
		}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.apiKey))

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    functionID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to execute: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    true,
		}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result TriggerExecutionResult
	result.TriggerID = functionID
	result.HTTPStatusCode = resp.StatusCode
	result.DurationMs = time.Since(start).Milliseconds()
	result.ExecutedAt = time.Now()
	result.ResponseSize = len(body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Retryable = false
		var output JSONMap
		if err := json.Unmarshal(body, &output); err == nil {
			result.Output = &output
		}
	} else if resp.StatusCode == 429 {
		result.Success = false
		result.ErrorMessage = "rate limited"
		result.Retryable = true
	} else if resp.StatusCode >= 500 {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("server error: HTTP %d", resp.StatusCode)
		result.Retryable = true
	} else {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("client error: HTTP %d - %s", resp.StatusCode, string(body))
		result.Retryable = false
	}

	return &result, nil
}

// ExecuteTrigger executes a trigger by calling the target function
func (e *FunctionTriggerExecutor) ExecuteTrigger(ctx context.Context, event QueuedTriggerEvent, trigger StateTrigger) (*TriggerExecutionResult, error) {
	start := time.Now()

	// Build payload
	payload := e.buildPayload(event, trigger)

	url := e.GetTargetURL(trigger)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    trigger.ID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to marshal payload: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    false,
		}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    trigger.ID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    false,
		}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.apiKey))
	req.Header.Set("X-Trigger-ID", event.TriggerID.String())
	req.Header.Set("X-Correlation-ID", event.CorrelationID)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    trigger.ID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("request failed: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    true,
		}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := &TriggerExecutionResult{
		TriggerID:      trigger.ID,
		HTTPStatusCode: resp.StatusCode,
		DurationMs:     time.Since(start).Milliseconds(),
		ExecutedAt:     time.Now(),
		ResponseSize:   len(body),
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		result.Success = true
		result.Retryable = false
		var output JSONMap
		if err := json.Unmarshal(body, &output); err == nil {
			result.Output = &output
		}
	case resp.StatusCode == 429:
		result.Success = false
		result.ErrorMessage = "rate limited"
		result.Retryable = true
	case resp.StatusCode >= 500:
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("server error: HTTP %d", resp.StatusCode)
		result.Retryable = true
	default:
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("client error: HTTP %d", resp.StatusCode)
		result.Retryable = false
	}

	return result, nil
}

func (e *FunctionTriggerExecutor) buildPayload(event QueuedTriggerEvent, trigger StateTrigger) JSONMap {
	payload := JSONMap{
		"trigger_type":   trigger.TriggerType,
		"source_state":   event.StateID.String(),
		"key":            event.Key,
		"event_type":     event.EventType,
		"correlation_id": event.CorrelationID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"trigger_id":     trigger.ID.String(),
	}

	if trigger.IncludePrevious && event.OldValue != nil {
		payload["previous_value"] = *event.OldValue
	}
	if trigger.IncludeNew && event.NewValue != nil {
		payload["new_value"] = *event.NewValue
	}

	return payload
}

// WebhookTriggerExecutor executes triggers by calling external webhooks
type WebhookTriggerExecutor struct {
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewWebhookTriggerExecutor creates a new webhook trigger executor
func NewWebhookTriggerExecutor(logger *logrus.Logger) *WebhookTriggerExecutor {
	if logger == nil {
		logger = logrus.New()
	}
	return &WebhookTriggerExecutor{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetTargetURL returns the webhook URL from trigger condition
func (e *WebhookTriggerExecutor) GetTargetURL(trigger StateTrigger) string {
	if trigger.Condition != nil {
		if webhookURL, ok := trigger.Condition["webhook_url"].(string); ok && webhookURL != "" {
			return webhookURL
		}
	}
	return ""
}

// ExecuteFunction is not supported for webhooks
func (e *WebhookTriggerExecutor) ExecuteFunction(ctx context.Context, functionID uuid.UUID, payload JSONMap) (*TriggerExecutionResult, error) {
	return nil, fmt.Errorf("webhook executor does not support ExecuteFunction")
}

// ExecuteTrigger executes a trigger by calling a webhook URL
func (e *WebhookTriggerExecutor) ExecuteTrigger(ctx context.Context, event QueuedTriggerEvent, trigger StateTrigger) (*TriggerExecutionResult, error) {
	start := time.Now()

	webhookURL := e.GetTargetURL(trigger)
	if webhookURL == "" {
		return &TriggerExecutionResult{
			TriggerID:    trigger.ID,
			Success:      false,
			ErrorMessage: "no webhook_url configured in trigger condition",
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    false,
		}, fmt.Errorf("no webhook_url configured")
	}

	// Build payload
	payload := JSONMap{
		"trigger_type":   trigger.TriggerType,
		"source_state":   event.StateID.String(),
		"key":            event.Key,
		"event_type":     event.EventType,
		"correlation_id": event.CorrelationID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"trigger_id":     trigger.ID.String(),
		"tenant_id":      event.TenantID.String(),
	}

	if trigger.IncludePrevious && event.OldValue != nil {
		payload["previous_value"] = *event.OldValue
	}
	if trigger.IncludeNew && event.NewValue != nil {
		payload["new_value"] = *event.NewValue
	}

	// Add custom headers from trigger config
	headers := map[string]string{
		"Content-Type":     "application/json",
		"X-Trigger-ID":     trigger.ID.String(),
		"X-Correlation-ID": event.CorrelationID,
		"X-Event-Type":     event.EventType,
		"User-Agent":       "FunctionFly-TriggerEngine/1.0",
	}

	if trigger.Condition != nil {
		if customHeaders, ok := trigger.Condition["headers"].(map[string]interface{}); ok {
			for k, v := range customHeaders {
				if vs, ok := v.(string); ok {
					headers[k] = vs
				}
			}
		}
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    trigger.ID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to marshal payload: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    false,
		}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(jsonPayload))
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    trigger.ID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    false,
		}, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    trigger.ID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("webhook request failed: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
			Retryable:    true,
		}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := &TriggerExecutionResult{
		TriggerID:      trigger.ID,
		HTTPStatusCode: resp.StatusCode,
		DurationMs:     time.Since(start).Milliseconds(),
		ExecutedAt:     time.Now(),
		ResponseSize:   len(body),
	}

	// Webhooks follow similar retry logic
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		result.Success = true
		result.Retryable = false
		var output JSONMap
		if err := json.Unmarshal(body, &output); err == nil {
			result.Output = &output
		}
	case resp.StatusCode == 429:
		result.Success = false
		result.ErrorMessage = "rate limited by webhook"
		result.Retryable = true
	case resp.StatusCode >= 500:
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("webhook server error: HTTP %d", resp.StatusCode)
		result.Retryable = true
	default:
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("webhook client error: HTTP %d", resp.StatusCode)
		result.Retryable = false
	}

	return result, nil
}

// MultiExecutor chooses between function and webhook execution based on trigger config
type MultiExecutor struct {
	functionExecutor *FunctionTriggerExecutor
	webhookExecutor  *WebhookTriggerExecutor
	logger           *logrus.Logger
}

// NewMultiExecutor creates a new multi-executor
func NewMultiExecutor(functionExecutor *FunctionTriggerExecutor, webhookExecutor *WebhookTriggerExecutor, logger *logrus.Logger) *MultiExecutor {
	return &MultiExecutor{
		functionExecutor: functionExecutor,
		webhookExecutor:  webhookExecutor,
		logger:           logger,
	}
}

// GetTargetURL returns the appropriate target URL
func (e *MultiExecutor) GetTargetURL(trigger StateTrigger) string {
	// Check if webhook URL is configured
	if url := e.webhookExecutor.GetTargetURL(trigger); url != "" {
		return url
	}
	return e.functionExecutor.GetTargetURL(trigger)
}

// ExecuteFunction delegates to the function executor
func (e *MultiExecutor) ExecuteFunction(ctx context.Context, functionID uuid.UUID, payload JSONMap) (*TriggerExecutionResult, error) {
	return e.functionExecutor.ExecuteFunction(ctx, functionID, payload)
}

// ExecuteTrigger chooses the right executor based on trigger configuration
func (e *MultiExecutor) ExecuteTrigger(ctx context.Context, event QueuedTriggerEvent, trigger StateTrigger) (*TriggerExecutionResult, error) {
	// Check if this is a webhook trigger
	if e.webhookExecutor.GetTargetURL(trigger) != "" {
		return e.webhookExecutor.ExecuteTrigger(ctx, event, trigger)
	}
	// Otherwise use function executor
	return e.functionExecutor.ExecuteTrigger(ctx, event, trigger)
}

// ============================================
// Trigger Engine Configuration
// ============================================

// TriggerEngineConfig holds configuration for the trigger engine
type TriggerEngineConfig struct {
	// Polling interval for checking new events
	PollInterval time.Duration
	// Maximum concurrent executions
	MaxConcurrent int
	// Batch size for processing events
	BatchSize int
	// Enable/disable the engine
	Enabled bool
	// Retry configuration
	MaxRetries       int
	RetryBackoffBase time.Duration
	// Dead letter queue settings
	DeadLetterEnabled bool
	// Per-tenant rate limit (triggers per minute)
	TenantRateLimit int
}

// DefaultTriggerEngineConfig returns the default configuration
func DefaultTriggerEngineConfig() TriggerEngineConfig {
	return TriggerEngineConfig{
		PollInterval:      5 * time.Second,
		MaxConcurrent:     10,
		BatchSize:         100,
		Enabled:           true,
		MaxRetries:        3,
		RetryBackoffBase:  5 * time.Second,
		DeadLetterEnabled: true,
		TenantRateLimit:   100, // 100 triggers per minute per tenant
	}
}

// ============================================
// Trigger Engine
// ============================================

// TriggerEngine processes state triggers
type TriggerEngine struct {
	db       *gorm.DB
	config   TriggerEngineConfig
	executor TriggerExecutor
	logger   *logrus.Logger

	// Rate limiting - per trigger
	rateLimitLock sync.Mutex
	lastExecution map[uuid.UUID]time.Time
	minInterval   time.Duration

	// Per-tenant rate limiting
	tenantRateLimitLock sync.Mutex
	tenantLastExecution map[uuid.UUID]time.Time
	tenantMinInterval   time.Duration

	// Execution tracking
	mu          sync.Mutex
	executing   map[uuid.UUID]bool
	workerCount int
	stopChan    chan struct{}

	// Event queue channel
	eventChan chan QueuedTriggerEvent
	workers   sync.WaitGroup
}

// NewTriggerEngine creates a new trigger engine
func NewTriggerEngine(
	db *gorm.DB,
	config TriggerEngineConfig,
	executor TriggerExecutor,
	logger *logrus.Logger,
) *TriggerEngine {
	if config.PollInterval == 0 {
		config = DefaultTriggerEngineConfig()
	}
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 10
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryBackoffBase == 0 {
		config.RetryBackoffBase = 5 * time.Second
	}
	if config.TenantRateLimit == 0 {
		config.TenantRateLimit = 100
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &TriggerEngine{
		db:                  db,
		config:              config,
		executor:            executor,
		logger:              logger,
		lastExecution:       make(map[uuid.UUID]time.Time),
		minInterval:         time.Minute, // Minimum 1 minute between same trigger
		tenantLastExecution: make(map[uuid.UUID]time.Time),
		tenantMinInterval:   time.Duration(60/config.TenantRateLimit) * time.Second,
		executing:           make(map[uuid.UUID]bool),
		workerCount:         0,
		stopChan:            make(chan struct{}),
		eventChan:           make(chan QueuedTriggerEvent, config.BatchSize*2),
	}
}

// Start starts the trigger engine
func (e *TriggerEngine) Start(ctx context.Context) {
	if !e.config.Enabled {
		e.logger.Info("Trigger engine is disabled")
		return
	}

	e.logger.WithFields(logrus.Fields{
		"pollInterval":    e.config.PollInterval,
		"maxConcurrent":   e.config.MaxConcurrent,
		"maxRetries":      e.config.MaxRetries,
		"tenantRateLimit": e.config.TenantRateLimit,
	}).Info("Starting trigger engine")

	// Start worker pool
	for i := 0; i < e.config.MaxConcurrent; i++ {
		e.workers.Add(1)
		go e.worker(ctx)
	}

	// Start the polling worker for legacy trigger processing
	go e.pollingWorker(ctx)

	// Start retry worker for failed events
	go e.retryWorker(ctx)

	e.logger.Info("Trigger engine started")
}

// Stop stops the trigger engine
func (e *TriggerEngine) Stop() {
	e.logger.Info("Stopping trigger engine")
	close(e.stopChan)
	e.workers.Wait()
}

// worker processes events from the event channel
func (e *TriggerEngine) worker(ctx context.Context) {
	defer e.workers.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case event := <-e.eventChan:
			e.processQueuedEvent(ctx, event)
		}
	}
}

// pollingWorker is a single worker that processes legacy triggers
func (e *TriggerEngine) pollingWorker(ctx context.Context) {
	ticker := time.NewTicker(e.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.processLegacyTriggers(ctx)
		}
	}
}

// retryWorker handles retrying failed events
func (e *TriggerEngine) retryWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check for retries every 30s
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.processRetries(ctx)
		}
	}
}

// processQueuedEvent processes a single event from the queue
func (e *TriggerEngine) processQueuedEvent(ctx context.Context, event QueuedTriggerEvent) {
	// Get trigger details
	var trigger StateTrigger
	err := e.db.WithContext(ctx).First(&trigger, "id = ?", event.TriggerID).Error
	if err != nil {
		e.logger.WithError(err).WithField("trigger_id", event.TriggerID).Error("Failed to fetch trigger for queued event")
		return
	}

	// Check rate limits
	if !e.checkRateLimit(trigger.ID) {
		e.logger.WithField("trigger_id", trigger.ID).Debug("Trigger rate limited")
		return
	}

	if !e.checkTenantRateLimit(event.TenantID) {
		e.logger.WithField("tenant_id", event.TenantID).Debug("Tenant rate limited")
		return
	}

	// Execute the trigger
	result, err := e.executor.ExecuteTrigger(ctx, event, trigger)

	// Record execution
	e.recordExecutionLog(ctx, event, trigger, result, err)

	if err != nil || !result.Success {
		// Handle failure
		e.handleExecutionFailure(ctx, event, result, err)
	} else {
		// Mark as completed
		e.markEventCompleted(ctx, event, result)
		e.updateRateLimit(trigger.ID)
		e.updateTenantRateLimit(event.TenantID)
	}
}

// handleExecutionFailure handles failed executions with retry logic
func (e *TriggerEngine) handleExecutionFailure(ctx context.Context, event QueuedTriggerEvent, result *TriggerExecutionResult, execErr error) {
	newAttempt := event.Attempts + 1
	maxAttempts := event.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = e.config.MaxRetries
	}

	errorMsg := ""
	if execErr != nil {
		errorMsg = execErr.Error()
	} else if result != nil {
		errorMsg = result.ErrorMessage
	}

	retryable := true
	if result != nil && !result.Retryable {
		retryable = false
	}

	if newAttempt >= maxAttempts || !retryable {
		// Move to dead letter queue
		if e.config.DeadLetterEnabled {
			e.moveToDeadLetter(ctx, event, errorMsg, newAttempt)
		}
		// Mark as failed
		e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("id = ?", event.ID).Updates(map[string]interface{}{
			"status":        "dead_letter",
			"attempts":      newAttempt,
			"error_message": errorMsg,
			"last_error_at": time.Now(),
		})
	} else {
		// Schedule retry with exponential backoff
		backoff := e.config.RetryBackoffBase * time.Duration(1<<uint(newAttempt-1))
		nextAttempt := time.Now().Add(backoff)

		e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("id = ?", event.ID).Updates(map[string]interface{}{
			"status":          "pending",
			"attempts":        newAttempt,
			"error_message":   errorMsg,
			"last_error_at":   time.Now(),
			"next_attempt_at": nextAttempt,
		})

		e.logger.WithFields(logrus.Fields{
			"event_id":     event.ID,
			"attempt":      newAttempt,
			"max_attempts": maxAttempts,
			"next_attempt": nextAttempt,
			"error":        errorMsg,
		}).Info("Scheduled trigger retry")
	}
}

// moveToDeadLetter moves a failed event to the dead letter queue
func (e *TriggerEngine) moveToDeadLetter(ctx context.Context, event QueuedTriggerEvent, finalError string, failedAttempts int) {
	dlq := TriggerDeadLetter{
		OriginalEventID: event.ID,
		TriggerID:       event.TriggerID,
		StateID:         event.StateID,
		TenantID:        event.TenantID,
		EventType:       event.EventType,
		Key:             event.Key,
		Payload:         event.Payload,
		FinalError:      finalError,
		FailedAttempts:  failedAttempts,
		CorrelationID:   event.CorrelationID,
		CanRetry:        true,
	}

	if err := e.db.WithContext(ctx).Create(&dlq).Error; err != nil {
		e.logger.WithError(err).Error("Failed to create dead letter entry")
	}

	e.logger.WithFields(logrus.Fields{
		"event_id":        event.ID,
		"trigger_id":      event.TriggerID,
		"failed_attempts": failedAttempts,
	}).Warn("Event moved to dead letter queue")
}

// markEventCompleted marks an event as successfully completed
func (e *TriggerEngine) markEventCompleted(ctx context.Context, event QueuedTriggerEvent, result *TriggerExecutionResult) {
	now := time.Now()
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("id = ?", event.ID).Updates(map[string]interface{}{
		"status":        "completed",
		"executed_at":   now,
		"completed_at":  now,
		"error_message": nil,
	})

	// Update trigger last triggered time
	e.db.WithContext(ctx).Model(&StateTrigger{}).Where("id = ?", event.TriggerID).Update("last_triggered_at", &now)
}

// recordExecutionLog records an execution to the log
func (e *TriggerEngine) recordExecutionLog(ctx context.Context, event QueuedTriggerEvent, trigger StateTrigger, result *TriggerExecutionResult, execErr error) {
	status := "success"
	if execErr != nil || (result != nil && !result.Success) {
		if result != nil && result.HTTPStatusCode == 429 {
			status = "rate_limited"
		} else {
			status = "error"
		}
	}

	var errorMsg *string
	if execErr != nil {
		s := execErr.Error()
		errorMsg = &s
	} else if result != nil && result.ErrorMessage != "" {
		errorMsg = &result.ErrorMessage
	}

	var httpStatus, respSize *int
	if result != nil {
		h := result.HTTPStatusCode
		httpStatus = &h
		r := result.ResponseSize
		respSize = &r
	}

	durationMs := int64(0)
	if result != nil {
		durationMs = result.DurationMs
	}

	log := TriggerExecutionLog{
		TriggerID:        event.TriggerID,
		StateID:          event.StateID,
		TenantID:         event.TenantID,
		QueuedEventID:    &event.ID,
		EventType:        event.EventType,
		Key:              event.Key,
		TargetType:       e.getTargetType(trigger),
		TargetURL:        e.executor.GetTargetURL(trigger),
		PayloadSizeBytes: len([]byte(fmt.Sprintf("%v", event.Payload))),
		Status:           status,
		HTTPStatusCode:   httpStatus,
		ResponseSize:     respSize,
		DurationMs:       durationMs,
		ErrorMessage:     errorMsg,
		RetryCount:       event.Attempts,
		CorrelationID:    event.CorrelationID,
	}

	if err := e.db.WithContext(ctx).Create(&log).Error; err != nil {
		e.logger.WithError(err).Warn("Failed to record trigger execution log")
	}
}

func (e *TriggerEngine) getTargetType(trigger StateTrigger) string {
	// Check if it's a webhook trigger
	if trigger.Condition != nil {
		if webhookURL, ok := trigger.Condition["webhook_url"].(string); ok && webhookURL != "" {
			return "webhook"
		}
	}
	return "function"
}

// processRetries handles retrying failed events
func (e *TriggerEngine) processRetries(ctx context.Context) {
	var events []QueuedTriggerEvent
	err := e.db.WithContext(ctx).
		Where("status = ? AND next_attempt_at <= ?", "pending", time.Now()).
		Order("next_attempt_at ASC").
		Limit(e.config.BatchSize).
		Find(&events).Error

	if err != nil {
		e.logger.WithError(err).Error("Failed to fetch events for retry")
		return
	}

	for _, event := range events {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		default:
		}

		// Mark as processing
		e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("id = ?", event.ID).Update("status", "processing")

		// Send to event channel
		e.eventChan <- event
	}
}

// processLegacyTriggers processes legacy triggers (backward compatibility)
func (e *TriggerEngine) processLegacyTriggers(ctx context.Context) {
	// Find active triggers
	var triggers []StateTrigger
	err := e.db.WithContext(ctx).
		Model(&StateTrigger{}).
		Where("is_active = ?", true).
		Find(&triggers).Error

	if err != nil {
		e.logger.WithError(err).Error("Failed to fetch active triggers")
		return
	}

	for _, trigger := range triggers {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		default:
		}

		// Check rate limit
		if !e.checkRateLimit(trigger.ID) {
			continue
		}

		// Check if already executing
		e.mu.Lock()
		if e.executing[trigger.ID] {
			e.mu.Unlock()
			continue
		}
		e.executing[trigger.ID] = true
		e.mu.Unlock()

		// Execute trigger
		go e.executeLegacyTrigger(ctx, trigger)
	}
}

// executeLegacyTrigger executes a legacy trigger
func (e *TriggerEngine) executeLegacyTrigger(ctx context.Context, trigger StateTrigger) {
	defer func() {
		e.mu.Lock()
		delete(e.executing, trigger.ID)
		e.mu.Unlock()
	}()

	// Get state changes since last trigger
	events, err := e.getPendingEvents(ctx, trigger)
	if err != nil {
		e.logger.WithError(err).Error("Failed to get pending events for trigger")
		return
	}

	if len(events) == 0 {
		return
	}

	// Queue each event for processing
	for _, event := range events {
		// Create queued event
		key := ""
		if event.Key != nil {
			key = *event.Key
		}

		queuedEvent := QueuedTriggerEvent{
			ID:            uuid.New(),
			TriggerID:     trigger.ID,
			StateID:       event.StateID,
			TenantID:      trigger.TenantID,
			Key:           key,
			EventType:     event.EventType,
			OldValue:      event.PreviousValue,
			NewValue:      event.NewValue,
			CorrelationID: event.CorrelationID,
			Payload:       e.buildLegacyPayload(trigger, event),
			Status:        "pending",
			MaxAttempts:   e.config.MaxRetries,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Save to queue
		if err := e.db.WithContext(ctx).Create(&queuedEvent).Error; err != nil {
			e.logger.WithError(err).Error("Failed to queue event")
			continue
		}

		// Send to event channel
		select {
		case e.eventChan <- queuedEvent:
		default:
			// Channel full, event will be picked up by retry worker
			e.logger.WithField("event_id", queuedEvent.ID).Warn("Event channel full, event queued in DB")
		}
	}

	// Update trigger last triggered time
	now := time.Now()
	e.db.Model(&StateTrigger{}).Where("id = ?", trigger.ID).Update("last_triggered_at", &now)
	e.updateRateLimit(trigger.ID)
}

func (e *TriggerEngine) getPendingEvents(ctx context.Context, trigger StateTrigger) ([]StateEvent, error) {
	var events []StateEvent

	query := e.db.WithContext(ctx).
		Model(&StateEvent{}).
		Where("state_id = ?", trigger.SourceStateID)

	// Filter by key pattern if specified
	if trigger.KeyPattern != nil && *trigger.KeyPattern != "" {
		pattern := *trigger.KeyPattern
		// Simple glob matching - convert * to % for SQL LIKE
		likePattern := strings.ReplaceAll(pattern, "*", "%")
		query = query.Where("key LIKE ?", likePattern)
	}

	// Filter by time (since last triggered)
	if trigger.LastTriggeredAt != nil {
		query = query.Where("timestamp > ?", trigger.LastTriggeredAt)
	}

	err := query.
		Order("timestamp ASC").
		Limit(e.config.BatchSize).
		Find(&events).Error

	return events, err
}

func (e *TriggerEngine) buildLegacyPayload(trigger StateTrigger, event StateEvent) JSONMap {
	payload := JSONMap{
		"trigger_type":   trigger.TriggerType,
		"source_state":   event.StateID.String(),
		"event_type":     event.EventType,
		"correlation_id": event.CorrelationID,
		"timestamp":      event.Timestamp,
	}

	if event.Key != nil {
		payload["key"] = *event.Key
	}

	if trigger.IncludePrevious && event.PreviousValue != nil {
		payload["previous_value"] = *event.PreviousValue
	}
	if trigger.IncludeNew && event.NewValue != nil {
		payload["new_value"] = *event.NewValue
	}

	return payload
}

// checkRateLimit checks if a trigger can be executed based on rate limiting
func (e *TriggerEngine) checkRateLimit(triggerID uuid.UUID) bool {
	e.rateLimitLock.Lock()
	defer e.rateLimitLock.Unlock()

	lastExec, exists := e.lastExecution[triggerID]
	if exists {
		timeSinceLastExec := time.Since(lastExec)
		if timeSinceLastExec < e.minInterval {
			return false
		}
	}
	return true
}

// updateRateLimit updates the last execution time for a trigger
func (e *TriggerEngine) updateRateLimit(triggerID uuid.UUID) {
	e.rateLimitLock.Lock()
	defer e.rateLimitLock.Unlock()
	e.lastExecution[triggerID] = time.Now()
}

// checkTenantRateLimit checks if a tenant can execute triggers based on rate limiting
func (e *TriggerEngine) checkTenantRateLimit(tenantID uuid.UUID) bool {
	e.tenantRateLimitLock.Lock()
	defer e.tenantRateLimitLock.Unlock()

	lastExec, exists := e.tenantLastExecution[tenantID]
	if exists {
		timeSinceLastExec := time.Since(lastExec)
		if timeSinceLastExec < e.tenantMinInterval {
			return false
		}
	}
	return true
}

// updateTenantRateLimit updates the last execution time for a tenant
func (e *TriggerEngine) updateTenantRateLimit(tenantID uuid.UUID) {
	e.tenantRateLimitLock.Lock()
	defer e.tenantRateLimitLock.Unlock()
	e.tenantLastExecution[tenantID] = time.Now()
}

// ProcessStateChange is called when a state value changes to check for triggers
func (e *TriggerEngine) ProcessStateChange(ctx context.Context, stateID uuid.UUID, key string, eventType string, oldValue, newValue *JSONMap) error {
	if !e.config.Enabled {
		return nil
	}

	// Find triggers that match this state change
	var triggers []StateTrigger
	err := e.db.WithContext(ctx).
		Model(&StateTrigger{}).
		Where("source_state_id = ? AND is_active = ?", stateID, true).
		Find(&triggers).Error

	if err != nil {
		return err
	}

	if len(triggers) == 0 {
		return nil
	}

	e.logger.WithFields(logrus.Fields{
		"state_id":      stateID,
		"key":           key,
		"event_type":    eventType,
		"trigger_count": len(triggers),
	}).Debug("Processing state change for triggers")

	for _, trigger := range triggers {
		// Check if trigger type matches
		if !e.matchesTriggerType(trigger.TriggerType, eventType) {
			continue
		}

		// Check if key matches
		if trigger.KeyPattern != nil && !e.matchesKeyPattern(*trigger.KeyPattern, key) {
			continue
		}

		// Check condition if specified
		if !e.evaluateCondition(trigger.Condition, oldValue, newValue) {
			continue
		}

		// Queue the event
		e.queueEventFromStateChange(ctx, trigger, stateID, key, eventType, oldValue, newValue)
	}

	return nil
}

// queueEventFromStateChange creates and queues an event from a state change
func (e *TriggerEngine) queueEventFromStateChange(ctx context.Context, trigger StateTrigger, stateID uuid.UUID, key, eventType string, oldValue, newValue *JSONMap) {
	// Build payload
	payload := JSONMap{
		"trigger_type":  trigger.TriggerType,
		"source_state":  stateID.String(),
		"key":           key,
		"event_type":    eventType,
		"correlationID": uuid.New().String(),
	}

	if trigger.IncludePrevious && oldValue != nil {
		payload["previous_value"] = *oldValue
	}
	if trigger.IncludeNew && newValue != nil {
		payload["new_value"] = *newValue
	}

	// Create queued event
	event := QueuedTriggerEvent{
		ID:            uuid.New(),
		TriggerID:     trigger.ID,
		StateID:       stateID,
		TenantID:      trigger.TenantID,
		Key:           key,
		EventType:     eventType,
		OldValue:      oldValue,
		NewValue:      newValue,
		CorrelationID: payload["correlationID"].(string),
		Payload:       payload,
		Status:        "pending",
		MaxAttempts:   e.config.MaxRetries,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to database queue
	if err := e.db.WithContext(ctx).Create(&event).Error; err != nil {
		e.logger.WithError(err).WithField("trigger_id", trigger.ID).Error("Failed to queue trigger event")
		return
	}

	// Send to event channel for immediate processing
	select {
	case e.eventChan <- event:
		e.logger.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"trigger_id": trigger.ID,
			"state_id":   stateID,
		}).Debug("Queued trigger event for immediate processing")
	default:
		// Channel full, event will be picked up by retry worker
		e.logger.WithField("event_id", event.ID).Warn("Event channel full, event queued in DB for later processing")
	}
}

// matchesTriggerType checks if the event type matches the trigger type
func (e *TriggerEngine) matchesTriggerType(triggerType, eventType string) bool {
	switch triggerType {
	case "on_write":
		return eventType == "set" || eventType == "patch"
	case "on_delete":
		return eventType == "delete"
	case "on_read":
		// Read triggers are handled differently
		return false
	case "on_condition":
		return true
	default:
		return false
	}
}

// matchesKeyPattern checks if a key matches a glob pattern
func (e *TriggerEngine) matchesKeyPattern(pattern, key string) bool {
	// Convert glob to regex
	regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(pattern, ".", `\.`), "*", ".*") + "$"
	return matchRegex(regexPattern, key)
}

// matchRegex matches text against a regex pattern (e.g. from matchesKeyPattern). Returns false if the pattern is invalid.
func matchRegex(pattern, text string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(text)
}

// evaluateCondition evaluates a trigger condition
func (e *TriggerEngine) evaluateCondition(condition JSONMap, oldValue, newValue *JSONMap) bool {
	if condition == nil || len(condition) == 0 {
		return true
	}

	// Simple condition evaluation - check for field changes
	if changeField, ok := condition["changed_field"].(string); ok {
		oldVal := getNestedValue(oldValue, changeField)
		newVal := getNestedValue(newValue, changeField)
		return oldVal != newVal
	}

	// Check for value thresholds
	if threshold, ok := condition["threshold"].(map[string]interface{}); ok {
		if field, ok := threshold["field"].(string); ok {
			if newVal := getNestedValue(newValue, field); newVal != nil {
				if num, ok := newVal.(float64); ok {
					if min, ok := threshold["min"].(float64); ok && num < min {
						return false
					}
					if max, ok := threshold["max"].(float64); ok && num > max {
						return false
					}
				}
			}
		}
	}

	return true
}

// getNestedValue gets a nested value from a map using dot notation
func getNestedValue(m *JSONMap, path string) interface{} {
	if m == nil {
		return nil
	}

	parts := strings.Split(path, ".")
	current := interface{}(*m)

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// GetStats returns trigger engine statistics
func (e *TriggerEngine) GetStats() map[string]interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()

	stats := map[string]interface{}{
		"workerCount":     e.workerCount,
		"executingCount":  len(e.executing),
		"enabled":         e.config.Enabled,
		"pollInterval":    e.config.PollInterval,
		"maxConcurrent":   e.config.MaxConcurrent,
		"maxRetries":      e.config.MaxRetries,
		"tenantRateLimit": e.config.TenantRateLimit,
	}

	return stats
}

// GetTriggerStats returns statistics for all triggers
func (e *TriggerEngine) GetTriggerStats(ctx context.Context) (map[string]interface{}, error) {
	var totalTriggers, activeTriggers, inactiveTriggers int64

	e.db.Model(&StateTrigger{}).Count(&totalTriggers)
	e.db.Model(&StateTrigger{}).Where("is_active = ?", true).Count(&activeTriggers)
	e.db.Model(&StateTrigger{}).Where("is_active = ?", false).Count(&inactiveTriggers)

	// Get queue stats
	var pendingCount, processingCount, failedCount, deadLetterCount int64
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "pending").Count(&pendingCount)
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "processing").Count(&processingCount)
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "failed").Count(&failedCount)
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "dead_letter").Count(&deadLetterCount)

	// Get dead letter stats
	var dlqCount int64
	e.db.WithContext(ctx).Model(&TriggerDeadLetter{}).Count(&dlqCount)

	stats := map[string]interface{}{
		"totalTriggers":    totalTriggers,
		"activeTriggers":   activeTriggers,
		"inactiveTriggers": inactiveTriggers,
		"pendingEvents":    pendingCount,
		"processingEvents": processingCount,
		"failedEvents":     failedCount,
		"deadLetterEvents": deadLetterCount,
		"dlqEntries":       dlqCount,
	}

	return stats, nil
}

// GetQueueStats returns detailed queue statistics
func (e *TriggerEngine) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Status counts
	var counts struct {
		Pending    int64
		Processing int64
		Completed  int64
		Failed     int64
		DeadLetter int64
	}

	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "pending").Count(&counts.Pending)
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "processing").Count(&counts.Processing)
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "completed").Count(&counts.Completed)
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "failed").Count(&counts.Failed)
	e.db.WithContext(ctx).Model(&QueuedTriggerEvent{}).Where("status = ?", "dead_letter").Count(&counts.DeadLetter)

	stats["byStatus"] = counts

	// Average execution time
	var avgDuration float64
	e.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(duration_ms), 0)
		FROM trigger_execution_logs
		WHERE created_at > NOW() - INTERVAL '1 hour'
	`).Scan(&avgDuration)
	stats["avgExecutionDurationMs"] = avgDuration

	// Events by tenant (top 10)
	var tenantStats []struct {
		TenantID     uuid.UUID `gorm:"column:tenant_id"`
		PendingCount int64     `gorm:"column:pending_count"`
	}
	e.db.WithContext(ctx).Raw(`
		SELECT tenant_id, COUNT(*) as pending_count
		FROM trigger_event_queue
		WHERE status = 'pending'
		GROUP BY tenant_id
		ORDER BY pending_count DESC
		LIMIT 10
	`).Scan(&tenantStats)
	stats["topTenantsByPending"] = tenantStats

	return stats, nil
}

// RetryDeadLetterEvent attempts to retry a dead letter event
func (e *TriggerEngine) RetryDeadLetterEvent(ctx context.Context, dlqID uuid.UUID) error {
	var dlq TriggerDeadLetter
	if err := e.db.WithContext(ctx).First(&dlq, "id = ?", dlqID).Error; err != nil {
		return fmt.Errorf("dead letter entry not found: %w", err)
	}

	if !dlq.CanRetry {
		return fmt.Errorf("dead letter entry is marked as non-retryable")
	}

	// Create new queued event from dead letter
	event := QueuedTriggerEvent{
		ID:            uuid.New(),
		TriggerID:     dlq.TriggerID,
		StateID:       dlq.StateID,
		TenantID:      dlq.TenantID,
		Key:           dlq.Key,
		EventType:     dlq.EventType,
		Payload:       dlq.Payload,
		CorrelationID: dlq.CorrelationID,
		Status:        "pending",
		MaxAttempts:   e.config.MaxRetries,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := e.db.WithContext(ctx).Create(&event).Error; err != nil {
		return fmt.Errorf("failed to create retry event: %w", err)
	}

	// Update dead letter entry
	now := time.Now()
	dlq.RetriedAt = &now
	dlq.RetriedSuccess = false // Will be updated when processed
	if err := e.db.WithContext(ctx).Save(&dlq).Error; err != nil {
		e.logger.WithError(err).Warn("Failed to update dead letter entry after retry")
	}

	// Send to event channel
	select {
	case e.eventChan <- event:
	default:
		// Channel full, event will be picked up by retry worker
	}

	return nil
}

// PurgeDeadLetter removes old dead letter entries
func (e *TriggerEngine) PurgeDeadLetter(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	result := e.db.WithContext(ctx).Where("created_at < ? AND (retried_success = ? OR can_retry = ?)", cutoff, true, false).Delete(&TriggerDeadLetter{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// PurgeCompletedEvents removes old completed queue entries
func (e *TriggerEngine) PurgeCompletedEvents(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	result := e.db.WithContext(ctx).Where("status = ? AND completed_at < ?", "completed", cutoff).Delete(&QueuedTriggerEvent{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
