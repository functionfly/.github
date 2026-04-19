import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ROUTES } from '@/lib/constants';
import { cn } from '@/lib/utils';
import type { AppStatus, BackendStatus } from '@/types';
import { useAppStatus } from '@/hooks';
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
  ExternalLink,
  MoreVertical,
  RefreshCw,
  Server,
  Settings,
  Shield,
  XCircle,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

/** Match app id segment when it is a UUID (canonical URL uses slug). */
const APP_PATH_UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// ─── Sub-components ──────────────────────────────────────────────────────────

function StatusDot({ ok, state }: { ok?: boolean; state?: string }) {
  if (state === 'open')
    return <span className="w-2.5 h-2.5 rounded-full bg-red-500 animate-pulse" />;
  if (state === 'half-open')
    return <span className="w-2.5 h-2.5 rounded-full bg-amber-500 animate-pulse" />;
  if (ok === true) return <span className="w-2.5 h-2.5 rounded-full bg-emerald-500" />;
  if (ok === false) return <span className="w-2.5 h-2.5 rounded-full bg-red-500" />;
  return <span className="w-2.5 h-2.5 rounded-full bg-muted-foreground/40" />;
}

function BackendStatusCard({ backendStatus }: { backendStatus: BackendStatus }) {
  const { backend, circuitState, latestHealthCheck } = backendStatus;

  const isHealthy = latestHealthCheck?.ok === true;
  const isUnhealthy = latestHealthCheck?.ok === false;
  const circuitOpen = circuitState?.state === 'open';
  const circuitHalfOpen = circuitState?.state === 'half-open';

  const statusLabel = circuitOpen
    ? 'Circuit Open'
    : circuitHalfOpen
      ? 'Half-Open'
      : isHealthy
        ? 'Healthy'
        : isUnhealthy
          ? 'Unhealthy'
          : 'Unknown';

  const statusVariant =
    circuitOpen || isUnhealthy
      ? 'destructive'
      : circuitHalfOpen
        ? 'warning'
        : isHealthy
          ? 'success'
          : 'outline';

  return (
    <div className="flex items-center gap-4 p-4 rounded-xl border border-border/50 bg-card hover:border-border transition-colors">
      <StatusDot ok={latestHealthCheck?.ok} state={circuitState?.state} />

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-medium text-sm capitalize">{backend.provider}</span>
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
        <p className="text-xs text-muted-foreground truncate mt-1 font-mono">{backend.url}</p>
      </div>

      <div className="flex items-center gap-3 text-xs text-muted-foreground shrink-0">
        {latestHealthCheck && (
          <div className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            <span>{latestHealthCheck.latencyMs}ms</span>
          </div>
        )}
        {latestHealthCheck?.ok ? (
          <CheckCircle2 className="w-4 h-4 text-emerald-500" />
        ) : latestHealthCheck?.ok === false ? (
          <XCircle className="w-4 h-4 text-red-500" />
        ) : null}
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
  icon: React.ElementType;
  trend?: 'up' | 'down' | 'neutral';
  color?: 'default' | 'success' | 'warning' | 'danger';
}) {
  const colorMap = {
    default: 'text-brand-500 bg-brand-500/10',
    success: 'text-emerald-500 bg-emerald-500/10',
    warning: 'text-amber-500 bg-amber-500/10',
    danger: 'text-red-500 bg-red-500/10',
  };

  return (
    <Card className="border-border/50">
      <CardContent className="p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <p className="text-xs text-muted-foreground font-medium uppercase tracking-wide mb-2">
              {label}
            </p>
            <p className="text-3xl font-bold text-foreground tabular-nums">{value}</p>
            {sub && <p className="text-xs text-muted-foreground mt-1">{sub}</p>}
          </div>
          <div
            className={cn('w-10 h-10 rounded-xl flex items-center justify-center', colorMap[color])}
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
      {/* Header skeleton */}
      <div className="flex items-center gap-4 animate-pulse">
        <div className="w-9 h-9 rounded-lg bg-muted" />
        <div className="w-14 h-14 rounded-2xl bg-muted" />
        <div className="space-y-2 flex-1">
          <div className="h-6 bg-muted rounded w-48" />
          <div className="h-4 bg-muted rounded w-32" />
        </div>
      </div>
      {/* Stats skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-28 rounded-xl bg-muted animate-pulse" />
        ))}
      </div>
      {/* Content skeleton */}
      <div className="h-64 rounded-xl bg-muted animate-pulse" />
    </div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="w-16 h-16 rounded-2xl bg-destructive/10 border border-destructive/20 flex items-center justify-center mb-5">
        <AlertCircle className="w-7 h-7 text-destructive" />
      </div>
      <h3 className="text-lg font-semibold mb-2">Failed to load app</h3>
      <p className="text-sm text-muted-foreground mb-6 max-w-sm">{message}</p>
      <Button variant="outline" onClick={onRetry} className="gap-2">
        <RefreshCw className="w-4 h-4" />
        Try Again
      </Button>
    </div>
  );
}

