'use client';

import { useUniversalRuntimeStore } from '@/stores/universalRuntimeStore';
import { cn } from '@/lib/utils';
import {
  Activity,
  AlertTriangle,
  Box,
  Cloud,
  Cpu,
  Gauge,
  Globe,
  Layers,
  Network,
  Settings,
  Zap,
} from 'lucide-react';

interface UniversalRuntimeIntegrationProps {
  className?: string;
}

export function UniversalRuntimeIntegration({ className }: UniversalRuntimeIntegrationProps) {
  const {
    runtimes,
    activeView,
    executionMode,
    metrics,
    alerts,
    cloudProviders,
    capabilities,
    selectedRuntimeId,
    selectRuntime,
    dismissAlert,
  } = useUniversalRuntimeStore();

  const activeAlerts = alerts.filter((a) => !a.dismissed);

  return (
    <div className={cn('space-y-4', className)}>
      <div className="aviation-panel p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-aviation-text-secondary">Runtime Overview</h3>
          <span className="aviation-badge">{runtimes.length} Runtimes</span>
        </div>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          <div className="aviation-stat">
            <Box className="w-4 h-4 text-aviation-cyan mb-1" />
            <span className="aviation-stat-label">WASM</span>
            <span className="aviation-stat-value">
              {runtimes.filter((r) => r.type === 'wasm').length}
            </span>
          </div>
          <div className="aviation-stat">
            <Cpu className="w-4 h-4 text-aviation-amber mb-1" />
            <span className="aviation-stat-label">GPU</span>
            <span className="aviation-stat-value">
              {runtimes.filter((r) => r.type === 'gpu').length}
            </span>
          </div>
          <div className="aviation-stat">
            <Zap className="w-4 h-4 text-aviation-green mb-1" />
            <span className="aviation-stat-label">Serverless</span>
            <span className="aviation-stat-value">
              {runtimes.filter((r) => r.type === 'serverless').length}
            </span>
          </div>
          <div className="aviation-stat">
            <Globe className="w-4 h-4 text-aviation-purple mb-1" />
            <span className="aviation-stat-label">Edge</span>
            <span className="aviation-stat-value">
              {runtimes.filter((r) => r.type === 'edge').length}
            </span>
          </div>
        </div>
      </div>

      <div className="aviation-panel p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-aviation-text-secondary">Performance Metrics</h3>
          <Activity className="w-4 h-4 text-aviation-text-muted" />
        </div>
        <div className="grid grid-cols-3 gap-4">
          <div className="aviation-metric-card">
            <span className="aviation-metric-label">Throughput</span>
            <span className="aviation-metric-value">{metrics.throughput.toFixed(1)}</span>
            <span className="aviation-metric-unit">req/s</span>
          </div>
          <div className="aviation-metric-card">
            <span className="aviation-metric-label">Avg Latency</span>
            <span className="aviation-metric-value">{metrics.averageLatency.toFixed(0)}</span>
            <span className="aviation-metric-unit">ms</span>
          </div>
          <div className="aviation-metric-card">
            <span className="aviation-metric-label">Success Rate</span>
            <span className="aviation-metric-value">
              {metrics.totalRequests > 0
                ? ((metrics.successfulRequests / metrics.totalRequests) * 100).toFixed(1)
                : '0.0'}
            </span>
            <span className="aviation-metric-unit">%</span>
          </div>
        </div>
      </div>

      <div className="aviation-panel p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-aviation-text-secondary">Cloud Providers</h3>
          <Cloud className="w-4 h-4 text-aviation-text-muted" />
        </div>
        {cloudProviders.length > 0 ? (
          <div className="space-y-2">
            {cloudProviders.map((provider) => (
              <div
                key={provider.id}
                className={cn(
                  'aviation-list-item',
                  provider.status === 'connected' && 'border-l-2 border-l-aviation-green'
                )}
              >
                <span className="font-medium">{provider.name}</span>
                <span className="text-sm text-aviation-text-muted">{provider.region}</span>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-aviation-text-muted">No cloud providers configured</p>
        )}
      </div>

      <div className="aviation-panel p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-aviation-text-secondary">Alerts</h3>
          {activeAlerts.length > 0 && (
            <span className="aviation-badge-error">{activeAlerts.length}</span>
          )}
        </div>
        {activeAlerts.length > 0 ? (
          <div className="space-y-2">
            {activeAlerts.map((alert) => (
              <div
                key={alert.id}
                className={cn(
                  'aviation-alert',
                  alert.severity === 'critical' && 'border-l-4 border-l-aviation-red',
                  alert.severity === 'warning' && 'border-l-4 border-l-aviation-amber'
                )}
              >
                <AlertTriangle className="w-4 h-4 flex-shrink-0" />
                <span className="flex-1 text-sm">{alert.message}</span>
                <button
                  onClick={() => dismissAlert(alert.id)}
                  className="aviation-btn-ghost text-xs"
                >
                  Dismiss
                </button>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-aviation-text-muted">No active alerts</p>
        )}
      </div>

      <div className="aviation-panel p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-aviation-text-secondary">Capabilities</h3>
          <Layers className="w-4 h-4 text-aviation-text-muted" />
        </div>
        <div className="grid grid-cols-2 gap-2">
          {capabilities.length > 0 ? (
            capabilities.map((cap) => (
              <div
                key={cap.id}
                className={cn(
                  'aviation-list-item text-xs',
                  cap.supported ? 'text-aviation-green' : 'text-aviation-text-muted'
                )}
              >
                {cap.name}
                {cap.performance && (
                  <span className="text-[10px] ml-1">({cap.performance})</span>
                )}
              </div>
            ))
          ) : (
            <p className="text-sm text-aviation-text-muted col-span-2">No capabilities loaded</p>
          )}
        </div>
      </div>
    </div>
  );
}

export default UniversalRuntimeIntegration;
