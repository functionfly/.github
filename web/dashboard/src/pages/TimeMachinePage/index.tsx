import './styles.css';

import { useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  History,
  Plus,
  Clock,
  CheckCircle2,
  Loader2,
  ArrowRight,
} from 'lucide-react';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
} from '@/components/containment';
import { useReplays, useTimeMachineLimits } from '@/hooks/useTimeMachine';
import { usePlan } from '@/hooks/usePlan';
import { cn, formatDateTime } from '@/lib/utils';

const STATUS_MAP: Record<string, 'live' | 'pending' | 'revoked'> = {
  completed: 'live',
  failed: 'revoked',
  cancelled: 'revoked',
  pending: 'pending',
  scanning: 'pending',
  replaying: 'pending',
  diffing: 'pending',
  reconciling: 'pending',
};

const STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  scanning: 'Scanning',
  replaying: 'Replaying',
  diffing: 'Diffing',
  reconciling: 'Reconciling',
  completed: 'Completed',
  failed: 'Failed',
  cancelled: 'Cancelled',
};

function getStatusPill(status: string) {
  const mapped = STATUS_MAP[status] ?? 'pending';
  const label = STATUS_LABELS[status] ?? status;
  return { mapped, label };
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

  return (
    <div className="tm-page">
      <PageGrid />

      {/* Hero Chamber */}
      <Chamber className="tm-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE TM-01" secondary="Time Machine" position="top-right" />

        <div className="tm-hero__header">
          <div className="tm-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="tm-hero__title">Time Machine</h1>
          </div>
          <p className="tm-hero__subtitle">
            Rewind and fix production bugs as if they never happened
          </p>
          <div className="tm-hero__actions">
            <Link to="/time-machine/new">
              <SealedButton iconLeft={<Plus className="h-4 w-4" />}>
                New Replay
              </SealedButton>
            </Link>
          </div>
        </div>

        <GaugeStrip>
          <Gauge isFirst data={{ value: stats.active, label: 'Active Replays' }} />
          <Gauge data={{ value: stats.total, label: 'Total Replays' }} />
          <Gauge data={{ value: stats.executionsReplayed.toLocaleString(), label: 'Executions Replayed' }} />
          <Gauge data={{ value: stats.bugsFixed, label: 'Bugs Fixed' }} />
        </GaugeStrip>
      </Chamber>

      {/* Loading */}
      {isLoading && (
        <div className="tm-loading">
          <Loader2 className="tm-loading__spinner" />
        </div>
      )}

      {/* Empty State */}
      {!isLoading && replays.length === 0 && (
        <Chamber className="tm-empty">
          <History className="tm-empty__icon" />
          <h3 className="tm-empty__title">No replays yet</h3>
          <p className="tm-empty__desc">
            Create your first replay to rewind and fix production bugs. Select a time window and a target version, and the Time Machine will replay all executions to find differences.
          </p>
        </Chamber>
      )}

      {/* Replay Table */}
      {!isLoading && replays.length > 0 && (
        <Chamber className="tm-table-chamber">
          <CornerBrace position="tr" />
          <CornerBrace position="bl" />

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
                  <th className="tm-th-right">Executions</th>
                  <th className="tm-th-right">Changed</th>
                  <th>Created</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {replays.map((replay) => {
                  const { mapped, label } = getStatusPill(replay.status);
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
                        <StatusPill status={mapped} label={label} />
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
                          <ArrowRight className="h-4 w-4" />
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </Chamber>
      )}

      {/* Plan Limits */}
      {limits && (
        <Chamber className="tm-limits">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag primary="PLAN" secondary={displayName} position="top-right" />

          <div className="tm-limits__header">
            <Clock className="tm-limits__icon" />
            <h2 className="tm-limits__title">Plan Limits — {displayName}</h2>
          </div>

          <div className="tm-limits-grid">
            <LimitItem label="Replay Window" value={limits.unlimited ? 'Unlimited' : `${limits.replay_window_hours}h`} highlight={limits.unlimited} />
            <LimitItem label="Max Executions per Replay" value={limits.unlimited ? 'Unlimited' : limits.max_executions_per_replay.toLocaleString()} highlight={limits.unlimited} />
            <LimitItem label="Concurrent Replays" value={limits.unlimited ? 'Unlimited' : limits.max_concurrent_replays} highlight={limits.unlimited} />
            <LimitItem label="Data Retention" value={`${limits.data_retention_days} days`} />
            <LimitItem label="Auto Reconciliation" value={limits.auto_reconciliation ? 'Enabled' : 'Not available'} highlight={limits.auto_reconciliation} disabled={!limits.auto_reconciliation} />
            <LimitItem label="Audit Certificates" value={limits.audit_certificates ? 'Included' : 'Not available'} highlight={limits.audit_certificates} disabled={!limits.audit_certificates} />
          </div>
        </Chamber>
      )}
    </div>
  );
}

function LimitItem({ label, value, highlight, disabled }: { label: string; value: string | number; highlight?: boolean; disabled?: boolean }) {
  return (
    <div className="tm-limit-item">
      <p className="tm-limit-label">{label}</p>
      <p className={cn('tm-limit-value', highlight && 'tm-limit-value--ok', disabled && 'tm-limit-value--disabled')}>
        {value}
      </p>
    </div>
  );
}
