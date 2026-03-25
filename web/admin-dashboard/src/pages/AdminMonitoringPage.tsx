import type { RealtimeEvent } from '@/hooks/useRealtimeSubscription';
import { useStatusWebSocket } from '@/hooks/useStatusWebSocket';
import { adminApiClient } from '@/lib/api/adminClient';
import { publicApiClient } from '@/lib/api/publicClient';
import { API_ROUTES } from '@/lib/constants';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Activity,
  AlertTriangle,
  Bell,
  CheckCircle,
  Plus,
  Server,
  Wifi,
  WifiOff,
  XCircle,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';

// Types
interface Alert {
  id: string;
  name: string;
  condition: string;
  severity: 'critical' | 'warning' | 'info';
  enabled: boolean;
  created_at: string;
}

interface Metric {
  name: string;
  value: number;
  unit: string;
  timestamp: string;
}

interface HealthCheck {
  service: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  latency_ms: number;
  last_check: string;
  message?: string;
}

interface MonitoringData {
  alerts: Alert[];
  metrics: Metric[];
  health_checks: HealthCheck[];
}

/**
 * Admin Monitoring Page
 * Manages alerts, metrics, and health checks
 */
export function AdminMonitoringPage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<'alerts' | 'metrics' | 'health'>('alerts');

  // Fetch monitoring data
  const { data, isLoading } = useQuery<MonitoringData>({
    queryKey: ['monitoring-data'],
    queryFn: async () => {
      const [alertsRes, metricsRes, healthRes] = await Promise.all([
        publicApiClient.get<{ alerts: Alert[] }>(API_ROUTES.ADMIN_MONITORING_ALERTS),
        publicApiClient.get<{ metrics: Metric[] }>(API_ROUTES.ADMIN_MONITORING_METRICS),
        publicApiClient.get<{ checks: HealthCheck[] }>(API_ROUTES.ADMIN_MONITORING_HEALTH),
      ]);
      return {
        alerts: alertsRes.data?.alerts || [],
        metrics: metricsRes.data?.metrics || [],
        health_checks: healthRes.data?.checks || [],
      };
    },
    refetchInterval: 30000,
  });

  // Toggle alert mutation
  const toggleAlertMutation = useMutation({
    mutationFn: async ({ id, enabled }: { id: string; enabled: boolean }) => {
      return adminApiClient.patch(`/monitoring/alerts/${id}`, { enabled });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitoring-data'] });
    },
  });

  // WebSocket connection status
  const [isRealtimeConnected, setIsRealtimeConnected] = useState(false);

  // Handle real-time status updates
  const handleStatusUpdate = useCallback(
    (event: RealtimeEvent) => {
      // Update health checks in real-time
      if (event.type === 'health_update' && event.checks) {
        queryClient.setQueryData<MonitoringData>(['monitoring-data'], (old) => {
          if (!old) return old;
          return { ...old, health_checks: event.checks as HealthCheck[] };
        });
      }
    },
    [queryClient]
  );

  // Handle real-time incident updates
  const handleIncidentUpdate = useCallback(
    (_event: RealtimeEvent) => {
      // Invalidate to refetch when incidents change
      queryClient.invalidateQueries({ queryKey: ['monitoring-data'] });
    },
    [queryClient]
  );

  // Set up WebSocket for real-time updates
  const { isConnected, error: wsError } = useStatusWebSocket({
    enabled: true,
    onStatusUpdate: handleStatusUpdate,
    onIncidentUpdate: handleIncidentUpdate,
  });

  // Track WebSocket connection state
  useEffect(() => {
    setIsRealtimeConnected(isConnected);
  }, [isConnected]);

  // Auto-refresh data when WebSocket disconnects (fallback to polling).
  // Only fire after the connection was previously established to avoid
  // invalidating on the initial render before the socket has ever connected.
  const wasConnectedRef = useRef(false);
  useEffect(() => {
    if (isRealtimeConnected) {
      wasConnectedRef.current = true;
    } else if (wasConnectedRef.current && !wsError) {
      queryClient.invalidateQueries({ queryKey: ['monitoring-data'] });
    }
  }, [isRealtimeConnected, wsError, queryClient]);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-blue-100 text-blue-800';
    }
  };

  const getHealthIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'degraded':
        return <AlertTriangle className="w-5 h-5 text-yellow-500" />;
      default:
        return <XCircle className="w-5 h-5 text-red-500" />;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Monitoring</h1>
          <p className="text-gray-600 mt-1">Manage alerts, metrics, and health checks</p>
        </div>
        {/* Real-time connection status */}
        <div className="flex items-center gap-2">
          {isRealtimeConnected ? (
            <span className="flex items-center gap-1 text-sm text-green-600">
              <Wifi className="w-4 h-4" />
              Live
            </span>
          ) : (
            <span className="flex items-center gap-1 text-sm text-gray-500">
              <WifiOff className="w-4 h-4" />
              Polling
            </span>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('alerts')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'alerts'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Bell className="w-4 h-4 inline mr-2" />
            Alerts
          </button>
          <button
            onClick={() => setActiveTab('metrics')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'metrics'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Activity className="w-4 h-4 inline mr-2" />
            Metrics
          </button>
          <button
            onClick={() => setActiveTab('health')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'health'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Server className="w-4 h-4 inline mr-2" />
            Health Checks
          </button>
        </nav>
      </div>

      {/* Alerts Tab */}
      {activeTab === 'alerts' && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
              <Plus className="w-4 h-4" />
              Create Alert
            </button>
          </div>

          <div className="bg-white rounded-lg shadow overflow-hidden">
            {isLoading ? (
              <div className="p-8 text-center text-gray-500">Loading...</div>
            ) : data?.alerts?.length === 0 ? (
              <div className="p-8 text-center text-gray-500">No alerts configured</div>
            ) : (
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Condition
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Severity
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Created
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {data?.alerts?.map((alert) => (
                    <tr key={alert.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {alert.name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {alert.condition}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getSeverityColor(alert.severity)}`}
                        >
                          {alert.severity}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <button
                          onClick={() =>
                            toggleAlertMutation.mutate({ id: alert.id, enabled: !alert.enabled })
                          }
                          className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${alert.enabled ? 'bg-green-600' : 'bg-gray-200'}`}
                        >
                          <span
                            className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${alert.enabled ? 'translate-x-5' : 'translate-x-0'}`}
                          />
                        </button>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(alert.created_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Metrics Tab */}
      {activeTab === 'metrics' && (
        <div className="bg-white rounded-lg shadow p-6">
          {isLoading ? (
            <div className="text-center text-gray-500">Loading...</div>
          ) : data?.metrics?.length === 0 ? (
            <div className="text-center text-gray-500">No metrics available</div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {data?.metrics?.map((metric, idx) => (
                <div key={idx} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-medium text-gray-900">{metric.name}</h3>
                    <Activity className="w-4 h-4 text-gray-400" />
                  </div>
                  <div className="mt-2">
                    <span className="text-2xl font-bold">{metric.value}</span>
                    <span className="text-sm text-gray-500 ml-1">{metric.unit}</span>
                  </div>
                  <div className="mt-2 text-xs text-gray-500">
                    Updated: {new Date(metric.timestamp).toLocaleString()}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Health Checks Tab */}
      {activeTab === 'health' && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          {isLoading ? (
            <div className="p-8 text-center text-gray-500">Loading...</div>
          ) : data?.health_checks?.length === 0 ? (
            <div className="p-8 text-center text-gray-500">No health checks available</div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Service
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Latency
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Last Check
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Message
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {data?.health_checks?.map((check) => (
                  <tr key={check.service}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {check.service}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        {getHealthIcon(check.status)}
                        <span className="text-sm text-gray-900 capitalize">{check.status}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {check.latency_ms}ms
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(check.last_check).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {check.message || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}

export default AdminMonitoringPage;
