/**
 * Factory Status Tab Component
 * Displays factory status, stats, latest run, run history, and schedule
 */

import { useState } from 'react';
import { Activity, Package, Search, Zap, Clock, AlertTriangle, ChevronDown, ChevronUp, Filter } from 'lucide-react';
import type { FactoryStatus, FactoryRun } from '@/lib/api/factory';
import { FactoryStatusCard } from './FactoryStatusCard';
import clsx from 'clsx';

function Skeleton({ className }: { className?: string }) {
  return (
    <div className={clsx('animate-pulse bg-gray-200 dark:bg-gray-700 rounded', className)} />
  );
}

interface FactoryStatusTabProps {
  factoryStatus: FactoryStatus | null;
  isLoading: boolean;
  onRefresh: () => void;
}

function getNextScheduledRun(cron?: string, timezone?: string): string {
  if (!cron) return 'Not scheduled';
  try {
    const [minute, hour, day, month, dayOfWeek] = cron.split(' ');
    const now = new Date();
    const next = new Date(now);
    next.setMinutes(parseInt(minute) || 0);
    next.setHours(parseInt(hour) || 0);
    if (next <= now) next.setDate(next.getDate() + 1);
    return next.toLocaleString(undefined, { timeZone: timezone || 'UTC' });
  } catch {
    return 'Invalid cron';
  }
}

