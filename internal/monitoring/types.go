package monitoring

import "time"

// AlertRule defines conditions for automatic alert generation
type AlertRule struct {
	ID          string
	Name        string
	Description string
	Severity    string
	Condition   AlertCondition
	Cooldown    time.Duration // Minimum time between alerts of this type
}

// AlertCondition defines when an alert should be triggered
type AlertCondition struct {
	MetricType      string
	Operator        string // "gt", "lt", "eq", "ne", "eq_str", "ne_str"
	Threshold       float64
	StringValue     string // For string-based comparisons
	Duration        time.Duration // How long condition must be true
	Labels          map[string]interface{} // Additional filters
}