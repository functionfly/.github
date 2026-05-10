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
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { EmptyState } from '@/components/ui';
import { useReplays, useTimeMachineLimits } from '@/hooks/useTimeMachine';
import { usePlan } from '@/hooks/usePlan';
import { cn, formatDateTime } from '@/lib/utils';

const STATUS_CONFIG: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' | 'success' | 'warning' | 'error'; icon: React.ReactNode }> = {
  pending: { label: 'Pending', variant: 'warning', icon: <Clock className="w-3 h-3" /> },
  scanning: { label: 'Scanning', variant: 'default', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  replaying: { label: 'Replaying', variant: 'default', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  diffing: { label: 'Diffing', variant: 'default', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  reconciling: { label: 'Reconciling', variant: 'default', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
  completed: { label: 'Completed', variant: 'success', icon: <CheckCircle2 className="w-3 h-3" /> },
  failed: { label: 'Failed', variant: 'error', icon: <XCircle className="w-3 h-3" /> },
  cancelled: { label: 'Cancelled', variant: 'secondary', icon: <AlertCircle className="w-3 h-3" /> },
};

function getStatusConfig(status: string) {
  return STATUS_CONFIG[status] ?? { label: status, variant: 'secondary' as const, icon: null };
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
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-primary tracking-tight">
            Time Machine
          </h1>
          <p className="text-text-secondary mt-1">
            Rewind and fix production bugs as if they never happened
          </p>
        </div>
        <Button asChild>
          <Link to="/time-machine/new">
            <Plus className="w-4 h-4 mr-2" />
            New Replay
          </Link>
        </Button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat) => (
          <Card key={stat.title} className="border-theme bg-card">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
                {stat.icon}
                {stat.title}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
              ) : (
                <p className="text-2xl font-semibold text-text-primary">
                  {stat.value.toLocaleString()}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-text-muted" />
        </div>
      )}

      {!isLoading && replays.length === 0 && (
        <EmptyState
          icon={<History className="h-8 w-8" />}
          title="No replays yet"
          description="Create your first replay to rewind and fix production bugs. Select a time window and a target version, and the Time Machine will replay all executions to find differences."
          action={
            <Button asChild>
              <Link to="/time-machine/new">
                <Plus className="w-4 h-4 mr-2" />
                Create First Replay
              </Link>
            </Button>
          }
        />
      )}

      {!isLoading && replays.length > 0 && (
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base">Replay Jobs</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/10">
                    <th className="text-left py-3 px-4 text-text-secondary font-medium">Function</th>
                    <th className="text-left py-3 px-4 text-text-secondary font-medium">Time Window</th>
                    <th className="text-left py-3 px-4 text-text-secondary font-medium">Status</th>
                    <th className="text-left py-3 px-4 text-text-secondary font-medium">Progress</th>
                    <th className="text-right py-3 px-4 text-text-secondary font-medium">Executions</th>
                    <th className="text-right py-3 px-4 text-text-secondary font-medium">Changed</th>
                    <th className="text-left py-3 px-4 text-text-secondary font-medium">Created</th>
                    <th className="w-10 py-3 px-4" />
                  </tr>
                </thead>
                <tbody>
                  {replays.map((replay) => {
                    const statusCfg = getStatusConfig(replay.status);
                    return (
                      <tr
                        key={replay.id}
                        className="border-b border-white/5 hover:bg-white/[0.02] transition-colors"
                      >
                        <td className="py-3 px-4">
                          <Link
                            to={`/time-machine/${replay.id}`}
                            className="text-text-primary hover:text-brand-500 font-medium transition-colors"
                          >
                            {replay.function_id}
                          </Link>
                        </td>
                        <td className="py-3 px-4 text-text-secondary whitespace-nowrap">
                          {new Date(replay.window_start).toLocaleDateString()} –{' '}
                          {new Date(replay.window_end).toLocaleDateString()}
                        </td>
                        <td className="py-3 px-4">
                          <Badge variant={statusCfg.variant} className="gap-1">
                            {statusCfg.icon}
                            {statusCfg.label}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 min-w-[120px]">
                          <div className="flex items-center gap-2">
                            <Progress value={replay.progress_percent} className="h-1.5 flex-1" />
                            <span className="text-xs text-text-muted w-8 text-right">
                              {replay.progress_percent}%
                            </span>
                          </div>
                        </td>
                        <td className="py-3 px-4 text-right text-text-primary font-mono tabular-nums">
                          {replay.total_executions_replayed.toLocaleString()}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <span
                            className={cn(
                              'font-mono tabular-nums',
                              replay.total_executions_changed > 0
                                ? 'text-amber-500'
                                : 'text-text-muted'
                            )}
                          >
                            {replay.total_executions_changed.toLocaleString()}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-text-secondary whitespace-nowrap">
                          {formatDateTime(replay.created_at)}
                        </td>
                        <td className="py-3 px-4">
                          <Link
                            to={`/time-machine/${replay.id}`}
                            className="text-text-muted hover:text-text-primary transition-colors"
                          >
                            <ArrowRight className="w-4 h-4" />
                          </Link>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {limits && (
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Clock className="w-4 h-4 text-text-muted" />
              Plan Limits — {displayName}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div className="space-y-1">
                <p className="text-sm text-text-secondary">Replay Window</p>
                <p className="text-lg font-semibold text-text-primary">
                  {limits.unlimited ? 'Unlimited' : `${limits.replay_window_hours}h`}
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-text-secondary">Max Executions per Replay</p>
                <p className="text-lg font-semibold text-text-primary">
                  {limits.unlimited ? 'Unlimited' : limits.max_executions_per_replay.toLocaleString()}
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-text-secondary">Concurrent Replays</p>
                <p className="text-lg font-semibold text-text-primary">
                  {limits.unlimited ? 'Unlimited' : limits.max_concurrent_replays}
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-text-secondary">Data Retention</p>
                <p className="text-lg font-semibold text-text-primary">
                  {limits.data_retention_days} days
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-text-secondary">Auto Reconciliation</p>
                <p className="text-lg font-semibold text-text-primary">
                  {limits.auto_reconciliation ? (
                    <span className="text-emerald-500">Enabled</span>
                  ) : (
                    <span className="text-text-muted">Not available</span>
                  )}
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-text-secondary">Audit Certificates</p>
                <p className="text-lg font-semibold text-text-primary">
                  {limits.audit_certificates ? (
                    <span className="text-emerald-500">Included</span>
                  ) : (
                    <span className="text-text-muted">Not available</span>
                  )}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
