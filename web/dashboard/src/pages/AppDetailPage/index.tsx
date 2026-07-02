import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ROUTES } from '@/lib/constants';
import { useAuthStore } from '@/stores/authStore';
import type { AppStatus, BackendStatus, Deployment } from '@/types';
import type { SecretMetadata } from '@/types/vault';
import type { APIKey } from '@/types/api-key';
import {
  useAppAnalytics,
  useAppStatus,
  useAppSecrets,
  useCreateBackend,
  useDeleteApp,
  useDeleteBackend,
  useUpdateApp,
  useDeployments,
  useProvisioningStatus,
  useRollbackDeployment,
  useVaultSecrets,
  useCreateSecret,
  useDeleteSecret,
  useAPIKeys,
  useCreateAPIKey,
  useDeleteAPIKey,
  useRotateAPIKey,
  usePageTitle,
  appKeys,
} from '@/hooks';
import {
  Activity,
  AlertCircle,
  ArrowLeft,
  BarChart3,
  Building2,
  CheckCircle2,
  ChevronRight,
  Clock,
  Copy,
  DollarSign,
  ExternalLink,
  Eye,
  EyeOff,
  Globe,
  Key,
  Loader2,
  Lock,
  MoreVertical,
  Plus,
  RefreshCw,
  RotateCw,
  Server,
  Settings,
  Shield,
  Trash2,
  TrendingDown,
  Upload,
  XCircle,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { AreaChart } from '@/components/ui/chart-area';
import { LineChart } from '@/components/ui/chart-line';
import { BarChart } from '@/components/ui/chart-bar';
import { AddBackendDialog } from './AddBackendDialog';
import { apiKeysApi } from '@/api/apikeys';

const APP_PATH_UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/** Normalise snake_case API fields to camelCase. */
function normaliseApp(raw: Record<string, unknown>) {
  return {
    id: String(raw.id ?? ''),
    name: String(raw.name ?? ''),
    slug: String(raw.slug ?? ''),
    tenantId: String(raw.tenantId ?? raw.tenant_id ?? ''),
    deployUrl: String(raw.deployUrl ?? raw.deploy_url ?? ''),
    createdAt: String(raw.createdAt ?? raw.created_at ?? ''),
  };
}

/** Safely parse an ISO date, returning a fallback for invalid values. */
function safeDate(val: unknown): Date {
  if (!val) return new Date();
  const d = new Date(String(val));
  return isNaN(d.getTime()) ? new Date() : d;
}

// ─── Sub-components ──────────────────────────────────────────────────────────

function StatusDot({ ok, state }: { ok?: boolean; state?: string }) {
  if (state === 'open')
    return <span className="w-2.5 h-2.5 rounded-full animate-pulse" style={{ background: 'var(--status-revoked)' }} />;
  if (state === 'half-open')
    return <span className="w-2.5 h-2.5 rounded-full animate-pulse" style={{ background: 'var(--status-pending)' }} />;
  if (ok === true) return <span className="w-2.5 h-2.5 rounded-full" style={{ background: 'var(--status-ok)' }} />;
  if (ok === false) return <span className="w-2.5 h-2.5 rounded-full" style={{ background: 'var(--status-revoked)' }} />;
  return <span className="w-2.5 h-2.5 rounded-full" style={{ background: 'var(--text-faint)', opacity: 0.4 }} />;
}

// Check if the health check response indicates a degraded state (fallback probe succeeded)
function isDegradedHealthCheck(latestHealthCheck?: { ok?: boolean; statusCode?: number; errorMessage?: string }): boolean {
  if (!latestHealthCheck || !latestHealthCheck.ok) return false;
  // 404 from /healthz means fallback was used → degraded
  if (latestHealthCheck.statusCode === 404) return true;
  if (latestHealthCheck.errorMessage?.includes('fallback')) return true;
  return false;
}

function BackendStatusCard({
  backendStatus,
  onDelete,
}: {
  backendStatus: BackendStatus;
  onDelete?: (id: string) => void;
}) {
  const { backend, circuitState, latestHealthCheck } = backendStatus;

  const isHealthy = latestHealthCheck?.ok === true;
  const isDegraded = isHealthy && isDegradedHealthCheck(latestHealthCheck);
  const isUnhealthy = latestHealthCheck?.ok === false;
  const circuitOpen = circuitState?.state === 'open';
  const circuitHalfOpen = circuitState?.state === 'half-open';

  const statusLabel = circuitOpen
    ? 'Circuit Open'
    : circuitHalfOpen
      ? 'Half-Open'
      : isDegraded
        ? 'Degraded'
        : isHealthy
          ? 'Healthy'
          : isUnhealthy
            ? 'Unhealthy'
            : 'Unknown';

  const statusVariant =
    circuitOpen || isUnhealthy
      ? 'destructive'
      : circuitHalfOpen || isDegraded
        ? 'warning'
        : isHealthy
          ? 'success'
          : 'outline';

  return (
    <div className="flex items-center gap-4 p-4 rounded-[var(--radius-lg)] transition-colors" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
      <StatusDot ok={latestHealthCheck?.ok} state={circuitState?.state} />

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-medium text-sm capitalize" style={{ color: 'var(--text)' }}>{backend.provider}</span>
          <Badge variant="outline" className="text-xs font-mono">
            {backend.region}
          </Badge>
          <Badge
            variant={statusVariant as 'destructive' | 'warning' | 'success' | 'outline'}
            className="text-xs"
          >
            {statusLabel}
          </Badge>
        </div>
        <p className="text-xs truncate mt-1 font-mono" style={{ color: 'var(--text-faint)' }}>{backend.url}</p>
      </div>

      <div className="flex items-center gap-3 text-xs shrink-0" style={{ color: 'var(--text-faint)' }}>
        {latestHealthCheck && (
          <div className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            <span>{latestHealthCheck.latencyMs}ms</span>
          </div>
        )}
        {latestHealthCheck?.ok ? (
          <CheckCircle2 className="w-4 h-4" style={{ color: 'var(--status-ok)' }} />
        ) : latestHealthCheck?.ok === false ? (
          <XCircle className="w-4 h-4" style={{ color: 'var(--status-revoked)' }} />
        ) : null}
        {onDelete && (
          <button
            onClick={() => onDelete(backend.id)}
            className="ml-1 transition-colors hover:opacity-80"
            style={{ color: 'var(--status-revoked)' }}
            aria-label={`Remove ${backend.provider} backend`}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  sub,
  icon: Icon,
  trend,
  color = 'default',
}: {
  label: string;
  value: string | number;
  sub?: string;
  icon: React.ComponentType<{ className?: string }>;
  trend?: 'up' | 'down' | 'neutral';
  color?: 'default' | 'success' | 'warning' | 'danger';
}) {
  const colorMap: Record<string, { color: string; bg: string }> = {
    default: { color: 'var(--status-ok)', bg: 'rgba(143, 255, 208, 0.06)' },
    success: { color: 'var(--status-ok)', bg: 'rgba(143, 255, 208, 0.06)' },
    warning: { color: 'var(--status-pending)', bg: 'rgba(232, 196, 104, 0.06)' },
    danger: { color: 'var(--status-revoked)', bg: 'rgba(255, 107, 107, 0.06)' },
  };

  return (
    <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
      <CardContent className="p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <p className="text-xs font-medium uppercase tracking-wide mb-2" style={{ color: 'var(--text-faint)' }}>
              {label}
            </p>
            <p className="text-3xl font-bold tabular-nums" style={{ color: 'var(--text)' }}>{value}</p>
            {sub && <p className="text-xs mt-1" style={{ color: 'var(--text-faint)' }}>{sub}</p>}
          </div>
          <div
            className="w-10 h-10 rounded-[var(--radius-lg)] flex items-center justify-center"
            style={{ color: colorMap[color].color, background: colorMap[color].bg }}
          >
            <Icon className="w-5 h-5" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function LoadingState() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4 animate-pulse">
        <div className="w-9 h-9 rounded-[var(--radius)]" style={{ background: 'var(--panel-raised)' }} />
        <div className="w-14 h-14 rounded-[var(--radius-lg)]" style={{ background: 'var(--panel-raised)' }} />
        <div className="space-y-2 flex-1">
          <div className="h-6 rounded w-48" style={{ background: 'var(--panel-raised)' }} />
          <div className="h-4 rounded w-32" style={{ background: 'var(--panel-raised)' }} />
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-28 rounded-[var(--radius-lg)] animate-pulse" style={{ background: 'var(--panel-raised)' }} />
        ))}
      </div>
      <div className="h-64 rounded-[var(--radius-lg)] animate-pulse" style={{ background: 'var(--panel-raised)' }} />
    </div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="w-16 h-16 rounded-[var(--radius-lg)] flex items-center justify-center mb-5" style={{ background: 'rgba(255, 107, 107, 0.06)', border: '1px solid rgba(255, 107, 107, 0.2)' }}>
        <AlertCircle className="w-7 h-7" style={{ color: 'var(--status-revoked)' }} />
      </div>
      <h3 className="text-lg font-semibold mb-2" style={{ fontFamily: 'var(--font-display)' }}>Failed to load app</h3>
      <p className="text-sm mb-6 max-w-sm" style={{ color: 'var(--text-faint)' }}>{message}</p>
      <Button variant="outline" onClick={onRetry} className="gap-2">
        <RefreshCw className="w-4 h-4" />
        Try Again
      </Button>
    </div>
  );
}

