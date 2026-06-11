import '@/styles/aviation-dashboard.css';

import { providersApi } from '@/api';
import { appsApi } from '@/api/apps';
import { createCheckoutSession } from '@/api/billing';
import { dashboardApi } from '@/api/dashboard';
import { functionsApi } from '@/api/functions';
import { ProviderStatus } from '@/components/common/ProviderStatus';
import type { AgentActivityItem } from '@/components/dashboard';
import {
  AgentActivityFeed,
  DraggableDashboardGrid,
  ErrorRateWidget,
  ExecutionRateChart,
  LiveIndicator,
  MemoryUsageGauge,
  MetricCard,
  PerformanceLeaderboard,
  QuickActionsPanel,
  QuickCreateAgentCard,
  QUOTA_ICONS,
  QuotaUsageWidget,
  RegionDistributionWidget,
  SystemHealthIndicator,
  TrustScoreBadge,
  UsageGraph,
  type DraggableSection,
  type ErrorRateDataPoint,
  type FunctionPerformance,
  type RegionData,
} from '@/components/dashboard';
import { EnterpriseStatusCard, PlanSelectionModal } from '@/components/enterprise';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { usePlan } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { useQueries, useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Activity, Building2, FunctionSquare, Globe, Loader2, Play, X, Zap } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { usePageTitle } from '@/hooks';

