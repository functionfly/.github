package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/captcha"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ExecutionSecurityConfig holds configuration for execution security features
type ExecutionSecurityConfig struct {
	// Rate limiting
	DefaultDailyLimit   int     `json:"default_daily_limit"`
	DefaultHourlyLimit  int     `json:"default_hourly_limit"`
	DefaultMinuteLimit  int     `json:"default_minute_limit"`
	AnonymousMultiplier float64 `json:"anonymous_multiplier"` // Higher limits for anonymous users

	// Throttling
	ThrottleMultiplier  float64       `json:"throttle_multiplier"`
	MaxThrottleDuration time.Duration `json:"max_throttle_duration"`
	BlockThreshold      int           `json:"block_threshold"` // Suspicious activity score threshold for blocking

	// CAPTCHA
	CaptchaRequiredScore  int           `json:"captcha_required_score"`
	CaptchaValidityWindow time.Duration `json:"captcha_validity_window"`

	// Abuse detection
	AbuseDetectionEnabled bool    `json:"abuse_detection_enabled"`
	ErrorRateThreshold    float64 `json:"error_rate_threshold"`
	RateSpikeThreshold    float64 `json:"rate_spike_threshold"`
	SuspiciousInputScore  int     `json:"suspicious_input_score"`

	// Resource limits
	DefaultMaxMemoryMB  int `json:"default_max_memory_mb"`
	DefaultMaxCPUTimeMs int `json:"default_max_cpu_time_ms"`
	DefaultTimeoutMs    int `json:"default_timeout_ms"`

	// Input validation
	InputValidationEnabled bool `json:"input_validation_enabled"`
	StrictSchemaValidation bool `json:"strict_schema_validation"`
}

// ExecutionSecurityMiddleware provides comprehensive security for function executions
type ExecutionSecurityMiddleware struct {
	db             *gorm.DB
	config         *ExecutionSecurityConfig
	captchaService *captcha.CaptchaService
	logger         *logrus.Logger
}

// getUserIDFromContext extracts user ID from request context (set by auth middleware)
func getUserIDFromContext(r *http.Request) *uuid.UUID {
	claims := GetUserFromContext(r)
	if claims == nil {
		return nil
	}
	id := claims.UserID
	return &id
}

// NewExecutionSecurityMiddleware creates a new execution security middleware
func NewExecutionSecurityMiddleware(db *gorm.DB, config *ExecutionSecurityConfig, captchaService *captcha.CaptchaService) *ExecutionSecurityMiddleware {
	if config == nil {
		config = &ExecutionSecurityConfig{
			DefaultDailyLimit:      1000,
			DefaultHourlyLimit:     100,
			DefaultMinuteLimit:     10,
			AnonymousMultiplier:    0.1, // Anonymous users get 10% of normal limits
			ThrottleMultiplier:     2.0,
			MaxThrottleDuration:    time.Hour * 24,
			BlockThreshold:         100,
			CaptchaRequiredScore:   50,
			CaptchaValidityWindow:  time.Hour * 24,
			AbuseDetectionEnabled:  true,
			ErrorRateThreshold:     0.8,
			RateSpikeThreshold:     5.0,
			SuspiciousInputScore:   25,
			DefaultMaxMemoryMB:     128,
			DefaultMaxCPUTimeMs:    30000,
			DefaultTimeoutMs:       30000,
			InputValidationEnabled: true,
			StrictSchemaValidation: false,
		}
	}

	if captchaService == nil {
		captchaService = captcha.NewCaptchaService(logrus.New())
	}

	return &ExecutionSecurityMiddleware{
		db:             db,
		config:         config,
		captchaService: captchaService,
		logger:         logrus.New(),
	}
}

