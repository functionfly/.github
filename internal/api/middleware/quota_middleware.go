package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// QuotaMiddleware provides real-time quota enforcement for registry executions
// This ensures users get immediate feedback when approaching limits
type QuotaMiddleware struct {
	tracker   services.RealtimeUsageTrackerInterface
	repo      storage.Repository
	logger    *logrus.Logger
	enabled   bool
	mode      string // "enforce", "warn", "log"
}

// Quota mode constants
const (
	QuotaModeEnforce = "enforce"
	QuotaModeWarn    = "warn"
	QuotaModeLog     = "log"
)

// QuotaMiddlewareConfig holds configuration for the quota middleware
type QuotaMiddlewareConfig struct {
	// Enabled controls whether quota enforcement is active
	Enabled bool

	// Mode determines enforcement behavior:
	// - "enforce": Block requests when quota exceeded (default)
	// - "warn": Allow requests but add warning headers
	// - "log": Only log violations without blocking
	Mode string

	// AllowPublicFunctions allows public functions to execute even if quota exceeded
	AllowPublicFunctions bool
}

// DefaultQuotaMiddlewareConfig returns default configuration
func DefaultQuotaMiddlewareConfig() *QuotaMiddlewareConfig {
	return &QuotaMiddlewareConfig{
		Enabled:              true,
		Mode:                 "enforce",
		AllowPublicFunctions: true,
	}
}

// NewQuotaMiddleware creates a new quota enforcement middleware
func NewQuotaMiddleware(
	tracker services.RealtimeUsageTrackerInterface,
	repo storage.Repository,
	config *QuotaMiddlewareConfig,
) *QuotaMiddleware {
	if config == nil {
		config = DefaultQuotaMiddlewareConfig()
	}

	enabled := config.Enabled
	if tracker != nil {
		enabled = enabled && tracker.IsEnabled()
	}

	return &QuotaMiddleware{
		tracker: tracker,
		repo:    repo,
		logger:  logrus.New(),
		enabled: enabled,
		mode:    config.Mode,
	}
}

// QuotaContextKey is the context key for quota status
type QuotaContextKey struct{}

// Middleware returns the HTTP middleware handler
func (m *QuotaMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only enforce quotas on registry execution endpoints
		if !m.shouldEnforce(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract tenant ID from context (set by auth middleware)
		tenantID, err := m.extractTenantID(r)
		if err != nil {
			m.logger.WithError(err).Debug("Could not extract tenant ID, skipping quota check")
			next.ServeHTTP(w, r)
			return
		}

		// Check if this is a public function execution (might be exempt)
		if m.isPublicFunctionExecution(r) {
			// Public functions are allowed to execute regardless of quota
			next.ServeHTTP(w, r)
			return
		}

		// Perform real-time quota check BEFORE execution
		ctx := r.Context()
		result, err := m.tracker.RecordExecution(ctx, tenantID, "")
		if err != nil {
			m.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Quota check failed")
			// On error, allow execution (fail open) but log
			next.ServeHTTP(w, r)
			return
		}

		// Add quota headers to response
		m.addQuotaHeaders(w, result.Status)

		// Handle quota check result
		if !result.Allowed {
			switch m.mode {
			case "enforce":
				m.rejectRequest(w, result)
				return
			case "warn":
				// Add warning header but continue
				w.Header().Set("X-Quota-Warning", "Quota exceeded: "+result.Reason)
			case "log":
				m.logger.WithFields(logrus.Fields{
					"tenant_id": tenantID,
					"reason":    result.Reason,
				}).Warn("Quota would be exceeded (logging only)")
			}
		}

		// Add quota status to context for downstream handlers
		ctx = context.WithValue(ctx, QuotaContextKey{}, result.Status)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// shouldEnforce determines if quota enforcement should apply to this request
func (m *QuotaMiddleware) shouldEnforce(r *http.Request) bool {
	if !m.enabled {
		return false
	}

	// Only enforce on POST/PUT/PATCH methods (mutations and executions)
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		return false
	}

	// Only enforce on registry execution endpoints
	path := r.URL.Path
	return strings.Contains(path, "/registry/") &&
		(strings.Contains(path, "/execute") ||
			strings.Contains(path, "/run") ||
			strings.Contains(path, "/invoke"))
}

// extractTenantID extracts the tenant ID from the request context
func (m *QuotaMiddleware) extractTenantID(r *http.Request) (uuid.UUID, error) {
	// Try to get from context (set by auth middleware)
	if tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
		return tenantID, nil
	}

	// Try to get from user context
	if user, ok := r.Context().Value("user").(*storage.User); ok && user.TenantID != uuid.Nil {
		return user.TenantID, nil
	}

	// Try to extract from JWT claims
	if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
		if tenantIDStr, ok := claims["tenant_id"].(string); ok {
			return uuid.Parse(tenantIDStr)
		}
	}

	return uuid.Nil, nil
}

// isPublicFunctionExecution checks if this is a public function execution request
func (m *QuotaMiddleware) isPublicFunctionExecution(r *http.Request) bool {
	// Check header for public function indicator
	if r.Header.Get("X-Public-Function") == "true" {
		return true
	}

	// Check query parameter
	if r.URL.Query().Get("public") == "true" {
		return true
	}

	return false
}

