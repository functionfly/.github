/**
 * VaultUsageChart - Visual representation of secret access patterns over time
 *
 * Renders an area chart showing secret access frequency and patterns,
 * helping identify usage trends and potential anomalies.
 *
 * @example
 * ```tsx
 * // Basic usage with data
 * <VaultUsageChart
 *   data={[
 *     { date: "2024-03-01", accesses: 150 },
 *     { date: "2024-03-02", accesses: 230 },
 *     { date: "2024-03-03", accesses: 180 },
 *   ]}
 * />
 *
 * // With custom height and title
 * <VaultUsageChart
 *   data={usageData}
 *   title="Daily Access Patterns"
 *   height={300}
 * />
 *
 * // Loading state
 * <VaultUsageChart isLoading />
 *
 * // Empty state (no data)
 * <VaultUsageChart data={[]} />
 * ```
 */

import { useMemo } from "react";
import { Activity, TrendingUp, TrendingDown, Minus } from "lucide-react";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";

// Visx imports for chart rendering
import {
  XYChart,
  AreaSeries,
  Axis,
  Tooltip,
  buildChartTheme,
} from "@visx/xychart";
import { curveMonotoneX } from "@visx/curve";

/** Single data point for the usage chart */
export interface UsageDataPoint {
  /** Date string in ISO format or date-only format */
  date: string;
  /** Number of access events for that date */
  accesses: number;
}

/** Trend direction for the summary indicator */
export type TrendDirection = "up" | "down" | "neutral";

export interface VaultUsageChartProps {
  /** Array of usage data points */
  data?: UsageDataPoint[];
  /** Chart title displayed in the header */
  title?: string;
  /** Whether the component is in loading state */
  isLoading?: boolean;
  /** Chart height in pixels */
  height?: number;
  /** Show trend indicator comparing first and last data points */
  showTrend?: boolean;
  /** Additional CSS classes */
  className?: string;
}

/** Default chart theme with CSS variable colors */
const chartTheme = buildChartTheme({
  backgroundColor: "transparent",
  colors: ["var(--color-brand-500)"],
  gridColor: "var(--border-subtle)",
  gridColorDark: "var(--border-subtle)",
  svgLabelBig: { fill: "var(--color-text-primary)" },
  svgLabelSmall: { fill: "var(--color-text-secondary)" },
  tickLength: 4,
});

/**
 * Calculate trend direction from data
 */
function calculateTrend(data: UsageDataPoint[]): TrendDirection {
  if (data.length < 2) return "neutral";
  const first = data[0].accesses;
  const last = data[data.length - 1].accesses;
  const diff = last - first;
  if (diff > first * 0.1) return "up";
  if (diff < -first * 0.1) return "down";
  return "neutral";
}

/**
 * Format date for tooltip display
 */
function formatTooltipDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

/**
 * Skeleton loader for the chart
 */
function VaultUsageChartSkeleton({ className }: { className?: string }) {
  return (
    <Card className={cn("overflow-hidden", className)}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <Skeleton className="h-5 w-40" />
          <Skeleton className="h-5 w-20" />
        </div>
      </CardHeader>
      <CardContent>
        <Skeleton className="w-full" style={{ height: 200 }} />
      </CardContent>
    </Card>
  );
}

/**
 * Empty state when no data is available
 */
function VaultUsageChartEmpty() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <Activity className="h-10 w-10 text-(--color-text-muted) mb-3" />
      <h4 className="text-sm font-medium text-(--color-text-secondary)">
        No Usage Data
      </h4>
      <p className="text-xs text-(--color-text-muted) mt-1">
        Secret access patterns will appear here once data is available.
      </p>
    </div>
  );
}

/**
 * Trend indicator component
 */
function TrendIndicator({
  direction,
  value,
}: {
  direction: TrendDirection;
  value: number;
}) {
  const config = {
    up: { icon: TrendingUp, color: "text-green-500", bg: "bg-green-500/10" },
    down: { icon: TrendingDown, color: "text-red-500", bg: "bg-red-500/10" },
    neutral: { icon: Minus, color: "text-(--color-text-muted)", bg: "bg-(--color-bg-tertiary)" },
  };

  const { icon: Icon, color, bg } = config[direction];
  const sign = direction === "up" ? "+" : direction === "down" ? "" : "";

  return (
    <Badge
      variant="outline"
      className={cn("gap-1 font-medium", bg, color)}
    >
      <Icon className="h-3 w-3" />
      {sign}
      {value}%
    </Badge>
  );
}