// RateLimitAndQuotaMiddleware enforces per-user execution quotas and rate limits
func (esm *ExecutionSecurityMiddleware) RateLimitAndQuotaMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Bypass rate limiting in development mode (never in production)
		if (os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development") && os.Getenv("PRODUCTION_ENV") != "true" {
			next.ServeHTTP(w, r)
			return
		}

		userID := getUserIDFromContext(r)
		ipAddress := getClientIP(r)

		// Get or create quota record
		quota, err := esm.getOrCreateQuota(userID, ipAddress)
		if err != nil {
			esm.logger.WithError(err).Error("Failed to get/create quota record")
			esm.respondRateLimited(w, "Internal error")
			return
		}

		// Update counters and check limits
		if err := esm.updateAndCheckQuota(quota); err != nil {
			esm.logger.WithFields(logrus.Fields{
				"user_id": userID,
				"ip":      ipAddress,
				"error":   err.Error(),
			}).Warn("Quota check failed")

			// Record security event
			esm.recordSecurityEvent(userID, &ipAddress, "quota_exceeded", "warning", err.Error(), nil)

			esm.respondRateLimited(w, err.Error())
			return
		}

		// Check for progressive throttling
		if quota.IsThrottled {
			if time.Now().Before(*quota.ThrottleUntil) {
				esm.respondRateLimited(w, "Temporarily throttled due to suspicious activity")
				return
			}
			// Throttle period expired, reset
			quota.IsThrottled = false
			quota.ThrottleUntil = nil
			esm.db.Save(quota)
		}

		// Check for blocking
		if quota.BlockUntil != nil && time.Now().Before(*quota.BlockUntil) {
			esm.respondBlocked(w, "Account temporarily blocked due to abuse")
			return
		}

		next.ServeHTTP(w, r)
	}
}

// CaptchaRequiredMiddleware checks if CAPTCHA is required for anonymous executions
func (esm *ExecutionSecurityMiddleware) CaptchaRequiredMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Bypass CAPTCHA check in development mode (never in production)
		if os.Getenv("DEVELOPMENT") == "true" && os.Getenv("PRODUCTION_ENV") != "true" {
			next.ServeHTTP(w, r)
			return
		}

		userID := getUserIDFromContext(r)

		// Skip CAPTCHA check for authenticated users
		if userID != nil {
			next.ServeHTTP(w, r)
			return
		}

		ipAddress := getClientIP(r)
		quota, err := esm.getOrCreateQuota(userID, ipAddress)
		if err != nil {
			esm.logger.WithError(err).Error("Failed to get quota for CAPTCHA check")
			esm.respondError(w, http.StatusInternalServerError, "Internal error")
			return
		}

		// Check if CAPTCHA is required
		if quota.CaptchaRequired {
			// Validate CAPTCHA token from request
			captchaToken := r.Header.Get("X-Captcha-Token")
			if captchaToken == "" {
				captchaToken = r.URL.Query().Get("captcha_token")
			}

			if captchaToken == "" {
				esm.respondCaptchaRequired(w, "CAPTCHA verification required for anonymous executions")
				return
			}

			// Validate CAPTCHA token
			if !esm.validateCaptchaToken(captchaToken) {
				quota.SuspiciousActivityScore += 10
				esm.db.Save(quota)

				esm.recordSecurityEvent(nil, &ipAddress, "captcha_failed", "warning", "Invalid CAPTCHA token", nil)
				esm.respondCaptchaRequired(w, "Invalid CAPTCHA token")
				return
			}

			// CAPTCHA passed, reduce suspicious score
			if quota.SuspiciousActivityScore > 0 {
				quota.SuspiciousActivityScore = int(math.Max(0, float64(quota.SuspiciousActivityScore-5)))
				now := time.Now()
				quota.LastCaptchaCompleted = &now
				esm.db.Save(quota)
			}
		}

		next.ServeHTTP(w, r)
	}
}

// AbuseDetectionMiddleware monitors for abuse patterns and applies progressive throttling
func (esm *ExecutionSecurityMiddleware) AbuseDetectionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !esm.config.AbuseDetectionEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Bypass abuse detection in development mode (never in production)
		if os.Getenv("DEVELOPMENT") == "true" && os.Getenv("PRODUCTION_ENV") != "true" {
			next.ServeHTTP(w, r)
			return
		}

		userID := getUserIDFromContext(r)
		ipAddress := getClientIP(r)

		// Analyze request patterns
		if patterns := esm.detectAbusePatterns(r, userID, ipAddress); len(patterns) > 0 {
			for _, pattern := range patterns {
				esm.handleAbusePattern(pattern, userID, ipAddress)
			}
		}

		next.ServeHTTP(w, r)
	}
}

