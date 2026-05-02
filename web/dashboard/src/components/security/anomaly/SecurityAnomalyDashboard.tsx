/**
 * SecurityAnomalyDashboard Component
 * Real-time security anomaly monitoring and alerting for user dashboard
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Shield, AlertTriangle, AlertCircle, CheckCircle, Clock,
  RefreshCw, Filter, Eye, EyeOff, Bell, X, ChevronRight
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

interface SecurityAnomaly {
  id: string;
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  detectedAt: string;
  status: 'active' | 'acknowledged' | 'resolved' | 'dismissed';
  source?: string;
  metadata?: Record<string, unknown>;
}

interface SecurityMetrics {
  totalAnomalies: number;
  activeAnomalies: number;
  resolvedToday: number;
  avgResponseTime: number;
}

interface SecurityAnomalyDashboardProps {
  tenantId?: string;
}

export function SecurityAnomalyDashboard({ tenantId }: SecurityAnomalyDashboardProps) {
  const [anomalies, setAnomalies] = useState<SecurityAnomaly[]>([]);
  const [metrics, setMetrics] = useState<SecurityMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [severityFilter, setSeverityFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [showAcknowledged, setShowAcknowledged] = useState(false);

  const fetchAnomalies = useCallback(async () => {
    setIsLoading(true);
    try {
      // Placeholder - would call actual API
      // In real implementation, this would use security hooks
      setMetrics({
        totalAnomalies: 0,
        activeAnomalies: 0,
        resolvedToday: 0,
        avgResponseTime: 0,
      });
    } catch (err) {
      console.error('Failed to fetch security anomalies:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAnomalies();
    const interval = setInterval(fetchAnomalies, 30000);
    return () => clearInterval(interval);
  }, [fetchAnomalies]);

  const filteredAnomalies = anomalies.filter(anomaly => {
    if (severityFilter !== 'all' && anomaly.severity !== severityFilter) return false;
    if (statusFilter !== 'all' && anomaly.status !== statusFilter) return false;
    if (!showAcknowledged && anomaly.status === 'acknowledged') return false;
    return true;
  });

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-600 text-white';
      case 'high':
        return 'bg-orange-500 text-white';
      case 'medium':
        return 'bg-yellow-500 text-black';
      case 'low':
        return 'bg-blue-500 text-white';
      default:
        return 'bg-gray-500 text-white';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
      case 'high':
        return <AlertTriangle className="h-4 w-4" />;
      case 'medium':
        return <AlertCircle className="h-4 w-4" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300';
      case 'acknowledged':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300';
      case 'resolved':
        return 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300';
      case 'dismissed':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900/50 dark:text-gray-300';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const acknowledgeAnomaly = (id: string) => {
    setAnomalies(prev =>
      prev.map(a => a.id === id ? { ...a, status: 'acknowledged' as const } : a)
    );
  };

  const resolveAnomaly = (id: string) => {
    setAnomalies(prev =>
      prev.map(a => a.id === id ? { ...a, status: 'resolved' as const } : a)
    );
  };

  const dismissAnomaly = (id: string) => {
    setAnomalies(prev =>
      prev.map(a => a.id === id ? { ...a, status: 'dismissed' as const } : a)
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Shield className="h-5 w-5 text-emerald-600" />
              <CardTitle>Security Anomaly Dashboard</CardTitle>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={fetchAnomalies}
              disabled={isLoading}
              className="flex items-center gap-2"
            >
              <RefreshCw className={cn('h-4 w-4', isLoading && 'animate-spin')} />
              Refresh
            </Button>
          </div>
          <CardDescription>
            Monitor and respond to security anomalies in real-time
          </CardDescription>
        </CardHeader>
      </Card>

      {/* Metrics Overview */}
      {metrics && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Total Anomalies</p>
                  <p className="text-2xl font-bold">{metrics.totalAnomalies}</p>
                </div>
                <Shield className="h-8 w-8 text-gray-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Active</p>
                  <p className="text-2xl font-bold text-red-600">{metrics.activeAnomalies}</p>
                </div>
                <AlertTriangle className="h-8 w-8 text-red-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Resolved Today</p>
                  <p className="text-2xl font-bold text-green-600">{metrics.resolvedToday}</p>
                </div>
                <CheckCircle className="h-8 w-8 text-green-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Avg Response Time</p>
                  <p className="text-2xl font-bold">{metrics.avgResponseTime}m</p>
                </div>
                <Clock className="h-8 w-8 text-blue-400" />
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center gap-4 flex-wrap">
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">Severity:</span>
              <select
                value={severityFilter}
                onChange={(e) => setSeverityFilter(e.target.value)}
                className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-lg"
              >
                <option value="all">All</option>
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>

            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">Status:</span>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-lg"
              >
                <option value="all">All</option>
                <option value="active">Active</option>
                <option value="acknowledged">Acknowledged</option>
                <option value="resolved">Resolved</option>
              </select>
            </div>

            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowAcknowledged(!showAcknowledged)}
              className={cn('flex items-center gap-2', showAcknowledged && 'bg-blue-50 border-blue-300')}
            >
              {showAcknowledged ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
              {showAcknowledged ? 'Showing' : 'Hiding'} Acknowledged
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Anomaly List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            Anomalies ({filteredAnomalies.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {filteredAnomalies.length === 0 ? (
            <div className="text-center py-12">
              <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
              <h3 className="text-lg font-medium">All Clear</h3>
              <p className="text-gray-500 dark:text-gray-400">
                No security anomalies matching your filters
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredAnomalies.map((anomaly) => (
                <div
                  key={anomaly.id}
                  className={cn(
                    'p-4 border rounded-lg transition-colors',
                    anomaly.status === 'active'
                      ? 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10'
                      : 'border-gray-200 dark:border-gray-700'
                  )}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <div className={cn(
                        'p-2 rounded-lg',
                        anomaly.severity === 'critical' || anomaly.severity === 'high'
                          ? 'bg-red-100 dark:bg-red-900/50'
                          : anomaly.severity === 'medium'
                          ? 'bg-yellow-100 dark:bg-yellow-900/50'
                          : 'bg-blue-100 dark:bg-blue-900/50'
                      )}>
                        {getSeverityIcon(anomaly.severity)}
                      </div>
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <Badge className={getSeverityColor(anomaly.severity)}>
                            {anomaly.severity}
                          </Badge>
                          <span className="font-medium">{anomaly.type}</span>
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                          {anomaly.description}
                        </p>
                        <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
                          <span>Detected: {new Date(anomaly.detectedAt).toLocaleString()}</span>
                          {anomaly.source && <span>Source: {anomaly.source}</span>}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <Badge className={getStatusColor(anomaly.status)}>
                        {anomaly.status}
                      </Badge>

                      {anomaly.status === 'active' && (
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => acknowledgeAnomaly(anomaly.id)}
                          >
                            Acknowledge
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => resolveAnomaly(anomaly.id)}
                            className="text-green-600 hover:text-green-700"
                          >
                            Resolve
                          </Button>
                        </div>
                      )}

                      {anomaly.status === 'acknowledged' && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => resolveAnomaly(anomaly.id)}
                          className="text-green-600 hover:text-green-700"
                        >
                          Resolve
                        </Button>
                      )}

                      {anomaly.status !== 'resolved' && anomaly.status !== 'dismissed' && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => dismissAnomaly(anomaly.id)}
                          className="text-gray-500 hover:text-gray-600"
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default SecurityAnomalyDashboard;
