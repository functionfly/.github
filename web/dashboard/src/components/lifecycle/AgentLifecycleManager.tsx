/**
 * AgentLifecycleManager Component
 * UI for managing agent lifecycle: start, stop, restart, pause, resume
 * Used in the user-facing dashboard (web/dashboard)
 */

import { useState } from 'react';
import { 
  Play, Pause, Square, RotateCw, CheckCircle, XCircle, 
  Clock, AlertTriangle, Activity
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useAgent, useAgentUsage } from '@/hooks';
import { agentApi } from '@/api/agent';
import { cn } from '@/lib/utils';

interface AgentLifecycleManagerProps {
  agentId: string;
  agentName: string;
  currentStatus: string;
  onStatusChange?: (newStatus: string) => void;
}

type LifecycleAction = 'start' | 'pause' | 'stop' | 'restart';

interface LifecycleState {
  isLoading: boolean;
  lastAction: LifecycleAction | null;
  error: string | null;
}

export function AgentLifecycleManager({
  agentId,
  agentName,
  currentStatus,
  onStatusChange,
}: AgentLifecycleManagerProps) {
  const [state, setState] = useState<LifecycleState>({
    isLoading: false,
    lastAction: null,
    error: null,
  });

  const { data: agent } = useAgent(agentId);
  const { data: usage } = useAgentUsage(agentId);

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
      case 'pending':
        return 'bg-blue-500';
      default:
        return 'bg-gray-500';
    }
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'running':
        return 'success';
      case 'paused':
      case 'suspended':
        return 'warning';
      case 'stopped':
      case 'terminated':
        return 'error';
      default:
        return 'secondary';
    }
  };

  const handleLifecycleAction = async (action: LifecycleAction) => {
    setState({ isLoading: true, lastAction: action, error: null });

    try {
      switch (action) {
        case 'stop':
          await agentApi.triggerKillSwitch(agentId, 'User requested stop');
          break;
        case 'pause':
          await agentApi.updatePolicy(agentId, { agentId, maxExecutionDepth: 0 });
          break;
        case 'restart':
          await agentApi.triggerKillSwitch(agentId, 'User requested restart');
          await agentApi.startSession(agentId);
          break;
        case 'start':
          await agentApi.startSession(agentId);
          break;
      }
      setState({ isLoading: false, lastAction: null, error: null });
      onStatusChange?.(action === 'stop' ? 'stopped' : action === 'pause' ? 'paused' : 'active');
    } catch (err) {
      setState({
        isLoading: false,
        lastAction: null,
        error: err instanceof Error ? err.message : 'Action failed'
      });
    }
  };

  const isActionEnabled = (action: LifecycleAction): boolean => {
    if (state.isLoading) return false;
    const status = currentStatus.toLowerCase();
    
    switch (action) {
      case 'start':
        return status === 'stopped' || status === 'paused';
      case 'pause':
        return status === 'running' || status === 'active';
      case 'stop':
        return status === 'running' || status === 'active' || status === 'pending';
      case 'restart':
        return status === 'running' || status === 'active';
      default:
        return false;
    }
  };

  const canShowLifecycleControls = () => {
    const status = currentStatus.toLowerCase();
    return ['active', 'running', 'paused', 'stopped', 'pending', 'suspended'].includes(status);
  };

  if (!canShowLifecycleControls()) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={cn('w-3 h-3 rounded-full', getStatusColor(currentStatus))} />
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Agent Lifecycle
            </CardTitle>
          </div>
          <Badge variant={getStatusBadgeVariant(currentStatus)}>
            {currentStatus}
          </Badge>
        </div>
        <CardDescription>
          Manage lifecycle operations for {agentName}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Current Status */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <div className="space-y-1">
            <p className="text-xs text-gray-500 dark:text-gray-400">Status</p>
            <p className="font-medium capitalize">{currentStatus}</p>
          </div>
          {usage && usage.usage && (
            <>
              <div className="space-y-1">
                <p className="text-xs text-gray-500 dark:text-gray-400">Calls/min</p>
                <p className="font-medium">{usage.usage.callsThisMinute}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-gray-500 dark:text-gray-400">Memory</p>
                <p className="font-medium">{usage.usage.callsToday} calls today</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-gray-500 dark:text-gray-400">Spend Today</p>
                <p className="font-medium">${usage.usage.spendTodayUSD.toFixed(4)}</p>
              </div>
            </>
          )}
        </div>

        {/* Error Display */}
        {state.error && (
          <div className="flex items-center gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-800 dark:text-red-300">
            <AlertTriangle className="h-4 w-4" />
            <span className="text-sm">{state.error}</span>
          </div>
        )}

        {/* Lifecycle Actions */}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleLifecycleAction('start')}
            disabled={!isActionEnabled('start')}
            className="flex items-center gap-2"
          >
            <Play className="h-4 w-4" />
            Start
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={() => handleLifecycleAction('pause')}
            disabled={!isActionEnabled('pause')}
            className="flex items-center gap-2"
          >
            <Pause className="h-4 w-4" />
            Pause
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={() => handleLifecycleAction('restart')}
            disabled={!isActionEnabled('restart')}
            className="flex items-center gap-2"
          >
            <RotateCw className={cn('h-4 w-4', state.isLoading && state.lastAction === 'restart' && 'animate-spin')} />
            Restart
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={() => handleLifecycleAction('stop')}
            disabled={!isActionEnabled('stop')}
            className="flex items-center gap-2 text-red-600 hover:text-red-700 hover:border-red-300"
          >
            <Square className="h-4 w-4" />
            Stop
          </Button>
        </div>

        {/* Loading State */}
        {state.isLoading && (
          <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <Clock className="h-4 w-4 animate-spin" />
            Processing {state.lastAction}...
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default AgentLifecycleManager;
