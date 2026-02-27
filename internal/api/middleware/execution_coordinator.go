package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/captcha"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// ExecutionCoordinatorMiddleware coordinates all execution security features
type ExecutionCoordinatorMiddleware struct {
	securityMW *ExecutionSecurityMiddleware
}

// NewExecutionCoordinatorMiddleware creates a new execution coordinator middleware
func NewExecutionCoordinatorMiddleware(db *gorm.DB, config *ExecutionSecurityConfig, captchaService *captcha.CaptchaService) *ExecutionCoordinatorMiddleware {
	if config == nil {
		config = &ExecutionSecurityConfig{
			DefaultDailyLimit:      1000,
			DefaultHourlyLimit:     100,
			DefaultMinuteLimit:     10,
			AnonymousMultiplier:    0.1,
			ThrottleMultiplier:     2.0,
			MaxThrottleDuration:    24 * 60 * 60 * 1000, // 24 hours in milliseconds
			BlockThreshold:         100,
			CaptchaRequiredScore:   50,
			CaptchaValidityWindow:  24 * 60 * 60 * 1000, // 24 hours in milliseconds
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

	securityMW := NewExecutionSecurityMiddleware(db, config, captchaService)

	return &ExecutionCoordinatorMiddleware{
		securityMW: securityMW,
	}
}

// SecureExecution wraps an execution handler with comprehensive security checks
func (ecm *ExecutionCoordinatorMiddleware) SecureExecution(functionID uuid.UUID, version string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Apply security middleware in the correct order

			// 1. Rate limiting and quota checking (first line of defense)
			rateLimitMW := ecm.securityMW.RateLimitAndQuotaMiddleware
			handler := rateLimitMW(func(w http.ResponseWriter, r *http.Request) {
				// 2. CAPTCHA verification for anonymous users
				captchaMW := ecm.securityMW.CaptchaRequiredMiddleware
				captchaHandler := captchaMW(func(w http.ResponseWriter, r *http.Request) {
					// 3. Abuse pattern detection
					abuseMW := ecm.securityMW.AbuseDetectionMiddleware
					abuseHandler := abuseMW(func(w http.ResponseWriter, r *http.Request) {
						// 4. Input validation against function schema
						inputValidationMW := ecm.securityMW.InputValidationMiddleware(functionID, version)
						inputHandler := inputValidationMW(func(w http.ResponseWriter, r *http.Request) {
							// 5. Resource limits enforcement
							resourceMW := ecm.securityMW.ResourceLimitsMiddleware(functionID, version)
							resourceHandler := resourceMW(next)
							resourceHandler.ServeHTTP(w, r)
						})
						inputHandler.ServeHTTP(w, r)
					})
					abuseHandler.ServeHTTP(w, r)
				})
				captchaHandler.ServeHTTP(w, r)
			})

			handler.ServeHTTP(w, r)
		}
	}
}

// GetCaptchaChallenge returns a CAPTCHA challenge for the client
func (ecm *ExecutionCoordinatorMiddleware) GetCaptchaChallenge(providerName string) (*captcha.CaptchaChallenge, error) {
	return ecm.securityMW.GetCaptchaChallenge(providerName)
}

// GetExecutionLimits returns the current execution limits for a user/IP
func (ecm *ExecutionCoordinatorMiddleware) GetExecutionLimits(r *http.Request) (map[string]interface{}, error) {
	userID := getUserIDFromContext(r)
	ipAddress := getClientIP(r)

	quota, err := ecm.securityMW.getOrCreateQuota(userID, ipAddress)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"daily_limit":     quota.DailyExecutionLimit,
		"daily_used":      quota.DailyExecutions,
		"hourly_limit":    quota.HourlyExecutionLimit,
		"hourly_used":     quota.HourlyExecutions,
		"minute_limit":    quota.MinuteExecutionLimit,
		"minute_used":     quota.MinuteExecutions,
		"is_throttled":    quota.IsThrottled,
		"is_blocked":      quota.BlockUntil != nil,
		"captcha_required": quota.CaptchaRequired,
		"suspicious_score": quota.SuspiciousActivityScore,
	}, nil
}

// CreateExecutionSecurityRoutes adds security-related routes to the router
func (ecm *ExecutionCoordinatorMiddleware) CreateExecutionSecurityRoutes(router *mux.Router) {
	// Route for getting CAPTCHA challenges
	router.HandleFunc("/captcha/challenge", func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		if provider == "" {
			provider = "recaptcha_v3" // default
		}

		challenge, err := ecm.GetCaptchaChallenge(provider)
		if err != nil {
			http.Error(w, "Failed to generate CAPTCHA challenge", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(challenge)
	}).Methods("GET", "OPTIONS")

	// Route for getting current execution limits
	router.HandleFunc("/execution/limits", func(w http.ResponseWriter, r *http.Request) {
		limits, err := ecm.GetExecutionLimits(r)
		if err != nil {
			http.Error(w, "Failed to get execution limits", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(limits)
	}).Methods("GET", "OPTIONS")
}