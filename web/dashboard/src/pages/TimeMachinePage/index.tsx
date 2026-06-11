import './styles.css';

import { useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  History,
  Plus,
  Clock,
  AlertCircle,
  CheckCircle2,
  XCircle,
  Loader2,
  ArrowRight,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useReplays, useTimeMachineLimits } from '@/hooks/useTimeMachine';
import { usePlan } from '@/hooks/usePlan';
import { cn, formatDateTime } from '@/lib/utils';

const STATUS_CONFIG: Record<string, { label: string; icon: React.ReactNode }> = {
  pending: { label: 'Pending', icon: <Clock className="w-3 h-3" /> },
  scanning: { label: 'Scanning', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  replaying: { label: 'Replaying', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  diffing: { label: 'Diffing', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  reconciling: { label: 'Reconciling', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  completed: { label: 'Completed', icon: <CheckCircle2 className="w-3 h-3" /> },
  failed: { label: 'Failed', icon: <XCircle className="w-3 h-3" /> },
  cancelled: { label: 'Cancelled', icon: <AlertCircle className="w-3 h-3" /> },
};

function getStatusConfig(status: string) {
  return STATUS_CONFIG[status] ?? { label: status, icon: null };
}

export function TimeMachinePage() {
  const { data: replaysData, isLoading } = useReplays();
  const { data: limits } = useTimeMachineLimits();
  const { displayName } = usePlan();

  const replays = replaysData?.items ?? [];

  const stats = useMemo(() => {
    const active = replays.filter(
      (r) => !['completed', 'failed', 'cancelled'].includes(r.status)
    ).length;
    const total = replays.length;
    const executionsReplayed = replays.reduce(
      (sum, r) => sum + r.total_executions_replayed,
      0
    );
    const bugsFixed = replays.filter((r) => r.status === 'completed').length;
    return { active, total, executionsReplayed, bugsFixed };
  }, [replays]);

  const statCards = [
    { title: 'Active Replays', value: stats.active, icon: <Loader2 className="w-5 h-5 text-blue-500" /> },
    { title: 'Total Replays', value: stats.total, icon: <History className="w-5 h-5 text-brand-500" /> },
    { title: 'Executions Replayed', value: stats.executionsReplayed, icon: <Clock className="w-5 h-5 text-purple-500" /> },
    { title: 'Bugs Fixed', value: stats.bugsFixed, icon: <CheckCircle2 className="w-5 h-5 text-emerald-500" /> },
  ];

  return (
    <div className="tm-container">
      <div className="tm-header">
        <div className="tm-header-content">
          <h1 className="tm-title">
            <span className="tm-title-icon">
              <History className="w-5 h-5" />
            </span>
            Time Machine
          </h1>
          <p className="tm-subtitle">
            Rewind and fix production bugs as if they never happened
          </p>
        </div>
        <Button asChild className="tm-button">
          <Link to="/time-machine/new">
            <Plus className="w-4 h-4 mr-2" />
            New Replay
          </Link>
        </Button>
      </div>

      <div className="tm-stats-grid">
        {statCards.map((stat, idx) => (
          <div key={stat.title} className={cn("tm-stat-card", idx === 0 && "first")}>
            <div className="tm-stat-header">
              <span className="tm-stat-icon">{stat.icon}</span>
              <span className="tm-stat-label">{stat.title}</span>
            </div>
            <div className="tm-stat-value">
              {isLoading ? (
                <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
              ) : (
                stat.value.toLocaleString()
              )}
            </div>
          </div>
        ))}
      </div>

      {isLoading && (
        <div className="tm-loading">
          <Loader2 className="tm-loading-spinner" />
        </div>
      )}

      {!isLoading && replays.length === 0 && (
        <div className="tm-empty">
          <History className="tm-empty-icon" />
          <h3 className="tm-empty-title">No replays yet</h3>
          <p className="tm-empty-description">
            Create your first replay to rewind and fix production bugs. Select a time window and a target version, and the Time Machine will replay all executions to find differences.
          </p>
        </div>
      )}

      {!isLoading && replays.length > 0 && (
        <div className="tm-table-container">
          <div className="tm-table-header">
            <h2 className="tm-table-title">Replay Jobs</h2>
          </div>
          <div className="tm-table-wrapper">
            <table className="tm-table">
              <thead>
                <tr>
                  <th>Function</th>
                  <th>Time Window</th>
                  <th>Status</th>
                  <th>Progress</th>
                  <th className="text-right">Executions</th>
                  <th className="text-right">Changed</th>
                  <th>Created</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {replays.map((replay) => {
                  const statusCfg = getStatusConfig(replay.status);
                  const statusClass = replay.status === 'pending' ? 'tm-status-pending' :
                    ['scanning', 'replaying', 'diffing', 'reconciling'].includes(replay.status) ? 'tm-status-active' :
                    replay.status === 'completed' ? 'tm-status-completed' :
                    replay.status === 'failed' ? 'tm-status-failed' : 'tm-status-pending';
                  return (
                    <tr key={replay.id}>
                      <td>
                        <Link to={`/time-machine/${replay.id}`} className="tm-function-id">
                          {replay.function_id}
                        </Link>
                      </td>
                      <td className="tm-time-window">
                        {new Date(replay.window_start).toLocaleDateString()} –{' '}
                        {new Date(replay.window_end).toLocaleDateString()}
                      </td>
                      <td>
                        <span className={cn("tm-status", statusClass)}>
                          {statusCfg.icon}
                          {statusCfg.label}
                        </span>
                      </td>
                      <td>
                        <div className="tm-progress">
                          <div className="tm-progress-bar">
                            <div
                              className="tm-progress-fill"
                              style={{ width: `${replay.progress_percent}%` }}
                            />
                          </div>
                          <span className="tm-progress-value">{replay.progress_percent}%</span>
                        </div>
                      </td>
                      <td className="tm-numeric">
                        {replay.total_executions_replayed.toLocaleString()}
                      </td>
                      <td className="tm-numeric">
                        <span className={cn(
                          replay.total_executions_changed > 0 ? 'tm-numeric-highlight' : 'tm-numeric-muted'
                        )}>
                          {replay.total_executions_changed.toLocaleString()}
                        </span>
                      </td>
                      <td className="tm-time-window">
                        {formatDateTime(replay.created_at)}
                      </td>
                      <td className="tm-action">
                        <Link to={`/time-machine/${replay.id}`} className="tm-action-link">
                          <ArrowRight className="w-4 h-4" />
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {limits && (
        <div className="tm-limits-panel">
          <div className="tm-limits-header">
            <Clock className="tm-limits-icon" />
            <h2 className="tm-limits-title">Plan Limits — {displayName}</h2>
          </div>
          <div className="tm-limits-content">
            <div className="tm-limits-grid">
              <div className="tm-limit-item">
                <p className="tm-limit-label">Replay Window</p>
                <p className={cn("tm-limit-value", limits.unlimited && "tm-limit-value-unlimited")}>
                  {limits.unlimited ? 'Unlimited' : `${limits.replay_window_hours}h`}
                </p>
              </div>
              <div className="tm-limit-item">
                <p className="tm-limit-label">Max Executions per Replay</p>
                <p className={cn("tm-limit-value", limits.unlimited && "tm-limit-value-unlimited")}>
                  {limits.unlimited ? 'Unlimited' : limits.max_executions_per_replay.toLocaleString()}
                </p>
              </div>
              <div className="tm-limit-item">
                <p className="tm-limit-label">Concurrent Replays</p>
                <p className={cn("tm-limit-value", limits.unlimited && "tm-limit-value-unlimited")}>
                  {limits.unlimited ? 'Unlimited' : limits.max_concurrent_replays}
                </p>
              </div>
              <div className="tm-limit-item">
                <p className="tm-limit-label">Data Retention</p>
                <p className="tm-limit-value">{limits.data_retention_days} days</p>
              </div>
              <div className="tm-limit-item">
                <p className="tm-limit-label">Auto Reconciliation</p>
                <p className={cn("tm-limit-value", !limits.auto_reconciliation && "tm-limit-value-disabled")}>
                  {limits.auto_reconciliation ? (
                    <span className="tm-status tm-status-completed">Enabled</span>
                  ) : (
                    <span className="tm-limit-value-disabled">Not available</span>
                  )}
                </p>
              </div>
              <div className="tm-limit-item">
                <p className="tm-limit-label">Audit Certificates</p>
                <p className={cn("tm-limit-value", !limits.audit_certificates && "tm-limit-value-disabled")}>
                  {limits.audit_certificates ? (
                    <span className="tm-status tm-status-completed">Included</span>
                  ) : (
                    <span className="tm-limit-value-disabled">Not available</span>
                  )}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
