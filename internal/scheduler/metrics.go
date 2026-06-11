package scheduler

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SchedulerMetrics struct {
	TotalChecks       prometheus.Gauge
	TotalAlertsSent   prometheus.Gauge
	TotalErrors       prometheus.Gauge
	LastCheckTime     prometheus.Gauge
	LastCheckLatency  prometheus.Gauge
	CheckDuration     prometheus.Histogram
	WalletsChecked    prometheus.Gauge
	LowBalanceCount   prometheus.Gauge
	AlertsSent        prometheus.Gauge
	AlertsThrottled   prometheus.Gauge
	EmailsSent        prometheus.Gauge
	InAppNotifsSent   prometheus.Gauge
	AutoTopupAlerts   prometheus.Gauge

	registry prometheus.Registerer
	mu        sync.RWMutex
}

func NewSchedulerMetrics() *SchedulerMetrics {
	return &SchedulerMetrics{
		TotalChecks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_checks_total",
			Help:      "Total number of low balance checks performed",
		}),
		TotalAlertsSent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_alerts_sent_total",
			Help:      "Total number of low balance alerts sent",
		}),
		TotalErrors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_errors_total",
			Help:      "Total number of low balance check errors",
		}),
		LastCheckTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_last_check_timestamp",
			Help:      "Unix timestamp of last low balance check",
		}),
		LastCheckLatency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_last_check_latency_seconds",
			Help:      "Latency of last low balance check in seconds",
		}),
		CheckDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_check_duration_seconds",
			Help:      "Duration of low balance check operations",
			Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60},
		}),
		WalletsChecked: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_wallets_checked",
			Help:      "Number of wallets checked in last run",
		}),
		LowBalanceCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_count",
			Help:      "Number of wallets with low balance in last run",
		}),
		AlertsSent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_alerts_sent",
			Help:      "Number of alerts sent in last run",
		}),
		AlertsThrottled: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_alerts_throttled",
			Help:      "Number of alerts throttled in last run",
		}),
		EmailsSent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_emails_sent",
			Help:      "Number of emails sent in last run",
		}),
		InAppNotifsSent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_in_app_notifications_sent",
			Help:      "Number of in-app notifications sent in last run",
		}),
		AutoTopupAlerts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "scheduler",
			Name:      "low_balance_auto_topup_alerts",
			Help:      "Number of auto-topup alerts in last run",
		}),
	}
}

func (m *SchedulerMetrics) Register(r prometheus.Registerer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.registry = r

	metrics := []prometheus.Collector{
		m.TotalChecks,
		m.TotalAlertsSent,
		m.TotalErrors,
		m.LastCheckTime,
		m.LastCheckLatency,
		m.CheckDuration,
		m.WalletsChecked,
		m.LowBalanceCount,
		m.AlertsSent,
		m.AlertsThrottled,
		m.EmailsSent,
		m.InAppNotifsSent,
		m.AutoTopupAlerts,
	}

	for _, metric := range metrics {
		if err := r.Register(metric); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
				continue
			}
			return err
		}
	}

	return nil
}

func (m *SchedulerMetrics) Unregister() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.registry == nil {
		return
	}

	m.registry.Unregister(m.TotalChecks)
	m.registry.Unregister(m.TotalAlertsSent)
	m.registry.Unregister(m.TotalErrors)
	m.registry.Unregister(m.LastCheckTime)
	m.registry.Unregister(m.LastCheckLatency)
	m.registry.Unregister(m.CheckDuration)
	m.registry.Unregister(m.WalletsChecked)
	m.registry.Unregister(m.LowBalanceCount)
	m.registry.Unregister(m.AlertsSent)
	m.registry.Unregister(m.AlertsThrottled)
	m.registry.Unregister(m.EmailsSent)
	m.registry.Unregister(m.InAppNotifsSent)
	m.registry.Unregister(m.AutoTopupAlerts)
}

func (m *SchedulerMetrics) UpdateFromMetrics(internal *LowBalanceMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalChecks.Set(float64(internal.TotalChecks))
	m.TotalAlertsSent.Set(float64(internal.TotalAlertsSent))
	m.TotalErrors.Set(float64(internal.TotalErrors))
	if !internal.LastCheckTime.IsZero() {
		m.LastCheckTime.Set(float64(internal.LastCheckTime.Unix()))
	}
	m.LastCheckLatency.Set(internal.LastCheckLatency.Seconds())
}

func (m *SchedulerMetrics) UpdateFromResult(result *LowBalanceCheckResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if result == nil {
		return
	}

	m.CheckDuration.Observe(result.CheckDuration.Seconds())
	m.WalletsChecked.Set(float64(result.WalletsChecked))
	m.LowBalanceCount.Set(float64(result.LowBalanceCount))
	m.AlertsSent.Set(float64(result.AlertsSent))
	m.AlertsThrottled.Set(float64(result.AlertsThrottled))
	m.EmailsSent.Set(float64(result.EmailsSent))
	m.InAppNotifsSent.Set(float64(result.InAppNotifsSent))
	m.AutoTopupAlerts.Set(float64(result.AutoTopupAlerts))
}

var globalSchedulerMetrics *SchedulerMetrics
var schedulerMetricsOnce sync.Once

func GetSchedulerMetrics() *SchedulerMetrics {
	schedulerMetricsOnce.Do(func() {
		globalSchedulerMetrics = NewSchedulerMetrics()
	})
	return globalSchedulerMetrics
}

func RegisterSchedulerMetrics(r prometheus.Registerer) error {
	return GetSchedulerMetrics().Register(r)
}