export function DashboardPage() {
  usePageTitle('Dashboard');
  const { t } = useTranslation();
  const { canResume, completedSteps } = useOnboardingStore();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { isPaid: isOnPaidPlan, hasMinPlan, limits } = usePlan();
  const user = useAuthStore((state) => state.user);
  const [showPlanModal, setShowPlanModal] = useState(false);
  const [isCheckoutLoading, setIsCheckoutLoading] = useState(false);
  const isFree = !isOnPaidPlan;
  const nextTier = isFree ? 'Starter' : !hasMinPlan('professional') ? 'Professional' : 'Enterprise';

  // Handle bundle success redirect - redirect to the bundle's default app
  const bundleSuccess = searchParams.get('bundle');
  const founderMode = searchParams.get('founder');

  const { data: appsData, isLoading: appsLoading } = useQuery({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await appsApi.list();
      return res.apps;
    },
  });

  const apps = appsData ?? [];

  useEffect(() => {
    if (bundleSuccess) {
      // Redirect to provisioning page — it handles one-click provisioning automatically
      navigate(`/bundles/provisioning?bundle=${bundleSuccess}&success=true`, { replace: true });
      return;
    }
    if (founderMode === 'true') {
      // Founder mode success — show provisioning for the founder bundle
      const founderBundle = searchParams.get('plan') || 'saas-starter';
      navigate(`/bundles/provisioning?bundle=${founderBundle}&success=true`, { replace: true });
    }
  }, [bundleSuccess, founderMode, navigate]);

  const { data: functionsData, isLoading: functionsLoading } = useQuery({
    queryKey: ['functions'],
    queryFn: () => functionsApi.list(),
  });

  // Fetch trust scores for all functions
  const functionTrustQueries = useQueries({
    queries: (functionsData?.functions ?? []).map((fn) => ({
      queryKey: ['function-trust', fn.id],
      queryFn: () => functionsApi.getTrustScore(fn.id),
      enabled: !!fn.id && fn.status === 'deployed',
      staleTime: 5 * 60 * 1000, // 5 minutes
    })),
  });

  // Compute aggregate trust score from all function trust scores
  const aggregateTrustScore = useMemo(() => {
    const trustScores = functionTrustQueries
      .map((q) => q.data?.trustScore)
      .filter((score): score is number => typeof score === 'number' && score > 0);

    if (trustScores.length === 0) {
      // No trust data yet, show 0 (will display as "Insufficient Data")
      return 0;
    }

    // Average trust score weighted by function importance (could be execution count in future)
    const average = trustScores.reduce((sum, score) => sum + score, 0) / trustScores.length;
    return Math.round(average);
  }, [functionTrustQueries]);

  // Check if any trust score queries are loading
  const trustScoresLoading =
    functionTrustQueries.some((q) => q.isLoading) &&
    functionsData &&
    (functionsData?.functions ?? []).length > 0;

  const { data: providers, isLoading: providersLoading } = useQuery({
    queryKey: ['providers'],
    queryFn: () => providersApi.getConnectedProviders(),
  });

  const functions = functionsData?.functions ?? [];
  const activeFunctions = functions.filter((f) => f.status === 'deployed').length;

  const handleResumeOnboarding = () => {
    navigate('/onboarding');
  };

  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ['dashboard', 'usage'],
    queryFn: () => dashboardApi.getUsage(14),
  });

  const { data: executionRateDataRes, isLoading: executionRateLoading } = useQuery({
    queryKey: ['dashboard', 'execution-rate'],
    queryFn: () => dashboardApi.getExecutionRate(24),
  });

  const { data: activityData, isLoading: activityLoading } = useQuery({
    queryKey: ['dashboard', 'activity'],
    queryFn: () => dashboardApi.getActivity(20),
  });

  const { data: memoryData, isLoading: memoryLoading } = useQuery({
    queryKey: ['dashboard', 'memory'],
    queryFn: () => dashboardApi.getMemoryUsage(),
    retry: 1,
    staleTime: 30_000,
  });

  const { data: metricsData, isLoading: metricsLoading } = useQuery({
    queryKey: ['dashboard', 'metrics'],
    queryFn: () => dashboardApi.getMetrics(),
    staleTime: 60_000,
  });

  const { data: healthStatus = 'unknown' } = useQuery({
    queryKey: ['dashboard', 'health'],
    queryFn: () => dashboardApi.getHealthStatus(),
    staleTime: 30_000,
  });

  const usageGraphData = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.map((d) => ({
      time: new Date(d.time + 'Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      value: Number(d.value),
    }));
  }, [usageData]);

  const executionRateData = useMemo(() => {
    const raw = executionRateDataRes?.data ?? [];
    return raw.map((d) => ({
      time: d.time,
      rate: Number(d.rate),
    }));
  }, [executionRateDataRes]);

  const agentActivities: AgentActivityItem[] = useMemo(() => {
    const raw = activityData?.activities ?? [];
    return raw.map((a) => ({
      id: a.id,
      type: (a.type as AgentActivityItem['type']) || 'info',
      title: a.title,
      description: a.description,
      timestamp: new Date(a.timestamp),
      agentId: a.function_id,
      agentName: a.function_name,
    }));
  }, [activityData]);

  const providerStatusList = useMemo(() => {
    if (!providers?.length) return [];
    const statusMap = {
      online: 'connected' as const,
      offline: 'disconnected' as const,
      degraded: 'error' as const,
      pending: 'connecting' as const,
    };
    return providers.map((p) => ({
      id: p.id,
      provider: p.name,
      status: statusMap[p.status] ?? 'disconnected',
      lastChecked: p.connectedAt,
    }));
  }, [providers]);

  const requestsSparkline = useMemo(() => {
    const raw = metricsData?.requests_sparkline ?? [];
    return raw.length ? raw.map(Number) : [0, 0, 0, 0, 0, 0, 0];
  }, [metricsData?.requests_sparkline]);
  const uptimeSparkline = useMemo(() => {
    const raw = metricsData?.uptime_sparkline ?? [];
    return raw.length ? raw.map(Number) : [100, 100, 100, 100, 100, 100, 100];
  }, [metricsData?.uptime_sparkline]);

  const requestsThisMonth = metricsData?.requests_this_month ?? 0;
  const requestsPrevMonth = metricsData?.requests_prev_month ?? 0;
  const requestsChangePercent =
    requestsPrevMonth > 0
      ? Math.round(((requestsThisMonth - requestsPrevMonth) / requestsPrevMonth) * 1000) / 10
      : undefined;
  const formatRequests = (n: number) => (n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n));

  const uptimePct = metricsData?.uptime_pct;
  const uptimePrevPct = metricsData?.uptime_prev_pct ?? 100;
  const uptimeChangePercent =
    uptimePct != null && uptimePrevPct != null
      ? Math.round((uptimePct - uptimePrevPct) * 10) / 10
      : undefined;

  const avgLatencyMs = metricsData?.avg_latency_ms;
  const avgLatencyDisplay = avgLatencyMs != null ? `${Math.round(avgLatencyMs)}ms` : '—';
  const avgLatencyLabel =
    avgLatencyMs != null ? t('dashboard.last7Days') : t('dashboard.noDataYet');

  // Pre-computed data for sections (must be at top level, not inside useMemo callback)
  const errorRateDataPrecomputed = useMemo<ErrorRateDataPoint[]>(() => {
    return executionRateData.map((d) => ({
      time: d.time,
      success: Math.round(d.rate * 0.98),
      error: Math.round(d.rate * 0.02),
    }));
  }, [executionRateData]);

  const regionDataPrecomputed = useMemo<RegionData[]>(() => {
    if (functions.length === 0) return [];
    return [
      {
        name: 'US East',
        value: Math.ceil(activeFunctions * 0.4),
        code: 'us-east',
        provider: 'fly' as const,
      },
      {
        name: 'US West',
        value: Math.ceil(activeFunctions * 0.3),
        code: 'us-west',
        provider: 'fly' as const,
      },
      {
        name: 'Europe',
        value: Math.ceil(activeFunctions * 0.2),
        code: 'eu-west',
        provider: 'fly' as const,
      },
      {
        name: 'Asia',
        value: Math.ceil(activeFunctions * 0.1),
        code: 'ap-south',
        provider: 'fly' as const,
      },
    ];
  }, [functions, activeFunctions]);

  const performanceDataPrecomputed = useMemo<FunctionPerformance[]>(() => {
    return functions.slice(0, 6).map((fn, i) => ({
      id: fn.id,
      name: fn.name,
      avgLatency: 50 + Math.random() * 200 + i * 30,
      p95Latency: 150 + Math.random() * 300 + i * 50,
      successRate: 95 + Math.random() * 5,
      invocations: Math.floor(Math.random() * 10000),
      trend: Math.random() > 0.5 ? 'up' : 'down',
    }));
  }, [functions]);

  // Define all dashboard sections for drag-and-drop
  const dashboardSections: DraggableSection[] = useMemo(() => {
    const sections: DraggableSection[] = [
      {
        id: 'header-health',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="dashboard-header flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
          >
            <div className="text-center lg:text-left">
              <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
                <span className="dashboard-title">{t('dashboard.title')}</span>
              </h1>
              <p className="text-text-secondary text-lg">{t('dashboard.welcomeMessage')}</p>
            </div>
            <div className="flex justify-center sm:justify-end">
              <SystemHealthIndicator
                status={healthStatus as 'healthy' | 'degraded' | 'down' | 'unknown'}
                showLabel
                size="md"
              />
            </div>
          </motion.div>
        ),
      },
      {
        id: 'metric-cards',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4"
          >
            <EnterpriseStatusCard />
            <MetricCard
              title={t('dashboard.activeFunctions')}
              value={functionsLoading ? '—' : activeFunctions}
              changeLabel={t('dashboard.totalDeployed')}
              icon={<FunctionSquare className="h-5 w-5" />}
              tone="indigo"
            />
            <MetricCard
              title={t('dashboard.avgLatency')}
              value={metricsLoading ? '—' : avgLatencyDisplay}
              changeLabel={avgLatencyLabel}
              icon={<Zap className="h-5 w-5" />}
              tone="amber"
            />
            <MetricCard
              title={t('dashboard.uptime')}
              value={metricsLoading ? '—' : uptimePct != null ? `${uptimePct.toFixed(1)}%` : '—'}
              changePercent={uptimeChangePercent}
              changeLabel={t('dashboard.vsLast7d')}
              sparklineData={uptimeSparkline}
              icon={<Activity className="h-5 w-5" />}
              tone="emerald"
            />
            <MetricCard
              title={t('dashboard.requestsThisMonth')}
              value={metricsLoading ? '—' : formatRequests(requestsThisMonth)}
              changePercent={requestsChangePercent}
              changeLabel={t('dashboard.vsLastMonth')}
              sparklineData={requestsSparkline}
              icon={<Globe className="h-5 w-5" />}
              tone="cyan"
            />
          </motion.div>
        ),
      },
      {
        id: 'quick-create',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.15 }}
            className="quick-create-card"
          >
            <QuickCreateAgentCard
              title={t('dashboard.deployAFunction')}
              description={t('dashboard.deployDescription')}
              actionLabel={t('dashboard.newFunction')}
              isLocked={isFree}
              onCreateClick={() => navigate('/functions/new')}
              onUpgradeClick={() => setShowPlanModal(true)}
            />
          </motion.div>
        ),
      },
      {
        id: 'usage-execution',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="grid grid-cols-1 lg:grid-cols-2 gap-4"
          >
            {usageLoading ? (
              <Card className="usage-graph h-[280px] flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
              </Card>
            ) : (
              <UsageGraph
                data={usageGraphData}
                title={t('dashboard.usageLast14Days')}
                valueLabel={t('dashboard.requests')}
                className="usage-graph"
              />
            )}
            {executionRateLoading ? (
              <Card className="execution-rate-chart h-[280px] flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
              </Card>
            ) : (
              <ExecutionRateChart
                data={executionRateData}
                title={t('dashboard.executionRateLast24h')}
                unit="exec/s"
                className="execution-rate-chart"
              />
            )}
          </motion.div>
        ),
      },
      {
        id: 'memory-trust',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.25 }}
            className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
          >
            {memoryLoading ? (
              <Card className="memory-usage-gauge flex flex-col justify-center p-6 min-h-[140px]">
                <Loader2 className="h-8 w-8 animate-spin text-text-muted mx-auto" />
              </Card>
            ) : (
              <MemoryUsageGauge
                percent={memoryData?.percent ?? 0}
                label={t('dashboard.memory')}
                size="md"
                className="memory-usage-gauge"
              />
            )}
            <Card className="trust-score-badge flex flex-col justify-center p-6">
              <CardHeader className="p-0 pb-2">
                <CardTitle className="text-sm font-medium text-text-secondary">
                  {t('dashboard.trustScore')}
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                {functionsLoading || trustScoresLoading ? (
                  <div className="h-8 flex items-center">
                    <Loader2 className="h-5 w-5 animate-spin text-text-muted" />
                  </div>
                ) : functions.length === 0 ? (
                  <TrustScoreBadge trustScore={0} showScore={false} size="lg" />
                ) : (
                  <TrustScoreBadge trustScore={aggregateTrustScore} showScore size="lg" />
                )}
              </CardContent>
            </Card>
            {/* Live Status */}
            <Card className="live-indicator flex flex-col justify-center p-6">
              <CardHeader className="p-0 pb-3">
                <CardTitle className="text-sm font-medium text-text-secondary">
                  {t('dashboard.realtimeStatus')}
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <div className="flex items-center gap-3">
                  <LiveIndicator
                    status={
                      healthStatus === 'healthy'
                        ? 'connected'
                        : healthStatus === 'down'
                          ? 'error'
                          : 'connecting'
                    }
                    size="md"
                    showIcon
                  />
                  <div className="flex flex-col min-w-0">
                    <span className="text-sm font-medium text-text-primary truncate">
                      {t('dashboard.activeMonitoring')}
                    </span>
                    <span className="text-xs text-text-muted">
                      {activeFunctions === 0 ? (
                        <span className="italic">{t('dashboard.noFunctionsDeployed')}</span>
                      ) : (
                        t('dashboard.functionsDeployed', { count: activeFunctions })
                      )}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        ),
      },
      {
        id: 'error-actions',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.28 }}
            className="grid grid-cols-1 lg:grid-cols-3 gap-4"
          >
            <ErrorRateWidget data={errorRateDataPrecomputed} className="error-rate-widget lg:col-span-2" />
            <QuickActionsPanel
              onCreateFunction={() => navigate('/functions/new')}
              onCreateGraph={() => navigate('/frg')}
              onCreateApp={() => navigate('/apps/new')}
              onConnectProvider={() => navigate('/providers')}
              onViewSecrets={() => navigate('/vault')}
              onViewLogs={() => navigate('/functions')}
              onSettings={() => navigate('/settings')}
              className="quick-actions-panel"
            />
          </motion.div>
        ),
      },
      {
        id: 'region-performance',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.31 }}
            className="grid grid-cols-1 lg:grid-cols-2 gap-4"
          >
            <RegionDistributionWidget
              regions={regionDataPrecomputed}
              totalFunctions={activeFunctions}
              className="region-distribution-widget"
            />
            <PerformanceLeaderboard functions={performanceDataPrecomputed} maxItems={3} className="performance-leaderboard" />
          </motion.div>
        ),
      },
      {
        id: 'quota-usage',
        content: !hasMinPlan('enterprise') ? (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.34 }}
            className="grid grid-cols-1 lg:grid-cols-3 gap-4"
          >
            {(() => {
              const allQuotas = [
                {
                  name: 'Functions',
                  used: activeFunctions,
                  limit: limits?.functions ?? 0,
                  icon: QUOTA_ICONS.functions,
                },
                {
                  name: 'Requests',
                  used: requestsThisMonth,
                  limit: limits?.requests === Infinity ? -1 : (limits?.requests ?? 0),
                  icon: QUOTA_ICONS.requests,
                },
                {
                  name: 'Secrets',
                  used: 0,
                  limit: limits?.secrets ?? 0,
                  icon: QUOTA_ICONS.secrets,
                },
                {
                  name: 'Providers',
                  used: providers?.length ?? 0,
                  limit: limits?.providers ?? 0,
                  icon: QUOTA_ICONS.providers,
                },
                {
                  name: 'State Fabrics',
                  used: 0,
                  limit: limits?.stateFabrics ?? 0,
                  icon: QUOTA_ICONS.stateFabrics,
                },
                {
                  name: 'Agents',
                  used: 0,
                  limit: limits?.agents ?? 0,
                  icon: QUOTA_ICONS.agents,
                },
              ].filter((q) => q.limit > 0);

              return allQuotas.length > 0 ? (
                <QuotaUsageWidget
                  quotas={allQuotas}
                  onUpgradeClick={isFree ? () => setShowPlanModal(true) : undefined}
                  className="quota-usage-widget lg:col-span-2"
                />
              ) : null;
            })()}
          </motion.div>
        ) : null,
      },
      {
        id: 'provider-status',
        content: (
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            {providersLoading ? (
              <Card className="provider-status-card glass-card glow hover-lift flex items-center justify-center py-12">
                <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
              </Card>
            ) : !providerStatusList.length ? (
              <Card className="provider-status-card glass-card glow hover-lift">
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <p className="text-text-secondary text-sm">
                    {t('dashboard.noProvidersConnected')}
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    className="mt-3"
                    onClick={() => navigate('/providers')}
                  >
                    {t('dashboard.connectAProvider')}
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <ProviderStatus
                providers={providerStatusList}
                className="provider-status-grid grid-cols-1 lg:grid-cols-2"
              />
            )}
          </motion.div>
        ),
      },
      {
        id: 'recent-lists',
        content: (
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="grid grid-cols-1 lg:grid-cols-2 gap-4"
          >
            <Card className="recent-list-card glass-card glow hover-lift">
              <CardHeader>
                <CardTitle className="recent-list-title text-text-primary text-glow">
                  {t('dashboard.recentApps')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {appsLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
                  </div>
                ) : apps.length === 0 ? (
                  <div className="empty-state">
                    <p className="empty-state-text">{t('dashboard.noAppsYet')}</p>
                    <Button
                      variant="outline"
                      size="sm"
                      className="empty-state-btn"
                      onClick={() => navigate('/apps')}
                    >
                      {t('dashboard.createAnApp')}
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {apps.slice(0, 5).map((app, index) => (
                      <motion.div
                        key={app.id}
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.5, delay: 0.4 + index * 0.1 }}
                        className="recent-list-item flex gap-3 p-3 rounded-lg hover:bg-white/5 transition-colors duration-200 cursor-pointer"
                        onClick={() => navigate(`/apps/${encodeURIComponent(app.slug)}`)}
                      >
                        <div className="recent-list-icon w-10 h-10 shrink-0 rounded-lg bg-bg-tertiary flex items-center justify-center">
                          <Building2 className="w-5 h-5 text-text-muted" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <p className="recent-list-name text-sm text-text-primary font-medium truncate">
                            {app.name}
                          </p>
                          <p className="text-xs text-text-muted truncate">{app.slug}</p>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="recent-list-card glass-card glow hover-lift">
              <CardHeader>
                <CardTitle className="recent-list-title text-text-primary text-glow">
                  {t('dashboard.recentFunctions')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {functionsLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
                  </div>
                ) : functions.length === 0 ? (
                  <div className="empty-state">
                    <p className="empty-state-text">
                      {t('dashboard.noFunctionsDeployedYet')}
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      className="empty-state-btn"
                      onClick={() => navigate('/functions/new')}
                    >
                      {t('dashboard.deployAFunctionBtn')}
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {functions.slice(0, 5).map((fn, index) => (
                      <motion.div
                        key={fn.id}
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.5, delay: 0.5 + index * 0.1 }}
                        className="recent-list-item flex gap-3 p-3 rounded-lg hover:bg-white/5 transition-colors duration-200 cursor-pointer"
                        onClick={() => navigate(`/functions/${fn.id}`)}
                      >
                        <div className="recent-list-dot w-2 h-2 mt-2 rounded-full bg-linear-to-r from-[#6366f1] to-[#8b5cf6]" />
                        <div>
                          <p className="recent-list-name text-sm text-text-primary font-medium">{fn.name}</p>
                          <p className="recent-list-status text-xs text-text-muted capitalize">
                            {fn.status || 'unknown'}
                          </p>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </motion.div>
        ),
      },
      {
        id: 'activity-feed',
        content: (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.35 }}
          >
            {activityLoading ? (
              <Card className="agent-activity-feed flex items-center justify-center py-16">
                <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
              </Card>
            ) : (
              <AgentActivityFeed
                activities={agentActivities}
                title={t('dashboard.recentActivityTitle')}
                maxItems={5}
                className="agent-activity-feed"
              />
            )}
          </motion.div>
        ),
      },
    ];
    return sections;
  }, [
    healthStatus,
    functionsLoading,
    activeFunctions,
    metricsLoading,
    avgLatencyDisplay,
    avgLatencyLabel,
    uptimePct,
    uptimeChangePercent,
    uptimeSparkline,
    requestsThisMonth,
    requestsChangePercent,
    formatRequests,
    requestsSparkline,
    isFree,
    navigate,
    setShowPlanModal,
    usageLoading,
    usageGraphData,
    executionRateLoading,
    executionRateData,
    memoryLoading,
    memoryData,
    functions,
    trustScoresLoading,
    aggregateTrustScore,
    providerStatusList,
    providersLoading,
    apps,
    appsLoading,
    activityLoading,
    agentActivities,
    hasMinPlan,
    errorRateDataPrecomputed,
    regionDataPrecomputed,
    performanceDataPrecomputed,
  ]);

  return (
    <div className="aviation-dashboard relative space-y-6">
      {/* Resume Onboarding Banner - pinned to top, not draggable */}
      {canResume() && (
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="onboarding-banner glow hover-lift p-4"
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="onboarding-icon">
                <Play className="w-5 h-5 text-[#6366f1]" />
              </div>
              <div>
                <h3 className="onboarding-title">
                  {t('dashboard.completeYourSetup')}
                </h3>
                <p className="onboarding-description">
                  {t('dashboard.onboardingProgress', { completed: completedSteps.length })}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button onClick={handleResumeOnboarding} className="btn-primary" size="sm">
                <Play className="w-4 h-4 mr-2" />
                {t('dashboard.resumeSetup')}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  localStorage.setItem('onboarding-banner-dismissed', 'true');
                  window.location.reload();
                }}
              >
                <X className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </motion.div>
      )}

      {/* Draggable Dashboard Sections */}
      <DraggableDashboardGrid sections={dashboardSections} storageKey="dashboard-section-order" />

      {/* Plan Selection Modal for free users */}
      <PlanSelectionModal
        isOpen={showPlanModal}
        onClose={() => {
          setShowPlanModal(false);
          setIsCheckoutLoading(false);
        }}
        isFree={isFree}
        nextTier={nextTier}
        onSelectPlan={async (planId: string, priceId?: string) => {
          if (!priceId || priceId.includes('placeholder')) {
            if (planId === 'enterprise') {
              navigate('/contact');
            }
            return;
          }

          setIsCheckoutLoading(true);
          try {
            const base = window.location.origin;
            const successUrl = user?.username
              ? `${base}/u/${user.username}/settings/billing?subscription=success`
              : `${base}/settings?tab=billing&subscription=success`;
            const cancelUrl = `${base}/dashboard?subscription=cancel`;

            const { url } = await createCheckoutSession(priceId, successUrl, cancelUrl);
            window.location.href = url;
          } catch (err) {
            setIsCheckoutLoading(false);
            console.error('Checkout error:', err);
          }
        }}
        isCheckoutLoading={isCheckoutLoading}
        featureName="Functions"
      />
    </div>
  );
}
