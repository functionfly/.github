import './styles.css';

import { Link, useParams } from 'react-router-dom';
import {
  History,
  Clock,
  Loader2,
  XCircle,
  CheckCircle2,
  AlertCircle,
  ArrowLeft,
  GitBranch,
  Play,
  Ban,
  AlertTriangle,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { EmptyState } from '@/components/ui';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  useReplay,
  useReplayItems,
  useDiffSummary,
  useCancelReplay,
  useStartReconciliation,
  useTimeMachineLimits,
} from '@/hooks/useTimeMachine';
import { cn, formatDateTime } from '@/lib/utils';
import { useState } from 'react';

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

const ACTIVE_STATUSES = new Set(['pending', 'scanning', 'replaying', 'diffing', 'reconciling']);

function getStatusConfig(status: string) {
  return STATUS_CONFIG[status] ?? { label: status, variant: 'secondary' as const, icon: null };
}

function DiffTypeBadge({ type }: { type: string | null }) {
  const config: Record<string, { variant: 'default' | 'secondary' | 'destructive' | 'success' | 'warning' | 'error'; label: string }> = {
    identical: { variant: 'secondary', label: 'Identical' },
    minor: { variant: 'warning', label: 'Minor' },
    major: { variant: 'destructive', label: 'Major' },
    breaking: { variant: 'error', label: 'Breaking' },
    error: { variant: 'error', label: 'Error' },
  };
  const cfg = config[type ?? ''] ?? { variant: 'secondary' as const, label: type ?? 'Unknown' };
  return <Badge variant={cfg.variant}>{cfg.label}</Badge>;
}

function PreviewBlock({ label, data }: { label: string; data: unknown }) {
  const text = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
  const truncated = text && text.length > 200 ? text.slice(0, 200) + '…' : text;
  return (
    <div className="space-y-1">
      <p className="text-xs font-medium text-text-secondary uppercase tracking-wide">{label}</p>
      <pre className="text-xs text-text-primary bg-bg-secondary/60 rounded-md p-2 overflow-x-auto max-h-24 font-mono">
        {truncated ?? '—'}
      </pre>
    </div>
  );
}

