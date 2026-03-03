import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
} from 'recharts';
import { TrendingUp, TrendingDown, Calendar, Clock } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Sparkline } from '@/components/common/Sparkline';
import type { UptimeMetrics } from '@/api/status';

interface UptimeChartProps {
  metrics: UptimeMetrics | null;
  isLoading?: boolean;
}

type TimeRange = 30 | 90 | 365;

const timeRangeLabels: Record<TimeRange, string> = {
  30: 'Last 30 Days',
  90: 'Last 90 Days',
  365: 'Last Year',
};

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{ value: number; payload: { date: string; uptime: number } }>;
  label?: string;
}

function CustomTooltip({ active, payload, label }: CustomTooltipProps) {
  if (active && payload && payload.length) {
    const data = payload[0].payload;
    return (
      <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3 shadow-lg">
        <p className="text-sm font-medium text-text-primary">
          {new Date(data.date).toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
          })}
        </p>
        <p className="mt-1 text-sm">
          <span className="text-text-muted">Uptime: </span>
          <span
            className={cn(
              'font-medium',
              data.uptime >= 99.9
                ? 'text-emerald-400'
                : data.uptime >= 99
                ? 'text-amber-400'
                : 'text-red-400'
            )}
          >
            {data.uptime.toFixed(3)}%
          </span>
        </p>
      </div>
    );
  }
  return null;
}

function MetricCard({
  title,
  value,
  change,
  trend,
  icon: Icon,
}: {
  title: string;
  value: string;
  change?: number;
  trend?: 'up' | 'down' | 'neutral';
  icon: React.ComponentType<{ className?: string }>;
}) {
  return (
    <div className="rounded-lg bg-bg-tertiary/50 p-4">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs text-text-muted">{title}</p>
          <p className="mt-1 text-2xl font-bold text-text-primary">{value}</p>
          {change !== undefined && (
            <div
              className={cn(
                'mt-1 flex items-center gap-1 text-xs',
                trend === 'up' && 'text-emerald-400',
                trend === 'down' && 'text-red-400',
                trend === 'neutral' && 'text-text-muted'
              )}
            >
              {trend === 'up' && <TrendingUp className="h-3 w-3" />}
              {trend === 'down' && <TrendingDown className="h-3 w-3" />}
              <span>{change > 0 ? '+' : ''}{change}%</span>
            </div>
          )}
        </div>
        <div className="rounded-lg bg-bg-secondary p-2">
          <Icon className="h-4 w-4 text-text-muted" />
        </div>
      </div>
    </div>
  );
}

function UptimeBar({
  label,
  uptime,
  index,
}: {
  label: string;
  uptime: number;
  index: number;
}) {
  const getColor = (uptime: number) => {
    if (uptime >= 99.9) return 'bg-emerald-500';
    if (uptime >= 99) return 'bg-amber-500';
    return 'bg-red-500';
  };

  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.05, duration: 0.3 }}
      className="flex items-center gap-3"
    >
      <span className="w-32 truncate text-sm text-text-secondary">{label}</span>
      <div className="flex-1">
        <div className="h-2 w-full overflow-hidden rounded-full bg-bg-tertiary">
          <motion.div
            className={cn('h-full rounded-full', getColor(uptime))}
            initial={{ width: 0 }}
            animate={{ width: `${uptime}%` }}
            transition={{ duration: 0.5, delay: index * 0.05 }}
          />
        </div>
      </div>
      <span
        className={cn(
          'w-16 text-right text-sm font-medium',
          uptime >= 99.9 ? 'text-emerald-400' : uptime >= 99 ? 'text-amber-400' : 'text-red-400'
        )}
      >
        {uptime.toFixed(2)}%
      </span>
    </motion.div>
  );
}

