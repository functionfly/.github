// Simple monitoring and alerting system
// In production, integrate with services like DataDog, New Relic, or Sentry

interface Metric {
  name: string
  value: number
  timestamp: number
  tags?: Record<string, string>
}

interface Alert {
  id: string
  level: 'info' | 'warning' | 'error' | 'critical'
  message: string
  timestamp: number
  resolved?: boolean
  resolvedAt?: number
}

class Monitor {
  private metrics: Metric[] = []
  private alerts: Alert[] = []
  private maxMetrics = 1000
  private maxAlerts = 100

  // Record a metric
  recordMetric(name: string, value: number, tags?: Record<string, string>) {
    const metric: Metric = {
      name,
      value,
      timestamp: Date.now(),
      tags,
    }

    this.metrics.push(metric)

    // Keep only recent metrics
    if (this.metrics.length > this.maxMetrics) {
      this.metrics = this.metrics.slice(-this.maxMetrics)
    }

    // Log significant metrics
    if (value > 1000 || name.includes('error')) {
      console.log(`📊 Metric: ${name} = ${value}`, tags)
    }
  }

  // Record an event
  recordEvent(event: string, data?: any) {
    console.log(`📝 Event: ${event}`, data)
    this.recordMetric(`event.${event}`, 1, data)
  }

  // Create an alert
  alert(level: Alert['level'], message: string, data?: any) {
    const alert: Alert = {
      id: `alert_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      level,
      message,
      timestamp: Date.now(),
    }

    this.alerts.push(alert)

    // Keep only recent alerts
    if (this.alerts.length > this.maxAlerts) {
      this.alerts = this.alerts.slice(-this.maxAlerts)
    }

    // Log alerts
    const emoji = {
      info: 'ℹ️',
      warning: '⚠️',
      error: '❌',
      critical: '🚨',
    }[level]

    console.log(`${emoji} Alert [${level.toUpperCase()}]: ${message}`, data)

    // In production, send to external monitoring service
    // this.sendToMonitoringService(alert, data)
  }

  // Get recent metrics
  getMetrics(since?: number): Metric[] {
    if (!since) return this.metrics

    return this.metrics.filter(m => m.timestamp >= since)
  }

  // Get active alerts
  getActiveAlerts(): Alert[] {
    return this.alerts.filter(a => !a.resolved)
  }

  // Resolve an alert
  resolveAlert(alertId: string) {
    const alert = this.alerts.find(a => a.id === alertId)
    if (alert && !alert.resolved) {
      alert.resolved = true
      alert.resolvedAt = Date.now()
      console.log(`✅ Alert resolved: ${alert.message}`)
    }
  }

  // Health check
  getHealthStatus() {
    const activeAlerts = this.getActiveAlerts()
    const criticalAlerts = activeAlerts.filter(a => a.level === 'critical')
    const errorAlerts = activeAlerts.filter(a => a.level === 'error')

    return {
      status: criticalAlerts.length > 0 ? 'critical' :
              errorAlerts.length > 0 ? 'error' :
              activeAlerts.length > 0 ? 'warning' : 'healthy',
      activeAlerts: activeAlerts.length,
      metrics: this.metrics.length,
      lastMetric: this.metrics[this.metrics.length - 1]?.timestamp,
    }
  }
}

// Global monitor instance
export const monitor = new Monitor()

// Convenience functions
export const recordMetric = (name: string, value: number, tags?: Record<string, string>) =>
  monitor.recordMetric(name, value, tags)

export const recordEvent = (event: string, data?: any) =>
  monitor.recordEvent(event, data)

export const alert = (level: Alert['level'], message: string, data?: any) =>
  monitor.alert(level, message, data)

export const getHealthStatus = () => monitor.getHealthStatus()