function EmptyTabState({ icon: Icon, title, description, action }: { icon: React.ComponentType<{ className?: string }>; title: string; description: string; action?: React.ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center rounded-[var(--radius-lg)]" style={{ border: '1px dashed var(--panel-edge)' }}>
      <div className="w-14 h-14 rounded-[var(--radius-lg)] flex items-center justify-center mb-4" style={{ background: 'var(--panel)' }}>
        <Icon className="w-6 h-6" style={{ color: 'var(--text-faint)', opacity: 0.6 }} />
      </div>
      <h4 className="font-medium mb-1">{title}</h4>
      <p className="text-sm mb-4 max-w-xs" style={{ color: 'var(--text-faint)' }}>{description}</p>
      {action}
    </div>
  );
}

// ─── Overview Tab ─────────────────────────────────────────────────────────────

function OverviewTab({ data }: { data: AppStatus }) {
  const { app, backends } = data;
  const user = useAuthStore((s) => s.user);
  const healthyCount = backends.filter((b) => b.latestHealthCheck?.ok).length;
  const openCircuits = backends.filter((b) => b.circuitState?.state === 'open').length;
  const deployUrl = app.deployUrl || (app as unknown as Record<string, string>).deploy_url || '';

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard label="Total Backends" value={backends.length} sub="connected providers" icon={Server} color="default" />
        <StatCard label="Healthy" value={healthyCount} sub={`of ${backends.length} operational`} icon={CheckCircle2} color="success" />
        <StatCard label="Open Circuits" value={openCircuits} sub={openCircuits > 0 ? 'requires attention' : 'all circuits closed'} icon={Shield} color={openCircuits > 0 ? 'danger' : 'success'} />
      </div>

      <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
        <CardHeader className="pb-3">
          <CardTitle className="text-base" style={{ fontFamily: 'var(--font-display)' }}>App Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-xs uppercase tracking-wide mb-1" style={{ color: 'var(--text-faint)' }}>App ID</p>
              <div className="flex items-center gap-2">
                <code className="font-mono text-xs px-2 py-1 rounded-[var(--radius-sm)] truncate max-w-[160px]" style={{ background: 'var(--panel)', color: 'var(--text-dim)' }}>{app.id}</code>
                <button onClick={() => { navigator.clipboard.writeText(app.id); toast.success('App ID copied'); }} className="transition-colors" style={{ color: 'var(--text-faint)' }} aria-label="Copy app ID">
                  <Copy className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-1" style={{ color: 'var(--text-faint)' }}>Slug</p>
              <code className="font-mono text-xs px-2 py-1 rounded-[var(--radius-sm)]" style={{ background: 'var(--panel)', color: 'var(--text-dim)' }}>{app.slug}</code>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-1" style={{ color: 'var(--text-faint)' }}>Owner</p>
              <p className="text-sm font-medium" style={{ color: 'var(--text)' }}>{user?.name || user?.username || '—'}</p>
              {user?.email && <p className="text-xs mt-0.5 font-mono" style={{ color: 'var(--text-faint)' }}>{user.email}</p>}
              {user?.username && user?.name && <p className="text-xs mt-0.5 font-mono" style={{ color: 'var(--text-faint)' }}>@{user.username}</p>}
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-1" style={{ color: 'var(--text-faint)' }}>Plan</p>
              <Badge variant="outline" className="text-xs capitalize">{user?.plan || 'free'}</Badge>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-1" style={{ color: 'var(--text-faint)' }}>Created</p>
              <p className="text-sm" style={{ color: 'var(--text)' }}>{safeDate(app.createdAt || (app as unknown as Record<string, string>).created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-1" style={{ color: 'var(--text-faint)' }}>Tenant</p>
              <code className="font-mono text-xs px-2 py-1 rounded-[var(--radius-sm)] truncate max-w-[160px] block" style={{ background: 'var(--panel)', color: 'var(--text-dim)' }}>{app.tenantId || (app as unknown as Record<string, string>).tenant_id || '—'}</code>
            </div>
            {deployUrl && (
              <div className="col-span-2">
                <p className="text-xs uppercase tracking-wide mb-1" style={{ color: 'var(--text-faint)' }}>Production URL</p>
                <div className="flex items-center gap-2">
                  <a href={deployUrl} target="_blank" rel="noopener noreferrer" className="text-sm font-mono transition-colors" style={{ color: 'var(--accent)' }}>{deployUrl}</a>
                  <button onClick={() => { navigator.clipboard.writeText(deployUrl); toast.success('URL copied'); }} className="transition-colors" style={{ color: 'var(--text-faint)' }} aria-label="Copy production URL">
                    <Copy className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// ─── Backends Tab ─────────────────────────────────────────────────────────────

function BackendsTab({ data, appId }: { data: AppStatus; appId: string }) {
  const { backends } = data;
  const [addOpen, setAddOpen] = useState(false);
  const deleteBackend = useDeleteBackend(appId);
  const queryClient = useQueryClient();

  const handleAddSuccess = useCallback(() => {
    setAddOpen(false);
    queryClient.invalidateQueries({ queryKey: appKeys.all });
  }, [queryClient]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm" style={{ color: 'var(--text-faint)' }}>
          {backends.length} backend{backends.length !== 1 ? 's' : ''} configured
        </p>
        <Button variant="outline" size="sm" className="gap-2" onClick={() => setAddOpen(true)}>
          <Server className="w-3.5 h-3.5" />
          Add Backend
        </Button>
      </div>

      {backends.length > 0 ? (
        <div className="space-y-3">
          {backends.map((backendStatus) => (
            <BackendStatusCard
              key={backendStatus.backend.id}
              backendStatus={backendStatus}
              onDelete={(id) => deleteBackend.mutate(id)}
            />
          ))}
        </div>
      ) : (
        <EmptyTabState
          icon={Server}
          title="No backends configured"
          description="Add a backend provider to start deploying your app to the edge."
          action={
            <Button variant="outline" size="sm" className="gap-2" onClick={() => setAddOpen(true)}>
              <Server className="w-3.5 h-3.5" />
              Add Backend
            </Button>
          }
        />
      )}

      <AddBackendDialog
        appId={appId}
        open={addOpen}
        onOpenChange={setAddOpen}
        onSuccess={handleAddSuccess}
      />
    </div>
  );
}

// ─── Analytics Tab ────────────────────────────────────────────────────────────

function AnalyticsTab({ appId }: { appId: string }) {
  const [timeRange, setTimeRange] = useState<'24h' | '7d' | '30d'>('7d');
  const days = timeRange === '24h' ? 1 : timeRange === '7d' ? 7 : 30;

  const { data, isLoading } = useAppAnalytics(appId, days);

  const requestsChartData = useMemo(() => {
    if (!data?.requestsOverTime) return [];
    return data.requestsOverTime.map((d) => ({
      time: new Date(d.timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      success: d.success,
      errors: d.errors,
    }));
  }, [data]);

  const latencyChartData = useMemo(() => {
    if (!data?.latencyOverTime) return [];
    return data.latencyOverTime.map((d) => ({
      time: new Date(d.timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      p50: d.p50Ms,
      p95: d.p95Ms,
      p99: d.p99Ms,
    }));
  }, [data]);

  const errorChartData = useMemo(() => {
    if (!data?.topErrors) return [];
    return data.topErrors.map((e) => ({ name: `${e.statusCode}`, count: e.count }));
  }, [data]);

  const backendChartData = useMemo(() => {
    if (!data?.backendBreakdown) return [];
    return data.backendBreakdown.map((b) => ({ name: b.provider, requests: b.requests, avgLatency: Math.round(b.avgLatencyMs) }));
  }, [data]);

  const summary = data?.summary;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="w-6 h-6 animate-spin" style={{ color: 'var(--text-faint)' }} />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <p className="text-sm" style={{ color: 'var(--text-faint)' }}>
          Analytics for the last {timeRange === '24h' ? '24 hours' : timeRange === '7d' ? '7 days' : '30 days'}
        </p>
        <div className="inline-flex gap-1 p-1 rounded-[var(--radius-sm)]" style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}>
          {(['24h', '7d', '30d'] as const).map((range) => (
            <button key={range} onClick={() => setTimeRange(range)} className="px-3 py-1.5 text-xs font-medium rounded-[var(--radius-sm)] transition-all" style={{ background: timeRange === range ? 'var(--panel-raised)' : 'transparent', color: timeRange === range ? 'var(--text)' : 'var(--text-faint)', boxShadow: timeRange === range ? '0 1px 3px rgba(0,0,0,0.2)' : 'none' }}>
              {range}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <StatCard label="Total Requests" value={summary ? summary.totalRequests.toLocaleString() : '0'} icon={Activity} color="default" />
        <StatCard label="Avg Latency" value={summary ? `${Math.round(summary.avgLatencyMs)}ms` : '0ms'} icon={Zap} color="default" />
        <StatCard label="Error Rate" value={summary ? `${(summary.errorRate * 100).toFixed(1)}%` : '0%'} icon={TrendingDown} color={summary && summary.errorRate > 0.05 ? 'danger' : 'success'} />
        <StatCard label="Success Rate" value={summary ? `${(summary.successRate * 100).toFixed(1)}%` : '0%'} icon={CheckCircle2} color="success" />
      </div>

      {summary && (summary.totalExecutions > 0 || summary.totalCostCents > 0) && (
        <div className="grid grid-cols-2 gap-4">
          {summary.totalExecutions > 0 && <StatCard label="Function Executions" value={summary.totalExecutions.toLocaleString()} icon={Server} color="default" />}
          {summary.totalCostCents > 0 && <StatCard label="Total Cost" value={`$${(summary.totalCostCents / 100).toFixed(2)}`} icon={DollarSign} color="default" />}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {requestsChartData.length > 0 ? (
          <AreaChart data={requestsChartData} series={[{ key: 'success', name: 'Success', color: '#10b981', gradient: { from: '#10b981', to: '#10b98100' } }, { key: 'errors', name: 'Errors', color: '#ef4444', gradient: { from: '#ef4444', to: '#ef444400' } }]} title="Requests Over Time" xAxisKey="time" height={280} stacked showLegend />
        ) : (
          <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}><CardContent className="flex items-center justify-center h-[280px]"><p className="text-sm" style={{ color: 'var(--text-faint)' }}>No request data yet</p></CardContent></Card>
        )}
        {latencyChartData.length > 0 ? (
          <LineChart data={latencyChartData} series={[{ key: 'p50', name: 'P50', color: '#10b981' }, { key: 'p95', name: 'P95', color: '#f59e0b' }, { key: 'p99', name: 'P99', color: '#ef4444' }]} title="Latency Distribution" xAxisKey="time" height={280} showLegend yAxisFormatter={(v) => `${v}ms`} />
        ) : (
          <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}><CardContent className="flex items-center justify-center h-[280px]"><p className="text-sm" style={{ color: 'var(--text-faint)' }}>No latency data yet</p></CardContent></Card>
        )}
        {errorChartData.length > 0 ? (
          <BarChart data={errorChartData} series={[{ key: 'count', name: 'Errors', color: '#ef4444' }]} title="Error Breakdown" xAxisKey="name" height={280} showLabels />
        ) : (
          <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}><CardContent className="flex items-center justify-center h-[280px]"><div className="text-center"><CheckCircle2 className="w-8 h-8 mx-auto mb-2" style={{ color: 'var(--status-ok)' }} /><p className="text-sm" style={{ color: 'var(--text-faint)' }}>No errors in this period</p></div></CardContent></Card>
        )}
        {backendChartData.length > 0 ? (
          <BarChart data={backendChartData} series={[{ key: 'requests', name: 'Requests', color: '#6366f1' }]} title="Backend Performance" xAxisKey="name" height={280} showLabels />
        ) : (
          <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}><CardContent className="flex items-center justify-center h-[280px]"><p className="text-sm" style={{ color: 'var(--text-faint)' }}>No backend data yet</p></CardContent></Card>
        )}
      </div>

      {summary && summary.p95LatencyMs > 0 && (
        <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
          <CardHeader className="pb-3"><CardTitle className="text-base" style={{ fontFamily: 'var(--font-display)' }}>Latency Percentiles</CardTitle></CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-4">
              <div className="text-center"><p className="text-2xl font-bold tabular-nums" style={{ color: 'var(--text)' }}>{Math.round(summary.avgLatencyMs)}ms</p><p className="text-xs mt-1" style={{ color: 'var(--text-faint)' }}>Average</p></div>
              <div className="text-center"><p className="text-2xl font-bold tabular-nums" style={{ color: 'var(--status-pending)' }}>{summary.p95LatencyMs}ms</p><p className="text-xs mt-1" style={{ color: 'var(--text-faint)' }}>P95</p></div>
              <div className="text-center"><p className="text-2xl font-bold tabular-nums" style={{ color: 'var(--status-revoked)' }}>{summary.p99LatencyMs}ms</p><p className="text-xs mt-1" style={{ color: 'var(--text-faint)' }}>P99</p></div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

// ─── Deployments Tab ──────────────────────────────────────────────────────────

const PROVISIONING_COMPONENT_META: Record<string, { label: string; icon: typeof Upload }> = {
  user_db: { label: 'User Database', icon: Server },
  auth: { label: 'Authentication', icon: Shield },
  payments: { label: 'Payments', icon: DollarSign },
  email_workflows: { label: 'Email Workflows', icon: Upload },
  analytics: { label: 'Analytics', icon: BarChart3 },
};

function DeploymentsTab({ appId }: { appId: string }) {
  const { data, isLoading, refetch } = useDeployments(appId, 20);
  const rollback = useRollbackDeployment();
  const deployments = data?.deployments ?? [];
  const { data: provisioningData } = useProvisioningStatus();

  const provisioningResult = provisioningData && 'components' in provisioningData ? provisioningData : null;
  const provisioningComponents = provisioningResult?.components ?? {};
  const hasProvisioning = Object.keys(provisioningComponents).length > 0;

  const statusColor = (s: Deployment['status']) => {
    switch (s) {
      case 'success': return 'var(--status-ok)';
      case 'failed': return 'var(--status-revoked)';
      case 'rolled_back': return 'var(--status-pending)';
      default: return 'var(--text-faint)';
    }
  };

  const provisioningStatusColor = (s: string) => {
    switch (s) {
      case 'active': return 'var(--status-ok)';
      case 'failed': return 'var(--status-revoked)';
      case 'provisioning': return 'var(--accent)';
      default: return 'var(--text-faint)';
    }
  };

  if (isLoading) {
    return <div className="flex items-center justify-center py-16"><Loader2 className="w-6 h-6 animate-spin" style={{ color: 'var(--text-faint)' }} /></div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm" style={{ color: 'var(--text-faint)' }}>{deployments.length} deployment{deployments.length !== 1 ? 's' : ''}</p>
        <Button variant="outline" size="sm" className="gap-2" onClick={() => refetch()}>
          <RefreshCw className="w-3.5 h-3.5" />
          Refresh
        </Button>
      </div>

      {deployments.length > 0 ? (
        <div className="space-y-2">
          {deployments.map((d) => (
            <div key={d.id} className="flex items-center gap-4 p-4 rounded-[var(--radius-lg)]" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
              <div className="w-2.5 h-2.5 rounded-full shrink-0" style={{ background: statusColor(d.status) }} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm" style={{ color: 'var(--text)' }}>{d.provider}</span>
                  <Badge variant="outline" className="text-xs font-mono">{d.region}</Badge>
                  <Badge variant="outline" className="text-xs capitalize">{d.status.replace('_', ' ')}</Badge>
                </div>
                <p className="text-xs mt-1 font-mono" style={{ color: 'var(--text-faint)' }}>{d.id}</p>
              </div>
              <div className="flex items-center gap-2 shrink-0 text-xs" style={{ color: 'var(--text-faint)' }}>
                <span>{safeDate(d.createdAt || d.created_at).toLocaleString()}</span>
                {d.status === 'success' && (
                  <Button variant="ghost" size="sm" className="gap-1 h-7 px-2" onClick={() => rollback.mutate(d.id)}>
                    <RotateCw className="w-3 h-3" />
                    Rollback
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      ) : hasProvisioning ? (
        <div className="space-y-2">
          <div className="flex items-center gap-2 mb-3">
            <Badge variant="outline" className="text-xs capitalize" style={{ background: provisioningResult?.status === 'active' ? 'rgba(143,255,208,0.06)' : undefined, color: provisioningResult?.status === 'active' ? 'var(--status-ok)' : undefined, borderColor: provisioningResult?.status === 'active' ? 'rgba(143,255,208,0.3)' : undefined }}>
              Bundle: {provisioningResult?.bundle_slug || 'unknown'}
            </Badge>
            {provisioningResult?.duration_ms ? (
              <span className="text-xs" style={{ color: 'var(--text-faint)' }}>Provisioned in {provisioningResult.duration_ms < 1000 ? `${provisioningResult.duration_ms}ms` : `${(provisioningResult.duration_ms / 1000).toFixed(1)}s`}</span>
            ) : null}
          </div>
          {Object.entries(provisioningComponents).map(([key, state]) => {
            const meta = PROVISIONING_COMPONENT_META[key] || { label: key, icon: Upload };
            const Icon = meta.icon;
            const status = state.status;
            return (
              <div key={key} className="flex items-center gap-4 p-4 rounded-[var(--radius-lg)]" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
                <div className="w-2.5 h-2.5 rounded-full shrink-0" style={{ background: provisioningStatusColor(status) }} />
                <div className="w-8 h-8 rounded-[var(--radius)] flex items-center justify-center shrink-0" style={{ background: status === 'active' ? 'rgba(143,255,208,0.08)' : status === 'failed' ? 'rgba(255,99,99,0.08)' : 'var(--panel)' }}>
                  <Icon className="w-4 h-4" style={{ color: status === 'active' ? 'var(--status-ok)' : status === 'failed' ? 'var(--status-revoked)' : 'var(--text-faint)' }} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm" style={{ color: 'var(--text)' }}>{meta.label}</span>
                    <Badge variant="outline" className="text-xs capitalize">{status}</Badge>
                  </div>
                  {state.error && <p className="text-xs mt-1" style={{ color: 'var(--status-revoked)' }}>{state.error}</p>}
                </div>
                <div className="flex items-center gap-2 shrink-0 text-xs" style={{ color: 'var(--text-faint)' }}>
                  <span>{safeDate(state.timestamp).toLocaleString()}</span>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <EmptyTabState icon={Upload} title="No deployments yet" description="Deploy your app to see deployment history here." />
      )}
    </div>
  );
}

// ─── Secrets Tab (Vault integration) ──────────────────────────────────────────

function SecretsTab({ appId }: { appId: string }) {
  const { data: vaultData, isLoading: vaultLoading } = useVaultSecrets();
  const { data: appSecretsData, isLoading: appSecretsLoading } = useAppSecrets(appId);
  const createSecret = useCreateSecret();
  const deleteSecret = useDeleteSecret();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newValue, setNewValue] = useState('');
  const [showValue, setShowValue] = useState<Record<string, boolean>>({});

  const secrets = vaultData?.secrets ?? [];
  const appSecrets = appSecretsData?.secrets ?? [];
  const isLoading = vaultLoading || appSecretsLoading;

  const handleCreate = useCallback(() => {
    if (!newName.trim() || !newValue.trim()) {
      toast.error('Name and value are required');
      return;
    }
    createSecret.mutate(
      { name: newName, secret_type: 'api_key', encrypted_data: { ciphertext: newValue, iv: '', salt: '', tag: '', key_version: 1 } },
      { onSuccess: () => { setShowCreate(false); setNewName(''); setNewValue(''); } }
    );
  }, [newName, newValue, createSecret]);

  if (isLoading) {
    return <div className="flex items-center justify-center py-16"><Loader2 className="w-6 h-6 animate-spin" style={{ color: 'var(--text-faint)' }} /></div>;
  }

  return (
    <div className="space-y-6">
      {/* Vault secrets */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-sm font-medium" style={{ color: 'var(--text)' }}>Vault Secrets</h3>
            <p className="text-xs mt-0.5" style={{ color: 'var(--text-faint)' }}>Encrypted secrets stored in the zero-knowledge vault</p>
          </div>
          <Button variant="outline" size="sm" className="gap-2" onClick={() => setShowCreate(true)}>
            <Plus className="w-3.5 h-3.5" />
            Add Secret
          </Button>
        </div>

        {secrets.length > 0 ? (
          <div className="space-y-2">
            {secrets.map((s: SecretMetadata) => (
              <div key={s.id} className="flex items-center gap-4 p-3 rounded-[var(--radius-lg)]" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
                <Lock className="w-4 h-4 shrink-0" style={{ color: 'var(--text-faint)' }} />
                <div className="flex-1 min-w-0">
                  <span className="text-sm font-medium" style={{ color: 'var(--text)' }}>{s.name}</span>
                  <div className="flex items-center gap-2 mt-0.5">
                    <Badge variant="outline" className="text-xs">{s.secret_type}</Badge>
                    {s.current_version && <span className="text-xs" style={{ color: 'var(--text-faint)' }}>v{s.current_version}</span>}
                    <span className="text-xs" style={{ color: 'var(--text-faint)' }}>{s.access_count} accesses</span>
                  </div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => deleteSecret.mutate(s.id)} aria-label={`Delete ${s.name}`}>
                    <Trash2 className="w-3.5 h-3.5" style={{ color: 'var(--status-revoked)' }} />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <EmptyTabState icon={Lock} title="No vault secrets" description="Store encrypted secrets for this app in the zero-knowledge vault." />
        )}
      </div>

      {/* Provider-level secrets (e.g. Fly.io env) */}
      {appSecrets.length > 0 && (
        <div className="space-y-4">
          <div>
            <h3 className="text-sm font-medium" style={{ color: 'var(--text)' }}>Provider Secrets</h3>
            <p className="text-xs mt-0.5" style={{ color: 'var(--text-faint)' }}>Environment variables set on your deployment provider</p>
          </div>
          <div className="space-y-2">
            {appSecrets.map((s) => (
              <div key={s.key} className="flex items-center gap-3 p-3 rounded-[var(--radius-lg)]" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
                <code className="text-xs font-mono font-medium" style={{ color: 'var(--text)' }}>{s.key}</code>
                <div className="flex-1" />
                <code className="text-xs font-mono" style={{ color: 'var(--text-faint)' }}>
                  {showValue[s.key] ? (s.value ?? '••••••••') : '••••••••'}
                </code>
                <button onClick={() => setShowValue((p) => ({ ...p, [s.key]: !p[s.key] }))} className="transition-colors" style={{ color: 'var(--text-faint)' }}>
                  {showValue[s.key] ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Create secret dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Vault Secret</DialogTitle>
            <DialogDescription>Store a new encrypted secret in the vault. The value is encrypted client-side before storage.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="secret-name">Name</Label>
              <Input id="secret-name" placeholder="DATABASE_URL" value={newName} onChange={(e) => setNewName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="secret-value">Value</Label>
              <Input id="secret-value" type="password" placeholder="secret value" value={newValue} onChange={(e) => setNewValue(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={createSecret.isPending}>
              {createSecret.isPending ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Lock className="w-4 h-4 mr-2" />}
              Store Secret
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ─── API Keys Tab ─────────────────────────────────────────────────────────────

function ApiKeysTab({ appId }: { appId: string }) {
  const { data: keysData, isLoading } = useAPIKeys();
  const createKey = useCreateAPIKey();
  const deleteKey = useDeleteAPIKey();
  const rotateKey = useRotateAPIKey();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  const allKeys: APIKey[] = Array.isArray((keysData as { data?: APIKey[] })?.data)
    ? (keysData as { data: APIKey[] }).data
    : Array.isArray(keysData) ? keysData as unknown as APIKey[] : [];

  const appKeys = allKeys.filter((k) =>
    k.key_type === 'function' || k.key_type === 'agent' || k.key_type === 'runtime'
  );

  const handleCreate = useCallback(async () => {
    if (!newName.trim()) {
      toast.error('Key name is required');
      return;
    }
    try {
      const res = await apiKeysApi.create({
        name: newName,
        description: newDescription || undefined,
        key_type: 'function',
      });
      const key = (res as unknown as { data?: { key?: string; key_prefix?: string } })?.data;
      if (key?.key) {
        setCreatedKey(key.key);
      } else {
        setCreatedKey(null);
        setShowCreate(false);
      }
      setNewName('');
      setNewDescription('');
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to create API key');
    }
  }, [newName, newDescription]);

  if (isLoading) {
    return <div className="flex items-center justify-center py-16"><Loader2 className="w-6 h-6 animate-spin" style={{ color: 'var(--text-faint)' }} /></div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm" style={{ color: 'var(--text-faint)' }}>{appKeys.length} key{appKeys.length !== 1 ? 's' : ''}</p>
        <Button variant="outline" size="sm" className="gap-2" onClick={() => { setCreatedKey(null); setShowCreate(true); }}>
          <Key className="w-3.5 h-3.5" />
          Generate Key
        </Button>
      </div>

      {appKeys.length > 0 ? (
        <div className="space-y-2">
          {appKeys.map((k) => (
            <div key={k.id} className="flex items-center gap-4 p-3 rounded-[var(--radius-lg)]" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
              <Key className="w-4 h-4 shrink-0" style={{ color: k.is_active ? 'var(--text-faint)' : 'var(--status-revoked)' }} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium" style={{ color: 'var(--text)' }}>{k.name}</span>
                  {!k.is_active && <Badge variant="destructive" className="text-xs">Revoked</Badge>}
                  <Badge variant="outline" className="text-xs">{k.key_type}</Badge>
                </div>
                <div className="flex items-center gap-2 mt-0.5">
                  <code className="text-xs font-mono" style={{ color: 'var(--text-faint)' }}>{k.key_prefix}••••••••</code>
                  {k.last_used_at && <span className="text-xs" style={{ color: 'var(--text-faint)' }}>used {new Date(k.last_used_at).toLocaleDateString()}</span>}
                </div>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => rotateKey.mutate({ id: k.id, data: {} })} aria-label={`Rotate ${k.name}`}>
                  <RotateCw className="w-3.5 h-3.5" />
                </Button>
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => deleteKey.mutate(k.id)} aria-label={`Delete ${k.name}`}>
                  <Trash2 className="w-3.5 h-3.5" style={{ color: 'var(--status-revoked)' }} />
                </Button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <EmptyTabState icon={Key} title="No API keys" description="Generate API keys to authenticate requests to your app." />
      )}

      {/* Create dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{createdKey ? 'API Key Created' : 'Generate API Key'}</DialogTitle>
            <DialogDescription>
              {createdKey ? 'Copy this key now — it will not be shown again.' : 'Create a new API key for programmatic access.'}
            </DialogDescription>
          </DialogHeader>
          {createdKey ? (
            <div className="space-y-4 py-4">
              <div className="p-3 rounded-[var(--radius)] font-mono text-sm break-all" style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)', color: 'var(--text)' }}>
                {createdKey}
              </div>
              <Button variant="outline" className="w-full gap-2" onClick={() => { navigator.clipboard.writeText(createdKey); toast.success('Key copied'); }}>
                <Copy className="w-4 h-4" />
                Copy to Clipboard
              </Button>
            </div>
          ) : (
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="key-name">Name</Label>
                <Input id="key-name" placeholder="Production API Key" value={newName} onChange={(e) => setNewName(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="key-desc">Description (optional)</Label>
                <Input id="key-desc" placeholder="Used for server-side requests" value={newDescription} onChange={(e) => setNewDescription(e.target.value)} />
              </div>
            </div>
          )}
          <DialogFooter>
            {createdKey ? (
              <Button onClick={() => { setCreatedKey(null); setShowCreate(false); }}>Done</Button>
            ) : (
              <>
                <Button variant="outline" onClick={() => setShowCreate(false)}>Cancel</Button>
                <Button onClick={handleCreate} disabled={createKey.isPending}>
                  {createKey.isPending ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Key className="w-4 h-4 mr-2" />}
                  Generate
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ─── Settings Tab ─────────────────────────────────────────────────────────────

function SettingsTab({ data, appId }: { data: AppStatus; appId: string }) {
  const { app } = data;
  const navigate = useNavigate();
  const updateApp = useUpdateApp(appId);
  const deleteApp = useDeleteApp();

  const [name, setName] = useState(app.name);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState('');

  const handleSaveName = useCallback(() => {
    if (!name.trim()) {
      toast.error('App name cannot be empty');
      return;
    }
    if (name === app.name) return;
    updateApp.mutate({ name: name.trim() });
  }, [name, app.name, updateApp]);

  const handleDelete = useCallback(() => {
    if (deleteConfirm !== app.slug) return;
    deleteApp.mutate(appId, {
      onSuccess: () => navigate(ROUTES.APPS),
    });
  }, [deleteConfirm, app.slug, appId, deleteApp, navigate]);

  return (
    <div className="space-y-6">
      {/* General */}
      <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
        <CardHeader className="pb-3">
          <CardTitle className="text-base" style={{ fontFamily: 'var(--font-display)' }}>General Settings</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label>App Name</Label>
            <div className="flex gap-2">
              <Input value={name} onChange={(e) => setName(e.target.value)} className="font-mono" />
              <Button variant="outline" size="sm" onClick={handleSaveName} disabled={updateApp.isPending || name === app.name}>
                {updateApp.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Save'}
              </Button>
            </div>
          </div>
          <div className="space-y-1.5">
            <Label>App Slug</Label>
            <Input value={app.slug} disabled className="font-mono" />
            <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
              Used in URLs: <code className="font-mono">/apps/{app.slug}</code> — contact support to change.
            </p>
          </div>
          <div className="space-y-1.5">
            <Label>Deploy URL</Label>
            <div className="flex items-center gap-2">
              <Input value={app.deployUrl || (app as unknown as Record<string, string>).deploy_url || ''} disabled className="font-mono" />
              <Button variant="ghost" size="icon" className="shrink-0" onClick={() => { navigator.clipboard.writeText(app.deployUrl || (app as unknown as Record<string, string>).deploy_url); toast.success('Copied'); }}>
                <Copy className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Custom Domains */}
      <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
        <CardHeader className="pb-3">
          <CardTitle className="text-base" style={{ fontFamily: 'var(--font-display)' }}>Custom Domains</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm mb-3" style={{ color: 'var(--text-faint)' }}>
            Configure custom domains for your app on the <Link to="/settings" className="underline" style={{ color: 'var(--accent)' }}>Settings → Domains</Link> page.
          </p>
          <div className="flex items-center gap-3 p-3 rounded-[var(--radius)]" style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}>
            <Globe className="w-4 h-4" style={{ color: 'var(--text-faint)' }} />
            <div className="flex-1">
              <code className="text-sm font-mono" style={{ color: 'var(--text)' }}>{app.slug}.functionfly.com</code>
              <p className="text-xs mt-0.5" style={{ color: 'var(--text-faint)' }}>Default domain (always active)</p>
            </div>
            <CheckCircle2 className="w-4 h-4" style={{ color: 'var(--status-ok)' }} />
          </div>
        </CardContent>
      </Card>

      {/* Webhooks */}
      <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
        <CardHeader className="pb-3">
          <CardTitle className="text-base" style={{ fontFamily: 'var(--font-display)' }}>Webhooks</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm mb-3" style={{ color: 'var(--text-faint)' }}>
            Configure webhooks for deployment events, health alerts, and more on the <Link to="/settings" className="underline" style={{ color: 'var(--accent)' }}>Settings → Webhooks</Link> page.
          </p>
          <EmptyTabState icon={Zap} title="No webhooks configured" description="Set up webhooks to receive notifications about app events." />
        </CardContent>
      </Card>

      {/* Danger Zone */}
      <Card style={{ background: 'rgba(255, 107, 107, 0.03)', borderColor: 'rgba(255, 107, 107, 0.2)', borderRadius: 'var(--radius-lg)' }}>
        <CardHeader className="pb-3">
          <CardTitle className="text-base" style={{ fontFamily: 'var(--font-display)', color: 'var(--status-revoked)' }}>Danger Zone</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-medium" style={{ color: 'var(--text)' }}>Delete this app</p>
              <p className="text-xs mt-0.5" style={{ color: 'var(--text-faint)' }}>
                Permanently delete this app and all its backends, deployments, and secrets. This action cannot be undone.
              </p>
            </div>
            <Button variant="destructive" size="sm" onClick={() => setShowDeleteDialog(true)}>
              Delete App
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Delete confirmation dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete App</DialogTitle>
            <DialogDescription>
              This will permanently delete <strong>{app.name}</strong> and all associated backends, deployments, and secrets. Type <code className="font-mono">{app.slug}</code> to confirm.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Input placeholder={app.slug} value={deleteConfirm} onChange={(e) => setDeleteConfirm(e.target.value)} className="font-mono" />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setShowDeleteDialog(false); setDeleteConfirm(''); }}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteConfirm !== app.slug || deleteApp.isPending}>
              {deleteApp.isPending ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Trash2 className="w-4 h-4 mr-2" />}
              Delete Forever
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function AppDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('overview');

  const { data, isLoading, error, refetch } = useAppStatus(slug || '');

  const appName = (data as AppStatus)?.app?.name;
  usePageTitle(appName ? `Apps / ${appName}` : 'App');

  useEffect(() => {
    if (!data?.app?.slug || !slug) return;
    const segment = decodeURIComponent(slug);
    if (APP_PATH_UUID_RE.test(segment)) {
      navigate(`/apps/${encodeURIComponent(data.app.slug)}`, { replace: true });
    }
  }, [data?.app?.slug, slug, navigate]);

  if (isLoading) return <LoadingState />;

  if (error || !data) {
    return (
      <ErrorState
        message={error instanceof Error ? error.message : 'Failed to load app details'}
        onRetry={() => refetch()}
      />
    );
  }

  const { app, backends } = data as AppStatus;
  const healthyCount = backends.filter((b) => b.latestHealthCheck?.ok).length;
  const overallHealthy = backends.length === 0 || healthyCount === backends.length;

  const tabs = [
    { id: 'overview', label: 'Overview', icon: Activity },
    { id: 'backends', label: 'Backends', icon: Server, count: backends.length },
    { id: 'analytics', label: 'Analytics', icon: BarChart3 },
    { id: 'deployments', label: 'Deployments', icon: Upload },
    { id: 'secrets', label: 'Secrets', icon: Lock },
    { id: 'api-keys', label: 'API Keys', icon: Key },
    { id: 'settings', label: 'Settings', icon: Settings },
  ];

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-faint)' }} aria-label="Breadcrumb">
        <Link to={ROUTES.APPS} className="transition-colors" style={{ color: 'var(--text-faint)' }}>Apps</Link>
        <ChevronRight className="w-3.5 h-3.5" />
        <span className="font-medium truncate" style={{ color: 'var(--text)' }}>{app.name}</span>
      </nav>

      {/* Header */}
      <div className="flex items-start gap-4">
        <Link to={ROUTES.APPS} aria-label="Back to apps">
          <Button variant="ghost" size="icon" className="shrink-0 mt-1">
            <ArrowLeft className="w-4 h-4" />
          </Button>
        </Link>

        <div className="flex items-start gap-4 flex-1 min-w-0">
          <div className="w-14 h-14 rounded-[var(--radius-lg)] flex items-center justify-center shrink-0" style={{ background: 'rgba(143, 255, 208, 0.08)', border: '1px solid rgba(143, 255, 208, 0.15)' }}>
            <Building2 className="w-7 h-7" style={{ color: 'var(--status-ok)' }} />
          </div>

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl font-bold" style={{ color: 'var(--text)', fontFamily: 'var(--font-display)' }}>{app.name}</h1>
              <Badge variant={overallHealthy ? 'default' : 'destructive'} className="text-xs" style={overallHealthy ? { background: 'rgba(143,255,208,0.06)', color: 'var(--status-ok)', borderColor: 'rgba(143,255,208,0.3)' } : undefined}>
                <span className="w-1.5 h-1.5 rounded-full mr-1.5" style={{ background: overallHealthy ? 'var(--status-ok)' : 'var(--status-revoked)' }} />
                {overallHealthy ? 'Healthy' : 'Degraded'}
              </Badge>
            </div>
            <p className="text-sm font-mono mt-1" style={{ color: 'var(--text-faint)' }}>{app.slug}</p>
            {(app.deployUrl || (app as unknown as Record<string, string>).deploy_url) && (
              <div className="flex items-center gap-2 mt-2">
                <ExternalLink className="w-3.5 h-3.5" style={{ color: 'var(--text-faint)' }} />
                <a href={app.deployUrl || (app as unknown as Record<string, string>).deploy_url} target="_blank" rel="noopener noreferrer" className="text-sm font-mono truncate transition-colors" style={{ color: 'var(--accent)' }}>{app.deployUrl || (app as unknown as Record<string, string>).deploy_url}</a>
                <button onClick={() => { navigator.clipboard.writeText(app.deployUrl || (app as unknown as Record<string, string>).deploy_url); toast.success('Deploy URL copied'); }} className="transition-colors" style={{ color: 'var(--text-faint)' }} aria-label="Copy deploy URL">
                  <Copy className="w-3.5 h-3.5" />
                </button>
                <Button variant="outline" size="sm" className="gap-1.5 shrink-0" onClick={() => window.open(app.deployUrl || (app as unknown as Record<string, string>).deploy_url, '_blank', 'noopener,noreferrer')}>
                  <ExternalLink className="w-3.5 h-3.5" />
                  Open
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <Button variant="outline" size="sm" className="gap-2" onClick={() => refetch()}>
            <RefreshCw className="w-3.5 h-3.5" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon" aria-label="More actions">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuItem onClick={() => { navigator.clipboard.writeText(app.id); toast.success('App ID copied'); }} className="gap-2">
                <Copy className="w-4 h-4" />
                Copy App ID
              </DropdownMenuItem>
              <DropdownMenuItem className="gap-2" onClick={() => setActiveTab('deployments')}>
                <Upload className="w-4 h-4" />
                View Deployments
              </DropdownMenuItem>
              <DropdownMenuItem className="gap-2" onClick={() => setActiveTab('secrets')}>
                <Lock className="w-4 w-4" />
                Manage Secrets
              </DropdownMenuItem>
              <DropdownMenuItem className="gap-2" onClick={() => navigate(`/apps/${app.slug}/bundle`)}>
                <Settings className="w-4 w-4" />
                Bundle Config
              </DropdownMenuItem>
              <DropdownMenuItem className="gap-2" onClick={() => navigate(`${ROUTES.API_KEYS}?type=function`)}>
                <Key className="w-4 h-4" />
                API Keys
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="gap-2" onClick={() => setActiveTab('settings')}>
                <Settings className="w-4 h-4" />
                Settings
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Tabs */}
      <div>
        <div className="inline-flex gap-1 p-1 rounded-[var(--radius)] overflow-x-auto" style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)' }} role="tablist">
          {tabs.map(({ id, label, icon: Icon, count }) => {
            const isActive = activeTab === id;
            if (id === 'api-keys') {
              return (
                <button key={id} role="tab" aria-selected={false} onClick={() => navigate(`${ROUTES.API_KEYS}?type=function`)} className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-[var(--radius-sm)] transition-all whitespace-nowrap" style={{ background: 'transparent', color: 'var(--text-faint)', fontFamily: 'var(--font-body)' }}>
                  <Icon className="w-3.5 h-3.5" />
                  {label}
                </button>
              );
            }
            return (
              <button key={id} role="tab" aria-selected={isActive} onClick={() => setActiveTab(id)} className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-[var(--radius-sm)] transition-all whitespace-nowrap" style={{ background: isActive ? 'var(--panel-raised)' : 'transparent', color: isActive ? 'var(--text)' : 'var(--text-faint)', boxShadow: isActive ? '0 1px 3px rgba(0,0,0,0.2)' : 'none', fontFamily: 'var(--font-body)' }}>
                <Icon className="w-3.5 h-3.5" />
                {label}
                {typeof count === 'number' && count > 0 && (
                  <span className="text-xs rounded-full px-1.5 py-0.5 font-mono" style={{ background: 'var(--panel)', color: 'var(--text-faint)' }}>{count}</span>
                )}
              </button>
            );
          })}
        </div>

        <div className="mt-6">
          {activeTab === 'overview' && <OverviewTab data={data as AppStatus} />}
          {activeTab === 'backends' && <BackendsTab data={data as AppStatus} appId={(data as AppStatus).app.id} />}
          {activeTab === 'analytics' && <AnalyticsTab appId={(data as AppStatus).app.id} />}
          {activeTab === 'deployments' && <DeploymentsTab appId={(data as AppStatus).app.id} />}
          {activeTab === 'secrets' && <SecretsTab appId={(data as AppStatus).app.id} />}
          {activeTab === 'settings' && <SettingsTab data={data as AppStatus} appId={(data as AppStatus).app.id} />}
        </div>
      </div>
    </div>
  );
}
