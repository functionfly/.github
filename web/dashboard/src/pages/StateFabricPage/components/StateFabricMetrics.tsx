import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useStateFabricMetrics } from '@/hooks/useStateFabric';
import type { StateFabricMetrics as Metrics } from '@/types';
import { Activity, BarChart3, Clock, Database, TrendingDown, TrendingUp, Zap } from 'lucide-react';
import { useState } from 'react';
import {
  Area,
  AreaChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

interface StateFabricMetricsProps {
  fabricId: string;
  /** Optional initial metrics from parent (e.g. for header); component fetches its own with timeRange when mounted. */
  metrics?: Metrics | undefined;
}

export function StateFabricMetrics({ fabricId, metrics: metricsProp }: StateFabricMetricsProps) {
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d' | '30d'>('24h');
  const { data: metricsFromApi, isLoading } = useStateFabricMetrics(fabricId, timeRange);
  const metrics = metricsFromApi ?? metricsProp;
  const historyData = metricsFromApi?.history ?? [];

  const formatNumber = (num: number | undefined) => {
    if (num === undefined) return 'N/A';
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
  };

  const hasHistory = historyData.length > 0;

  if (isLoading && !metrics) {
    return (
      <div className="flex items-center justify-center py-16 text-text-muted">Loading metrics…</div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Time Range Selector */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-text-primary">Metrics Dashboard</h3>
        <div className="flex gap-2">
          {(['1h', '24h', '7d', '30d'] as const).map((range) => (
            <Badge
              key={range}
              variant={timeRange === range ? 'default' : 'outline'}
              className="cursor-pointer"
              onClick={() => setTimeRange(range)}
            >
              {range === '1h'
                ? '1 Hour'
                : range === '24h'
                  ? '24 Hours'
                  : range === '7d'
                    ? '7 Days'
                    : '30 Days'}
            </Badge>
          ))}
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Total Operations</p>
                <p className="text-2xl font-bold text-text-primary">
                  {formatNumber(metrics?.totalOperations)}
                </p>
              </div>
              <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center">
                <Activity className="w-5 h-5 text-blue-400" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-2 text-sm">
              <TrendingUp className="w-4 h-4 text-green-400" />
              <span className="text-green-400">+12.5%</span>
              <span className="text-text-muted">vs last period</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Operations/sec</p>
                <p className="text-2xl font-bold text-text-primary">
                  {metrics?.operationsPerSecond?.toFixed(1) || '0.0'}
                </p>
              </div>
              <div className="w-10 h-10 rounded-lg bg-green-500/10 flex items-center justify-center">
                <Zap className="w-5 h-5 text-green-400" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-2 text-sm">
              <TrendingUp className="w-4 h-4 text-green-400" />
              <span className="text-green-400">+5.2%</span>
              <span className="text-text-muted">vs last period</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Avg Latency</p>
                <p className="text-2xl font-bold text-text-primary">
                  {metrics?.averageLatency?.toFixed(0) || '0'} ms
                </p>
              </div>
              <div className="w-10 h-10 rounded-lg bg-yellow-500/10 flex items-center justify-center">
                <Clock className="w-5 h-5 text-yellow-400" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-2 text-sm">
              <TrendingDown className="w-4 h-4 text-green-400" />
              <span className="text-green-400">-8.1%</span>
              <span className="text-text-muted">vs last period</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Error Rate</p>
                <p className="text-2xl font-bold text-text-primary">
                  {((metrics?.errorRate || 0) * 100).toFixed(2)}%
                </p>
              </div>
              <div className="w-10 h-10 rounded-lg bg-red-500/10 flex items-center justify-center">
                <Database className="w-5 h-5 text-red-400" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-2 text-sm">
              <TrendingDown className="w-4 h-4 text-green-400" />
              <span className="text-green-400">-2.3%</span>
              <span className="text-text-muted">vs last period</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Operations Chart */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Operations Over Time</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[300px] flex items-center justify-center">
              {hasHistory ? (
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={historyData}>
                    <defs>
                      <linearGradient id="colorOps" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis dataKey="time" stroke="rgba(255,255,255,0.5)" fontSize={12} />
                    <YAxis stroke="rgba(255,255,255,0.5)" fontSize={12} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--bg-secondary)',
                        border: '1px solid var(--border-subtle)',
                      }}
                    />
                    <Area
                      type="monotone"
                      dataKey="operations"
                      stroke="#6366f1"
                      fillOpacity={1}
                      fill="url(#colorOps)"
                      isAnimationActive={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex flex-col items-center gap-2 text-text-muted text-sm">
                  <BarChart3 className="w-10 h-10 opacity-50" />
                  <span>No historical data for this range</span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Latency Chart */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Latency Over Time</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[300px] flex items-center justify-center">
              {hasHistory ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={historyData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis dataKey="time" stroke="rgba(255,255,255,0.5)" fontSize={12} />
                    <YAxis stroke="rgba(255,255,255,0.5)" fontSize={12} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--bg-secondary)',
                        border: '1px solid var(--border-subtle)',
                      }}
                    />
                    <Line
                      type="monotone"
                      dataKey="latency"
                      stroke="#10b981"
                      strokeWidth={2}
                      dot={false}
                      isAnimationActive={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex flex-col items-center gap-2 text-text-muted text-sm">
                  <BarChart3 className="w-10 h-10 opacity-50" />
                  <span>No historical data for this range</span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Storage Usage */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Storage Usage</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between mb-4">
            <div>
              <p className="text-sm text-text-muted">Total Storage Used</p>
              <p className="text-2xl font-bold text-text-primary">
                {formatNumber(metrics?.storageUsed)} bytes
              </p>
            </div>
            {metrics?.cacheHitRate !== undefined && (
              <div className="text-right">
                <p className="text-sm text-text-muted">Cache Hit Rate</p>
                <p className="text-2xl font-bold text-text-primary">
                  {(metrics.cacheHitRate * 100).toFixed(1)}%
                </p>
              </div>
            )}
          </div>
          <div className="h-[200px] flex items-center justify-center">
            {hasHistory ? (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={historyData}>
                  <defs>
                    <linearGradient id="colorStorage" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                  <XAxis dataKey="time" stroke="rgba(255,255,255,0.5)" fontSize={12} />
                  <YAxis stroke="rgba(255,255,255,0.5)" fontSize={12} />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--bg-secondary)',
                      border: '1px solid var(--border-subtle)',
                    }}
                  />
                  <Area
                    type="monotone"
                    dataKey="errors"
                    stroke="#8b5cf6"
                    fillOpacity={1}
                    fill="url(#colorStorage)"
                    isAnimationActive={false}
                  />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex flex-col items-center gap-2 text-text-muted text-sm">
                <BarChart3 className="w-10 h-10 opacity-50" />
                <span>No historical data for this range</span>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
