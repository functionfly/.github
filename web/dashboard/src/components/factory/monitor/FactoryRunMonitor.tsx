/**
 * FactoryRunMonitor Component
 * Real-time factory run monitoring and control for user dashboard
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Activity, Play, RefreshCw, CheckCircle, XCircle, Clock,
  AlertTriangle, ChevronDown, ChevronUp, Filter, Pause
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { factoryApi, type FactoryStatus, type FactoryRun } from '@/api/factory';
import { cn } from '@/lib/utils';

interface FactoryRunMonitorProps {
  agentId?: string;
}

export function FactoryRunMonitor({ agentId }: FactoryRunMonitorProps) {
  const [factoryStatus, setFactoryStatus] = useState<FactoryStatus | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [runHistoryExpanded, setRunHistoryExpanded] = useState(false);
  const [runStatusFilter, setRunStatusFilter] = useState<string>('all');
  const [autoRefresh, setAutoRefresh] = useState(true);

  const fetchFactoryStatus = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await factoryApi.getStatus();
      setFactoryStatus(data);
    } catch (err) {
      console.error('Failed to fetch factory status:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchFactoryStatus();
    if (autoRefresh) {
      const interval = setInterval(fetchFactoryStatus, 10000);
      return () => clearInterval(interval);
    }
  }, [fetchFactoryStatus, autoRefresh]);

  const handleRunPipeline = async () => {
    try {
      await factoryApi.triggerPipelineRun();
      fetchFactoryStatus();
    } catch (err) {
      console.error('Failed to trigger pipeline:', err);
    }
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'completed':
      case 'healthy':
        return 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300';
      case 'running':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300';
      case 'failed':
        return 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300';
      default:
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300';
    }
  };

  const calculateSuccessRate = (run?: FactoryRun) => {
    if (!run || run.functions_created === 0) return 0;
    const total = run.functions_created + run.functions_failed;
    if (total === 0) return 0;
    return Math.round((run.functions_created / total) * 100);
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Activity className="h-5 w-5" />
              <CardTitle>Factory Run Monitor</CardTitle>
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
                onClick={fetchFactoryStatus}
                disabled={isLoading}
                className="flex items-center gap-2"
              >
                <RefreshCw className={cn('h-4 w-4', isLoading && 'animate-spin')} />
                Refresh
              </Button>
              <Button
                size="sm"
                onClick={handleRunPipeline}
                className="flex items-center gap-2"
              >
                <Play className="h-4 w-4" />
                Run Pipeline
              </Button>
            </div>
          </div>
          <CardDescription>
            Monitor AI Function Factory pipeline runs
          </CardDescription>
        </CardHeader>
      </Card>

      {/* Stats Overview */}
      {factoryStatus && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Total Runs</p>
                  <p className="text-2xl font-bold">{factoryStatus.totals.runs}</p>
                </div>
                <div className="p-2 bg-blue-500/10 rounded-lg">
                  <Activity className="h-5 w-5 text-blue-600" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Published Functions</p>
                  <p className="text-2xl font-bold text-green-600">{factoryStatus.totals.published}</p>
                </div>
                <div className="p-2 bg-green-500/10 rounded-lg">
                  <CheckCircle className="h-5 w-5 text-green-600" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Opportunities</p>
                  <p className="text-2xl font-bold">{factoryStatus.totals.opportunities}</p>
                </div>
                <div className="p-2 bg-purple-500/10 rounded-lg">
                  <AlertTriangle className="h-5 w-5 text-purple-600" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Auto-Publish</p>
                  <p className={cn(
                    'text-2xl font-bold',
                    factoryStatus.totals.autopublish ? 'text-green-600' : 'text-yellow-600'
                  )}>
                    {factoryStatus.totals.autopublish ? 'On' : 'Off'}
                  </p>
                </div>
                <div className={cn(
                  'p-2 rounded-lg',
                  factoryStatus.totals.autopublish ? 'bg-green-500/10' : 'bg-yellow-500/10'
                )}>
                  {factoryStatus.totals.autopublish ? (
                    <CheckCircle className="h-5 w-5 text-green-600" />
                  ) : (
                    <XCircle className="h-5 w-5 text-yellow-600" />
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Latest Run */}
      {factoryStatus?.latest_run && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <Clock className="h-5 w-5" />
                Latest Pipeline Run
              </CardTitle>
              <span className={cn(
                'px-3 py-1 rounded-full text-sm font-medium',
                getStatusColor(factoryStatus.latest_run.status)
              )}>
                {factoryStatus.latest_run.status}
              </span>
            </div>
            <CardDescription>
              Run ID: {factoryStatus.latest_run.id}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Started</p>
                <p className="font-medium">{formatDate(factoryStatus.latest_run.started_at)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Completed</p>
                <p className="font-medium">
                  {factoryStatus.latest_run.completed_at
                    ? formatDate(factoryStatus.latest_run.completed_at)
                    : 'In Progress'}
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Success Rate</p>
                <p className={cn(
                  'font-medium',
                  calculateSuccessRate(factoryStatus.latest_run) >= 80 ? 'text-green-600' :
                  calculateSuccessRate(factoryStatus.latest_run) >= 50 ? 'text-yellow-600' : 'text-red-600'
                )}>
                  {calculateSuccessRate(factoryStatus.latest_run)}%
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Functions Created</p>
                <p className="font-medium text-blue-600">{factoryStatus.latest_run.functions_created}</p>
              </div>
            </div>

            <div className="grid grid-cols-5 gap-4 pt-4 border-t">
              <div className="text-center">
                <p className="text-2xl font-bold">{factoryStatus.latest_run.opportunities_found}</p>
                <p className="text-xs text-gray-500">Found</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-green-600">{factoryStatus.latest_run.opportunities_approved}</p>
                <p className="text-xs text-gray-500">Approved</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-red-600">{factoryStatus.latest_run.opportunities_rejected}</p>
                <p className="text-xs text-gray-500">Rejected</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-blue-600">{factoryStatus.latest_run.functions_created}</p>
                <p className="text-xs text-gray-500">Created</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-red-600">{factoryStatus.latest_run.functions_failed}</p>
                <p className="text-xs text-gray-500">Failed</p>
              </div>
            </div>

            {factoryStatus.latest_run.error && (
              <div className="mt-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                <div className="flex items-center gap-2 text-red-800 dark:text-red-300">
                  <AlertTriangle className="h-4 w-4" />
                  <span>{factoryStatus.latest_run.error}</span>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Configuration */}
      {factoryStatus?.config && (
        <Card>
          <CardHeader>
            <CardTitle>Configuration</CardTitle>
            <CardDescription>Current factory settings</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Discovery Batch Size</p>
                <p className="font-medium">{factoryStatus.config.discovery_batch_size}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Min Quality Score</p>
                <p className="font-medium">{factoryStatus.config.minimum_quality_score}%</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Min Test Score</p>
                <p className="font-medium">{factoryStatus.config.minimum_test_score}%</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500">Max Opportunities</p>
                <p className="font-medium">{factoryStatus.config.max_opportunities_per_run}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

export default FactoryRunMonitor;