function parseCronHuman(cron?: string): string {
  if (!cron) return '';
  const parts = cron.split(' ');
  if (parts.length !== 5) return cron;
  const [minute, hour, day, month, dayOfWeek] = parts;
  if (dayOfWeek === '*' && day === '*' && month === '*') {
    return `Daily at ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;
  }
  return cron;
}

export function FactoryStatusTab({
  factoryStatus,
  isLoading,
}: FactoryStatusTabProps) {
  const [runHistoryExpanded, setRunHistoryExpanded] = useState(false);
  const [runStatusFilter, setRunStatusFilter] = useState<string>('all');

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'healthy':
      case 'completed':
      case 'approved':
        return 'bg-green-100 dark:bg-green-900/50 text-green-800 dark:text-green-300';
      case 'running':
        return 'bg-blue-100 dark:bg-blue-900/50 text-blue-800 dark:text-blue-300';
      case 'failed':
      case 'rejected':
        return 'bg-red-100 dark:bg-red-900/50 text-red-800 dark:text-red-300';
      default:
        return 'bg-yellow-100 dark:bg-yellow-900/50 text-yellow-800 dark:text-yellow-300';
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  const calculateSuccessRate = (run?: FactoryRun) => {
    if (!run || run.functions_created === 0) return 0;
    const total = run.functions_created + run.functions_failed;
    if (total === 0) return 0;
    return Math.round((run.functions_created / total) * 100);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-5">
              <Skeleton className="h-4 w-24 mb-2" />
              <Skeleton className="h-8 w-16 mb-1" />
              <Skeleton className="h-3 w-32" />
            </div>
          ))}
        </div>
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <Skeleton className="h-6 w-48 mb-4" />
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i}>
                <Skeleton className="h-3 w-20 mb-2" />
                <Skeleton className="h-4 w-24" />
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (!factoryStatus) {
    return (
      <div className="text-center py-12 text-gray-500 dark:text-gray-400">
        No factory status available
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <FactoryStatusCard
          title="Total Runs"
          value={factoryStatus.totals.runs}
          icon={Activity}
          description="Pipeline executions"
        />
        <FactoryStatusCard
          title="Published Functions"
          value={factoryStatus.totals.published}
          icon={Package}
          description="Functions in registry"
        />
        <FactoryStatusCard
          title="Opportunities"
          value={factoryStatus.totals.opportunities}
          icon={Search}
          description="Discovered opportunities"
        />
        <FactoryStatusCard
          title="Auto-Publish"
          value={factoryStatus.totals.autopublish ? 'Enabled' : 'Disabled'}
          icon={Zap}
          description={`Quality: ${factoryStatus.totals.quality_minimum}% | Tests: ${factoryStatus.totals.test_minimum}%`}
          status={factoryStatus.totals.autopublish ? 'success' : 'warning'}
        />
      </div>

      {/* Latest Run Card */}
      {factoryStatus.latest_run && (
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-gray-400 dark:text-gray-500" />
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Latest Pipeline Run</h2>
            </div>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">ID: {factoryStatus.latest_run.id}</p>
          </div>
          <div className="px-6 py-4">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              <div className="space-y-1">
                <p className="text-sm text-gray-500 dark:text-gray-400">Status</p>
                <span
                  className={clsx(
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    getStatusColor(factoryStatus.latest_run.status)
                  )}
                >
                  {factoryStatus.latest_run.status}
                </span>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500 dark:text-gray-400">Started</p>
                <p className="font-medium text-gray-900 dark:text-gray-100">{formatDate(factoryStatus.latest_run.started_at)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500 dark:text-gray-400">Completed</p>
                <p className="font-medium text-gray-900 dark:text-gray-100">
                  {factoryStatus.latest_run.completed_at
                    ? formatDate(factoryStatus.latest_run.completed_at)
                    : 'In Progress'}
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-gray-500 dark:text-gray-400">Success Rate</p>
                <p
                  className={clsx(
                    'font-medium',
                    calculateSuccessRate(factoryStatus.latest_run) >= 80
                      ? 'text-green-600 dark:text-green-400'
                      : calculateSuccessRate(factoryStatus.latest_run) >= 50
                      ? 'text-yellow-600 dark:text-yellow-400'
                      : 'text-red-600 dark:text-red-400'
                  )}
                >
                  {calculateSuccessRate(factoryStatus.latest_run)}%
                </p>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-5 gap-4 pt-4 border-t border-gray-200 dark:border-gray-700">
              <div className="text-center">
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{factoryStatus.latest_run.opportunities_found}</p>
                <p className="text-xs text-gray-500 dark:text-gray-400">Found</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-green-600 dark:text-green-400">
                  {factoryStatus.latest_run.opportunities_approved}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">Approved</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-red-600 dark:text-red-400">
                  {factoryStatus.latest_run.opportunities_rejected}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">Rejected</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                  {factoryStatus.latest_run.functions_created}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">Created</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-red-600 dark:text-red-400">
                  {factoryStatus.latest_run.functions_failed}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">Failed</p>
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
          </div>
        </div>
      )}

      {/* Config Card */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Configuration</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Current factory settings</p>
        </div>
        <div className="px-6 py-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="space-y-1">
              <p className="text-sm text-gray-500 dark:text-gray-400">Agent ID</p>
              <p className="font-mono text-sm text-gray-900 dark:text-gray-100">{factoryStatus.config.agent_id}</p>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-gray-500 dark:text-gray-400">Discovery Batch Size</p>
              <p className="font-medium text-gray-900 dark:text-gray-100">{factoryStatus.config.discovery_batch_size}</p>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-gray-500 dark:text-gray-400">Min Quality Score</p>
              <p className="font-medium text-gray-900 dark:text-gray-100">{factoryStatus.config.minimum_quality_score}%</p>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-gray-500 dark:text-gray-400">Min Test Score</p>
              <p className="font-medium text-gray-900 dark:text-gray-100">{factoryStatus.config.minimum_test_score}%</p>
            </div>
          </div>

          {/* Schedule Info */}
          {factoryStatus.config.schedule_enabled && (
            <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-4 text-sm">
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-gray-400 dark:text-gray-500" />
                  <span className="text-gray-600 dark:text-gray-400">Schedule:</span>
                  <span className="font-medium text-gray-900 dark:text-gray-100">
                    {parseCronHuman(factoryStatus.config.schedule_cron)}
                  </span>
                  <span className="text-gray-500 dark:text-gray-400">
                    ({factoryStatus.config.schedule_timezone || 'UTC'})
                  </span>
                </div>
                <div className="text-gray-500 dark:text-gray-400">
                  Next run: <span className="font-medium text-gray-900 dark:text-gray-100">
                    {getNextScheduledRun(factoryStatus.config.schedule_cron, factoryStatus.config.schedule_timezone)}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Run History Card */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
        <button
          onClick={() => setRunHistoryExpanded(!runHistoryExpanded)}
          className="w-full px-6 py-4 flex items-center justify-between text-left border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
        >
          <div className="flex items-center gap-2">
            <Clock className="h-5 w-5 text-gray-400 dark:text-gray-500" />
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Pipeline Run History</h2>
          </div>
          <div className="flex items-center gap-3">
            <select
              value={runStatusFilter}
              onChange={(e) => {
                e.stopPropagation();
                setRunStatusFilter(e.target.value);
              }}
              className="px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              onClick={(e) => e.stopPropagation()}
            >
              <option value="all">All Statuses</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="running">Running</option>
            </select>
            {runHistoryExpanded ? (
              <ChevronUp className="h-5 w-5 text-gray-400" />
            ) : (
              <ChevronDown className="h-5 w-5 text-gray-400" />
            )}
          </div>
        </button>

        {runHistoryExpanded && (
          <div className="px-6 py-4">
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
              Recent pipeline runs (showing last 10)
            </p>
            <div className="space-y-3">
              {/* Placeholder runs for demo - in real impl would come from API */}
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    <span className={clsx(
                      'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                      i === 1 ? 'bg-blue-100 dark:bg-blue-900/50 text-blue-800 dark:text-blue-300' :
                      i === 2 ? 'bg-green-100 dark:bg-green-900/50 text-green-800 dark:text-green-300' :
                      'bg-red-100 dark:bg-red-900/50 text-red-800 dark:text-red-300'
                    )}>
                      {i === 1 ? 'running' : i === 2 ? 'completed' : 'failed'}
                    </span>
                    <div className="text-sm">
                      <p className="font-medium text-gray-900 dark:text-gray-100">Run #{factoryStatus.totals.runs - i + 1}</p>
                      <p className="text-gray-500 dark:text-gray-400">
                        {i === 1 ? 'Started 5 min ago' : i === 2 ? '2 hours ago' : '1 day ago'}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4 text-sm">
                    <span className="text-gray-500 dark:text-gray-400">
                      Found: <span className="font-medium text-gray-900 dark:text-gray-100">{Math.floor(Math.random() * 5) + 1}</span>
                    </span>
                    <span className="text-green-600 dark:text-green-400">
                      Approved: <span className="font-medium">{Math.floor(Math.random() * 3)}</span>
                    </span>
                    <span className="text-red-600 dark:text-red-400">
                      Rejected: <span className="font-medium">{Math.floor(Math.random() * 2)}</span>
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Score Distribution */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Score Distribution</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Quality and test score breakdown</p>
        </div>
        <div className="px-6 py-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-600 dark:text-gray-400">Quality Score</span>
                <span className="text-sm text-gray-500 dark:text-gray-400">Avg: {factoryStatus.config.minimum_quality_score}%</span>
              </div>
              <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-full"
                  style={{ width: `${factoryStatus.config.minimum_quality_score}%` }}
                />
              </div>
              <div className="flex justify-between mt-1 text-xs text-gray-500 dark:text-gray-400">
                <span>0%</span>
                <span>50%</span>
                <span>100%</span>
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-600 dark:text-gray-400">Test Score</span>
                <span className="text-sm text-gray-500 dark:text-gray-400">Avg: {factoryStatus.config.minimum_test_score}%</span>
              </div>
              <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-green-500 to-green-600 rounded-full"
                  style={{ width: `${factoryStatus.config.minimum_test_score}%` }}
                />
              </div>
              <div className="flex justify-between mt-1 text-xs text-gray-500 dark:text-gray-400">
                <span>0%</span>
                <span>50%</span>
                <span>100%</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