// InputValidationMiddleware validates function inputs against schemas
func (esm *ExecutionSecurityMiddleware) InputValidationMiddleware(functionID uuid.UUID, version string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !esm.config.InputValidationEnabled {
				next.ServeHTTP(w, r)
				return
			}

			// Parse input from request body
			var inputData interface{}
			if err := json.NewDecoder(r.Body).Decode(&inputData); err != nil {
				esm.recordSecurityEvent(nil, nil, "input_validation_failed", "warning", "Invalid JSON input", nil)
				esm.respondError(w, http.StatusBadRequest, "Invalid JSON input")
				return
			}

			// Restore request body for next handler
			r.Body = &jsonReader{data: inputData}

			// Get function input schema
			schema, err := esm.getFunctionInputSchema(functionID, version)
			if err != nil {
				esm.logger.WithError(err).Warn("Failed to get input schema, skipping validation")
				next.ServeHTTP(w, r)
				return
			}

			// Validate input against schema
			if err := esm.validateInputAgainstSchema(inputData, schema); err != nil {
				userID := getUserIDFromContext(r)
				ipAddress := getClientIP(r)
				esm.recordSecurityEvent(userID, &ipAddress, "input_validation_failed", "warning", err.Error(), map[string]interface{}{
					"function_id": functionID,
					"version":     version,
					"input_size":  len(fmt.Sprintf("%v", inputData)),
				})

				if schema.IsStrict {
					esm.respondError(w, http.StatusBadRequest, fmt.Sprintf("Input validation failed: %s", err.Error()))
					return
				}
				// Log but allow for non-strict validation
				esm.logger.WithFields(logrus.Fields{
					"function_id": functionID,
					"version":     version,
					"error":       err.Error(),
				}).Warn("Input validation failed but allowing due to non-strict mode")
			}

			next.ServeHTTP(w, r)
		}
	}
}

// ResourceLimitsMiddleware enforces memory/CPU limits per execution
func (esm *ExecutionSecurityMiddleware) ResourceLimitsMiddleware(functionID uuid.UUID, version string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get function resource limits
			maxMemoryMB := esm.config.DefaultMaxMemoryMB
			maxCPUTimeMs := esm.config.DefaultMaxCPUTimeMs
			timeoutMs := esm.config.DefaultTimeoutMs

			// Try to get function-specific limits
			if limits := esm.getFunctionResourceLimits(functionID, version); limits != nil {
				maxMemoryMB = limits.MemoryMB
				maxCPUTimeMs = limits.TimeoutMs
				if limits.TimeoutMs > 0 {
					timeoutMs = limits.TimeoutMs
				}
			}

			// Set resource limits in context for execution
			ctx := r.Context()
			ctx = contextWithResourceLimits(ctx, maxMemoryMB, maxCPUTimeMs, timeoutMs)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}
	}
}

// Helper methods

