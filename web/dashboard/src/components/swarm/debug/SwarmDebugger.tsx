/**
 * SwarmDebugger Component
 * Real-time swarm visualization and debugging for user dashboard
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Activity, Users, MessageSquare, GitBranch, AlertTriangle,
  RefreshCw, Eye, EyeOff, Filter, Play, Pause
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { agentApi } from '@/api/agent';
import { cn } from '@/lib/utils';

interface ChildAgent {
  id: string;
  name: string;
  status: 'active' | 'suspended' | 'pending';
  swarmRole: 'worker' | 'manager' | 'infrastructure';
  trustScore: number;
  economicScore: number;
}

interface SwarmHealth {
  status: 'healthy' | 'degraded' | 'critical';
  healthScore: number;
  anomalies: Array<{
    type: string;
    severity: 'low' | 'medium' | 'high';
    description: string;
    timestamp: string;
  }>;
  children: number;
}

interface MessageFlow {
  id: string;
  fromAgentId: string;
  toAgentId: string;
  messageType: string;
  status: string;
  createdAt: string;
  deliveredAt?: string;
}

interface SwarmDebuggerProps {
  agentId: string;
  agentName: string;
}

export function SwarmDebugger({ agentId, agentName }: SwarmDebuggerProps) {
  const [isPaused, setIsPaused] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [selectedFilter, setSelectedFilter] = useState<string>('all');
  const [showAnomaliesOnly, setShowAnomaliesOnly] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  
  const [children, setChildren] = useState<ChildAgent[]>([]);
  const [messages, setMessages] = useState<MessageFlow[]>([]);
  const [swarmHealth, setSwarmHealth] = useState<SwarmHealth | null>(null);

  const fetchSwarmData = useCallback(async () => {
    if (!agentId || isPaused) return;
    
    setIsLoading(true);
    try {
      const [childrenRes, inboxRes, healthRes] = await Promise.allSettled([
        agentApi.getChildren(agentId),
        agentApi.getInbox(agentId),
        agentApi.checkSwarmHealth(agentId, { hours: 1 }),
      ]);

      if (childrenRes.status === 'fulfilled') {
        const childrenData = (childrenRes.value.children as unknown[]) || [];
        setChildren(childrenData.map(c => ({
          id: (c as { id?: string }).id || '',
          name: (c as { name: string }).name,
          status: ((c as { status: string }).status || 'active') as ChildAgent['status'],
          swarmRole: ((c as { swarm_role?: string }).swarm_role || 'worker') as ChildAgent['swarmRole'],
          trustScore: Number((c as { trust_score?: number }).trust_score) || 0,
          economicScore: Number((c as { economic_score?: number }).economic_score) || 0,
        })));
      }

      if (inboxRes.status === 'fulfilled') {
        setMessages((inboxRes.value.messages as MessageFlow[]) || []);
      }

      if (healthRes.status === 'fulfilled') {
        const health = healthRes.value;
        setSwarmHealth({
          status: (health.status as SwarmHealth['status']) || 'healthy',
          healthScore: health.health_score || 0,
          anomalies: (health.anomalies as SwarmHealth['anomalies']) || [],
          children: health.children || 0,
        });
      }
    } catch (err) {
      console.error('Failed to fetch swarm data:', err);
    } finally {
      setIsLoading(false);
    }
  }, [agentId, isPaused]);

  useEffect(() => {
    fetchSwarmData();
    if (autoRefresh && !isPaused) {
      const interval = setInterval(fetchSwarmData, 5000);
      return () => clearInterval(interval);
    }
  }, [fetchSwarmData, autoRefresh, isPaused]);

  const filteredChildren = children.filter(child => {
    if (selectedFilter !== 'all' && child.swarmRole !== selectedFilter) return false;
    if (showAnomaliesOnly && child.trustScore < 50) return true;
    if (showAnomaliesOnly) return false;
    return true;
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-500';
      case 'suspended': return 'bg-red-500';
      case 'pending': return 'bg-yellow-500';
      default: return 'bg-gray-500';
    }
  };

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'manager': return '👑';
      case 'worker': return '⚙️';
      case 'infrastructure': return '🏗️';
      default: return '•';
    }
  };

  const getHealthColor = (status: string) => {
    switch (status) {
      case 'healthy': return 'text-green-600';
      case 'degraded': return 'text-yellow-600';
      case 'critical': return 'text-red-600';
      default: return 'text-gray-600';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Controls */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Activity className="h-5 w-5" />
              <CardTitle>Swarm Debugger</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAutoRefresh(!autoRefresh)}
                className={cn('flex items-center gap-2', autoRefresh && 'bg-green-50 border-green-300')}
              >
                {autoRefresh ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
                {autoRefresh ? 'Auto' : 'Paused'}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={fetchSwarmData}
                disabled={isLoading}
                className="flex items-center gap-2"
              >
                <RefreshCw className={cn('h-4 w-4', isLoading && 'animate-spin')} />
                Refresh
              </Button>
            </div>
          </div>
          <CardDescription>
            Debug and monitor {agentName} swarm operations in real-time
          </CardDescription>
        </CardHeader>
      </Card>

      {/* Health Status */}
      {swarmHealth && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Swarm Health
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <p className="text-xs text-gray-500 dark:text-gray-400">Health Status</p>
                <p className={cn('text-xl font-bold', getHealthColor(swarmHealth.status))}>
                  {swarmHealth.status.toUpperCase()}
                </p>
              </div>
              <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <p className="text-xs text-gray-500 dark:text-gray-400">Health Score</p>
                <p className="text-xl font-bold">{swarmHealth.healthScore}%</p>
              </div>
              <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <p className="text-xs text-gray-500 dark:text-gray-400">Child Agents</p>
                <p className="text-xl font-bold">{swarmHealth.children}</p>
              </div>
              <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <p className="text-xs text-gray-500 dark:text-gray-400">Anomalies</p>
                <p className={cn(
                  'text-xl font-bold',
                  swarmHealth.anomalies.length > 0 ? 'text-red-600' : 'text-green-600'
                )}>
                  {swarmHealth.anomalies.length}
                </p>
              </div>
            </div>

            {/* Anomalies List */}
            {swarmHealth.anomalies.length > 0 && (
              <div className="mt-4 space-y-2">
                <h4 className="text-sm font-semibold flex items-center gap-2 text-red-600">
                  <AlertTriangle className="h-4 w-4" />
                  Detected Anomalies
                </h4>
                {swarmHealth.anomalies.map((anomaly, idx) => (
                  <div
                    key={idx}
                    className={cn(
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
                      <span className="text-xs text-gray-500">
                        {new Date(anomaly.timestamp).toLocaleTimeString()}
                      </span>
                    </div>
                    <p className="text-sm mt-1 text-gray-600 dark:text-gray-400">{anomaly.description}</p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Child Agents */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Child Agents ({filteredChildren.length})
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
              </select>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowAnomaliesOnly(!showAnomaliesOnly)}
                className={cn('flex items-center gap-2', showAnomaliesOnly && 'bg-yellow-50 border-yellow-300')}
              >
                {showAnomaliesOnly ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                Anomalies
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredChildren.length === 0 ? (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              No child agents found matching filters
            </div>
          ) : (
            <div className="space-y-3">
              {filteredChildren.map((child) => (
                <div
                  key={child.id}
                  className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50"
                >
                  <div className="flex items-center gap-4">
                    <div className={cn('w-3 h-3 rounded-full', getStatusColor(child.status))} />
                    <span className="text-2xl">{getRoleIcon(child.swarmRole)}</span>
                    <div>
                      <p className="font-medium">{child.name}</p>
                      <p className="text-xs text-gray-500">{child.swarmRole} • {child.id.slice(0, 8)}...</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-6">
                    <div className="text-center">
                      <p className="text-xs text-gray-500">Trust</p>
                      <p className={cn(
                        'font-medium',
                        child.trustScore >= 80 ? 'text-green-600' :
                        child.trustScore >= 50 ? 'text-yellow-600' : 'text-red-600'
                      )}>
                        {child.trustScore}%
                      </p>
                    </div>
                    <div className="text-center">
                      <p className="text-xs text-gray-500">Economic</p>
                      <p className="font-medium">{child.economicScore}%</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Message Flow */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MessageSquare className="h-5 w-5" />
            Recent Messages ({messages.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {messages.length === 0 ? (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              No messages in inbox
            </div>
          ) : (
            <div className="space-y-2">
              {messages.slice(0, 10).map((msg) => (
                <div
                  key={msg.id}
                  className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    <MessageSquare className="h-4 w-4 text-gray-400" />
                    <div>
                      <p className="text-sm font-medium">{msg.messageType}</p>
                      <p className="text-xs text-gray-500">
                        {msg.fromAgentId.slice(0, 8)}... → {msg.toAgentId.slice(0, 8)}...
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={msg.status === 'delivered' ? 'success' : 'secondary'}>
                      {msg.status}
                    </Badge>
                    <span className="text-xs text-gray-500">
                      {new Date(msg.createdAt).toLocaleTimeString()}
                    </span>
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

export default SwarmDebugger;
