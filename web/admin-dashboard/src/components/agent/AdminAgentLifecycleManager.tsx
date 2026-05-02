/**
 * AdminAgentLifecycleManager Component
 * Platform-level agent lifecycle management for admin dashboard
 */

import { useState } from 'react';
import {
  Play, Pause, Square, RotateCw, CheckCircle, XCircle,
  Clock, AlertTriangle, Activity, Shield, Users
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { useToastHelpers } from '@/components/ui/Toast';
import clsx from 'clsx';

interface AgentStatus {
  id: string;
  agentId: string;
  name: string;
  status: string;
  tenantId: string;
  tenantName: string;
  trustScore: number;
  createdAt: string;
  updatedAt: string;
}

interface PlatformAgentStats {
  totalAgents: number;
  activeAgents: number;
  pausedAgents: number;
  stoppedAgents: number;
  totalChildren: number;
}

interface AdminAgentLifecycleManagerProps {
  agentId?: string;
  agentData?: AgentStatus;
}

type LifecycleAction = 'start' | 'pause' | 'stop' | 'restart' | 'kill';

export function AdminAgentLifecycleManager({ agentId, agentData }: AdminAgentLifecycleManagerProps) {
  const queryClient = useQueryClient();
  const toast = useToastHelpers();
  const [selectedAction, setSelectedAction] = useState<LifecycleAction | null>(null);
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);

  const { data: stats } = useQuery<PlatformAgentStats>({
    queryKey: ['admin-platform-agent-stats'],
    queryFn: async () => {
      return {
        totalAgents: 0,
        activeAgents: 0,
        pausedAgents: 0,
        stoppedAgents: 0,
        totalChildren: 0,
      };
    },
    staleTime: 30000,
  });

  const killAgentMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/agents/${id}/kill`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-platform-agent-stats'] });
      toast.success('Agent terminated');
      setConfirmDialogOpen(false);
    },
    onError: (error: Error) => {
      toast.error(`Failed to terminate agent: ${error.message}`);
    },
  });

  const pauseAgentMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/agents/${id}/pause`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-platform-agent-stats'] });
      toast.success('Agent paused');
    },
    onError: (error: Error) => {
      toast.error(`Failed to pause agent: ${error.message}`);
    },
  });

  const resumeAgentMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/agents/${id}/resume`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-platform-agent-stats'] });
      toast.success('Agent resumed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resume agent: ${error.message}`);
    },
  });

  const handleAction = (action: LifecycleAction) => {
    if (action === 'kill') {
      setSelectedAction(action);
      setConfirmDialogOpen(true);
    } else if (agentId) {
      switch (action) {
        case 'stop':
        case 'pause':
          pauseAgentMutation.mutate(agentId);
          break;
        case 'start':
        case 'restart':
          resumeAgentMutation.mutate(agentId);
          break;
      }
    }
  };

  const confirmKill = () => {
    if (agentId) {
      killAgentMutation.mutate(agentId);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'running':
        return 'bg-green-500';
      case 'paused':
      case 'suspended':
        return 'bg-yellow-500';
      case 'stopped':
      case 'terminated':
        return 'bg-red-500';
      default:
        return 'bg-gray-500';
    }
  };

  return (
    <div className="space-y-6">
      {/* Platform Stats Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Platform Agent Overview
          </CardTitle>
          <CardDescription>
            Aggregated agent status across all tenants
          </CardDescription>
        </CardHeader>
        <CardContent>
          {stats ? (
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg text-center">
                <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.totalAgents}</p>
                <p className="text-xs text-gray-500 dark:text-gray-400">Total Agents</p>
              </div>
              <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-green-600 dark:text-green-400">{stats.activeAgents}</p>
                <p className="text-xs text-green-600 dark:text-green-400">Active</p>
              </div>
              <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{stats.pausedAgents}</p>
                <p className="text-xs text-yellow-600 dark:text-yellow-400">Paused</p>
              </div>
              <div className="p-4 bg-red-50 dark:bg-red-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-red-600 dark:text-red-400">{stats.stoppedAgents}</p>
                <p className="text-xs text-red-600 dark:text-red-400">Stopped</p>
              </div>
              <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg text-center">
                <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">{stats.totalChildren}</p>
                <p className="text-xs text-blue-600 dark:text-blue-400">Child Agents</p>
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-400"></div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Individual Agent Control */}
      {agentId && agentData && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className={clsx('w-3 h-3 rounded-full', getStatusColor(agentData.status))} />
                <CardTitle>{agentData.name}</CardTitle>
              </div>
              <Badge variant={agentData.status === 'active' ? 'success' : agentData.status === 'stopped' ? 'error' : 'secondary'}>
                {agentData.status}
              </Badge>
            </div>
            <CardDescription>
              Agent ID: {agentData.agentId} | Tenant: {agentData.tenantName}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <div className="space-y-1">
                <p className="text-xs text-gray-500 dark:text-gray-400">Trust Score</p>
                <p className={clsx(
                  'font-medium',
                  agentData.trustScore >= 80 ? 'text-green-600' :
                  agentData.trustScore >= 50 ? 'text-yellow-600' : 'text-red-600'
                )}>
                  {agentData.trustScore.toFixed(1)}%
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-gray-500 dark:text-gray-400">Created</p>
                <p className="font-medium text-sm">{new Date(agentData.createdAt).toLocaleDateString()}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-gray-500 dark:text-gray-400">Last Updated</p>
                <p className="font-medium text-sm">{new Date(agentData.updatedAt).toLocaleDateString()}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-gray-500 dark:text-gray-400">Tenant ID</p>
                <p className="font-mono text-xs">{agentData.tenantId.slice(0, 8)}...</p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleAction('start')}
                disabled={agentData.status === 'active'}
                className="flex items-center gap-2"
              >
                <Play className="h-4 w-4" />
                Resume
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() => handleAction('pause')}
                disabled={agentData.status !== 'active'}
                className="flex items-center gap-2"
              >
                <Pause className="h-4 w-4" />
                Pause
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() => handleAction('stop')}
                disabled={agentData.status === 'stopped'}
                className="flex items-center gap-2"
              >
                <Square className="h-4 w-4" />
                Stop
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() => handleAction('kill')}
                className="flex items-center gap-2 text-red-600 hover:bg-red-50 hover:border-red-300"
              >
                <XCircle className="h-4 w-4" />
                Kill
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Confirmation Dialog */}
      {confirmDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full">
            <div className="flex items-center gap-3 text-red-600 mb-4">
              <AlertTriangle className="h-6 w-6" />
              <h2 className="text-xl font-bold">Terminate Agent</h2>
            </div>
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              Are you sure you want to terminate this agent? This action cannot be undone.
            </p>
            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setConfirmDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={confirmKill}
                className="bg-red-600 text-white hover:bg-red-700"
              >
                {killAgentMutation.isPending ? 'Terminating...' : 'Terminate Agent'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default AdminAgentLifecycleManager;