func (esm *ExecutionSecurityMiddleware) getOrCreateQuota(userID *uuid.UUID, ipAddress string) (*storage.UserExecutionQuota, error) {
	var quota storage.UserExecutionQuota

	query := esm.db.Where("ip_address = ?", ipAddress)
	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}

	err := query.First(&quota).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new quota record
			quota = storage.UserExecutionQuota{
				UserID:               userID,
				IPAddress:            ipAddress,
				DailyExecutionLimit:  esm.config.DefaultDailyLimit,
				HourlyExecutionLimit: esm.config.DefaultHourlyLimit,
				MinuteExecutionLimit: esm.config.DefaultMinuteLimit,
				DailyResetAt:         time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour),
				HourlyResetAt:        time.Now().Add(time.Hour).Truncate(time.Hour),
				MinuteResetAt:        time.Now().Add(time.Minute).Truncate(time.Minute),
			}

			// Apply anonymous multiplier
			if userID == nil {
				quota.DailyExecutionLimit = int(float64(quota.DailyExecutionLimit) * esm.config.AnonymousMultiplier)
				quota.HourlyExecutionLimit = int(float64(quota.HourlyExecutionLimit) * esm.config.AnonymousMultiplier)
				quota.MinuteExecutionLimit = int(float64(quota.MinuteExecutionLimit) * esm.config.AnonymousMultiplier)
			}

			if err := esm.db.Create(&quota).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Reset counters if needed
	now := time.Now()
	if now.After(quota.DailyResetAt) {
		quota.DailyExecutions = 0
		quota.DailyResetAt = now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	}
	if now.After(quota.HourlyResetAt) {
		quota.HourlyExecutions = 0
		quota.HourlyResetAt = now.Add(time.Hour).Truncate(time.Hour)
	}
	if now.After(quota.MinuteResetAt) {
		quota.MinuteExecutions = 0
		quota.MinuteResetAt = now.Add(time.Minute).Truncate(time.Minute)
	}

	esm.db.Save(&quota)
	return &quota, nil
}

func (esm *ExecutionSecurityMiddleware) updateAndCheckQuota(quota *storage.UserExecutionQuota) error {
	now := time.Now()

	// Increment counters
	quota.DailyExecutions++
	quota.HourlyExecutions++
	quota.MinuteExecutions++

	// Check limits
	if quota.DailyExecutions > quota.DailyExecutionLimit {
		return fmt.Errorf("daily execution limit exceeded (%d/%d)", quota.DailyExecutions, quota.DailyExecutionLimit)
	}
	if quota.HourlyExecutions > quota.HourlyExecutionLimit {
		return fmt.Errorf("hourly execution limit exceeded (%d/%d)", quota.HourlyExecutions, quota.HourlyExecutionLimit)
	}
	if quota.MinuteExecutions > quota.MinuteExecutionLimit {
		return fmt.Errorf("per-minute execution limit exceeded (%d/%d)", quota.MinuteExecutions, quota.MinuteExecutionLimit)
	}

	// Check if CAPTCHA should be required
	if quota.SuspiciousActivityScore >= esm.config.CaptchaRequiredScore {
		quota.CaptchaRequired = true
	}

	// Apply progressive throttling if suspicious score is high
	if quota.SuspiciousActivityScore > 30 && !quota.IsThrottled {
		throttleDuration := time.Duration(quota.SuspiciousActivityScore) * time.Minute
		if throttleDuration > esm.config.MaxThrottleDuration {
			throttleDuration = esm.config.MaxThrottleDuration
		}
		throttleUntil := now.Add(throttleDuration)
		quota.IsThrottled = true
		quota.ThrottleUntil = &throttleUntil
	}

	// Apply blocking if score is very high
	if quota.SuspiciousActivityScore >= esm.config.BlockThreshold && quota.BlockUntil == nil {
		blockDuration := time.Duration(quota.SuspiciousActivityScore/10) * time.Hour
		if blockDuration > esm.config.MaxThrottleDuration {
			blockDuration = esm.config.MaxThrottleDuration
		}
		blockUntil := now.Add(blockDuration)
		quota.BlockUntil = &blockUntil
	}

	esm.db.Save(quota)
	return nil
}

func (esm *ExecutionSecurityMiddleware) detectAbusePatterns(r *http.Request, userID *uuid.UUID, ipAddress string) []*storage.AbusePattern {
	var patterns []*storage.AbusePattern

	// Check for suspicious input patterns
	body, _ := json.Marshal(r.Body)
	bodyStr := string(body)
	if esm.isSuspiciousInput(bodyStr) {
		patterns = append(patterns, &storage.AbusePattern{
			PatternType: "suspicious_input",
			Severity:    "medium",
			UserID:      userID,
			IPAddress:   ipAddress,
			Description: "Detected potentially malicious input patterns",
			PatternData: json.RawMessage(fmt.Sprintf(`{"input_length": %d, "suspicious_patterns": ["script_tags", "sql_injection"]} `, len(bodyStr))),
		})
	}

	// Check for rate spikes using historical execution data
	spikePatterns, err := esm.detectRateSpikes(r, userID, ipAddress)
	if err != nil {
		esm.logger.WithError(err).Warn("Failed to analyze rate spikes")
	} else {
		patterns = append(patterns, spikePatterns...)
	}

	return patterns
}

