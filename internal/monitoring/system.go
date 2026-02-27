package monitoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/cli"
)

// MonitoringSystem handles comprehensive monitoring and alerting
type MonitoringSystem struct {
	client      *cli.Client
	alertRules  []ThresholdAlertRule
	collectors  []MetricCollector
	mu          sync.RWMutex
}

// ThresholdAlertRule represents a threshold-based alerting rule
type ThresholdAlertRule struct {
	ID          string
	Name        string
	Description string
	Query       string
	Threshold   float64
	Operator    string // ">", "<", ">=", "<=", "==", "!="
	Severity    string // "info", "warning", "error", "critical"
	Cooldown    time.Duration
	Enabled     bool
}

// MetricCollector collects metrics from various sources
type MetricCollector interface {
	Collect(ctx context.Context) ([]Metric, error)
	Name() string
}

// Metric represents a single metric measurement
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Alert represents an active alert
type Alert struct {
	ID        string    `json:"id"`
	RuleID    string    `json:"rule_id"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Labels    map[string]string `json:"labels,omitempty"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// NewMonitoringSystem creates a new monitoring system
func NewMonitoringSystem(client *cli.Client) *MonitoringSystem {
	ms := &MonitoringSystem{
		client:     client,
		alertRules: make([]ThresholdAlertRule, 0),
		collectors: make([]MetricCollector, 0),
	}

	// Add default alert rules
	ms.addDefaultAlertRules()

	return ms
}

// addDefaultAlertRules adds default alerting rules
func (ms *MonitoringSystem) addDefaultAlertRules() {
	defaultRules := []ThresholdAlertRule{
		{
			ID:          "high-error-rate",
			Name:        "High Error Rate",
			Description: "Function error rate is above 5%",
			Query:       "error_rate",
			Threshold:   5.0,
			Operator:    ">",
			Severity:    "warning",
			Cooldown:    5 * time.Minute,
			Enabled:     true,
		},
		{
			ID:          "high-latency",
			Name:        "High Latency",
			Description: "Function latency is above 1000ms",
			Query:       "avg_latency_ms",
			Threshold:   1000.0,
			Operator:    ">=",
			Severity:    "warning",
			Cooldown:    5 * time.Minute,
			Enabled:     true,
		},
		{
			ID:          "critical-error-rate",
			Name:        "Critical Error Rate",
			Description: "Function error rate is above 10%",
			Query:       "error_rate",
			Threshold:   10.0,
			Operator:    ">",
			Severity:    "critical",
			Cooldown:    2 * time.Minute,
			Enabled:     true,
		},
	}

	for _, rule := range defaultRules {
		ms.AddAlertRule(rule)
	}
}

// AddAlertRule adds a new alert rule
func (ms *MonitoringSystem) AddAlertRule(rule ThresholdAlertRule) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.alertRules = append(ms.alertRules, rule)
	log.Printf("Added alert rule: %s", rule.Name)
}

// RemoveAlertRule removes an alert rule
func (ms *MonitoringSystem) RemoveAlertRule(ruleID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, rule := range ms.alertRules {
		if rule.ID == ruleID {
			ms.alertRules = append(ms.alertRules[:i], ms.alertRules[i+1:]...)
			log.Printf("Removed alert rule: %s", ruleID)
			return
		}
	}
}

// AddCollector adds a metric collector
func (ms *MonitoringSystem) AddCollector(collector MetricCollector) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.collectors = append(ms.collectors, collector)
	log.Printf("Added metric collector: %s", collector.Name())
}

// CollectMetrics collects metrics from all collectors
func (ms *MonitoringSystem) CollectMetrics(ctx context.Context) ([]Metric, error) {
	ms.mu.RLock()
	collectors := make([]MetricCollector, len(ms.collectors))
	copy(collectors, ms.collectors)
	ms.mu.RUnlock()

	var allMetrics []Metric

	for _, collector := range collectors {
		metrics, err := collector.Collect(ctx)
		if err != nil {
			log.Printf("Failed to collect metrics from %s: %v", collector.Name(), err)
			continue
		}
		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, nil
}

// CheckAlerts checks all alert rules against current metrics
func (ms *MonitoringSystem) CheckAlerts(metrics []Metric) []Alert {
	ms.mu.RLock()
	rules := make([]ThresholdAlertRule, len(ms.alertRules))
	copy(rules, ms.alertRules)
	ms.mu.RUnlock()

	var alerts []Alert

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Find matching metrics
		for _, metric := range metrics {
			if metric.Name == rule.Query {
				if ms.evaluateRule(rule, metric.Value) {
					alert := Alert{
						ID:        fmt.Sprintf("%s-%d", rule.ID, time.Now().Unix()),
						RuleID:    rule.ID,
						Message:   fmt.Sprintf("%s: %s (%.2f %s %.2f)", rule.Name, rule.Description, metric.Value, rule.Operator, rule.Threshold),
						Severity:  rule.Severity,
						Value:     metric.Value,
						Threshold: rule.Threshold,
						Labels:    metric.Labels,
						StartedAt: time.Now(),
					}
					alerts = append(alerts, alert)
				}
			}
		}
	}

	return alerts
}

// evaluateRule evaluates if a rule condition is met
func (ms *MonitoringSystem) evaluateRule(rule ThresholdAlertRule, value float64) bool {
	switch rule.Operator {
	case ">":
		return value > rule.Threshold
	case "<":
		return value < rule.Threshold
	case ">=":
		return value >= rule.Threshold
	case "<=":
		return value <= rule.Threshold
	case "==":
		return value == rule.Threshold
	case "!=":
		return value != rule.Threshold
	default:
		return false
	}
}

// GetLogs retrieves logs with filtering
func (ms *MonitoringSystem) GetLogs(author, name string, params map[string]string) ([]*cli.FunctionLogEntry, error) {
	return ms.client.GetFunctionLogs(author, name, params)
}

// GetMetrics retrieves detailed metrics
func (ms *MonitoringSystem) GetMetrics(author, name, period string) (*cli.DetailedMetrics, error) {
	return ms.client.GetDetailedMetrics(author, name, period)
}

// GetHealthStatus retrieves health status
func (ms *MonitoringSystem) GetHealthStatus(author, name string) (*cli.HealthStatus, error) {
	return ms.client.GetHealthStatus(author, name)
}

// StartMonitoring starts the monitoring system
func (ms *MonitoringSystem) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting monitoring system with %d collectors and %d alert rules", len(ms.collectors), len(ms.alertRules))

	for {
		select {
		case <-ctx.Done():
			log.Println("Monitoring system stopped")
			return
		case <-ticker.C:
			metrics, err := ms.CollectMetrics(ctx)
			if err != nil {
				log.Printf("Failed to collect metrics: %v", err)
				continue
			}

			alerts := ms.CheckAlerts(metrics)
			if len(alerts) > 0 {
				for _, alert := range alerts {
					log.Printf("ALERT [%s]: %s", alert.Severity, alert.Message)
				}
			}

			log.Printf("Collected %d metrics, %d alerts triggered", len(metrics), len(alerts))
		}
	}
}