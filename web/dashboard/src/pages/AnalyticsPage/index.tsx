import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks';
import { CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Globe, Clock, AlertTriangle, TrendingUp, Loader2, Rocket } from 'lucide-react';
import {
  useFunctions,
  useDashboardUsage,
  useDashboardExecutionRate,
  useDashboardMetrics,
} from '@/hooks';
import { UsageGraph } from '@/components/dashboard';
import { LineChart } from '@/components/common/LineChart';
import './styles.css';

export function AnalyticsPage() {
  usePageTitle('Analytics');
  const { t } = useTranslation();
  const [timeRange, setTimeRange] = useState<'24h' | '7d' | '30d'>('7d');

  // Map time range to days for API calls
  const days = timeRange === '24h' ? 1 : timeRange === '7d' ? 7 : 30;
  const hours = timeRange === '24h' ? 24 : timeRange === '7d' ? 168 : 720;

  // Use hooks instead of raw queries
  const { data: functionsData, isLoading: functionsLoading } = useFunctions();
  const { data: usageData, isLoading: usageLoading } = useDashboardUsage(days);
  const { data: executionRateData, isLoading: executionRateLoading } =
    useDashboardExecutionRate(hours);
  const { data: metricsData, isLoading: metricsLoading } = useDashboardMetrics();

  const functions = functionsData?.functions ?? [];
  const activeFunctions = functions.filter((f) => f.status === 'deployed').length;

  // Process usage data for chart
  const usageChartData = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.map((d) => ({
      time: new Date(d.time + 'Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      value: Number(d.value),
    }));
  }, [usageData]);

  // Process execution rate data for chart
  const executionChartData = useMemo(() => {
    const raw = executionRateData?.data ?? [];
    return raw.map((d) => ({
      time: d.time,
      value: Number(d.rate),
    }));
  }, [executionRateData]);

  // Calculate aggregate metrics from data
  const totalRequests = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.reduce((sum, d) => sum + Number(d.value), 0);
  }, [usageData]);

  const avgRequestsPerDay = useMemo(() => {
    const raw = usageData?.data ?? [];
    if (raw.length === 0) return 0;
    return Math.round(totalRequests / raw.length);
  }, [usageData, totalRequests]);

  // Latency and success/error rate from dashboard metrics API
  const avgLatency =
    metricsData?.avg_latency_ms != null ? Math.round(metricsData.avg_latency_ms) : undefined;
  const successRate =
    metricsData?.uptime_pct != null ? Math.round(metricsData.uptime_pct * 10) / 10 : undefined;
  const errorRate = successRate != null ? Math.round((100 - successRate) * 10) / 10 : undefined;

  const stats = [
    {
      title: t('analytics.totalRequests'),
      value: functionsLoading ? '—' : totalRequests > 0 ? totalRequests.toLocaleString() : '0',
      change: {
        value: 0,
        label:
          totalRequests > 0
            ? t('analytics.lastTimeRange', { timeRange })
            : t('analytics.noDataYet'),
      },
      icon: <Globe className="w-5 h-5 text-brand-500" />,
      trend: totalRequests > 0 ? ('up' as const) : ('neutral' as const),
    },
    {
      title: t('analytics.avgLatency'),
      value:
        functionsLoading || metricsLoading
          ? '—'
          : avgLatency != null && avgLatency >= 0
            ? `${avgLatency}ms`
            : '—',
      change: {
        value: 0,
        label: avgLatency != null ? t('analytics.last7d') : t('analytics.noDataYet'),
      },
      icon: <Clock className="w-5 h-5 text-brand-500" />,
      trend: avgLatency != null && avgLatency < 100 ? ('down' as const) : ('neutral' as const),
    },
    {
      title: t('analytics.errorRate'),
      value: functionsLoading || metricsLoading ? '—' : errorRate != null ? `${errorRate}%` : '—',
      change: {
        value: 0,
        label: errorRate != null ? t('analytics.last7d') : t('analytics.noDataYet'),
      },
      icon: <AlertTriangle className="w-5 h-5 text-error" />,
      trend: errorRate != null && errorRate < 1 ? ('down' as const) : ('neutral' as const),
    },
    {
      title: t('analytics.successRate'),
      value:
        functionsLoading || metricsLoading ? '—' : successRate != null ? `${successRate}%` : '—',
      change: {
        value: 0,
        label: successRate != null ? t('analytics.last7d') : t('analytics.noDataYet'),
      },
      icon: <TrendingUp className="w-5 h-5 text-success" />,
      trend: successRate != null && successRate > 99 ? ('up' as const) : ('neutral' as const),
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: 'var(--analytics-card-text)' }}>
            {t('analytics.title')}
          </h1>
          <p className="mt-1" style={{ color: 'var(--analytics-card-text-secondary)' }}>
            {t('analytics.subtitle')}
          </p>
        </div>
        <div className="analytics-time-toggle">
          {/* Time Range Controls */}
          {(['24h', '7d', '30d'] as const).map((range) => (
            <button
              key={range}
              className={`analytics-time-btn ${timeRange === range ? 'analytics-time-btn-active' : ''}`}
              onClick={() => setTimeRange(range)}
            >
              {range === '24h'
                ? t('analytics.hours24')
                : range === '7d'
                  ? t('analytics.days7')
                  : t('analytics.days30')}
            </button>
          ))}
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <div key={stat.title} className="analytics-stat-card aviation-panel-glow">
            <div className="flex items-start justify-between">
              <div>
                <div className="analytics-stat-value analytics-stat-value-primary">
                  {stat.value}
                </div>
                <div className="analytics-stat-label">{stat.title}</div>
                <div className="analytics-stat-change">{stat.change.label}</div>
              </div>
              <div
                className="analytics-stat-icon"
                style={{ background: 'var(--analytics-info-bg)' }}
              >
                {stat.icon}
              </div>
            </div>
            {/* Aviation corner accents */}
            <div className="aviation-corner aviation-corner-tl" />
            <div className="aviation-corner aviation-corner-tr" />
            <div className="aviation-corner aviation-corner-bl" />
            <div className="aviation-corner aviation-corner-br" />
          </div>
        ))}
      </div>

      {/* Loading or empty state */}
      {(functionsLoading || usageLoading || executionRateLoading || metricsLoading) && (
        <div className="analytics-loading-container">
          <Loader2 className="analytics-loading-spinner" />
        </div>
      )}

      {!functionsLoading &&
        !usageLoading &&
        !executionRateLoading &&
        !metricsLoading &&
        functions.length === 0 && (
          <div className="analytics-empty-state">
            <div className="analytics-empty-icon">
              <Rocket className="h-8 w-8" />
            </div>
            <h3 className="analytics-empty-title">{t('analytics.noFunctionsDeployed')}</h3>
            <p className="analytics-empty-description">{t('analytics.noFunctionsDescription')}</p>
            <Button asChild className="btn-bundle-primary">
              <a href="/functions/new">{t('analytics.deployFunction')}</a>
            </Button>
          </div>
        )}

      {/* Charts Grid - shown when functions exist */}
      {!functionsLoading &&
        !usageLoading &&
        !executionRateLoading &&
        !metricsLoading &&
        functions.length > 0 && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="analytics-card analytics-chart-card aviation-panel-glow">
              <div className="analytics-chart-header">
                <h3 className="analytics-chart-title">{t('analytics.requestsOverTime')}</h3>
              </div>
              <CardContent className="pb-2">
                {usageChartData.length > 0 ? (
                  <div className="h-[300px]">
                    <UsageGraph
                      data={usageChartData}
                      title=""
                      valueLabel={t('analytics.requests')}
                    />
                  </div>
                ) : (
                  <div className="analytics-chart-empty">
                    <p>{t('analytics.noRequestData')}</p>
                  </div>
                )}
              </CardContent>
              {/* Aviation corner accents */}
              <div className="aviation-corner aviation-corner-tl" />
              <div className="aviation-corner aviation-corner-tr" />
              <div className="aviation-corner aviation-corner-bl" />
              <div className="aviation-corner aviation-corner-br" />
            </div>

            <div className="analytics-card analytics-chart-card aviation-panel-glow">
              <div className="analytics-chart-header">
                <h3 className="analytics-chart-title">{t('analytics.executionRate')}</h3>
              </div>
              <CardContent className="pb-2">
                {executionChartData.length > 0 ? (
                  <div className="h-[300px]">
                    <LineChart
                      data={executionChartData}
                      series={[
                        {
                          key: 'value',
                          name: t('analytics.executions'),
                          color: '#f16c3b',
                        },
                      ]}
                      xAxisKey="time"
                      showLegend={false}
                    />
                  </div>
                ) : (
                  <div className="analytics-chart-empty">
                    <p>{t('analytics.noExecutionData')}</p>
                  </div>
                )}
              </CardContent>
              {/* Aviation corner accents */}
              <div className="aviation-corner aviation-corner-tl" />
              <div className="aviation-corner aviation-corner-tr" />
              <div className="aviation-corner aviation-corner-bl" />
              <div className="aviation-corner aviation-corner-br" />
            </div>
          </div>
        )}
    </div>
  );
}