func (esm *ExecutionSecurityMiddleware) detectRateSpikes(r *http.Request, userID *uuid.UUID, ipAddress string) ([]*storage.AbusePattern, error) {
	var patterns []*storage.AbusePattern

	// Check for user-based rate spikes
	if userID != nil {
		// Check 5-minute rate spike (unusual burst of activity)
		since5Min := time.Now().Add(-5 * time.Minute)
		var count5Min int64
		if err := esm.db.Model(&storage.RegistryFunctionExecution{}).
			Where("user_id = ? AND timestamp >= ?", *userID, since5Min).
			Count(&count5Min).Error; err == nil && count5Min > 50 { // More than 50 executions in 5 minutes
			ratePerMinute := float64(count5Min) / 5.0
			patterns = append(patterns, &storage.AbusePattern{
				PatternType: "rate_spike_user_5min",
				Severity:    "high",
				UserID:      userID,
				IPAddress:   ipAddress,
				Description: fmt.Sprintf("User exceeded 50 executions in 5 minutes (%.1f/min)", ratePerMinute),
				PatternData: json.RawMessage(fmt.Sprintf(`{"execution_count": %d, "rate_per_minute": %.2f, "window_minutes": 5}`, count5Min, ratePerMinute)),
			})
		}

		// Check for execution spikes compared to 24-hour baseline
		since24h := time.Now().Add(-24 * time.Hour)
		since10Min := time.Now().Add(-10 * time.Minute)

		var baselineCount int64
		if err := esm.db.Model(&storage.RegistryFunctionExecution{}).
			Where("user_id = ? AND timestamp >= ?", *userID, since24h).
			Count(&baselineCount).Error; err == nil && baselineCount > 0 {

			var recentCount int64
			if err := esm.db.Model(&storage.RegistryFunctionExecution{}).
				Where("user_id = ? AND timestamp >= ?", *userID, since10Min).
				Count(&recentCount).Error; err == nil {

				baselineRate := float64(baselineCount) / (24 * 60) // executions per minute over 24 hours
				recentRate := float64(recentCount) / 10.0          // executions per minute over 10 minutes

				if recentRate > baselineRate*3.0 { // 3x baseline rate
					patterns = append(patterns, &storage.AbusePattern{
						PatternType: "execution_spike_baseline",
						Severity:    "medium",
						UserID:      userID,
						IPAddress:   ipAddress,
						Description: fmt.Sprintf("Execution rate 3x above 24-hour baseline (%.1f/min vs %.3f/min baseline)", recentRate, baselineRate),
						PatternData: json.RawMessage(fmt.Sprintf(`{"recent_rate": %.2f, "baseline_rate": %.3f, "threshold_multiplier": 3.0, "recent_window_minutes": 10, "baseline_window_hours": 24}`, recentRate, baselineRate)),
					})
				}
			}
		}
	}

	// Check for IP-based rate spikes (for anonymous users or additional protection)
	since5Min := time.Now().Add(-5 * time.Minute)
	var ipCount5Min int64
	if err := esm.db.Model(&storage.RegistryFunctionExecution{}).
		Where("caller_ip = ? AND timestamp >= ?", ipAddress, since5Min).
		Count(&ipCount5Min).Error; err == nil && ipCount5Min > 100 { // More than 100 executions from same IP in 5 minutes
		ratePerMinute := float64(ipCount5Min) / 5.0
		patterns = append(patterns, &storage.AbusePattern{
			PatternType: "rate_spike_ip_5min",
			Severity:    "high",
			UserID:      userID,
			IPAddress:   ipAddress,
			Description: fmt.Sprintf("IP exceeded 100 executions in 5 minutes (%.1f/min)", ratePerMinute),
			PatternData: json.RawMessage(fmt.Sprintf(`{"execution_count": %d, "rate_per_minute": %.2f, "window_minutes": 5}`, ipCount5Min, ratePerMinute)),
		})
	}

	return patterns, nil
}

