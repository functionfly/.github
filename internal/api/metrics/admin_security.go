package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AdminLoginAttempts tracks admin login attempts by status
	AdminLoginAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "admin_login_attempts_total",
			Help: "Total number of admin login attempts",
		},
		[]string{"status"},
	)

	// AdminRateLimitHits tracks admin rate limit hits
	AdminRateLimitHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "admin_rate_limit_hits_total",
			Help: "Total number of admin rate limit hits",
		},
	)

	// AdminIPBlocks tracks admin IP blocks
	AdminIPBlocks = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "admin_ip_blocks_total",
			Help: "Total number of admin IP blocks",
		},
	)

	// AdminSessionValidations tracks admin session validations
	AdminSessionValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "admin_session_validations_total",
			Help: "Total number of admin session validations",
		},
		[]string{"status"},
	)

	// AdminCSRFViolations tracks admin CSRF violations
	AdminCSRFViolations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "admin_csrf_violations_total",
			Help: "Total number of admin CSRF violations",
		},
	)

	// AdminSecurityAlertsTriggered tracks security alerts triggered
	AdminSecurityAlertsTriggered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "admin_security_alerts_triggered_total",
			Help: "Total number of admin security alerts triggered",
		},
		[]string{"alert_type", "severity"},
	)

	// AdminSecurityAlertChecks tracks security alert checks
	AdminSecurityAlertChecks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "admin_security_alert_checks_total",
			Help: "Total number of admin security alert checks",
		},
		[]string{"alert_type"},
	)

	// AdminFailedLoginAttempts tracks failed admin login attempts by IP
	AdminFailedLoginAttempts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "admin_failed_login_attempts_current",
			Help: "Current number of failed admin login attempts in window",
		},
		[]string{"ip_address"},
	)

	// AdminSuspiciousActivity tracks suspicious activity indicators
	AdminSuspiciousActivity = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "admin_suspicious_activity_total",
			Help: "Total number of suspicious activity indicators",
		},
		[]string{"type", "severity"},
	)

	// AdminSessionAnomalies tracks session anomalies detected
	AdminSessionAnomalies = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "admin_session_anomalies_total",
			Help: "Total number of admin session anomalies detected",
		},
		[]string{"type"},
	)

	// AdminIPReputationScores tracks IP reputation scores
	AdminIPReputationScores = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "admin_ip_reputation_score",
			Help: "Current IP reputation score (-100 to 100)",
		},
		[]string{"ip_address"},
	)
)

// RecordLoginSuccess records a successful login
func RecordLoginSuccess() {
	AdminLoginAttempts.WithLabelValues("success").Inc()
}

// RecordLoginFailure records a failed login
func RecordLoginFailure() {
	AdminLoginAttempts.WithLabelValues("failure").Inc()
}

// RecordRateLimitHit records a rate limit hit
func RecordRateLimitHit() {
	AdminRateLimitHits.Inc()
}

// RecordIPBlock records an IP block event
func RecordIPBlock() {
	AdminIPBlocks.Inc()
}

// RecordSessionValidation records a session validation attempt
func RecordSessionValidation(success bool) {
	if success {
		AdminSessionValidations.WithLabelValues("success").Inc()
	} else {
		AdminSessionValidations.WithLabelValues("failure").Inc()
	}
}

// RecordCSRFViolation records a CSRF violation
func RecordCSRFViolation() {
	AdminCSRFViolations.Inc()
}

// RecordSecurityAlert records a triggered security alert
func RecordSecurityAlert(alertType, severity string) {
	AdminSecurityAlertsTriggered.WithLabelValues(alertType, severity).Inc()
}

// RecordSecurityAlertCheck records a security alert check
func RecordSecurityAlertCheck(alertType string) {
	AdminSecurityAlertChecks.WithLabelValues(alertType).Inc()
}