// addQuotaHeaders adds quota information to response headers
func (m *QuotaMiddleware) addQuotaHeaders(w http.ResponseWriter, status *services.RealtimeQuotaStatus) {
	if status == nil {
		return
	}

	w.Header().Set("X-Quota-Executions-Used", fmt.Sprintf("%d", status.ExecutionsUsed))
	w.Header().Set("X-Quota-Executions-Limit", fmt.Sprintf("%d", status.ExecutionsLimit))
	w.Header().Set("X-Quota-Executions-Percent", fmt.Sprintf("%.1f", status.ExecutionsPercent))
	w.Header().Set("X-Quota-Status", status.Status)

	if status.ComputeMsLimit > 0 {
		w.Header().Set("X-Quota-Compute-Used", fmt.Sprintf("%d", status.ComputeMsUsed))
		w.Header().Set("X-Quota-Compute-Limit", fmt.Sprintf("%d", status.ComputeMsLimit))
		w.Header().Set("X-Quota-Compute-Percent", fmt.Sprintf("%.1f", status.ComputeMsPercent))
	}
}

// rejectRequest rejects a request due to quota exceeded
func (m *QuotaMiddleware) rejectRequest(w http.ResponseWriter, result *services.QuotaCheckResult) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired) // 402 Payment Required

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "QUOTA_EXCEEDED",
			"message": result.Reason,
			"type":    "quota_exceeded",
		},
		"quota_status": result.Status,
		"upgrade_url":  "/settings/billing",
	}

	json.NewEncoder(w).Encode(response)
}

// formatPercent formats a percentage for headers
func formatPercent(p float64) string {
	return fmt.Sprintf("%.1f", p)
}

// GetQuotaStatusFromContext retrieves quota status from request context
func GetQuotaStatusFromContext(ctx context.Context) (*services.RealtimeQuotaStatus, bool) {
	status, ok := ctx.Value(QuotaContextKey{}).(*services.RealtimeQuotaStatus)
	return status, ok
}

// QuotaEnforcer provides a simpler interface for direct quota checks
type QuotaEnforcer struct {
	tracker services.RealtimeUsageTrackerInterface
	logger  *logrus.Logger
}

// NewQuotaEnforcer creates a new quota enforcer
func NewQuotaEnforcer(tracker services.RealtimeUsageTrackerInterface) *QuotaEnforcer {
	return &QuotaEnforcer{
		tracker: tracker,
		logger:  logrus.New(),
	}
}

// CheckAndRecord checks quota and records the execution attempt
// Returns (allowed, status, error)
func (e *QuotaEnforcer) CheckAndRecord(ctx context.Context, tenantID uuid.UUID) (bool, *services.RealtimeQuotaStatus, error) {
	if e.tracker == nil || !e.tracker.IsEnabled() {
		return true, nil, nil
	}

	result, err := e.tracker.RecordExecution(ctx, tenantID, "")
	if err != nil {
		e.logger.WithError(err).Warn("Quota check failed, allowing execution")
		return true, nil, err
	}

	return result.Allowed, result.Status, nil
}

// RecordComputeUsage records compute time usage
func (e *QuotaEnforcer) RecordComputeUsage(ctx context.Context, tenantID uuid.UUID, cpuTimeMs int) error {
	if e.tracker == nil || !e.tracker.IsEnabled() {
		return nil
	}

	return e.tracker.RecordComputeUsage(ctx, tenantID, cpuTimeMs)
}

// GetStatus returns current quota status
func (e *QuotaEnforcer) GetStatus(ctx context.Context, tenantID uuid.UUID) (*services.RealtimeQuotaStatus, error) {
	if e.tracker == nil || !e.tracker.IsEnabled() {
		return nil, nil
	}

	return e.tracker.GetQuotaStatus(ctx, tenantID)
}

// QuotaCheckMiddleware provides a lightweight check-only middleware
// that adds quota info headers without blocking requests
type QuotaCheckMiddleware struct {
	tracker services.RealtimeUsageTrackerInterface
}

// NewQuotaCheckMiddleware creates a middleware that adds quota info without enforcement
func NewQuotaCheckMiddleware(tracker services.RealtimeUsageTrackerInterface) *QuotaCheckMiddleware {
	return &QuotaCheckMiddleware{tracker: tracker}
}

// Middleware returns the HTTP middleware handler
func (m *QuotaCheckMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.tracker == nil || !m.tracker.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Extract tenant ID
		tenantID, err := m.extractTenantID(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Get quota status (this doesn't increment counters)
		ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
		defer cancel()

		status, err := m.tracker.GetQuotaStatus(ctx, tenantID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Add quota headers
		if status != nil {
			w.Header().Set("X-Quota-Executions-Percent", formatPercent(status.ExecutionsPercent))
			w.Header().Set("X-Quota-Compute-Percent", formatPercent(status.ComputeMsPercent))
			w.Header().Set("X-Quota-Status", status.Status)
		}

		next.ServeHTTP(w, r)
	})
}

func (m *QuotaCheckMiddleware) extractTenantID(r *http.Request) (uuid.UUID, error) {
	if tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
		return tenantID, nil
	}
	if user, ok := r.Context().Value("user").(*storage.User); ok && user.TenantID != uuid.Nil {
		return user.TenantID, nil
	}
	return uuid.Nil, nil
}
