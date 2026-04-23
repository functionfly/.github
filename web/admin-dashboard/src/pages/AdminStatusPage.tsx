import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import {
  Activity,
  Server,
  RefreshCw,
  CheckCircle2,
  XCircle,
  HelpCircle,
  Clock,
  Zap,
  Globe,
  TrendingUp,
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface HealthCheckResponse {
  status?: string;
  version?: string;
  timestamp?: string;
  checks?: Record<string, { healthy?: boolean; status?: string }>;
  edge?: EdgeStats;
}

/** Edge (edge.functionfly.com) stats from GET /v1/admin/health (includes edge) */
interface EdgeStats {
  probe_configured: boolean;
  health: string;
  uptime_ratio: number;
  total_requests: number;
  last_probe_at: string;
  last_probe_ok: boolean;
  last_probe_latency_ms: number;
  last_error?: string;
  probe_errors_total: number;
  nodes?: { host: string; region: string; status: string }[];
}

function StatusBadge({
  status,
  variant = 'light',
}: {
  status: string;
  variant?: 'light' | 'dark';
}) {
  const s = (status || 'unknown').toLowerCase();
  const healthy = s === 'healthy' || s === 'ok';
  const degraded = s === 'degraded' || s === 'warning';

  const darkClasses = healthy
    ? { label: status || 'Healthy', className: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40', Icon: CheckCircle2 }
    : degraded
      ? { label: status || 'Degraded', className: 'bg-amber-500/20 text-amber-300 border-amber-500/40', Icon: Activity }
      : s === 'unknown'
        ? { label: 'Unknown', className: 'bg-slate-500/20 text-slate-300 border-slate-500/40', Icon: HelpCircle }
        : { label: status || 'Unhealthy', className: 'bg-red-500/20 text-red-300 border-red-500/40', Icon: XCircle };

  const lightClasses = healthy
    ? { label: darkClasses.label, Icon: darkClasses.Icon, className: 'bg-emerald-100 text-emerald-800 border-emerald-200' }
    : degraded
      ? { label: darkClasses.label, Icon: darkClasses.Icon, className: 'bg-amber-100 text-amber-800 border-amber-200' }
      : s === 'unknown'
        ? { label: darkClasses.label, Icon: darkClasses.Icon, className: 'bg-slate-100 text-slate-600 border-slate-200' }
        : { label: darkClasses.label, Icon: darkClasses.Icon, className: 'bg-red-100 text-red-800 border-red-200' };

  const config = variant === 'dark' ? darkClasses : lightClasses;

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm font-medium font-mono ${config.className}`}
    >
      <config.Icon className="h-3.5 w-3.5 shrink-0" />
      {config.label}
    </span>
  );
}

export function AdminStatusPage() {
  const {
    data: healthResponse,
    isLoading,
    dataUpdatedAt,
    isFetching,
    refetch,
  } = useQuery({
    queryKey: ['admin-health-status'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<HealthCheckResponse>('/health');
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
    refetchInterval: 30000,
  });

  // Edge stats come from the same health response (GET /v1/admin/health) so no 404
  const rawHealth =
    healthResponse && 'data' in healthResponse && healthResponse.data != null
      ? (healthResponse as { data: HealthCheckResponse }).data
      : (healthResponse as HealthCheckResponse | null) ?? {};
  const health: HealthCheckResponse = rawHealth;
  const edgeStats = health.edge ?? null;

  if (isLoading) {
    return <LoadingScreen />;
  }

  const checks = health.checks || {};
  const entries = Object.entries(checks);
  const healthyCount = entries.filter(([, c]) => c?.healthy === true).length;
  const totalCount = entries.length;
  const overallStatus = health.status || 'unknown';
  const refetchAll = () => refetch();

  return (
    <div className="status-page space-y-8">
      {/* Page header */}
      <div className="flex flex-col gap-1">
        <div className="flex items-center gap-3">
          <h1 className="font-display text-3xl font-bold tracking-tight text-gray-900 dark:text-gray-100">
            Platform Status
          </h1>
          <span className="inline-flex items-center gap-1 rounded-full bg-slate-100 dark:bg-slate-800 px-2.5 py-0.5 text-xs font-medium text-slate-600 dark:text-slate-400">
            <Zap className="h-3 w-3" />
            Live
          </span>
        </div>
        <p className="text-gray-600 dark:text-gray-400">
          Live system health and subsystem checks.
        </p>
      </div>

      {/* Hero status card */}
      <div className="status-hero-card overflow-hidden rounded-xl border border-slate-200/80 p-6 text-white md:p-8">
        <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div className="flex flex-wrap items-center gap-4">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-white/10">
                <Server className="h-6 w-6 text-status-accent" />
              </div>
              <div>
                <p className="text-sm font-medium text-slate-400">Overall status</p>
                <div className="mt-1">
                  <StatusBadge status={overallStatus} variant="dark" />
                </div>
              </div>
            </div>
            <div className="border-l border-white/10 pl-4">
              <p className="text-xs font-medium uppercase tracking-wider text-slate-500">Version</p>
              <p className="font-mono text-sm font-medium text-slate-200">
                {health.version || '—'}
              </p>
            </div>
            <div className="border-l border-white/10 pl-4">
              <p className="text-xs font-medium uppercase tracking-wider text-slate-500">Last updated</p>
              <p className="flex items-center gap-1.5 font-mono text-sm text-slate-300">
                <Clock className="h-3.5 w-3.5" />
                {formatDistanceToNow(new Date(dataUpdatedAt), { addSuffix: true })}
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={() => refetchAll()}
            disabled={isFetching}
            className="inline-flex items-center gap-2 rounded-lg bg-white/10 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-white/20 disabled:opacity-60"
          >
            <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
            {isFetching ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </div>

      {/* Edge (edge.functionfly.com) monitoring */}
      <div className="overflow-hidden rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm">
        <div className="border-b border-gray-200 dark:border-gray-700 bg-gray-50/80 dark:bg-gray-800 px-6 py-4">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="font-display text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                <Globe className="h-5 w-5 text-sky-600 dark:text-sky-400" />
                Edge — edge.functionfly.com
              </h2>
              <p className="mt-0.5 text-sm text-gray-500 dark:text-gray-400">
                TLS edge and function execution endpoint. Probed by the orchestrator when EDGE_HEALTH_URL is set.
              </p>
            </div>
            <a
              href="https://edge.functionfly.com"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm font-medium text-sky-600 hover:text-sky-700 dark:text-sky-400 dark:hover:text-sky-300"
            >
              Open →
            </a>
          </div>
        </div>
        <div className="p-6">
          {isFetching && edgeStats === null ? (
            <div className="flex items-center gap-3 text-gray-500 dark:text-gray-400">
              <RefreshCw className="h-4 w-4 animate-spin" />
              <span className="text-sm">Loading edge status…</span>
            </div>
          ) : edgeStats === null ? (
            <div className="flex flex-col gap-2 text-gray-500 dark:text-gray-400">
              <p className="text-sm font-medium text-gray-900 dark:text-gray-100">Could not load edge status</p>
              <p className="text-xs">
                Check that you are logged in and the orchestrator is running. If the problem persists, check the browser console.
              </p>
            </div>
          ) : !edgeStats.probe_configured ? (
            <div className="flex flex-col gap-3 rounded-lg border border-amber-200 dark:border-amber-800 bg-amber-50/80 dark:bg-amber-900/20 p-4">
              <p className="text-sm font-medium text-amber-900 dark:text-amber-200">Edge monitoring not configured</p>
              <p className="text-xs text-amber-800 dark:text-amber-300">
                Set <code className="rounded bg-amber-100 dark:bg-amber-900/50 px-1 font-mono text-amber-900 dark:text-amber-200">EDGE_HEALTH_URL</code> in the orchestrator environment to enable probing of edge.functionfly.com.
              </p>
              <p className="text-xs text-amber-700 dark:text-amber-400">
                Example: add <code className="rounded bg-amber-100 dark:bg-amber-900/50 px-1 font-mono">EDGE_HEALTH_URL=https://edge.functionfly.com</code> to your <code className="rounded bg-amber-100 dark:bg-amber-900/50 px-1 font-mono">.env</code> and restart the orchestrator. The health monitor will then probe the edge every cycle.
              </p>
            </div>
          ) : (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Status</p>
                <div className="mt-1">
                  <StatusBadge status={edgeStats.health} variant="light" />
                </div>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Uptime (recent probes)</p>
                <p className="mt-1 font-mono text-sm font-medium text-gray-900 dark:text-gray-100">
                  {Math.round(edgeStats.uptime_ratio * 100)}%
                </p>
                <div className="status-summary-bar mt-1 w-full max-w-[120px]">
                  <div
                    className="status-summary-bar-fill bg-emerald-500"
                    style={{ width: `${edgeStats.uptime_ratio * 100}%` }}
                  />
                </div>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Last probe</p>
                <p className="mt-1 flex items-center gap-1.5 font-mono text-sm text-gray-700 dark:text-gray-300">
                  <Clock className="h-3.5 w-3.5" />
                  {edgeStats.last_probe_at
                    ? formatDistanceToNow(new Date(edgeStats.last_probe_at), { addSuffix: true })
                    : '—'}
                </p>
                {edgeStats.last_probe_latency_ms >= 0 && (
                  <p className="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                    Latency: <span className="font-mono">{edgeStats.last_probe_latency_ms}</span> ms
                  </p>
                )}
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 flex items-center gap-1">
                  <TrendingUp className="h-3.5 w-3.5" />
                  Total requests
                </p>
                <p className="mt-1 font-mono text-sm font-medium text-gray-900 dark:text-gray-100">
                  {edgeStats.total_requests.toLocaleString()}
                </p>
                {edgeStats.probe_errors_total > 0 && (
                  <p className="mt-0.5 text-xs text-amber-600 dark:text-amber-400">
                    {edgeStats.probe_errors_total} probe error(s)
                  </p>
                )}
              </div>
            </div>
          )}
          {edgeStats?.nodes && edgeStats.nodes.length > 0 && (
            <div className="mt-6">
              <h3 className="mb-3 text-sm font-semibold text-gray-700">Edge VPS nodes by region</h3>
              <div className="overflow-hidden rounded-lg border border-gray-200">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-gray-200 bg-gray-50">
                      <th className="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Region
                      </th>
                      <th className="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Node
                      </th>
                      <th className="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Status
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {edgeStats.nodes.map((node, i) => (
                      <tr key={i} className="bg-white hover:bg-gray-50/80">
                        <td className="px-4 py-3 font-medium text-gray-900">{node.region || '—'}</td>
                        <td className="px-4 py-3 font-mono text-gray-700">{node.host}</td>
                        <td className="px-4 py-3">
                          <StatusBadge
                            status={node.status || edgeStats.health}
                            variant="light"
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
          {edgeStats?.last_error && (
            <div className="mt-4 rounded-lg bg-red-50 border border-red-100 px-4 py-2">
              <p className="text-xs font-medium text-red-800">Last probe error</p>
              <p className="font-mono text-sm text-red-700">{edgeStats.last_error}</p>
            </div>
          )}
          <div className="mt-4 flex justify-end">
            <button
              type="button"
              onClick={() => refetch()}
              disabled={isFetching}
              className="inline-flex items-center gap-2 text-sm font-medium text-gray-600 hover:text-gray-900 disabled:opacity-60"
            >
              <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
              Refresh edge
            </button>
          </div>
        </div>
      </div>

      {/* Summary bar (when we have checks) */}
      {totalCount > 0 && (
        <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-gray-500" />
              <span className="text-sm font-medium text-gray-700">Components</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-sm text-gray-600">
                <span className="font-semibold text-emerald-600">{healthyCount}</span>
                <span className="text-gray-400"> / </span>
                <span className="font-mono">{totalCount}</span>
                <span className="text-gray-400"> healthy</span>
              </span>
              <div className="status-summary-bar w-24 overflow-hidden md:w-32">
                <div
                  className="status-summary-bar-fill bg-emerald-500"
                  style={{ width: `${totalCount ? (healthyCount / totalCount) * 100 : 0}%` }}
                />
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Health checks table */}
      <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
        <div className="border-b border-gray-200 bg-gray-50/80 px-6 py-4">
          <h2 className="font-display text-lg font-semibold text-gray-900">Health checks</h2>
          <p className="mt-0.5 text-sm text-gray-500">Subsystem status and readiness.</p>
        </div>
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-200 bg-gray-50/50">
              <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                Component
              </th>
              <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                Status
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {entries.length === 0 ? (
              <tr>
                <td colSpan={2} className="px-6 py-16 text-center">
                  <div className="mx-auto flex max-w-sm flex-col items-center gap-3 text-gray-500">
                    <div className="flex h-12 w-12 items-center justify-center rounded-full bg-gray-100">
                      <Server className="h-6 w-6 text-gray-400" />
                    </div>
                    <p className="text-sm font-medium">No health checks available</p>
                    <p className="text-xs text-gray-400">
                      The platform may not be exposing component-level checks yet.
                    </p>
                  </div>
                </td>
              </tr>
            ) : (
              entries.map(([name, check], i) => {
                const healthy = check?.healthy === true;
                const statusLabel = check?.status || (healthy ? 'healthy' : 'unhealthy');
                const rowKind = healthy ? 'healthy' : 'unhealthy';
                return (
                  <tr
                    key={name}
                    className={`status-row-border ${rowKind} animate-stagger-fade bg-white opacity-0 transition hover:bg-gray-50/80`}
                    style={{ animationDelay: `${i * 40}ms` }}
                  >
                    <td className="px-6 py-4">
                      <span className="font-medium text-gray-900">{name}</span>
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={statusLabel} variant="light" />
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
