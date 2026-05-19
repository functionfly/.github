import React, { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { Activity, Zap, Clock, Database, Globe, AlertCircle } from 'lucide-react'

export interface TelemetryMetric {
  id: string
  label: string
  value: string | number
  unit?: string
  trend?: 'up' | 'down' | 'stable'
  icon?: React.ReactNode
  status?: 'success' | 'warning' | 'error' | 'info'
}

export interface BottomTelemetryBarProps {
  metrics?: TelemetryMetric[]
  className?: string
  showTimestamp?: boolean
}

const defaultMetrics: TelemetryMetric[] = [
  { id: 'latency', label: 'Latency', value: '42', unit: 'ms', icon: <Clock className="w-3 h-3" />, status: 'success' },
  { id: 'requests', label: 'Requests', value: '1.2K', unit: '/hr', icon: <Activity className="w-3 h-3" />, status: 'info' },
  { id: 'cost', label: 'Cost', value: '$24.50', icon: <Zap className="w-3 h-3" />, status: 'warning' },
  { id: 'uptime', label: 'Uptime', value: '99.9', unit: '%', icon: <Globe className="w-3 h-3" />, status: 'success' },
]

/**
 * BottomTelemetryBar - Status bar showing telemetry metrics
 * Displays live metrics at the bottom of the workspace
 */
export function BottomTelemetryBar({
  metrics = defaultMetrics,
  className,
  showTimestamp = true,
}: BottomTelemetryBarProps) {
  const [timestamp, setTimestamp] = useState(new Date())

  useEffect(() => {
    const interval = setInterval(() => {
      setTimestamp(new Date())
    }, 1000)
    return () => clearInterval(interval)
  }, [])

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'success': return 'text-aviation-green'
      case 'warning': return 'text-aviation-amber'
      case 'error': return 'text-aviation-red'
      default: return 'text-aviation-cyan'
    }
  }

  return (
    <div
      className={cn(
        'h-8 px-4 border-t border-aviation-border-panel',
        'bg-aviation-bg-secondary/50 backdrop-blur-sm',
        'flex items-center justify-between text-xs',
        className
      )}
    >
      {/* Metrics */}
      <div className="flex items-center gap-4">
        {metrics.map(metric => (
          <div key={metric.id} className="flex items-center gap-1.5">
            {metric.icon && (
              <span className={cn(getStatusColor(metric.status))}>
                {metric.icon}
              </span>
            )}
            <span className="text-aviation-text-muted">{metric.label}:</span>
            <span className={cn('font-mono font-medium', getStatusColor(metric.status))}>
              {metric.value}{metric.unit && <span className="text-aviation-text-dim">{metric.unit}</span>}
            </span>
          </div>
        ))}
      </div>

      {/* Timestamp */}
      {showTimestamp && (
        <div className="text-aviation-text-muted font-mono">
          {timestamp.toLocaleTimeString()}
        </div>
      )}
    </div>
  )
}