export function UptimeChart({ metrics, isLoading }: UptimeChartProps) {
  const [selectedRange, setSelectedRange] = useState<TimeRange>(30);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-4 w-48" />
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} className="h-20" />
            ))}
          </div>
          <Skeleton className="h-[300px]" />
        </CardContent>
      </Card>
    );
  }

  if (!metrics || !metrics.daily_data) {
    return (
      <Card className="p-8 text-center">
        <p className="text-text-secondary">No uptime data available</p>
      </Card>
    );
  }

  const chartData = metrics.daily_data.map((day) => ({
    date: day.date,
    uptime: day.uptime,
    incidents: day.incidents,
  }));

  const componentUptime = Object.entries(metrics.by_component || {}).sort(
    ([, a], [, b]) => b - a
  );

  const providerUptime = Object.entries(metrics.by_provider || {}).sort(
    ([, a], [, b]) => b - a
  );

  const avgUptime =
    chartData.reduce((sum, day) => sum + day.uptime, 0) / chartData.length;

  const bestDay = chartData.reduce((best, day) =>
    day.uptime > best.uptime ? day : best
  );

  const worstDay = chartData.reduce((worst, day) =>
    day.uptime < worst.uptime ? day : worst
  );

  const totalIncidents = chartData.reduce((sum, day) => sum + day.incidents, 0);

  return (
    <section aria-label="Uptime Metrics">
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <CardTitle className="text-xl">Uptime History</CardTitle>
              <p className="mt-1 text-sm text-text-secondary">
                Platform availability over time
              </p>
            </div>
            <div className="flex gap-2">
              {[30, 90, 365].map((range) => (
                <Button
                  key={range}
                  variant={selectedRange === range ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setSelectedRange(range as TimeRange)}
                >
                  {range}d
                </Button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {/* Summary metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <MetricCard
              title="Overall Uptime"
              value={`${metrics.overall_uptime.toFixed(3)}%`}
              icon={Clock}
            />
            <MetricCard
              title="Best Day"
              value={`${bestDay.uptime.toFixed(3)}%`}
              icon={TrendingUp}
            />
            <MetricCard
              title="Worst Day"
              value={`${worstDay.uptime.toFixed(3)}%`}
              icon={TrendingDown}
            />
            <MetricCard
              title="Total Incidents"
              value={totalIncidents.toString()}
              icon={Calendar}
            />
          </div>

          {/* Uptime chart */}
          <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart
                data={chartData}
                margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
              >
                <defs>
                  <linearGradient id="uptimeGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="var(--border-subtle)"
                  vertical={false}
                />
                <XAxis
                  dataKey="date"
                  tickFormatter={(value) =>
                    new Date(value).toLocaleDateString('en-US', {
                      month: 'short',
                      day: 'numeric',
                    })
                  }
                  stroke="var(--text-muted)"
                  tick={{ fill: 'var(--text-muted)', fontSize: 12 }}
                  tickLine={false}
                />
                <YAxis
                  domain={[98, 100]}
                  stroke="var(--text-muted)"
                  tick={{ fill: 'var(--text-muted)', fontSize: 12 }}
                  tickFormatter={(value) => `${value}%`}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip content={<CustomTooltip />} />
                <Area
                  type="monotone"
                  dataKey="uptime"
                  stroke="#10b981"
                  strokeWidth={2}
                  fill="url(#uptimeGradient)"
                  isAnimationActive={true}
                  animationDuration={1000}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Component uptime breakdown */}
          {componentUptime.length > 0 && (
            <div className="mt-8 pt-6 border-t border-border-subtle">
              <h3 className="text-sm font-medium text-text-primary mb-4">
                Component Uptime
              </h3>
              <div className="space-y-3">
                {componentUptime.slice(0, 5).map(([name, uptime], index) => (
                  <UptimeBar
                    key={name}
                    label={name}
                    uptime={uptime}
                    index={index}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Provider uptime breakdown */}
          {providerUptime.length > 0 && (
            <div className="mt-6">
              <h3 className="text-sm font-medium text-text-primary mb-4">
                Provider Uptime
              </h3>
              <div className="space-y-3">
                {providerUptime.map(([name, uptime], index) => (
                  <UptimeBar
                    key={name}
                    label={name}
                    uptime={uptime}
                    index={index}
                  />
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </section>
  );
}

// Compact version for dashboard use
export function UptimeMiniChart({
  data,
  className,
}: {
  data: number[];
  className?: string;
}) {
  const avg = data.reduce((a, b) => a + b, 0) / data.length;
  const trend =
    data.length > 1
      ? data[data.length - 1] > data[data.length - 2]
        ? 'up'
        : 'down'
      : 'neutral';

  return (
    <div className={cn('flex items-center gap-3', className)}>
      <Sparkline
        data={data}
        width={80}
        height={30}
        color={avg >= 99.9 ? '#10b981' : avg >= 99 ? '#f59e0b' : '#ef4444'}
      />
      <div className="flex flex-col">
        <span className="text-lg font-semibold text-text-primary">
          {avg.toFixed(2)}%
        </span>
        <span
          className={cn(
            'text-xs',
            trend === 'up' && 'text-emerald-400',
            trend === 'down' && 'text-red-400'
          )}
        >
          {trend === 'up' ? '↑' : trend === 'down' ? '↓' : '→'} uptime
        </span>
      </div>
    </div>
  );
}