/**
 * VaultUsageChart component
 *
 * Renders an area chart visualizing secret access patterns over time
 * using Visx for rendering.
 */
export function VaultUsageChart({
  data = [],
  title = "Secret Access Patterns",
  isLoading = false,
  height = 240,
  showTrend = true,
  className,
}: VaultUsageChartProps) {
  const trend = useMemo(() => calculateTrend(data), [data]);

  // Calculate percentage change for trend display
  const trendValue = useMemo(() => {
    if (data.length < 2) return 0;
    const first = data[0].accesses;
    const last = data[data.length - 1].accesses;
    if (first === 0) return 0;
    return Math.round(((last - first) / first) * 100);
  }, [data]);

  // Calculate average accesses
  const averageAccesses = useMemo(() => {
    if (data.length === 0) return 0;
    const sum = data.reduce((acc, d) => acc + d.accesses, 0);
    return Math.round(sum / data.length);
  }, [data]);

  if (isLoading) {
    return <VaultUsageChartSkeleton className={className} />;
  }

  if (data.length === 0) {
    return (
      <Card className={cn("overflow-hidden", className)}>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium text-(--color-text-secondary)">
            {title}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <VaultUsageChartEmpty />
        </CardContent>
      </Card>
    );
  }

  // Transform data for Visx
  const chartData = data.map((d) => ({
    x: d.date,
    y: d.accesses,
  }));

  return (
    <Card
      className={cn(
        "overflow-hidden transition-all duration-200",
        "border border-(--border-subtle)",
        "bg-(--color-bg-primary)",
        className
      )}
    >
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className={cn(
                "flex h-8 w-8 items-center justify-center rounded-lg",
                "bg-(--color-brand-500)/10"
              )}
            >
              <Activity className="h-4 w-4 text-(--color-brand-500)" />
            </div>
            <div>
              <CardTitle className="text-sm font-medium text-(--color-text-primary)">
                {title}
              </CardTitle>
              <p className="text-xs text-(--color-text-muted)">
                Avg: {averageAccesses.toLocaleString()} accesses/day
              </p>
            </div>
          </div>
          {showTrend && data.length >= 2 && (
            <TrendIndicator direction={trend} value={Math.abs(trendValue)} />
          )}
        </div>
      </CardHeader>
      <CardContent>
        <div style={{ height }}>
          <XYChart
            theme={chartTheme}
            xScale={{ type: "band" }}
            yScale={{ type: "linear" }}
            height={height}
            margin={{ top: 10, right: 10, bottom: 30, left: 40 }}
          >
            <AreaSeries
              dataKey="accesses"
              data={chartData}
              xAccessor={(d) => d.x}
              yAccessor={(d) => d.y}
              curve={curveMonotoneX}
              fillOpacity={0.3}
              strokeWidth={2}
            />
            <Axis
              orientation="bottom"
              hideTicks
              tickFormat={(date: string) => {
                const d = new Date(date);
                return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
              }}
            />
            <Axis orientation="left" hideTicks />
            <Tooltip
              showSeriesGlyphs
              renderTooltip={({ tooltipData }) => {
                const datum = tooltipData?.nearestDatum?.datum as typeof chartData[0] | undefined;
                if (!datum) return null;
                return (
                  <div className="rounded-lg border border-(--border-subtle) bg-(--color-bg-primary) p-2 shadow-lg">
                    <div className="text-xs text-(--color-text-secondary)">
                      {formatTooltipDate(datum.x)}
                    </div>
                    <div className="text-sm font-medium text-(--color-text-primary)">
                      {datum.y.toLocaleString()} accesses
                    </div>
                  </div>
                );
              }}
            />
          </XYChart>
        </div>
      </CardContent>
    </Card>
  );
}

export default VaultUsageChart;
