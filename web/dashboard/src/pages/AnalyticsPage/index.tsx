import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks';
import { Globe, Clock, AlertTriangle, TrendingUp, Loader2, Rocket } from 'lucide-react';
import {
  useFunctions,
  useDashboardUsage,
  useDashboardExecutionRate,
  useDashboardMetrics,
} from '@/hooks';
import { UsageGraph } from '@/components/dashboard';
import { LineChart } from '@/components/common/LineChart';
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
import './styles.css';

export function AnalyticsPage() {
  usePageTitle('Analytics');
  const { t } = useTranslation();
  const [timeRange, setTimeRange] = useState<'24h' | '7d' | '30d'>('7d');

  const days = timeRange === '24h' ? 1 : timeRange === '7d' ? 7 : 30;
  const hours = timeRange === '24h' ? 24 : timeRange === '7d' ? 168 : 720;

  const { data: functionsData, isLoading: functionsLoading } = useFunctions();
  const { data: usageData, isLoading: usageLoading } = useDashboardUsage(days);
  const { data: executionRateData, isLoading: executionRateLoading } = useDashboardExecutionRate(hours);
  const { data: metricsData, isLoading: metricsLoading } = useDashboardMetrics();

  const functions = functionsData?.functions ?? [];
  const activeFunctions = functions.filter((f) => f.status === 'deployed').length;

  const usageChartData = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.map((d) => ({
      time: new Date(d.time + 'Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      value: Number(d.value),
    }));
  }, [usageData]);

  const executionChartData = useMemo(() => {
    const raw = executionRateData?.data ?? [];
    return raw.map((d) => ({ time: d.time, value: Number(d.rate) }));
  }, [executionRateData]);

  const totalRequests = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.reduce((sum, d) => sum + Number(d.value), 0);
  }, [usageData]);

  const avgLatency = metricsData?.avg_latency_ms != null ? Math.round(metricsData.avg_latency_ms) : undefined;
  const successRate = metricsData?.uptime_pct != null ? Math.round(metricsData.uptime_pct * 10) / 10 : undefined;
  const errorRate = successRate != null ? Math.round((100 - successRate) * 10) / 10 : undefined;

  const loading = functionsLoading || usageLoading || executionRateLoading || metricsLoading;

  return (
    <div className="an-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="an-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE AN-01" secondary="Analytics" position="top-right" />

        <div className="an-hero__header">
          <div className="an-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="an-hero__title">{t('analytics.title')}</h1>
          </div>
          <p className="an-hero__subtitle">{t('analytics.subtitle')}</p>
          <div className="an-hero__controls">
            <div className="an-time-toggle">
              {(['24h', '7d', '30d'] as const).map((range) => (
                <button
                  key={range}
                  className={`an-time-btn ${timeRange === range ? 'an-time-btn--active' : ''}`}
                  onClick={() => setTimeRange(range)}
                >
                  {range === '24h' ? t('analytics.hours24') : range === '7d' ? t('analytics.days7') : t('analytics.days30')}
                </button>
              ))}
            </div>
          </div>
        </div>

        <GaugeStrip>
          <Gauge isFirst data={{ value: loading ? '...' : totalRequests.toLocaleString(), label: t('analytics.totalRequests') }} />
          <Gauge data={{ value: loading ? '...' : avgLatency != null ? `${avgLatency}ms` : '—', label: t('analytics.avgLatency') }} />
          <Gauge data={{ value: loading ? '...' : errorRate != null ? `${errorRate}%` : '—', label: t('analytics.errorRate') }} />
          <Gauge data={{ value: loading ? '...' : successRate != null ? `${successRate}%` : '—', label: t('analytics.successRate') }} />
        </GaugeStrip>
      </Chamber>

      {/* Loading */}
      {loading && (
        <div className="an-loading">
          <Loader2 className="an-loading__spinner" />
        </div>
      )}

      {/* Empty State */}
      {!loading && functions.length === 0 && (
        <Chamber className="an-empty">
          <CornerBrace position="tr" />
          <CornerBrace position="bl" />
          <div className="an-empty__center">
            <Rocket className="an-empty__icon" />
            <h3 className="an-empty__title">{t('analytics.noFunctionsDeployed')}</h3>
            <p className="an-empty__desc">{t('analytics.noFunctionsDescription')}</p>
            <a href="/functions/new">
              <SealedButton>{t('analytics.deployFunction')}</SealedButton>
            </a>
          </div>
        </Chamber>
      )}

      {/* Charts */}
      {!loading && functions.length > 0 && (
        <div className="an-charts-grid">
          <Chamber className="an-chart-chamber">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="an-chart-header">
              <h3 className="an-chart-title">{t('analytics.requestsOverTime')}</h3>
            </div>
            {usageChartData.length > 0 ? (
              <div className="an-chart-area">
                <UsageGraph data={usageChartData} title="" valueLabel={t('analytics.requests')} />
              </div>
            ) : (
              <div className="an-chart-empty">
                <p>{t('analytics.noRequestData')}</p>
              </div>
            )}
          </Chamber>

          <Chamber className="an-chart-chamber">
            <CornerBrace position="tr" />
            <CornerBrace position="bl" />
            <div className="an-chart-header">
              <h3 className="an-chart-title">{t('analytics.executionRate')}</h3>
            </div>
            {executionChartData.length > 0 ? (
              <div className="an-chart-area">
                <LineChart
                  data={executionChartData}
                  series={[{ key: 'value', name: t('analytics.executions'), color: '#f16c3b' }]}
                  xAxisKey="time"
                  showLegend={false}
                />
              </div>
            ) : (
              <div className="an-chart-empty">
                <p>{t('analytics.noExecutionData')}</p>
              </div>
            )}
          </Chamber>
        </div>
      )}
    </div>
  );
}
