package state

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TriggerEvent represents an event that can trigger a function
type TriggerEvent struct {
	TriggerID     uuid.UUID `json:"trigger_id"`
	StateID       uuid.UUID `json:"state_id"`
	Key           string    `json:"key"`
	EventType     string    `json:"event_type"` // "set", "delete", "patch"
	OldValue      *JSONMap  `json:"old_value,omitempty"`
	NewValue      *JSONMap  `json:"new_value,omitempty"`
	CorrelationID string    `json:"correlation_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// TriggerExecutionResult represents the result of a trigger execution
type TriggerExecutionResult struct {
	TriggerID    uuid.UUID `json:"trigger_id"`
	Success      bool      `json:"success"`
	Output       *JSONMap  `json:"output,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	DurationMs   int64     `json:"duration_ms"`
	ExecutedAt   time.Time `json:"executed_at"`
}

// TriggerExecutor is an interface for executing trigger functions
type TriggerExecutor interface {
	ExecuteFunction(ctx context.Context, functionID uuid.UUID, payload JSONMap) (*TriggerExecutionResult, error)
}

// HTTPTriggerExecutor executes functions via HTTP
type HTTPTriggerExecutor struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	logger     *logrus.Logger
}

// NewHTTPTriggerExecutor creates a new HTTP trigger executor
func NewHTTPTriggerExecutor(baseURL, apiKey string, logger *logrus.Logger) *HTTPTriggerExecutor {
	return &HTTPTriggerExecutor{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
		logger: logger,
	}
}

// ExecuteFunction executes a function via HTTP
func (e *HTTPTriggerExecutor) ExecuteFunction(ctx context.Context, functionID uuid.UUID, payload JSONMap) (*TriggerExecutionResult, error) {
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
		}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return &TriggerExecutionResult{
			TriggerID:    functionID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			DurationMs:   time.Since(start).Milliseconds(),
			ExecutedAt:   time.Now(),
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
		}, err
	}
	defer resp.Body.Close()

	var result TriggerExecutionResult
	result.TriggerID = functionID
	result.DurationMs = time.Since(start).Milliseconds()
	result.ExecutedAt = time.Now()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		var output JSONMap
		if err := json.NewDecoder(resp.Body).Decode(&output); err == nil {
			result.Output = &output
		}
	} else {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return &result, nil
}

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
}

// DefaultTriggerEngineConfig returns the default configuration
func DefaultTriggerEngineConfig() TriggerEngineConfig {
	return TriggerEngineConfig{
		PollInterval:  5 * time.Second,
		MaxConcurrent: 10,
		BatchSize:     100,
		Enabled:       true,
	}
}

// TriggerEngine processes state triggers
type TriggerEngine struct {
	db       *gorm.DB
	config   TriggerEngineConfig
	executor TriggerExecutor
	logger   *logrus.Logger

	// Rate limiting
	rateLimitLock sync.Mutex
	lastExecution map[uuid.UUID]time.Time
	minInterval   time.Duration

	// Execution tracking
	mu          sync.Mutex
	executing   map[uuid.UUID]bool
	workerCount int
	stopChan    chan struct{}
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

	return &TriggerEngine{
		db:            db,
		config:        config,
		executor:      executor,
		logger:        logger,
		lastExecution: make(map[uuid.UUID]time.Time),
		minInterval:   time.Minute, // Minimum 1 minute between same trigger
		executing:     make(map[uuid.UUID]bool),
		workerCount:   0,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the trigger engine
func (e *TriggerEngine) Start(ctx context.Context) {
	if !e.config.Enabled {
		e.logger.Info("Trigger engine is disabled")
		return
	}

	e.logger.WithFields(logrus.Fields{
		"pollInterval":  e.config.PollInterval,
		"maxConcurrent": e.config.MaxConcurrent,
	}).Info("Starting trigger engine")

	// Start worker pool
	for i := 0; i < e.config.MaxConcurrent; i++ {
		go e.worker(ctx)
	}

	e.logger.Info("Trigger engine started")
}

// Stop stops the trigger engine
func (e *TriggerEngine) Stop() {
	e.logger.Info("Stopping trigger engine")
	close(e.stopChan)
}

// worker is a single worker that processes trigger events
func (e *TriggerEngine) worker(ctx context.Context) {
	e.mu.Lock()
	e.workerCount++
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.workerCount--
		e.mu.Unlock()
	}()

	ticker := time.NewTicker(e.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.processTriggers(ctx)
		}
	}
}

// processTriggers finds and processes pending triggers
func (e *TriggerEngine) processTriggers(ctx context.Context) {
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
		go e.executeTrigger(ctx, trigger)
	}
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

// executeTrigger executes a trigger's target function
func (e *TriggerEngine) executeTrigger(ctx context.Context, trigger StateTrigger) {
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

	// Build payload
	payload := e.buildPayload(trigger, events)

	// Get target function
	if trigger.TargetFunctionID == nil {
		e.logger.Warn("Trigger has no target function")
		return
	}

	// Execute function
	result, err := e.executor.ExecuteFunction(ctx, *trigger.TargetFunctionID, payload)
	if err != nil {
		e.logger.WithError(err).Error("Failed to execute trigger function")
	}

	// Update trigger last triggered time
	now := time.Now()
	e.db.Model(&StateTrigger{}).Where("id = ?", trigger.ID).Update("last_triggered_at", &now)

	// Update rate limit
	e.updateRateLimit(trigger.ID)

	// Log result
	if result != nil {
		e.logger.WithFields(logrus.Fields{
			"triggerID":    trigger.ID,
			"success":      result.Success,
			"durationMs":   result.DurationMs,
			"errorMessage": result.ErrorMessage,
		}).Info("Trigger executed")
	}
}

