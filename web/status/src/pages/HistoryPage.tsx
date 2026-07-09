import { Chamber, CornerBrace, StatusPill, FrameButton, SealedButton, PageGrid, ReducedMotionGate, AnnotationTag } from '@/components/containment';
import { Nav } from '@/components/Nav';
import { Footer } from '@/components/Footer';
import { UptimeBar } from '@/components/UptimeBar';
import {
  statusAPI,
  type Component,
  type IncidentsListResponse,
  type UptimeDataPoint,
} from '@/lib/api';
import { trackEvent } from '@/lib/analytics';
import { useQuery } from '@tanstack/react-query';
import { format, parseISO, subDays } from 'date-fns';
import { Activity, AlertTriangle, ArrowLeft, BarChart3, Calendar, Download, Globe, Server, TrendingUp } from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';

interface MonthlyStats {
  month: string;
  uptime: number;
  incidents: number;
  avgResponseTime: number;
}

function transformUptimeData(
  dataPoints: UptimeDataPoint[],
): Array<{
  date: string;
  status: 'operational' | 'degraded' | 'outage' | 'maintenance';
  uptime: number;
}> {
  if (!dataPoints || dataPoints.length === 0) {
    return Array.from({ length: 365 }, (_, i) => ({
      date: format(subDays(new Date(), 364 - i), 'MMM dd'),
      status: 'operational' as const,
      uptime: 99.97,
    }));
  }

  return dataPoints.map((point) => {
    const uptime = point.uptime_percent;
    let status: 'operational' | 'degraded' | 'outage' | 'maintenance' = 'operational';
    if (uptime < 95) status = 'outage';
    else if (uptime < 99) status = 'degraded';

    return {
      date: format(
        parseISO(point.timestamp.toString ? point.timestamp.toString() : new Date().toISOString()),
        'MMM dd',
      ),
      status,
      uptime,
    };
  });
}

function calculateMonthlyStats(
  components: Component[],
  incidents: IncidentsListResponse['incidents'],
): MonthlyStats[] {
  const now = new Date();
  const months: MonthlyStats[] = [];

  for (let i = 0; i < 6; i++) {
    const date = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const monthName = format(date, 'MMMM yyyy');
    const monthStart = new Date(date.getFullYear(), date.getMonth(), 1);
    const monthEnd = new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59);

    const monthIncidents = incidents.filter((inc) => {
      const incDate = parseISO(inc.created_at);
      return incDate >= monthStart && incDate <= monthEnd;
    });

    const avgUptime =
      components.length > 0
        ? components.reduce((acc, c) => acc + (c.uptime_30d || 99.97), 0) / components.length
        : 99.97;

    const avgLatency =
      components.length > 0
        ? Math.round(components.reduce((acc, c) => acc + (c.response_time_ms || 50), 0) / components.length)
        : 45;

    months.push({
      month: monthName,
      uptime: Number(avgUptime.toFixed(2)),
      incidents: monthIncidents.length,
      avgResponseTime: avgLatency,
    });
  }

  return months;
}

