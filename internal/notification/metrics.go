package notification

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type NotificationMetrics struct {
	QueueDepth       prometheus.Gauge
	QueueCapacity    prometheus.Gauge
	QueueEnqueued    prometheus.Counter
	QueueDropped     prometheus.Counter
	QueueSaturation  prometheus.Gauge
	DispatchDuration prometheus.Histogram
	DispatchTotal    *prometheus.CounterVec
	DispatchErrors   *prometheus.CounterVec
	ChannelDuration  *prometheus.HistogramVec
	ChannelTotal     *prometheus.CounterVec
	ChannelErrors    *prometheus.CounterVec
	RetryTotal       *prometheus.CounterVec
	EmailLatency     prometheus.Histogram
	EmailTotal       *prometheus.CounterVec
	EmailErrors      *prometheus.CounterVec
	WebhookLatency   prometheus.Histogram
	WebhookTotal     *prometheus.CounterVec
	WebhookErrors    *prometheus.CounterVec

	registry prometheus.Registerer
	mu        sync.RWMutex
}

func NewNotificationMetrics() *NotificationMetrics {
	return &NotificationMetrics{
		QueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "queue_depth",
			Help:      "Current number of notifications in the queue",
		}),
		QueueCapacity: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "queue_capacity",
			Help:      "Maximum queue capacity",
		}),
		QueueEnqueued: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "queue_enqueued_total",
			Help:      "Total number of notifications enqueued",
		}),
		QueueDropped: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "queue_dropped_total",
			Help:      "Total number of notifications dropped due to queue saturation",
		}),
		QueueSaturation: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "queue_saturation_percent",
			Help:      "Current queue saturation percentage",
		}),
		DispatchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "dispatch_duration_seconds",
			Help:      "Duration of notification dispatch",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		DispatchTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "dispatch_total",
			Help:      "Total number of notification dispatches",
		}, []string{"status"}),
		DispatchErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "dispatch_errors_total",
			Help:      "Total number of notification dispatch errors",
		}, []string{"error_type"}),
		ChannelDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "channel_duration_seconds",
			Help:      "Duration of channel send operations",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		}, []string{"channel"}),
		ChannelTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "channel_total",
			Help:      "Total number of channel send operations",
		}, []string{"channel", "status"}),
		ChannelErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "channel_errors_total",
			Help:      "Total number of channel send errors",
		}, []string{"channel", "error_type"}),
		RetryTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "notification",
			Name:      "retry_total",
			Help:      "Total number of retry attempts",
		}, []string{"channel", "status"}),
		EmailLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "email",
			Name:      "send_duration_seconds",
			Help:      "Duration of email send operations",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		}),
		EmailTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "email",
			Name:      "send_total",
			Help:      "Total number of email send operations",
		}, []string{"status"}),
		EmailErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "email",
			Name:      "send_errors_total",
			Help:      "Total number of email send errors",
		}, []string{"error_type"}),
		WebhookLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "webhook",
			Name:      "send_duration_seconds",
			Help:      "Duration of webhook send operations",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		}),
		WebhookTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "webhook",
			Name:      "send_total",
			Help:      "Total number of webhook send operations",
		}, []string{"status"}),
		WebhookErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "webhook",
			Name:      "send_errors_total",
			Help:      "Total number of webhook send errors",
		}, []string{"error_type"}),
	}
}

func (m *NotificationMetrics) Register(r prometheus.Registerer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.registry = r

	metrics := []prometheus.Collector{
		m.QueueDepth,
		m.QueueCapacity,
		m.QueueEnqueued,
		m.QueueDropped,
		m.QueueSaturation,
		m.DispatchDuration,
		m.DispatchTotal,
		m.DispatchErrors,
		m.ChannelDuration,
		m.ChannelTotal,
		m.ChannelErrors,
		m.RetryTotal,
		m.EmailLatency,
		m.EmailTotal,
		m.EmailErrors,
		m.WebhookLatency,
		m.WebhookTotal,
		m.WebhookErrors,
	}

	for _, metric := range metrics {
		if err := r.Register(metric); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
				continue
			}
			return err
		}
	}

	m.QueueCapacity.Set(1000)
	return nil
}

