import { enterpriseSlaApi } from '@/api/enterprise';
import { LineChart } from '@/components/common/LineChart';
import { PageLayout } from '@/components/layout/PageLayout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { Skeleton } from '@/components/ui/skeleton';
import { usePlan } from '@/hooks/usePlan';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { AlertCircle, AlertTriangle, BarChart3, CheckCircle, TrendingUp } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

/**
 * Enterprise SLA Dashboard Page
 * Shows uptime metrics, incident history, and SLA compliance
 */
function EnterpriseSLAPage() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();

  // Redirect non-enterprise users
  if (!isEnterprise) {
    return (
      <PageLayout title="SLA Dashboard">
        <div className="enterprise-sla-page">
          <Card className="sla-gate-card">
            <CardContent className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
                <TrendingUp className="w-8 h-8 text-amber-400" />
              </div>
              <h2 className="sla-gate-title text-xl font-semibold mb-2">Enterprise Feature</h2>
              <p className="sla-gate-desc mb-6 max-w-md">
                The SLA Dashboard is available exclusively for Enterprise plan customers. Upgrade to
                access detailed uptime metrics and SLA compliance reports.
              </p>
              <Button
                onClick={() => navigate('/pricing')}
                className="bg-gradient-to-r from-amber-500 to-yellow-500"
              >
                View Enterprise Plans
              </Button>
            </CardContent>
          </Card>
        </div>
      </PageLayout>
    );
  }

  const periodDays = 30;
  const {
    data: overview,
    isLoading: overviewLoading,
    error: overviewError,
  } = useQuery({
    queryKey: ['enterprise-sla-overview', periodDays],
    queryFn: () => enterpriseSlaApi.getOverview(periodDays),
    staleTime: 60 * 1000,
  });

  const {
    data: incidentsData,
    isLoading: incidentsLoading,
    error: incidentsError,
  } = useQuery({
    queryKey: ['enterprise-sla-incidents', periodDays],
    queryFn: () => enterpriseSlaApi.getIncidents({ limit: 20, days: periodDays }),
    staleTime: 60 * 1000,
  });

  const { data: uptimeHistoryData, isLoading: uptimeHistoryLoading } = useQuery({
    queryKey: ['enterprise-sla-uptime-history', periodDays],
    queryFn: () => enterpriseSlaApi.getUptimeHistory(periodDays),
    staleTime: 60 * 1000,
  });

  const incidents = incidentsData?.incidents ?? [];
  const uptimePoints = uptimeHistoryData?.points ?? [];
  const chartData = uptimePoints.map((p) => ({
    name: p.date,
    uptime: p.uptime_percent,
    incidents: p.incident_count,
  }));
  const isForbidden =
    overviewError &&
    (overviewError as { response?: { status?: number } })?.response?.status === 403;

  if (isForbidden) {
    return (
      <PageLayout title="SLA Dashboard">
        <div className="enterprise-sla-page">
          <Card className="sla-gate-card">
            <CardContent className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
                <TrendingUp className="w-8 h-8 text-amber-400" />
              </div>
              <h2 className="sla-gate-title text-xl font-semibold mb-2">Enterprise Feature</h2>
              <p className="sla-gate-desc mb-6 max-w-md">
                The SLA Dashboard is available exclusively for Enterprise plan customers. Upgrade to
                access detailed uptime metrics and SLA compliance reports.
              </p>
              <Button
                onClick={() => navigate('/pricing')}
                className="bg-gradient-to-r from-amber-500 to-yellow-500"
              >
                View Enterprise Plans
              </Button>
            </CardContent>
          </Card>
        </div>
      </PageLayout>
    );
  }

  if (overviewError && !overview) {
    return (
      <PageLayout title="SLA Dashboard">
        <div className="enterprise-sla-page">
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-16 text-center">
              <AlertCircle className="w-12 h-12 text-amber-500 mb-4" />
              <h2 className="text-lg font-semibold mb-2">Failed to load SLA data</h2>
              <p className="text-text-muted mb-4">
                We couldn’t load your uptime metrics. Please try again later.
              </p>
              <Button variant="outline" onClick={() => window.location.reload()}>
                Retry
              </Button>
            </CardContent>
          </Card>
        </div>
      </PageLayout>
    );
  }

  const showOverviewSkeleton = overviewLoading && !overview;
  const currentUptime = overview?.current_uptime_percent ?? 0;
  const slaTarget = overview?.sla_target_percent ?? 0;
  const incidentCount = overview?.incident_count ?? 0;
  const uptimeStatus: 'success' | 'warning' | 'error' = overview
    ? currentUptime >= slaTarget
      ? 'success'
      : currentUptime >= 99
        ? 'warning'
        : 'error'
    : 'warning';
  const incidentStatus: 'success' | 'warning' | 'error' = overview
    ? incidentCount === 0
      ? 'success'
      : incidentCount <= 2
        ? 'warning'
        : 'error'
    : 'warning';

  return (
    <PageLayout title="SLA Dashboard">
      <div className="enterprise-sla-page">
        <p className="enterprise-sla-subtitle">
          Monitor your service level agreements and uptime metrics
        </p>
        <div className="space-y-6">
          {/* SLA Overview Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {showOverviewSkeleton ? (
              <>
                <SLACardSkeleton index={0} />
                <SLACardSkeleton index={1} />
                <SLACardSkeleton index={2} />
              </>
            ) : (
              <>
                <SLACard
                  title="Current Uptime"
                  value={`${currentUptime.toFixed(2)}%`}
                  subtitle={`Last ${periodDays} days`}
                  icon={TrendingUp}
                  status={uptimeStatus}
                  index={0}
                />
                <SLACard
                  title="SLA Target"
                  value={`${slaTarget}%`}
                  subtitle="Guaranteed uptime"
                  icon={CheckCircle}
                  status="success"
                  index={1}
                />
                <SLACard
                  title="Incidents"
                  value={String(incidentCount)}
                  subtitle={`Last ${periodDays} days`}
                  icon={AlertCircle}
                  status={incidentStatus}
                  index={2}
                />
              </>
            )}
          </div>

          {/* Uptime History Chart */}
          <Card className="sla-uptime-card">
            <CardHeader>
              <CardTitle className="sla-card-title">Uptime History</CardTitle>
            </CardHeader>
            <CardContent>
              {uptimeHistoryLoading && !chartData.length ? (
                <div className="flex items-center justify-center py-12">
                  <LoadingSpinner size="md" />
                </div>
              ) : chartData.length === 0 ? (
                <div className="sla-uptime-empty">
                  <div className="sla-uptime-empty-icon">
                    <BarChart3 className="w-8 h-8" strokeWidth={1.75} />
                  </div>
                  <span className="sla-uptime-empty-title">No uptime data yet</span>
                  <p className="sla-uptime-empty-desc">
                    Daily uptime will appear here once data is available for the selected period.
                  </p>
                </div>
              ) : (
                <LineChart
                  data={chartData}
                  series={[
                    {
                      key: 'uptime',
                      name: 'Uptime %',
                      color: 'var(--chart-green, #22c55e)',
                      strokeWidth: 2,
                    },
                  ]}
                  xAxisKey="name"
                  height={280}
                  showLegend={true}
                  tooltipFormatter={(value) => [`${Number(value).toFixed(2)}%`, 'Uptime']}
                  yAxisFormatter={(v) => `${v}%`}
                  className="border-0 shadow-none bg-transparent"
                />
              )}
            </CardContent>
          </Card>

          {/* Recent Incidents */}
          <Card className="sla-incidents-card">
            <CardHeader>
              <CardTitle className="sla-card-title">Recent Incidents</CardTitle>
            </CardHeader>
            <CardContent>
              {incidentsError ? (
                <div className="flex items-center gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 p-4">
                  <AlertCircle className="w-5 h-5 text-amber-500 shrink-0" />
                  <p className="text-sm text-text-secondary">
                    Failed to load incidents. Try again later.
                  </p>
                </div>
              ) : incidentsLoading && !incidents.length ? (
                <div className="flex items-center justify-center py-8">
                  <LoadingSpinner size="md" />
                </div>
              ) : incidents.length === 0 ? (
                <div className="space-y-4">
                  <div className="sla-incident-row">
                    <div className="flex items-center gap-3">
                      <CheckCircle className="w-5 h-5 text-green-500 shrink-0" />
                      <div>
                        <p className="incident-title">No incidents reported</p>
                        <p className="incident-desc">Your services have been running smoothly</p>
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="space-y-4">
                  {incidents.map((inc) => (
                    <div key={inc.id} className="sla-incident-row">
                      <div className="flex items-center gap-3">
                        {inc.status === 'resolved' ? (
                          <CheckCircle className="w-5 h-5 text-green-500 shrink-0" />
                        ) : (
                          <AlertTriangle className="w-5 h-5 text-amber-500 shrink-0" />
                        )}
                        <div className="min-w-0 flex-1">
                          <p className="incident-title">{inc.title}</p>
                          <p className="incident-desc">
                            {inc.description || `${inc.severity} · ${inc.status}`}
                          </p>
                          <p className="text-xs text-text-muted mt-1">
                            {new Date(inc.created_at).toLocaleString()}
                            {inc.resolved_at &&
                              ` · Resolved ${new Date(inc.resolved_at).toLocaleString()}`}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </PageLayout>
  );
}

interface SLACardProps {
  title: string;
  value: string;
  subtitle: string;
  icon: typeof TrendingUp;
  status: 'success' | 'warning' | 'error';
  index?: number;
}

function SLACard({ title, value, subtitle, icon: Icon, status, index = 0 }: SLACardProps) {
  const iconWrapClass =
    status === 'success' ? 'success' : status === 'warning' ? 'warning' : 'error';

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.08, ease: 'easeOut' }}
    >
      <Card className="sla-metric-card">
        <CardContent className="p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="sla-label">{title}</p>
              <p className="sla-value mt-1">{value}</p>
              <p className="sla-subtitle">{subtitle}</p>
            </div>
            <div className={`sla-icon-wrap ${iconWrapClass}`}>
              <Icon className="w-5 h-5" />
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function SLACardSkeleton({ index }: { index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.08, ease: 'easeOut' }}
    >
      <Card className="sla-metric-card">
        <CardContent className="p-6">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-3">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-7 w-24" />
              <Skeleton className="h-4 w-40" />
            </div>
            <Skeleton className="h-10 w-10 rounded-lg" />
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

export { EnterpriseSLAPage };
export default EnterpriseSLAPage;