func (esm *ExecutionSecurityMiddleware) handleAbusePattern(pattern *storage.AbusePattern, userID *uuid.UUID, ipAddress string) {
	// Record the abuse pattern
	if err := esm.db.Create(pattern).Error; err != nil {
		esm.logger.WithError(err).Error("Failed to record abuse pattern")
		return
	}

	// Update user quota with suspicious activity
	quota, err := esm.getOrCreateQuota(userID, ipAddress)
	if err != nil {
		esm.logger.WithError(err).Error("Failed to get quota for abuse handling")
		return
	}

	// Increase suspicious score based on pattern severity
	scoreIncrease := 10
	switch pattern.Severity {
	case "low":
		scoreIncrease = 5
	case "medium":
		scoreIncrease = 15
	case "high":
		scoreIncrease = 30
	case "critical":
		scoreIncrease = 50
	}

	quota.SuspiciousActivityScore += scoreIncrease
	now := time.Now()
	quota.LastSuspiciousActivity = &now

	esm.db.Save(quota)

	// Take action based on severity
	switch pattern.Severity {
	case "high", "critical":
		esm.logger.WithFields(logrus.Fields{
			"user_id":      userID,
			"ip":           ipAddress,
			"pattern_type": pattern.PatternType,
			"severity":     pattern.Severity,
		}).Warn("High-severity abuse pattern detected, applying restrictions")
	}
}

func (esm *ExecutionSecurityMiddleware) validateCaptchaToken(token string) bool {
	if strings.TrimSpace(token) == "" {
		return false
	}

	// Use CAPTCHA service for validation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := esm.captchaService.VerifyToken(ctx, token, "", "")
	if err != nil {
		esm.logger.WithError(err).Error("CAPTCHA verification error")
		return false
	}

	// For reCAPTCHA v3, check score threshold
	if result.Score > 0 && result.Score < 0.5 {
		return false
	}

	return result.Success
}

// GetCaptchaChallenge generates a CAPTCHA challenge for the client
func (esm *ExecutionSecurityMiddleware) GetCaptchaChallenge(providerName string) (*captcha.CaptchaChallenge, error) {
	return esm.captchaService.GenerateChallenge(providerName)
}

func (esm *ExecutionSecurityMiddleware) getFunctionInputSchema(functionID uuid.UUID, version string) (*storage.FunctionInputSchema, error) {
	var schema storage.FunctionInputSchema

	// First get the function version ID
	var fnVersion storage.RegistryFunctionVersion
	var err error
	if version == "" {
		// Use latest version if no version specified
		err = esm.db.Where("function_id = ?", functionID).Order("published_at DESC").First(&fnVersion).Error
	} else {
		err = esm.db.Where("function_id = ? AND version = ?", functionID, version).First(&fnVersion).Error
	}
	if err != nil {
		return nil, err
	}

	err = esm.db.Where("function_version_id = ?", fnVersion.ID).First(&schema).Error
	return &schema, err
}

func (esm *ExecutionSecurityMiddleware) validateInputAgainstSchema(input interface{}, schema *storage.FunctionInputSchema) error {
	// Parse the JSON Schema
	var schemaData map[string]interface{}
	if err := json.Unmarshal(schema.Schema, &schemaData); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Convert input to the expected format
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		// Try to convert to map if it's not already
		inputBytes, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("failed to marshal input for validation: %w", err)
		}
		if err := json.Unmarshal(inputBytes, &inputMap); err != nil {
			return fmt.Errorf("input must be a valid JSON object: %w", err)
		}
	}

	// Validate against schema
	return esm.validateJSONObject(inputMap, schemaData)
}

