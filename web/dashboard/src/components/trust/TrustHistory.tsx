import { getTrustColorConfig, getTrustScoreBand } from '@/components/functions/TrustScoreBadge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { AlertTriangle, Shield, TrendingDown, TrendingUp } from 'lucide-react';
import { useMemo } from 'react';
import {
  Area,
  CartesianGrid,
  ComposedChart,
  Line,
  LineChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

/**
 * Data point for trust history
 */
export interface TrustHistoryDataPoint {
  /** Date/timestamp */
  date: string | Date;
  /** Trust score at this point */
  score: number;
  /** Optional: reliability component */
  reliability?: number;
  /** Optional: determinism component */
  determinism?: number;
  /** Optional: community component */
  community?: number;
}

/**
 * TrustHistory Component Props
 */
export interface TrustHistoryProps {
  /** Historical trust score data */
  data: TrustHistoryDataPoint[];
  /** Chart title */
  title?: string;
  /** Display variant */
  variant?: 'line' | 'area' | 'composed';
  /** Show components breakdown */
  showComponents?: boolean;
  /** Show trend line */
  showTrend?: boolean;
  /** Height of the chart */
  height?: number;
  /** Additional CSS classes */
  className?: string;
  /** Loading state */
  loading?: boolean;
}

/**
 * Format date for axis display
 */
function formatDate(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

/**
 * Custom tooltip for the chart
 */
function CustomTooltip({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: Array<{ value: number; dataKey: string; color: string }>;
  label?: string;
}) {
  if (!active || !payload?.length) return null;

  return (
    <div className="bg-background/95 backdrop-blur-sm border border-border rounded-lg p-3 shadow-lg">
      <p className="text-xs text-muted-foreground mb-2">{label}</p>
      {payload.map((entry, index) => (
        <div key={index} className="flex items-center gap-2 text-sm">
          <div className="w-2 h-2 rounded-full" style={{ backgroundColor: entry.color }} />
          <span className="text-muted-foreground capitalize">{entry.dataKey}:</span>
          <span className="font-semibold">{entry.value.toFixed(1)}%</span>
        </div>
      ))}
    </div>
  );
}

/**
 * TrustHistory Component
 *
 * Displays trust score over time using recharts.
 * Supports line chart, area chart, and composed chart (with component breakdown).
 *
 * @example
 * // Simple line chart
 * <TrustHistory data={trustData} title="Trust Score Over Time" />
 *
 * // Area chart with components
 * <TrustHistory
 *   data={trustData}
 *   variant="area"
 *   showComponents
 *   height={300}
 * />
 *
 * // Composed chart with trend
 * <TrustHistory data={trustData} variant="composed" showTrend />
 */
export function TrustHistory({
  data,
  title = 'Trust Score History',
  variant = 'line',
  showComponents = false,
  showTrend = false,
  height = 240,
  className,
  loading = false,
}: TrustHistoryProps) {
  // Process data for chart
  const chartData = useMemo(() => {
    return data.map((point) => ({
      ...point,
      date: formatDate(point.date),
      fullDate: point.date,
    }));
  }, [data]);

  // Calculate average score
  const averageScore = useMemo(() => {
    if (!data.length) return 0;
    return data.reduce((sum, point) => sum + point.score, 0) / data.length;
  }, [data]);

  // Calculate trend
  const trend = useMemo(() => {
    if (data.length < 2) return 0;
    const first = data[0].score;
    const last = data[data.length - 1].score;
    return last - first;
  }, [data]);

  // Get color based on current score
  const currentScore = data.length > 0 ? data[data.length - 1].score : 0;
  const band = getTrustScoreBand(currentScore);
  const colorConfig = getTrustColorConfig(band);

  if (loading) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader className="pb-2">
          <div className="h-5 w-32 bg-muted rounded animate-pulse" />
        </CardHeader>
        <CardContent>
          <div className="w-full bg-muted/30 rounded animate-pulse" style={{ height }} />
        </CardContent>
      </Card>
    );
  }

  if (!data.length) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center" style={{ height }}>
          <div className="text-center text-muted-foreground">
            <Shield className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">No trust history available</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const TrendIcon = trend > 0 ? TrendingUp : trend < 0 ? TrendingDown : AlertTriangle;
  const trendColorClass =
    trend > 0 ? 'text-emerald-400' : trend < 0 ? 'text-red-400' : 'text-gray-400';

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
          <div className="flex items-center gap-3 text-xs">
            {showTrend && (
              <div className={cn('flex items-center gap-1', trendColorClass)}>
                <TrendIcon className="h-3.5 w-3.5" />
                <span>
                  {trend > 0 ? '+' : ''}
                  {trend.toFixed(1)}%
                </span>
              </div>
            )}
            <span className="text-muted-foreground">Avg: {averageScore.toFixed(1)}%</span>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={height}>
          {variant === 'area' ? (
            <ComposedChart data={chartData} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
              <defs>
                <linearGradient id="trustGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={colorConfig.primary} stopOpacity={0.3} />
                  <stop offset="95%" stopColor={colorConfig.primary} stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted/30" />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 10, fill: 'currentColor' }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                domain={[0, 100]}
                tick={{ fontSize: 10, fill: 'currentColor' }}
                tickLine={false}
                axisLine={false}
              />
              <Tooltip content={<CustomTooltip />} />
              <ReferenceLine
                y={averageScore}
                stroke={colorConfig.primary}
                strokeDasharray="5 5"
                strokeOpacity={0.5}
              />
              <Area
                type="monotone"
                dataKey="score"
                stroke={colorConfig.primary}
                fill="url(#trustGradient)"
                strokeWidth={2}
              />
            </ComposedChart>
          ) : variant === 'composed' ? (
            <ComposedChart data={chartData} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted/30" />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 10, fill: 'currentColor' }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                domain={[0, 100]}
                tick={{ fontSize: 10, fill: 'currentColor' }}
                tickLine={false}
                axisLine={false}
              />
              <Tooltip content={<CustomTooltip />} />
              <ReferenceLine
                y={averageScore}
                stroke={colorConfig.primary}
                strokeDasharray="5 5"
                strokeOpacity={0.5}
              />
              <Line
                type="monotone"
                dataKey="score"
                stroke={colorConfig.primary}
                strokeWidth={2}
                dot={{ r: 3, fill: colorConfig.primary }}
              />
              {showComponents && (
                <>
                  <Line
                    type="monotone"
                    dataKey="reliability"
                    stroke="#3b82f6"
                    strokeWidth={1.5}
                    dot={{ r: 2 }}
                  />
                  <Line
                    type="monotone"
                    dataKey="determinism"
                    stroke="#8b5cf6"
                    strokeWidth={1.5}
                    dot={{ r: 2 }}
                  />
                  <Line
                    type="monotone"
                    dataKey="community"
                    stroke="#f59e0b"
                    strokeWidth={1.5}
                    dot={{ r: 2 }}
                  />
                </>
              )}
            </ComposedChart>
          ) : (
            <LineChart data={chartData} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted/30" />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 10, fill: 'currentColor' }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                domain={[0, 100]}
                tick={{ fontSize: 10, fill: 'currentColor' }}
                tickLine={false}
                axisLine={false}
              />
              <Tooltip content={<CustomTooltip />} />
              <ReferenceLine
                y={averageScore}
                stroke={colorConfig.primary}
                strokeDasharray="5 5"
                strokeOpacity={0.5}
              />
              <Line
                type="monotone"
                dataKey="score"
                stroke={colorConfig.primary}
                strokeWidth={2}
                dot={{ r: 3, fill: colorConfig.primary }}
                activeDot={{ r: 5 }}
              />
            </LineChart>
          )}
        </ResponsiveContainer>

        {showComponents && (
          <div className="flex items-center justify-center gap-4 mt-2 pt-2 border-t border-border/50">
            <div className="flex items-center gap-1.5 text-[10px]">
              <div className="w-2 h-2 rounded-full bg-blue-500" />
              <span className="text-muted-foreground">Reliability</span>
            </div>
            <div className="flex items-center gap-1.5 text-[10px]">
              <div className="w-2 h-2 rounded-full bg-violet-500" />
              <span className="text-muted-foreground">Determinism</span>
            </div>
            <div className="flex items-center gap-1.5 text-[10px]">
              <div className="w-2 h-2 rounded-full bg-amber-500" />
              <span className="text-muted-foreground">Community</span>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

/**
 * TrustHistorySkeleton Component
 * Loading placeholder for TrustHistory
 */
export function TrustHistorySkeleton({
  height = 240,
  className,
}: {
  height?: number;
  className?: string;
}) {
  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-2">
        <div className="h-5 w-40 bg-muted rounded animate-pulse" />
      </CardHeader>
      <CardContent>
        <div className="w-full bg-muted/30 rounded animate-pulse" style={{ height }} />
      </CardContent>
    </Card>
  );
}

export default TrustHistory;
