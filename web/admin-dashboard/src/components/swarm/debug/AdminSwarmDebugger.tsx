/**
 * AdminSwarmDebugger Component
 * Platform-level swarm visualization and debugging for admin dashboard
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Activity, Users, MessageSquare, GitBranch, AlertTriangle,
  RefreshCw, Eye, EyeOff, Filter, Play, Pause, Shield, Globe
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import clsx from 'clsx';

interface PlatformSwarmStatus {
  totalChildren: number;
  activeChildren: number;
  pausedChildren: number;
  totalMessages: number;
  pendingMessages: number;
  healthScore: number;
  anomalies: Array<{
    type: string;
    severity: string;
    description: string;
    timestamp: string;
    agentId?: string;
  }>;
}

interface ChildAgentInfo {
  id: string;
  name: string;
  status: string;
  swarmRole: string;
  parentAgentId: string;
  trustScore: number;
  tenantId: string;
}

interface AdminSwarmDebuggerProps {
  platformControllerId?: string;
}

export function AdminSwarmDebugger({ platformControllerId }: AdminSwarmDebuggerProps) {
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [selectedFilter, setSelectedFilter] = useState<string>('all');
  const [showAnomaliesOnly, setShowAnomaliesOnly] = useState(false);

  // Fetch platform swarm status
  const { data: swarmStatus, isLoading, refetch } = useQuery<PlatformSwarmStatus>({
    queryKey: ['admin-platform-swarm-status'],
    queryFn: async () => {
      try {
        const response = await adminApiClient.get<{ status: PlatformSwarmStatus }>('/platform/swarm/status');
        return response.data?.status ?? {
          totalChildren: 0,
          activeChildren: 0,
          pausedChildren: 0,
          totalMessages: 0,
          pendingMessages: 0,
          healthScore: 0,
          anomalies: [],
        };
      } catch {
        return {
          totalChildren: 0,
          activeChildren: 0,
          pausedChildren: 0,
          totalMessages: 0,
          pendingMessages: 0,
          healthScore: 0,
          anomalies: [],
        };
      }
    },
    staleTime: 10000,
    refetchInterval: autoRefresh ? 10000 : undefined,
  });

  // Fetch all child agents
  const { data: childrenData } = useQuery<{ children: ChildAgentInfo[] }>({
    queryKey: ['admin-platform-children'],
    queryFn: async () => {
      try {
        const response = await adminApiClient.get<{ children: ChildAgentInfo[] }>('/platform/swarm/children');
        return response.data ?? { children: [] };
      } catch {
        return { children: [] };
      }
    },
    staleTime: 15000,
    refetchInterval: autoRefresh ? 15000 : undefined,
  });

  const children = childrenData?.children ?? [];

  const filteredChildren = children.filter(child => {
    if (selectedFilter !== 'all' && child.swarmRole !== selectedFilter) return false;
    if (showAnomaliesOnly && child.trustScore < 50) return true;
    if (showAnomaliesOnly) return false;
    return true;
  });

  const getHealthColor = (score: number) => {
    if (score >= 80) return 'text-green-600';
    if (score >= 50) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active': return 'success';
      case 'paused': return 'warning';
      case 'stopped': return 'error';
      default: return 'secondary';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Globe className="h-5 w-5 text-blue-600" />
              <CardTitle>Platform Swarm Debugger</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAutoRefresh(!autoRefresh)}
                className={clsx('flex items-center gap-2', autoRefresh && 'bg-green-50 border-green-300')}
              >
                {autoRefresh ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
                {autoRefresh ? 'Auto' : 'Paused'}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => refetch()}
                disabled={isLoading}
                className="flex items-center gap-2"
              >
                <RefreshCw className={clsx('h-4 w-4', isLoading && 'animate-spin')} />
                Refresh
              </Button>
            </div>
          </div>
          <CardDescription>
            Monitor and debug platform-wide swarm operations
          </CardDescription>
        </CardHeader>
      </Card>

      {/* Platform Overview */}
      {swarmStatus && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              Platform Swarm Overview
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
              <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg text-center">
                <p className="text-2xl font-bold text-gray-900 dark:text-white">{swarmStatus.totalChildren}</p>
                <p className="text-xs text-gray-500">Total Children</p>
              </div>
              <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-green-600">{swarmStatus.activeChildren}</p>
                <p className="text-xs text-green-600">Active</p>
              </div>
              <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-yellow-600">{swarmStatus.pausedChildren}</p>
                <p className="text-xs text-yellow-600">Paused</p>
              </div>
              <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-blue-600">{swarmStatus.totalMessages}</p>
                <p className="text-xs text-blue-600">Total Messages</p>
              </div>
              <div className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-purple-600">{swarmStatus.pendingMessages}</p>
                <p className="text-xs text-purple-600">Pending</p>
              </div>
              <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg text-center">
                <p className={clsx('text-2xl font-bold', getHealthColor(swarmStatus.healthScore))}>
                  {swarmStatus.healthScore}%
                </p>
                <p className="text-xs text-gray-500">Health Score</p>
              </div>
            </div>

            {/* Anomalies */}
            {swarmStatus.anomalies.length > 0 && (
              <div className="mt-4 space-y-2">
                <h4 className="text-sm font-semibold flex items-center gap-2 text-red-600">
                  <AlertTriangle className="h-4 w-4" />
                  Detected Anomalies ({swarmStatus.anomalies.length})
                </h4>
                {swarmStatus.anomalies.map((anomaly, idx) => (
                  <div
                    key={idx}
                    className={clsx(
                      'p-3 rounded-lg border',
                      anomaly.severity === 'high' ? 'bg-red-50 border-red-200' :
                      anomaly.severity === 'medium' ? 'bg-yellow-50 border-yellow-200' :
                      'bg-gray-50 border-gray-200'
                    )}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge variant={anomaly.severity === 'high' ? 'error' : anomaly.severity === 'medium' ? 'warning' : 'secondary'}>
                          {anomaly.severity}
                        </Badge>
                        <span className="font-medium">{anomaly.type}</span>
                      </div>
                      <div className="flex items-center gap-2 text-xs text-gray-500">
                        {anomaly.agentId && <span>Agent: {anomaly.agentId.slice(0, 8)}...</span>}
                        <span>{new Date(anomaly.timestamp).toLocaleTimeString()}</span>
                      </div>
                    </div>
                    <p className="text-sm mt-1 text-gray-600">{anomaly.description}</p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* All Child Agents */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              All Child Agents ({filteredChildren.length})
            </CardTitle>
            <div className="flex items-center gap-2">
              <select
                value={selectedFilter}
                onChange={(e) => setSelectedFilter(e.target.value)}
                className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-lg"
              >
                <option value="all">All Roles</option>
                <option value="manager">Manager</option>
                <option value="worker">Worker</option>
                <option value="infrastructure">Infrastructure</option>
                <option value="scanner">Scanner</option>
              </select>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowAnomaliesOnly(!showAnomaliesOnly)}
                className={clsx('flex items-center gap-2', showAnomaliesOnly && 'bg-yellow-50 border-yellow-300')}
              >
                {showAnomaliesOnly ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                Anomalies
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredChildren.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No child agents found matching filters
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-700">
                    <th className="text-left py-3 px-4 text-sm font-semibold">Agent</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold">Role</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold">Status</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold">Trust Score</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold">Tenant</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredChildren.map((child) => (
                    <tr key={child.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-3">
                          <div className={clsx(
                            'w-2 h-2 rounded-full',
                            child.status === 'active' ? 'bg-green-500' :
                            child.status === 'paused' ? 'bg-yellow-500' : 'bg-gray-500'
                          )} />
                          <div>
                            <p className="font-medium">{child.name}</p>
                            <p className="text-xs text-gray-500 font-mono">{child.id.slice(0, 12)}...</p>
                          </div>
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <Badge variant="outline">{child.swarmRole}</Badge>
                      </td>
                      <td className="py-3 px-4">
                        <Badge variant={getStatusBadgeVariant(child.status)}>
                          {child.status}
                        </Badge>
                      </td>
                      <td className="py-3 px-4">
                        <span className={clsx(
                          'font-medium',
                          child.trustScore >= 80 ? 'text-green-600' :
                          child.trustScore >= 50 ? 'text-yellow-600' : 'text-red-600'
                        )}>
                          {child.trustScore.toFixed(1)}%
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <span className="text-xs font-mono text-gray-500">{child.tenantId.slice(0, 8)}...</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default AdminSwarmDebugger;