export function ReplayDetailPage() {
  const { id } = useParams<{ id: string }>();
  const replayId = id ?? '';

  const { data: replay, isLoading: replayLoading } = useReplay(replayId);
  const { data: itemsData, isLoading: itemsLoading } = useReplayItems(replayId);
  const { data: diffSummary, isLoading: diffLoading } = useDiffSummary(replayId);
  const { data: limits } = useTimeMachineLimits();
  const cancelReplay = useCancelReplay();
  const startReconciliation = useStartReconciliation();

  const [reconciliationMode, setReconciliationMode] = useState<'dry_run' | 'live'>('dry_run');
  const [showLiveConfirm, setShowLiveConfirm] = useState(false);

  const items = itemsData?.items ?? [];
  const changedItems = items.filter((item) => item.output_changed);
  const isActive = replay ? ACTIVE_STATUSES.has(replay.status) : false;
  const isCompleted = replay?.status === 'completed';
  const canReconcile = isCompleted && (limits?.auto_reconciliation || limits?.live_reconciliation);
  const canLiveReconcile = isCompleted && limits?.live_reconciliation;

  if (replayLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="w-8 h-8 animate-spin text-text-muted" />
      </div>
    );
  }

  if (!replay) {
    return (
      <div className="space-y-4">
        <Link
          to="/time-machine"
          className="inline-flex items-center gap-1.5 text-sm text-text-secondary hover:text-text-primary transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Time Machine
        </Link>
        <EmptyState
          icon={<AlertCircle className="h-8 w-8" />}
          title="Replay not found"
          description="The replay job you're looking for doesn't exist or has been removed."
          action={
            <Button asChild variant="outline">
              <Link to="/time-machine">View All Replays</Link>
            </Button>
          }
        />
      </div>
    );
  }

  const statusCfg = getStatusConfig(replay.status);

  const diffBreakdown = diffSummary?.breakdown;
  const diffTotal = diffSummary?.total_executions ?? 0;

  return (
    <div className="space-y-6">
      <Link
        to="/time-machine"
        className="inline-flex items-center gap-1.5 text-sm text-text-secondary hover:text-text-primary transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Time Machine
      </Link>

      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-text-primary tracking-tight">
              {replay.reason || 'Replay'}
            </h1>
            <Badge variant={statusCfg.variant} className="gap-1">
              {statusCfg.icon}
              {statusCfg.label}
            </Badge>
          </div>
          <p className="text-text-secondary text-sm flex items-center gap-2">
            <Clock className="w-3.5 h-3.5" />
            Created {formatDateTime(replay.created_at)}
            {replay.incident_url && (
              <>
                <span className="text-text-muted">·</span>
                <a
                  href={replay.incident_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-brand-500 hover:underline"
                >
                  View Incident
                </a>
              </>
            )}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {isActive && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => cancelReplay.mutate(replay.id)}
              disabled={cancelReplay.isPending}
            >
              {cancelReplay.isPending ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Ban className="w-4 h-4 mr-2" />
              )}
              Cancel
            </Button>
          )}
          {canReconcile && (
            <div className="flex items-center gap-2">
              {canLiveReconcile && (
                <Select
                  value={reconciliationMode}
                  onValueChange={(v) => setReconciliationMode(v as 'dry_run' | 'live')}
                >
                  <SelectTrigger className="w-[140px] h-9">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="dry_run">Dry Run</SelectItem>
                    <SelectItem value="live">Live</SelectItem>
                  </SelectContent>
                </Select>
              )}
              <Button
                size="sm"
                onClick={() => {
                  if (reconciliationMode === 'live' && canLiveReconcile) {
                    setShowLiveConfirm(true);
                  } else {
                    startReconciliation.mutate({ id: replay.id, dryRun: reconciliationMode === 'dry_run' });
                  }
                }}
                disabled={startReconciliation.isPending}
                variant={reconciliationMode === 'live' ? 'default' : 'default'}
                className={reconciliationMode === 'live' ? 'bg-amber-500 hover:bg-amber-600' : ''}
              >
                {startReconciliation.isPending ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <Play className="w-4 h-4 mr-2" />
                )}
                {reconciliationMode === 'live' ? 'Go Live' : 'Start Dry Run'}
              </Button>
            </div>
          )}
        </div>
      </div>

      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <History className="w-4 h-4 text-text-muted" />
            Progress
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between text-sm">
            <span className="text-text-secondary">
              Phase:{' '}
              <span className="text-text-primary font-medium">
                {replay.current_phase ?? 'Initializing'}
              </span>
            </span>
            <span className="text-text-primary font-semibold">{replay.progress_percent}%</span>
          </div>
          <Progress value={replay.progress_percent} className="h-2" />
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-2">
            {[
              { label: 'Found', value: replay.total_executions_found, color: 'text-text-primary' },
              { label: 'Replayed', value: replay.total_executions_replayed, color: 'text-blue-400' },
              { label: 'Changed', value: replay.total_executions_changed, color: 'text-amber-400' },
              { label: 'Failed', value: replay.total_executions_failed, color: 'text-red-400' },
            ].map((stat) => (
              <div key={stat.label} className="text-center space-y-1">
                <p className={cn('text-xl font-bold font-mono tabular-nums', stat.color)}>
                  {stat.value.toLocaleString()}
                </p>
                <p className="text-xs text-text-secondary">{stat.label}</p>
              </div>
            ))}
          </div>
          {replay.error_message && (
            <div className="flex items-start gap-2 p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-sm text-red-400">
              <XCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
              {replay.error_message}
            </div>
          )}
        </CardContent>
      </Card>

      {diffLoading ? (
        <Card className="border-theme bg-card">
          <CardContent className="flex items-center justify-center py-8">
            <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
          </CardContent>
        </Card>
      ) : diffSummary ? (
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <GitBranch className="w-4 h-4 text-text-muted" />
              Diff Summary
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
              {diffBreakdown &&
                (['identical', 'minor', 'major', 'breaking', 'error'] as const).map((type) => {
                  const count = diffBreakdown[type];
                  const pct = diffTotal > 0 ? Math.round((count / diffTotal) * 100) : 0;
                  return (
                    <div
                      key={type}
                      className="rounded-lg border border-white/10 bg-bg-secondary/40 p-3 text-center space-y-1"
                    >
                      <DiffTypeBadge type={type} />
                      <p className="text-lg font-bold text-text-primary font-mono tabular-nums">
                        {count}
                      </p>
                      <p className="text-xs text-text-muted">{pct}%</p>
                    </div>
                  );
                })}
            </div>
          </CardContent>
        </Card>
      ) : null}

      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base">Changed Executions</CardTitle>
        </CardHeader>
        <CardContent>
          {itemsLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
            </div>
          ) : changedItems.length === 0 ? (
            <p className="text-sm text-text-secondary text-center py-6">
              {items.length === 0
                ? 'No execution items to display yet.'
                : 'No executions changed — all outputs are identical.'}
            </p>
          ) : (
            <div className="space-y-3">
              {changedItems.map((item) => (
                <div
                  key={item.id}
                  className="rounded-lg border border-white/10 bg-bg-secondary/30 p-4 space-y-3"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <code className="text-xs font-mono text-text-secondary bg-bg-secondary rounded px-1.5 py-0.5">
                        {item.original_execution_id}
                      </code>
                      <DiffTypeBadge type={item.diff_type} />
                    </div>
                    <span className="text-xs text-text-muted">
                      {item.original_version} → {replay.target_version}
                    </span>
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <PreviewBlock label="Original Output" data={item.original_output} />
                    <PreviewBlock label="New Output" data={item.new_output} />
                  </div>
                  {item.diff_summary && (
                    <p className="text-xs text-text-secondary">{item.diff_summary}</p>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <AlertDialog open={showLiveConfirm} onOpenChange={setShowLiveConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-amber-500" />
              Live Reconciliation Warning
            </AlertDialogTitle>
            <AlertDialogDescription>
              You are about to run <strong>live reconciliation</strong>. This will modify actual execution
              outputs based on the replay results. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                setShowLiveConfirm(false);
                startReconciliation.mutate({ id: replay.id, dryRun: false });
              }}
              className="bg-amber-500 hover:bg-amber-600"
            >
              Confirm Live Reconciliation
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