// validateJSONObject validates a JSON object against a schema
func (esm *ExecutionSecurityMiddleware) validateJSONObject(obj map[string]interface{}, schema map[string]interface{}) error {
	// Check required fields
	if required, ok := schema["required"].([]interface{}); ok {
		for _, reqField := range required {
			fieldName, ok := reqField.(string)
			if !ok {
				continue
			}
			if _, exists := obj[fieldName]; !exists {
				return fmt.Errorf("required field '%s' is missing", fieldName)
			}
		}
	}

	// Check properties
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for fieldName, fieldSchema := range properties {
			fieldSchemaMap, ok := fieldSchema.(map[string]interface{})
			if !ok {
				continue
			}

			if fieldValue, exists := obj[fieldName]; exists {
				if err := esm.validateField(fieldValue, fieldSchemaMap); err != nil {
					return fmt.Errorf("field '%s': %w", fieldName, err)
				}
			}
		}
	}

	// Check additional properties
	if additionalProps, ok := schema["additionalProperties"].(bool); ok && !additionalProps {
		allowedFields := make(map[string]bool)
		if properties, ok := schema["properties"].(map[string]interface{}); ok {
			for fieldName := range properties {
				allowedFields[fieldName] = true
			}
		}

		for fieldName := range obj {
			if !allowedFields[fieldName] {
				return fmt.Errorf("additional property '%s' is not allowed", fieldName)
			}
		}
	}

	return nil
}

// validateField validates a single field against its schema
func (esm *ExecutionSecurityMiddleware) validateField(value interface{}, fieldSchema map[string]interface{}) error {
	// Check type
	expectedType, _ := fieldSchema["type"].(string)
	if expectedType != "" {
		if err := esm.validateType(value, expectedType); err != nil {
			return err
		}
	}

	// Check string constraints
	if expectedType == "string" || expectedType == "" {
		strVal, ok := value.(string)
		if ok {
			if minLength, ok := fieldSchema["minLength"].(float64); ok {
				if len(strVal) < int(minLength) {
					return fmt.Errorf("string length %d is less than minimum %d", len(strVal), int(minLength))
				}
			}
			if maxLength, ok := fieldSchema["maxLength"].(float64); ok {
				if len(strVal) > int(maxLength) {
					return fmt.Errorf("string length %d exceeds maximum %d", len(strVal), int(maxLength))
				}
			}
			if pattern, ok := fieldSchema["pattern"].(string); ok {
				// Compile and validate against regex pattern
				regex, err := regexp.Compile(pattern)
				if err != nil {
					return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
				}
				if !regex.MatchString(strVal) {
					return fmt.Errorf("string does not match required pattern '%s'", pattern)
				}
			}
		}
	}

	// Check numeric constraints
	if expectedType == "number" || expectedType == "integer" {
		numVal, ok := value.(float64)
		if ok {
			if minimum, ok := fieldSchema["minimum"].(float64); ok {
				if numVal < minimum {
					return fmt.Errorf("number %f is less than minimum %f", numVal, minimum)
				}
			}
			if maximum, ok := fieldSchema["maximum"].(float64); ok {
				if numVal > maximum {
					return fmt.Errorf("number %f exceeds maximum %f", numVal, maximum)
				}
			}
		}
	}

	// Check array constraints
	if expectedType == "array" {
		arrVal, ok := value.([]interface{})
		if ok {
			if minItems, ok := fieldSchema["minItems"].(float64); ok {
				if len(arrVal) < int(minItems) {
					return fmt.Errorf("array length %d is less than minimum %d", len(arrVal), int(minItems))
				}
			}
			if maxItems, ok := fieldSchema["maxItems"].(float64); ok {
				if len(arrVal) > int(maxItems) {
					return fmt.Errorf("array length %d exceeds maximum %d", len(arrVal), int(maxItems))
				}
			}

			// Validate array items if schema is provided
			if itemsSchema, ok := fieldSchema["items"].(map[string]interface{}); ok {
				for i, item := range arrVal {
					if err := esm.validateField(item, itemsSchema); err != nil {
						return fmt.Errorf("array item %d: %w", i, err)
					}
				}
			}
		}
	}

	return nil
}