export default function HistoryPage() {
  const [selectedRange, setSelectedRange] = useState<'30d' | '90d' | '6m' | '1y'>('1y');

  const { data: uptimeData, isLoading: isLoadingUptime } = useQuery({
    queryKey: ['uptime', 'all', '90d'],
    queryFn: () => statusAPI.getUptimeMetrics('all', '90d'),
  });

  const { data: componentsData, isLoading: isLoadingComponents } = useQuery({
    queryKey: ['components'],
    queryFn: () => statusAPI.getComponents(),
  });

  const { data: incidentsData, isLoading: isLoadingIncidents } = useQuery({
    queryKey: ['incidents', 'history'],
    queryFn: () => statusAPI.listIncidents({ limit: 100 }),
  });

  const ranges = [
    { label: '30d', value: '30d' as const },
    { label: '90d', value: '90d' as const },
    { label: '6m', value: '6m' as const },
    { label: '1y', value: '1y' as const },
  ];

  const isLoading = isLoadingUptime || isLoadingComponents || isLoadingIncidents;
  const components = componentsData?.components || [];
  const incidents = incidentsData?.incidents || [];
  const uptimePoints = uptimeData?.data_points || [];
  const yearlyUptime = transformUptimeData(uptimePoints);
  const monthlyStats = calculateMonthlyStats(components, incidents);

  const filteredUptime = yearlyUptime.slice(
    -(selectedRange === '30d' ? 30 : selectedRange === '90d' ? 90 : selectedRange === '6m' ? 180 : 365),
  );

  const overallUptime = uptimeData?.overall_uptime || 99.97;
  const totalIncidents = incidents.filter((inc) => inc.status === 'resolved').length;
  const activeIncidents = incidents.filter((inc) => inc.status !== 'resolved').length;
  const avgResponse =
    components.length > 0
      ? Math.round(components.reduce((acc, c) => acc + c.response_time_ms, 0) / components.length)
      : 45;

  const rangeButtons = ranges.map((range) => (
    <button
      key={range.value}
      onClick={() => {
        trackEvent('history_range_selected', { range: range.value });
        setSelectedRange(range.value);
      }}
      className="transition-all"
      style={{
        padding: '6px 12px',
        fontSize: '13px',
        fontWeight: 500,
        fontFamily: 'var(--font-mono)',
        borderRadius: 'var(--radius-sm)',
        border: 'none',
        cursor: 'pointer',
        background: selectedRange === range.value ? 'var(--steel)' : 'transparent',
        color: selectedRange === range.value ? 'var(--text)' : 'var(--text-faint)',
      }}
    >
      {range.label}
    </button>
  ));

  return (
    <ReducedMotionGate>
      <PageGrid />
      <div className="min-h-screen" style={{ background: 'var(--bg)', position: 'relative', zIndex: 1 }}>
        <Nav />

        <main style={{ paddingTop: '120px', paddingBottom: 'var(--space-8)' }}>
          <div
            className="mx-auto"
            style={{ maxWidth: '1180px', padding: '0 var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}
          >
            {/* Page Header */}
            <div className="text-center" style={{ marginBottom: 'var(--space-6)' }}>
              <Link
                to="/"
                className="inline-flex items-center gap-2 mb-4"
                style={{ color: 'var(--text-dim)', textDecoration: 'none', fontSize: '14px' }}
              >
                <ArrowLeft style={{ width: 14, height: 14 }} />
                Back to Status
              </Link>
              <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 'clamp(28px, 4vw, 48px)', fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.01em' }}>
                Status History
              </h1>
              <p style={{ fontSize: '15px', color: 'var(--text-dim)', maxWidth: '560px', margin: 'var(--space-3) auto 0', lineHeight: 1.6 }}>
                Historical uptime data for all FunctionFly services. {overallUptime.toFixed(2)}% uptime guarantee.
              </p>
            </div>

            {/* Overview Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {[
                { label: 'Overall Uptime', value: `${overallUptime.toFixed(2)}%`, icon: TrendingUp, color: 'var(--status-ok)' },
                { label: 'Avg Response', value: `${avgResponse}ms`, icon: Activity, color: 'var(--foil-a)' },
                { label: 'Total Incidents', value: totalIncidents.toString(), icon: AlertTriangle, color: 'var(--status-pending)' },
                { label: 'Active', value: activeIncidents.toString(), icon: Globe, color: activeIncidents > 0 ? 'var(--status-revoked)' : 'var(--status-ok)' },
              ].map((stat) => (
                <Chamber key={stat.label} nested>
                  <stat.icon style={{ width: 20, height: 20, color: stat.color, marginBottom: 'var(--space-2)' }} />
                  <div style={{ fontFamily: 'var(--font-mono)', fontSize: '26px', fontWeight: 500, color: 'var(--text)' }}>
                    {stat.value}
                  </div>
                  <div style={{ fontSize: '13px', color: 'var(--text-faint)', marginTop: 'var(--space-1)' }}>
                    {stat.label}
                  </div>
                </Chamber>
              ))}
            </div>

            {/* Uptime Chart */}
            <Chamber>
              <CornerBrace position="tr" />
              <CornerBrace position="bl" />
              <AnnotationTag label="HISTORY" detail="UPTIME" />

              <div className="flex items-center justify-between flex-wrap gap-3" style={{ marginBottom: 'var(--space-5)' }}>
                <div>
                  <h2 className="flex items-center gap-2" style={{ fontFamily: 'var(--font-display)', fontSize: '18px', fontWeight: 500, color: 'var(--text)' }}>
                    <BarChart3 style={{ width: 18, height: 18, color: 'var(--foil-a)' }} />
                    Uptime Overview
                  </h2>
                  <p style={{ fontSize: '13px', color: 'var(--text-dim)', marginTop: 'var(--space-1)' }}>
                    {selectedRange === '1y' ? 'Past year' : `Past ${selectedRange}`}
                  </p>
                </div>
                <div className="flex items-center gap-1 p-1" style={{ background: 'var(--panel-raised)', borderRadius: 'var(--radius)', border: '1px solid var(--panel-edge)' }}>
                  {rangeButtons}
                </div>
              </div>

              {isLoadingUptime ? (
                <div className="animate-pulse" style={{ height: 60, background: 'var(--panel-raised)', borderRadius: 'var(--radius)' }} />
              ) : (
                <UptimeBar segments={filteredUptime} />
              )}
            </Chamber>

            {/* Monthly Breakdown */}
            <section>
              <h2 className="flex items-center gap-2" style={{ fontFamily: 'var(--font-display)', fontSize: '22px', fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-5)' }}>
                <Calendar style={{ width: 20, height: 20, color: 'var(--foil-a)' }} />
                Monthly Breakdown
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {isLoading
                  ? Array.from({ length: 6 }).map((_, i) => (
                      <Chamber key={i} nested>
                        <div className="animate-pulse" style={{ height: 100, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
                      </Chamber>
                    ))
                  : monthlyStats.map((stat) => (
                      <Chamber key={stat.month} nested style={{ padding: 'var(--space-6)' }}>
                        <div className="flex items-center justify-between gap-3" style={{ marginBottom: 'var(--space-4)' }}>
                          <h3 style={{ fontWeight: 500, color: 'var(--text)', fontSize: '16px' }}>{stat.month}</h3>
                          <StatusPill
                            status={stat.incidents > 0 ? 'pending' : 'live'}
                            label={stat.incidents > 0 ? `${stat.incidents} incident${stat.incidents === 1 ? '' : 's'}` : 'No incidents'}
                          />
                        </div>
                        <div>
                          <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-1)' }}>
                            <span style={{ fontSize: '13px', color: 'var(--text-dim)' }}>Uptime</span>
                            <span style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--status-ok)' }}>{stat.uptime}%</span>
                          </div>
                          <div style={{ height: '4px', background: 'var(--panel-edge)', borderRadius: '2px', overflow: 'hidden' }}>
                            <div style={{ height: '100%', width: `${stat.uptime}%`, background: 'var(--status-ok)', borderRadius: '2px' }} />
                          </div>
                          <div className="flex items-center justify-between" style={{ marginTop: 'var(--space-3)' }}>
                            <span style={{ fontSize: '13px', color: 'var(--text-dim)' }}>Avg Latency</span>
                            <span style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>{stat.avgResponseTime}ms</span>
                          </div>
                        </div>
                      </Chamber>
                    ))}
              </div>
            </section>

            {/* Service Performance */}
            <Chamber>
              <CornerBrace position="tr" />
              <CornerBrace position="bl" />
              <h2 className="flex items-center gap-2" style={{ fontFamily: 'var(--font-display)', fontSize: '18px', fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-5)' }}>
                <Server style={{ width: 18, height: 18, color: 'var(--foil-a)' }} />
                Service Performance (Last 30 Days)
              </h2>

              <div className="overflow-x-auto">
                <table className="w-full" style={{ borderCollapse: 'collapse', fontSize: '13px' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid var(--panel-edge)' }}>
                      <th className="text-left uppercase" style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 500, letterSpacing: '0.06em', color: 'var(--text-faint)', padding: 'var(--space-3) var(--space-2)' }}>Service</th>
                      <th className="text-left uppercase" style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 500, letterSpacing: '0.06em', color: 'var(--text-faint)', padding: 'var(--space-3) var(--space-2)' }}>Status</th>
                      <th className="text-left uppercase hidden sm:table-cell" style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 500, letterSpacing: '0.06em', color: 'var(--text-faint)', padding: 'var(--space-3) var(--space-2)' }}>Latency</th>
                      <th className="text-right uppercase" style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 500, letterSpacing: '0.06em', color: 'var(--text-faint)', padding: 'var(--space-3) var(--space-2)' }}>Uptime</th>
                    </tr>
                  </thead>
                  <tbody>
                    {isLoadingComponents
                      ? Array.from({ length: 6 }).map((_, i) => (
                          <tr key={i} style={{ borderBottom: '1px solid var(--panel-edge)' }}>
                            <td style={{ padding: 'var(--space-3) var(--space-2)' }}><div className="animate-pulse" style={{ height: 16, width: 120, background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)' }} /></td>
                            <td style={{ padding: 'var(--space-3) var(--space-2)' }}><div className="animate-pulse" style={{ height: 16, width: 80, background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)' }} /></td>
                            <td className="hidden sm:table-cell" style={{ padding: 'var(--space-3) var(--space-2)' }}><div className="animate-pulse" style={{ height: 16, width: 60, background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)' }} /></td>
                            <td style={{ padding: 'var(--space-3) var(--space-2)' }}><div className="animate-pulse" style={{ height: 16, width: 60, background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)' }} /></td>
                          </tr>
                        ))
                      : components.map((service) => (
                          <tr key={service.id} style={{ borderBottom: '1px solid var(--panel-edge)' }}>
                            <td style={{ padding: 'var(--space-3) var(--space-2)', color: 'var(--text)', fontWeight: 500 }}>{service.name}</td>
                            <td style={{ padding: 'var(--space-3) var(--space-2)' }}>
                              <StatusPill
                                status={service.status === 'operational' ? 'live' : service.status === 'degraded' ? 'pending' : 'revoked'}
                                label={service.status === 'operational' ? 'OK' : service.status}
                              />
                            </td>
                            <td className="hidden sm:table-cell" style={{ padding: 'var(--space-3) var(--space-2)', fontFamily: 'var(--font-mono)', color: 'var(--text-dim)' }}>
                              {service.response_time_ms}ms
                            </td>
                            <td className="text-right" style={{ padding: 'var(--space-3) var(--space-2)' }}>
                              {service.uptime_30d != null ? (
                                <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 500, color: service.uptime_30d >= 99.98 ? 'var(--status-ok)' : service.uptime_30d >= 99.95 ? 'var(--status-pending)' : 'var(--status-revoked)' }}>
                                  {service.uptime_30d.toFixed(2)}%
                                </span>
                              ) : (
                                <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 500, color: 'var(--text-faint)' }}>N/A</span>
                              )}
                            </td>
                          </tr>
                        ))}
                  </tbody>
                </table>
              </div>
            </Chamber>

            {/* CTA */}
            <Chamber nested>
              <div className="flex flex-col md:flex-row items-center justify-between gap-6">
                <div className="text-center md:text-left">
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '18px', fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                    Need historical data?
                  </h3>
                  <p style={{ fontSize: '15px', color: 'var(--text-dim)' }}>
                    Export full uptime reports or access our API for programmatic access.
                  </p>
                </div>
                <div className="flex gap-3">
                  <FrameButton onClick={() => trackEvent('history_export_clicked')} iconLeft={<Download style={{ width: 14, height: 14 }} />}>
                    Export CSV
                  </FrameButton>
                  <SealedButton onClick={() => trackEvent('history_api_access_clicked')} iconLeft={<Globe style={{ width: 14, height: 14 }} />}>
                    API Access
                  </SealedButton>
                </div>
              </div>
            </Chamber>
          </div>
        </main>

        <Footer />
      </div>
    </ReducedMotionGate>
  );
}
