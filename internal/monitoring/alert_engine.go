package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// AlertEngine handles automatic alert generation based on monitoring rules
type AlertEngine struct {
	service      *Service
	alertRules   []AlertRule
	alertHistory map[string]time.Time // Track when alerts were last fired to prevent spam
	mu           sync.RWMutex
}

// NewAlertEngine creates a new alert engine with default rules
func NewAlertEngine(service *Service) *AlertEngine {
	engine := &AlertEngine{
		service:      service,
		alertHistory: make(map[string]time.Time),
	}

	// Define default alert rules
	engine.alertRules = []AlertRule{
		// Performance alerts
		{
			ID:          "high_error_rate",
			Name:        "High Error Rate",
			Description: "Error rate exceeds 5% for 5 minutes",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "error_rate",
				Operator:   "gt",
				Threshold:  0.05, // 5%
				Duration:   5 * time.Minute,
			},
			Cooldown: 10 * time.Minute,
		},
		{
			ID:          "backend_down",
			Name:        "Backend Unavailable",
			Description: "Backend has been unhealthy for 2 minutes",
			Severity:    "error",
			Condition: AlertCondition{
				MetricType: "health_status",
				Operator:   "eq",
				Threshold:  0, // 0 = unhealthy
				Duration:   2 * time.Minute,
			},
			Cooldown: 5 * time.Minute,
		},
		{
			ID:          "circuit_breaker_open",
			Name:        "Circuit Breaker Open",
			Description: "Circuit breaker has opened due to failures",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "circuit_state",
				Operator:   "eq_str",
				StringValue: "open", // Circuit breaker state
				Duration:   1 * time.Minute,
			},
			Cooldown: 15 * time.Minute,
		},
		{
			ID:          "high_latency",
			Name:        "High Response Latency",
			Description: "Average response time exceeds 1000ms for 3 minutes",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "response_time",
				Operator:   "gt",
				Threshold:  1000, // 1000ms
				Duration:   3 * time.Minute,
			},
			Cooldown: 10 * time.Minute,
		},
		// Security alerts
		{
			ID:          "suspicious_requests",
			Name:        "Suspicious Request Activity",
			Description: "High number of suspicious requests detected",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "suspicious_requests",
				Operator:   "gt",
				Threshold:  10,
				Duration:   1 * time.Minute,
			},
			Cooldown: 5 * time.Minute,
		},
		{
			ID:          "sql_injection_attempt",
			Name:        "SQL Injection Attempt Detected",
			Description: "Potential SQL injection attack detected",
			Severity:    "critical",
			Condition: AlertCondition{
				MetricType: "sql_injection_detected",
				Operator:   "gt",
				Threshold:  0,
				Duration:   1 * time.Minute,
			},
			Cooldown: 1 * time.Minute,
		},
		{
			ID:          "xss_attempt",
			Name:        "XSS Attempt Detected",
			Description: "Potential cross-site scripting attack detected",
			Severity:    "high",
			Condition: AlertCondition{
				MetricType: "xss_detected",
				Operator:   "gt",
				Threshold:  0,
				Duration:   1 * time.Minute,
			},
			Cooldown: 2 * time.Minute,
		},
		{
			ID:          "brute_force_attempt",
			Name:        "Brute Force Attack Detected",
			Description: "Multiple failed authentication attempts from single IP",
			Severity:    "high",
			Condition: AlertCondition{
				MetricType: "failed_auth_attempts",
				Operator:   "gt",
				Threshold:  5,
				Duration:   5 * time.Minute,
			},
			Cooldown: 10 * time.Minute,
		},
		{
			ID:          "unusual_traffic_spike",
			Name:        "Unusual Traffic Spike",
			Description: "Significant increase in traffic that may indicate DDoS attack",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "traffic_anomaly_score",
				Operator:   "gt",
				Threshold:  0.8,
				Duration:   2 * time.Minute,
			},
			Cooldown: 15 * time.Minute,
		},
		{
			ID:          "compliance_violation",
			Name:        "Compliance Violation Detected",
			Description: "Security compliance requirement has been violated",
			Severity:    "critical",
			Condition: AlertCondition{
				MetricType: "compliance_violation",
				Operator:   "gt",
				Threshold:  0,
				Duration:   1 * time.Minute,
			},
			Cooldown: 5 * time.Minute,
		},
		{
			ID:          "encryption_failure",
			Name:        "Database Encryption Failure",
			Description: "Database encryption is not functioning properly",
			Severity:    "critical",
			Condition: AlertCondition{
				MetricType: "encryption_status",
				Operator:   "eq",
				Threshold:  0, // 0 = failed
				Duration:   1 * time.Minute,
			},
			Cooldown: 2 * time.Minute,
		},
		// Local runtime alerts
		{
			ID:          "local_runtime_high_error_rate",
			Name:        "Local Runtime High Error Rate",
			Description: "Local runtime error rate exceeds 10% for 2 minutes",
			Severity:    "error",
			Condition: AlertCondition{
				MetricType: "local_runtime_error_rate",
				Operator:   "gt",
				Threshold:  0.10, // 10%
				Duration:   2 * time.Minute,
			},
			Cooldown: 5 * time.Minute,
		},
		{
			ID:          "local_runtime_high_latency",
			Name:        "Local Runtime High Latency",
			Description: "Local runtime execution latency exceeds 2000ms for 3 minutes",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "local_runtime_execution_duration",
				Operator:   "gt",
				Threshold:  2000, // 2000ms
				Duration:   3 * time.Minute,
			},
			Cooldown: 10 * time.Minute,
		},
		{
			ID:          "local_runtime_high_memory_usage",
			Name:        "Local Runtime High Memory Usage",
			Description: "Local runtime memory usage exceeds 500MB for 2 minutes",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "local_runtime_memory_usage",
				Operator:   "gt",
				Threshold:  500 * 1024 * 1024, // 500MB in bytes
				Duration:   2 * time.Minute,
			},
			Cooldown: 5 * time.Minute,
		},
		{
			ID:          "local_runtime_high_cpu_usage",
			Name:        "Local Runtime High CPU Usage",
			Description: "Local runtime CPU usage exceeds 80% for 2 minutes",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "local_runtime_cpu_usage",
				Operator:   "gt",
				Threshold:  80.0, // 80%
				Duration:   2 * time.Minute,
			},
			Cooldown: 5 * time.Minute,
		},
		{
			ID:          "local_runtime_low_throughput",
			Name:        "Local Runtime Low Throughput",
			Description: "Local runtime request throughput below 1 request/second for 5 minutes",
			Severity:    "warning",
			Condition: AlertCondition{
				MetricType: "local_runtime_request_throughput",
				Operator:   "lt",
				Threshold:  1.0, // 1 request/second
				Duration:   5 * time.Minute,
			},
			Cooldown: 10 * time.Minute,
		},
	}

	return engine
}

