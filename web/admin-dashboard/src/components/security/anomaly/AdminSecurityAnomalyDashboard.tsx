/**
 * AdminSecurityAnomalyDashboard Component
 * Platform-level security anomaly monitoring and alerting for admin dashboard
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Shield, AlertTriangle, AlertCircle, CheckCircle, Clock,
  RefreshCw, Filter, Eye, EyeOff, Bell, X, ChevronRight, Users, Activity
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { useToastHelpers } from '@/components/ui/Toast';
import clsx from 'clsx';

interface SecurityAnomaly {
  id: string;
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  detectedAt: string;
  status: 'active' | 'acknowledged' | 'resolved' | 'dismissed';
  tenantId?: string;
  tenantName?: string;
  agentId?: string;
  source?: string;
  metadata?: Record<string, unknown>;
}

interface SecurityMetrics {
  totalAnomalies: number;
  activeAnomalies: number;
  resolvedToday: number;
  avgResponseTime: number;
  criticalAlerts: number;
  affectedTenants: number;
}

interface AdminSecurityAnomalyDashboardProps {
  platformWide?: boolean;
}

export function AdminSecurityAnomalyDashboard({ platformWide = true }: AdminSecurityAnomalyDashboardProps) {
  const queryClient = useQueryClient();
  const toast = useToastHelpers();
  
  const [anomalies, setAnomalies] = useState<SecurityAnomaly[]>([]);
  const [metrics, setMetrics] = useState<SecurityMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [severityFilter, setSeverityFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [showAcknowledged, setShowAcknowledged] = useState(false);
  const [expandedAnomaly, setExpandedAnomaly] = useState<string | null>(null);

  const fetchAnomalies = useCallback(async () => {
    setIsLoading(true);
    try {
      // In real implementation, call admin API
      // GET /v1/admin/security/anomalies
      setMetrics({
        totalAnomalies: 0,
        activeAnomalies: 0,
        resolvedToday: 0,
        avgResponseTime: 0,
        criticalAlerts: 0,
        affectedTenants: 0,
      });
    } catch (err) {
      console.error('Failed to fetch security anomalies:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAnomalies();
    const interval = setInterval(fetchAnomalies, 15000);
    return () => clearInterval(interval);
  }, [fetchAnomalies]);

  // Acknowledge anomaly mutation
  const acknowledgeMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/security/anomalies/${id}/acknowledge`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-security-anomalies'] });
      toast.success('Anomaly acknowledged');
    },
    onError: (error: Error) => {
      toast.error(`Failed to acknowledge: ${error.message}`);
    },
  });

  // Resolve anomaly mutation
  const resolveMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/security/anomalies/${id}/resolve`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-security-anomalies'] });
      toast.success('Anomaly resolved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resolve: ${error.message}`);
    },
  });

  // Block tenant mutation
  const blockTenantMutation = useMutation({
    mutationFn: (tenantId: string) => adminApiClient.post(`/tenants/${tenantId}/block`),
    onSuccess: () => {
      toast.success('Tenant blocked due to security anomaly');
    },
    onError: (error: Error) => {
      toast.error(`Failed to block tenant: ${error.message}`);
    },
  });

  const filteredAnomalies = anomalies.filter(anomaly => {
    if (severityFilter !== 'all' && anomaly.severity !== severityFilter) return false;
    if (statusFilter !== 'all' && anomaly.status !== statusFilter) return false;
    if (!showAcknowledged && anomaly.status === 'acknowledged') return false;
    return true;
  });

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-600 text-white';
      case 'high': return 'bg-orange-500 text-white';
      case 'medium': return 'bg-yellow-500 text-black';
      case 'low': return 'bg-blue-500 text-white';
      default: return 'bg-gray-500 text-white';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300';
      case 'acknowledged': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300';
      case 'resolved': return 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300';
      case 'dismissed': return 'bg-gray-100 text-gray-800 dark:bg-gray-900/50 dark:text-gray-300';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const handleAcknowledge = (id: string) => {
    setAnomalies(prev =>
      prev.map(a => a.id === id ? { ...a, status: 'acknowledged' as const } : a)
    );
    acknowledgeMutation.mutate(id);
  };

  const handleResolve = (id: string) => {
    setAnomalies(prev =>
      prev.map(a => a.id === id ? { ...a, status: 'resolved' as const } : a)
    );
    resolveMutation.mutate(id);
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
              <RefreshCw className={clsx('h-4 w-4', isLoading && 'animate-spin')} />
              Refresh
            </Button>
          </div>
          <CardDescription>
            {platformWide
              ? 'Platform-wide security anomaly monitoring and response'
              : 'Tenant-level security anomaly monitoring'}
          </CardDescription>
        </CardHeader>
      </Card>

      {/* Metrics Overview */}
      {metrics && (
        <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
          <Card>
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500">Total</p>
                  <p className="text-2xl font-bold">{metrics.totalAnomalies}</p>
                </div>
                <Shield className="h-8 w-8 text-gray-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500">Active</p>
                  <p className="text-2xl font-bold text-red-600">{metrics.activeAnomalies}</p>
                </div>
                <AlertTriangle className="h-8 w-8 text-red-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500">Critical</p>
                  <p className="text-2xl font-bold text-red-600">{metrics.criticalAlerts}</p>
                </div>
                <AlertCircle className="h-8 w-8 text-red-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500">Resolved Today</p>
                  <p className="text-2xl font-bold text-green-600">{metrics.resolvedToday}</p>
                </div>
                <CheckCircle className="h-8 w-8 text-green-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500">Avg Response</p>
                  <p className="text-2xl font-bold">{metrics.avgResponseTime}m</p>
                </div>
                <Clock className="h-8 w-8 text-blue-400" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500">Affected Tenants</p>
                  <p className="text-2xl font-bold">{metrics.affectedTenants}</p>
                </div>
                <Users className="h-8 w-8 text-purple-400" />
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
              className={clsx('flex items-center gap-2', showAcknowledged && 'bg-blue-50 border-blue-300')}
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
              <p className="text-gray-500">
                No security anomalies matching your filters
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredAnomalies.map((anomaly) => (
                <div
                  key={anomaly.id}
                  className={clsx(
                    'p-4 border rounded-lg transition-colors cursor-pointer',
                    anomaly.status === 'active'
                      ? 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10'
                      : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                  )}
                  onClick={() => setExpandedAnomaly(expandedAnomaly === anomaly.id ? null : anomaly.id)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <div className={clsx(
                        'p-2 rounded-lg',
                        anomaly.severity === 'critical' || anomaly.severity === 'high'
                          ? 'bg-red-100 dark:bg-red-900/50'
                          : anomaly.severity === 'medium'
                          ? 'bg-yellow-100 dark:bg-yellow-900/50'
                          : 'bg-blue-100 dark:bg-blue-900/50'
                      )}>
                        {anomaly.severity === 'critical' || anomaly.severity === 'high' ? (
                          <AlertTriangle className="h-4 w-4 text-red-600" />
                        ) : (
                          <AlertCircle className="h-4 w-4 text-yellow-600" />
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <Badge className={getSeverityColor(anomaly.severity)}>
                            {anomaly.severity}
                          </Badge>
                          <span className="font-medium">{anomaly.type}</span>
                          {anomaly.tenantName && (
                            <Badge variant="outline">{anomaly.tenantName}</Badge>
                          )}
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                          {anomaly.description}
                        </p>
                        <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
                          <span>Detected: {new Date(anomaly.detectedAt).toLocaleString()}</span>
                          {anomaly.agentId && <span>Agent: {anomaly.agentId.slice(0, 8)}...</span>}
                          {anomaly.source && <span>Source: {anomaly.source}</span>}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <Badge className={getStatusColor(anomaly.status)}>
                        {anomaly.status}
                      </Badge>
                      <ChevronRight className={clsx(
                        'h-4 w-4 transition-transform',
                        expandedAnomaly === anomaly.id && 'rotate-90'
                      )} />
                    </div>
                  </div>

                  {/* Expanded Details */}
                  {expandedAnomaly === anomaly.id && (
                    <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                      <div className="grid grid-cols-2 gap-4 mb-4">
                        <div>
                          <p className="text-xs text-gray-500">Tenant ID</p>
                          <p className="font-mono text-sm">{anomaly.tenantId || 'N/A'}</p>
                        </div>
                        <div>
                          <p className="text-xs text-gray-500">Agent ID</p>
                          <p className="font-mono text-sm">{anomaly.agentId || 'N/A'}</p>
                        </div>
                      </div>

                      {anomaly.status === 'active' && (
                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleAcknowledge(anomaly.id);
                            }}
                          >
                            Acknowledge
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleResolve(anomaly.id);
                            }}
                            className="text-green-600 hover:bg-green-50"
                          >
                            Resolve
                          </Button>
                          {anomaly.tenantId && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation();
                                blockTenantMutation.mutate(anomaly.tenantId!);
                              }}
                              className="text-red-600 hover:bg-red-50"
                            >
                              Block Tenant
                            </Button>
                          )}
                        </div>
                      )}

                      {anomaly.status === 'acknowledged' && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleResolve(anomaly.id);
                          }}
                          className="text-green-600 hover:bg-green-50"
                        >
                          Resolve
                        </Button>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default AdminSecurityAnomalyDashboard;