func (m *NotificationMetrics) Unregister() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.registry == nil {
		return
	}

	m.registry.Unregister(m.QueueDepth)
	m.registry.Unregister(m.QueueCapacity)
	m.registry.Unregister(m.QueueEnqueued)
	m.registry.Unregister(m.QueueDropped)
	m.registry.Unregister(m.QueueSaturation)
	m.registry.Unregister(m.DispatchDuration)
	m.registry.Unregister(m.DispatchTotal)
	m.registry.Unregister(m.DispatchErrors)
	m.registry.Unregister(m.ChannelDuration)
	m.registry.Unregister(m.ChannelTotal)
	m.registry.Unregister(m.ChannelErrors)
	m.registry.Unregister(m.RetryTotal)
	m.registry.Unregister(m.EmailLatency)
	m.registry.Unregister(m.EmailTotal)
	m.registry.Unregister(m.EmailErrors)
	m.registry.Unregister(m.WebhookLatency)
	m.registry.Unregister(m.WebhookTotal)
	m.registry.Unregister(m.WebhookErrors)
}

func (m *NotificationMetrics) RecordQueueDepth(depth int) {
	m.QueueDepth.Set(float64(depth))
}

func (m *NotificationMetrics) RecordQueueSaturation(saturation float64) {
	m.QueueSaturation.Set(saturation)
}

func (m *NotificationMetrics) RecordEnqueue() {
	m.QueueEnqueued.Inc()
}

func (m *NotificationMetrics) RecordDropped() {
	m.QueueDropped.Inc()
}

func (m *NotificationMetrics) RecordDispatch(duration time.Duration, success bool) {
	m.DispatchDuration.Observe(duration.Seconds())
	if success {
		m.DispatchTotal.WithLabelValues("success").Inc()
	} else {
		m.DispatchTotal.WithLabelValues("failure").Inc()
	}
}

func (m *NotificationMetrics) RecordDispatchError(err error) {
	m.DispatchErrors.WithLabelValues(classifyError(err)).Inc()
}

func (m *NotificationMetrics) RecordChannelDuration(channel string, duration time.Duration) {
	m.ChannelDuration.WithLabelValues(channel).Observe(duration.Seconds())
}

func (m *NotificationMetrics) RecordChannelResult(channel, status string) {
	m.ChannelTotal.WithLabelValues(channel, status).Inc()
}

func (m *NotificationMetrics) RecordChannelError(channel, errType string) {
	m.ChannelErrors.WithLabelValues(channel, errType).Inc()
}

func (m *NotificationMetrics) RecordRetry(channel, status string) {
	m.RetryTotal.WithLabelValues(channel, status).Inc()
}

func (m *NotificationMetrics) RecordEmailLatency(duration time.Duration) {
	m.EmailLatency.Observe(duration.Seconds())
}

func (m *NotificationMetrics) RecordEmailResult(status string) {
	m.EmailTotal.WithLabelValues(status).Inc()
}

func (m *NotificationMetrics) RecordEmailError(errType string) {
	m.EmailErrors.WithLabelValues(errType).Inc()
}

func (m *NotificationMetrics) RecordWebhookLatency(duration time.Duration) {
	m.WebhookLatency.Observe(duration.Seconds())
}

func (m *NotificationMetrics) RecordWebhookResult(status string) {
	m.WebhookTotal.WithLabelValues(status).Inc()
}

func (m *NotificationMetrics) RecordWebhookError(errType string) {
	m.WebhookErrors.WithLabelValues(errType).Inc()
}

func classifyError(err error) string {
	if err == nil {
		return "none"
	}
	msg := err.Error()
	switch {
	case contains(msg, "connection") || contains(msg, "dial"):
		return "connection_error"
	case contains(msg, "timeout"):
		return "timeout"
	case contains(msg, "rate limit") || contains(msg, "429"):
		return "rate_limited"
	case contains(msg, "500") || contains(msg, "503"):
		return "server_error"
	case contains(msg, "401") || contains(msg, "unauthorized"):
		return "unauthorized"
	case contains(msg, "403") || contains(msg, "forbidden"):
		return "forbidden"
	case contains(msg, "404") || contains(msg, "not found"):
		return "not_found"
	default:
		return "unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

var globalNotificationMetrics *NotificationMetrics
var metricsOnce sync.Once

func GetNotificationMetrics() *NotificationMetrics {
	metricsOnce.Do(func() {
		globalNotificationMetrics = NewNotificationMetrics()
	})
	return globalNotificationMetrics
}

func RegisterNotificationMetrics(r prometheus.Registerer) error {
	return GetNotificationMetrics().Register(r)
}