// ProcessMetric evaluates alert rules against a new metric
func (ae *AlertEngine) ProcessMetric(metric *storage.PerformanceMetric) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	for _, rule := range ae.alertRules {
		if ae.shouldTriggerAlert(rule, metric) {
			go ae.triggerAlert(rule, metric)
		}
	}
}

// shouldTriggerAlert checks if an alert rule should trigger based on the metric
func (ae *AlertEngine) shouldTriggerAlert(rule AlertRule, metric *storage.PerformanceMetric) bool {
	// Check if metric type matches
	if metric.MetricType != rule.Condition.MetricType {
		return false
	}

	// Check cooldown period
	alertKey := rule.ID + ":" + metric.TenantID.String()
	if lastTriggered, exists := ae.alertHistory[alertKey]; exists {
		if time.Since(lastTriggered) < rule.Cooldown {
			return false // Still in cooldown
		}
	}

	// Check condition
	conditionMet := false
	switch rule.Condition.Operator {
	case "gt":
		conditionMet = metric.Value > rule.Condition.Threshold
	case "lt":
		conditionMet = metric.Value < rule.Condition.Threshold
	case "eq":
		conditionMet = metric.Value == rule.Condition.Threshold
	case "ne":
		conditionMet = metric.Value != rule.Condition.Threshold
	case "eq_str":
		conditionMet = metric.StringValue == rule.Condition.StringValue
	case "ne_str":
		conditionMet = metric.StringValue != rule.Condition.StringValue
	}

	return conditionMet
}