// getPendingEvents gets events that have occurred since the last trigger
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

// buildPayload builds the payload for the trigger function
func (e *TriggerEngine) buildPayload(trigger StateTrigger, events []StateEvent) JSONMap {
	payload := JSONMap{
		"trigger_type": trigger.TriggerType,
		"source_state": trigger.SourceStateID.String(),
		"events":       []JSONMap{},
	}

	eventsList := make([]JSONMap, len(events))
	for i, event := range events {
		eventMap := JSONMap{
			"id":            event.ID.String(),
			"event_type":    event.EventType,
			"key":           event.Key,
			"timestamp":     event.Timestamp,
			"correlationID": event.CorrelationID,
		}

		// Include previous value if requested
		if trigger.IncludePrevious && event.PreviousValue != nil {
			eventMap["previous_value"] = *event.PreviousValue
		}

		// Include new value if requested
		if trigger.IncludeNew && event.NewValue != nil {
			eventMap["new_value"] = *event.NewValue
		}

		eventsList[i] = eventMap
	}

	payload["events"] = eventsList
	return payload
}

// ProcessStateChange is called when a state value changes to check for triggers
func (e *TriggerEngine) ProcessStateChange(ctx context.Context, stateID uuid.UUID, key string, eventType string, oldValue, newValue *JSONMap) error {
	// Find triggers that match this state change
	var triggers []StateTrigger
	err := e.db.WithContext(ctx).
		Model(&StateTrigger{}).
		Where("source_state_id = ? AND is_active = ?", stateID, true).
		Find(&triggers).Error

	if err != nil {
		return err
	}

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

		// Queue trigger for execution
		go e.queueTrigger(ctx, trigger, stateID, key, eventType, oldValue, newValue)
	}

	return nil
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

// matchRegex is a simple regex matcher
func matchRegex(pattern, text string) bool {
	// Simple implementation - in production use regexp
	pattern = strings.TrimPrefix(pattern, "^")
	pattern = strings.TrimSuffix(pattern, "$")

	// Handle * wildcard
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, ".*") {
		parts := strings.Split(pattern, ".*")
		if len(parts) == 2 {
			return strings.HasPrefix(text, parts[0]) && strings.HasSuffix(text, parts[1])
		}
	}

	return pattern == text
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

// queueTrigger queues a trigger for execution
func (e *TriggerEngine) queueTrigger(ctx context.Context, trigger StateTrigger, stateID uuid.UUID, key, eventType string, oldValue, newValue *JSONMap) {
	// Check rate limit
	if !e.checkRateLimit(trigger.ID) {
		return
	}

	// Check if already executing
	e.mu.Lock()
	if e.executing[trigger.ID] {
		e.mu.Unlock()
		return
	}
	e.executing[trigger.ID] = true
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.executing, trigger.ID)
		e.mu.Unlock()
	}()

	// Build payload
	payload := JSONMap{
		"trigger_type":  trigger.TriggerType,
		"source_state":  stateID.String(),
		"key":           key,
		"event_type":    eventType,
		"correlationID": uuid.New().String(),
	}

	if trigger.IncludePrevious {
		payload["previous_value"] = oldValue
	}
	if trigger.IncludeNew {
		payload["new_value"] = newValue
	}

	// Execute
	if trigger.TargetFunctionID != nil {
		result, err := e.executor.ExecuteFunction(ctx, *trigger.TargetFunctionID, payload)
		if err != nil {
			e.logger.WithError(err).Error("Failed to execute trigger")
		}

		// Update last triggered
		now := time.Now()
		e.db.Model(&StateTrigger{}).Where("id = ?", trigger.ID).Update("last_triggered_at", &now)
		e.updateRateLimit(trigger.ID)

		if result != nil {
			e.logger.WithFields(logrus.Fields{
				"triggerID":  trigger.ID,
				"success":    result.Success,
				"durationMs": result.DurationMs,
			}).Info("Trigger executed")
		}
	}
}

// GetStats returns trigger engine statistics
func (e *TriggerEngine) GetStats() map[string]interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()

	stats := map[string]interface{}{
		"workerCount":    e.workerCount,
		"executingCount": len(e.executing),
		"enabled":        e.config.Enabled,
		"pollInterval":   e.config.PollInterval,
	}

	return stats
}

// GetTriggerStats returns statistics for all triggers
func (e *TriggerEngine) GetTriggerStats(ctx context.Context) (map[string]interface{}, error) {
	var totalTriggers, activeTriggers, inactiveTriggers int64

	e.db.Model(&StateTrigger{}).Count(&totalTriggers)
	e.db.Model(&StateTrigger{}).Where("is_active = ?", true).Count(&activeTriggers)
	e.db.Model(&StateTrigger{}).Where("is_active = ?", false).Count(&inactiveTriggers)

	stats := map[string]interface{}{
		"totalTriggers":    totalTriggers,
		"activeTriggers":   activeTriggers,
		"inactiveTriggers": inactiveTriggers,
	}

	return stats, nil
}