// ─── Overview Tab ─────────────────────────────────────────────────────────────

function OverviewTab({ data }: { data: AppStatus }) {
  const { app, backends } = data;
  const healthyCount = backends.filter((b) => b.latestHealthCheck?.ok).length;
  const openCircuits = backends.filter((b) => b.circuitState?.state === 'open').length;

  return (
    <div className="space-y-6">
      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          label="Total Backends"
          value={backends.length}
          sub="connected providers"
          icon={Server}
          color="default"
        />
        <StatCard
          label="Healthy"
          value={healthyCount}
          sub={`of ${backends.length} operational`}
          icon={CheckCircle2}
          color="success"
        />
        <StatCard
          label="Open Circuits"
          value={openCircuits}
          sub={openCircuits > 0 ? 'requires attention' : 'all circuits closed'}
          icon={Shield}
          color={openCircuits > 0 ? 'danger' : 'success'}
        />
      </div>

      {/* App Info */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base">App Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground text-xs uppercase tracking-wide mb-1">App ID</p>
              <div className="flex items-center gap-2">
                <code className="font-mono text-xs bg-muted px-2 py-1 rounded truncate max-w-[160px]">
                  {app.id}
                </code>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(app.id);
                    toast.success('App ID copied');
                  }}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                  aria-label="Copy app ID"
                >
                  <Copy className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
            <div>
              <p className="text-muted-foreground text-xs uppercase tracking-wide mb-1">Slug</p>
              <code className="font-mono text-xs bg-muted px-2 py-1 rounded">{app.slug}</code>
            </div>
            <div>
              <p className="text-muted-foreground text-xs uppercase tracking-wide mb-1">Created</p>
              <p className="text-sm">
                {new Date(app.createdAt).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </p>
            </div>
            <div>
              <p className="text-muted-foreground text-xs uppercase tracking-wide mb-1">Tenant</p>
              <code className="font-mono text-xs bg-muted px-2 py-1 rounded truncate max-w-[160px] block">
                {app.tenantId}
              </code>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// ─── Backends Tab ─────────────────────────────────────────────────────────────

function BackendsTab({ data }: { data: AppStatus }) {
  const { backends } = data;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          {backends.length} backend{backends.length !== 1 ? 's' : ''} configured
        </p>
        <Button variant="outline" size="sm" className="gap-2">
          <Server className="w-3.5 h-3.5" />
          Add Backend
        </Button>
      </div>

      {backends.length > 0 ? (
        <div className="space-y-3">
          {backends.map((backendStatus) => (
            <BackendStatusCard key={backendStatus.backend.id} backendStatus={backendStatus} />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center rounded-xl border border-dashed border-border/60">
          <div className="w-14 h-14 rounded-2xl bg-muted/50 flex items-center justify-center mb-4">
            <Server className="w-6 h-6 text-muted-foreground/60" />
          </div>
          <h4 className="font-medium mb-1">No backends configured</h4>
          <p className="text-sm text-muted-foreground mb-4 max-w-xs">
            Add a backend provider to start deploying your app to the edge.
          </p>
          <Button variant="outline" size="sm" className="gap-2">
            <Server className="w-3.5 h-3.5" />
            Add Backend
          </Button>
        </div>
      )}
    </div>
  );
}

// ─── Analytics Tab ────────────────────────────────────────────────────────────

function AnalyticsTab() {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="w-16 h-16 rounded-2xl bg-brand-500/10 border border-brand-500/20 flex items-center justify-center mb-5">
        <BarChart3 className="w-7 h-7 text-brand-500" />
      </div>
      <h4 className="font-semibold text-lg mb-2">Analytics Coming Soon</h4>
      <p className="text-sm text-muted-foreground max-w-sm">
        Request counts, error rates, latency charts, and more will be available here.
      </p>
    </div>
  );
}

// ─── Settings Tab ─────────────────────────────────────────────────────────────

function SettingsTab({ data }: { data: AppStatus }) {
  const { app } = data;

  return (
    <div className="space-y-6">
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base">General Settings</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium">App Name</label>
            <input
              type="text"
              defaultValue={app.name}
              className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background focus:outline-none focus:ring-2 focus:ring-brand-500/30 focus:border-brand-500/50 transition-colors"
              disabled
            />
            <p className="text-xs text-muted-foreground">Contact support to rename your app.</p>
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium">App Slug</label>
            <input
              type="text"
              defaultValue={app.slug}
              className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background font-mono focus:outline-none focus:ring-2 focus:ring-brand-500/30 focus:border-brand-500/50 transition-colors"
              disabled
            />
            <p className="text-xs text-muted-foreground">
              Used in URLs: <code className="font-mono">/apps/{app.slug}</code>
            </p>
          </div>
        </CardContent>
      </Card>

      <Card className="border-destructive/30 bg-destructive/5">
        <CardHeader className="pb-3">
          <CardTitle className="text-base text-destructive">Danger Zone</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-medium">Delete this app</p>
              <p className="text-xs text-muted-foreground mt-0.5">
                Permanently delete this app and all its backends. This action cannot be undone.
              </p>
            </div>
            <Button variant="destructive" size="sm" disabled>
              Delete App
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function AppDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('overview');

  // Use hook instead of raw query
  const { data, isLoading, error, refetch } = useAppStatus(slug || '');

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

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav
        className="flex items-center gap-2 text-sm text-muted-foreground"
        aria-label="Breadcrumb"
      >
        <Link to={ROUTES.APPS} className="hover:text-foreground transition-colors">
          Apps
        </Link>
        <ChevronRight className="w-3.5 h-3.5" />
        <span className="text-foreground font-medium truncate">{app.name}</span>
      </nav>

      {/* Header */}
      <div className="flex items-start gap-4">
        <Link to={ROUTES.APPS} aria-label="Back to apps">
          <Button variant="ghost" size="icon" className="shrink-0 mt-1">
            <ArrowLeft className="w-4 h-4" />
          </Button>
        </Link>

        <div className="flex items-start gap-4 flex-1 min-w-0">
          {/* App Icon */}
          <div className="w-14 h-14 rounded-2xl bg-brand-500/15 border border-brand-500/20 flex items-center justify-center shrink-0">
            <Building2 className="w-7 h-7 text-brand-500" />
          </div>

          {/* App Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl font-bold text-foreground">{app.name}</h1>
              <Badge
                variant={overallHealthy ? 'default' : 'destructive'}
                className={cn(
                  'text-xs',
                  overallHealthy
                    ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/30'
                    : ''
                )}
              >
                <span
                  className={cn(
                    'w-1.5 h-1.5 rounded-full mr-1.5',
                    overallHealthy ? 'bg-emerald-500' : 'bg-red-500'
                  )}
                />
                {overallHealthy ? 'Healthy' : 'Degraded'}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground font-mono mt-1">{app.slug}</p>
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
              <DropdownMenuItem
                onClick={() => {
                  navigator.clipboard.writeText(app.id);
                  toast.success('App ID copied');
                }}
                className="gap-2"
              >
                <Copy className="w-4 h-4" />
                Copy App ID
              </DropdownMenuItem>
              <DropdownMenuItem className="gap-2">
                <ExternalLink className="w-4 h-4" />
                View Deployments
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
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="w-full sm:w-auto">
          <TabsTrigger value="overview" className="gap-2">
            <Activity className="w-3.5 h-3.5" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="backends" className="gap-2">
            <Server className="w-3.5 h-3.5" />
            Backends
            {backends.length > 0 && (
              <span className="ml-1 text-xs bg-muted rounded-full px-1.5 py-0.5 font-mono">
                {backends.length}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="analytics" className="gap-2">
            <BarChart3 className="w-3.5 h-3.5" />
            Analytics
          </TabsTrigger>
          <TabsTrigger value="settings" className="gap-2">
            <Settings className="w-3.5 h-3.5" />
            Settings
          </TabsTrigger>
        </TabsList>

        <div className="mt-6">
          <TabsContent value="overview" className="mt-0">
            <OverviewTab data={data as AppStatus} />
          </TabsContent>
          <TabsContent value="backends" className="mt-0">
            <BackendsTab data={data as AppStatus} />
          </TabsContent>
          <TabsContent value="analytics" className="mt-0">
            <AnalyticsTab />
          </TabsContent>
          <TabsContent value="settings" className="mt-0">
            <SettingsTab data={data as AppStatus} />
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}