// triggerAlert creates and sends an alert
func (ae *AlertEngine) triggerAlert(rule AlertRule, metric *storage.PerformanceMetric) {
	metadata := map[string]interface{}{
		"rule_id":      rule.ID,
		"metric_type":  metric.MetricType,
		"operator":     rule.Condition.Operator,
		"triggered_at": time.Now(),
	}

	// Add value information based on metric type
	if metric.StringValue != "" {
		metadata["string_value"] = metric.StringValue
		metadata["expected_string_value"] = rule.Condition.StringValue
	} else {
		metadata["metric_value"] = metric.Value
		metadata["threshold"] = rule.Condition.Threshold
	}

	alert := &storage.Alert{
		AlertType: rule.ID,
		Severity:  rule.Severity,
		TenantID:  metric.TenantID,
		AppID:     metric.AppID,
		BackendID: metric.BackendID,
		Title:     rule.Name,
		Message:   ae.generateAlertMessage(rule, metric),
		Status:    "active",
		Metadata:  metadata,
	}

	// Record the alert
	ctx := context.Background()
	if err := ae.service.RecordAlert(ctx, alert); err != nil {
		logrus.WithError(err).WithField("rule_id", rule.ID).Error("Failed to record triggered alert")
		return
	}

	// Update alert history
	alertKey := rule.ID + ":" + metric.TenantID.String()
	ae.alertHistory[alertKey] = time.Now()

	logrus.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"rule_id":    rule.ID,
		"severity":   rule.Severity,
		"metric_value": metric.Value,
	}).Info("Alert triggered by rule")
}

// generateAlertMessage creates a human-readable alert message
func (ae *AlertEngine) generateAlertMessage(rule AlertRule, metric *storage.PerformanceMetric) string {
	var unit string
	switch metric.Unit {
	case "ms":
		unit = "milliseconds"
	case "percent":
		unit = "percent"
	case "requests_per_second":
		unit = "requests/second"
	case "state":
		unit = ""
	default:
		unit = metric.Unit
	}

	operatorDesc := ""
	var expectedValue interface{}
	var currentValue interface{}

	switch rule.Condition.Operator {
	case "gt":
		operatorDesc = "exceeded"
		expectedValue = rule.Condition.Threshold
		currentValue = metric.Value
	case "lt":
		operatorDesc = "fell below"
		expectedValue = rule.Condition.Threshold
		currentValue = metric.Value
	case "eq":
		operatorDesc = "equals"
		expectedValue = rule.Condition.Threshold
		currentValue = metric.Value
	case "ne":
		operatorDesc = "does not equal"
		expectedValue = rule.Condition.Threshold
		currentValue = metric.Value
	case "eq_str":
		operatorDesc = "is"
		expectedValue = rule.Condition.StringValue
		currentValue = metric.StringValue
		unit = "" // No unit for string values
	case "ne_str":
		operatorDesc = "is not"
		expectedValue = rule.Condition.StringValue
		currentValue = metric.StringValue
		unit = "" // No unit for string values
	}

	if unit == "" {
		return fmt.Sprintf("%s: %s %s %v (current: %v)",
			rule.Description,
			metric.MetricType,
			operatorDesc,
			expectedValue,
			currentValue)
	}

	return fmt.Sprintf("%s: %s %s %v %s (current: %v %s)",
		rule.Description,
		metric.MetricType,
		operatorDesc,
		expectedValue,
		unit,
		currentValue,
		unit)
}