// validateType checks if a value matches the expected JSON Schema type
func (esm *ExecutionSecurityMiddleware) validateType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
	case "integer":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected integer, got %T", value)
		}
		if num != float64(int64(num)) {
			return fmt.Errorf("expected integer, got float %f", num)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

func (esm *ExecutionSecurityMiddleware) getFunctionResourceLimits(functionID uuid.UUID, version string) *storage.RegistryFunctionVersion {
	var fnVersion storage.RegistryFunctionVersion
	var err error
	if version == "" {
		// Use latest version if no version specified
		err = esm.db.Where("function_id = ?", functionID).Order("published_at DESC").First(&fnVersion).Error
	} else {
		err = esm.db.Where("function_id = ? AND version = ?", functionID, version).First(&fnVersion).Error
	}
	if err != nil {
		return nil
	}
	return &fnVersion
}

func (esm *ExecutionSecurityMiddleware) isSuspiciousInput(input string) bool {
	suspiciousPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:", "onload=", "onerror=",
		"union select", "drop table", "alter table", "--", "#", "/*", "*/",
	}

	input = strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	return false
}

func (esm *ExecutionSecurityMiddleware) recordSecurityEvent(userID *uuid.UUID, ipAddress *string, eventType, severity, message string, eventData map[string]interface{}) {
	event := storage.ExecutionSecurityEvent{
		UserID:    userID,
		EventType: eventType,
		Severity:  severity,
		Message:   message,
	}

	if ipAddress != nil {
		event.IPAddress = *ipAddress
	}

	if eventData != nil {
		if dataJSON, err := json.Marshal(eventData); err == nil {
			event.EventData = dataJSON
		}
	}

	if err := esm.db.Create(&event).Error; err != nil {
		esm.logger.WithError(err).Error("Failed to record security event")
	}
}

// Response helpers

func (esm *ExecutionSecurityMiddleware) respondRateLimited(w http.ResponseWriter, message string) {
	w.Header().Set("X-RateLimit-Limit", "100")
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
	w.Header().Set("Retry-After", "60")
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":       "rate_limited",
		"message":     message,
		"retry_after": 60,
	})
}

func (esm *ExecutionSecurityMiddleware) respondBlocked(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "blocked",
		"message": message,
	})
}

func (esm *ExecutionSecurityMiddleware) respondCaptchaRequired(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Captcha-Required", "true")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "captcha_required",
		"message": message,
	})
}

func (esm *ExecutionSecurityMiddleware) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "internal_error",
		"message": message,
	})
}

// Context helpers

type resourceLimitsKey struct{}

type ResourceLimits struct {
	MaxMemoryMB  int
	MaxCPUTimeMs int
	TimeoutMs    int
}

func contextWithResourceLimits(ctx context.Context, maxMemoryMB, maxCPUTimeMs, timeoutMs int) context.Context {
	return context.WithValue(ctx, resourceLimitsKey{}, &ResourceLimits{
		MaxMemoryMB:  maxMemoryMB,
		MaxCPUTimeMs: maxCPUTimeMs,
		TimeoutMs:    timeoutMs,
	})
}

func GetResourceLimitsFromContext(ctx context.Context) *ResourceLimits {
	if limits, ok := ctx.Value(resourceLimitsKey{}).(*ResourceLimits); ok {
		return limits
	}
	return nil
}

// JSON reader helper that implements io.ReadCloser
type jsonReader struct {
	data    interface{}
	encoded []byte
	pos     int
}

func (jr *jsonReader) Read(p []byte) (n int, err error) {
	if jr.encoded == nil {
		jr.encoded, _ = json.Marshal(jr.data)
	}
	if jr.pos >= len(jr.encoded) {
		return 0, io.EOF
	}
	n = copy(p, jr.encoded[jr.pos:])
	jr.pos += n
	return n, nil
}

func (jr *jsonReader) Close() error {
	return nil
}
