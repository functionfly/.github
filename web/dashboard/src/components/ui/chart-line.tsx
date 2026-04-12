'use client';

import * as React from 'react';
import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  ReferenceLine,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface LineChartData {
  [key: string]: string | number;
}

export interface LineSeries {
  key: string;
  name: string;
  color?: string;
  strokeWidth?: number;
  dotted?: boolean;
  type?: 'monotone' | 'linear' | 'step' | 'natural';
  showDots?: boolean;
  hideInLegend?: boolean;
}

export interface LineChartProps {
  data: LineChartData[];
  series: LineSeries[];
  title?: string;
  description?: string;
  xAxisKey?: string;
  height?: number;
  showGrid?: boolean;
  showLegend?: boolean;
  className?: string;
  tooltipFormatter?: (value: any, name: string) => [string, string];
  yAxisFormatter?: (value: any) => string;
  xAxisFormatter?: (value: any) => string;
  referenceLines?: Array<{ y: number; label: string; color: string }>;
  syncId?: string;
}

const CustomTooltip = ({ active, payload, label, formatter }: any) => {
  if (active && payload && payload.length) {
    return (
      <div className="rounded-lg border border-border-default bg-card p-3 shadow-lg">
        <p className="mb-1 text-sm font-medium text-text-primary">{label}</p>
        {payload.map((entry: any, index: number) => {
          const [formattedValue, formattedName] = formatter
            ? formatter(entry.value, entry.name)
            : [entry.value, entry.name];

          return (
            <div key={index} className="flex items-center gap-2 text-sm">
              <div
                className="h-3 w-3 rounded-full"
                style={{ backgroundColor: entry.stroke || entry.color }}
              />
              <span className="text-text-secondary">{formattedName}:</span>
              <span className="font-medium text-text-primary">{formattedValue}</span>
            </div>
          );
        })}
      </div>
    );
  }
  return null;
};

const LineChartComponent = React.forwardRef<HTMLDivElement, LineChartProps>(
  (
    {
      data,
      series,
      title,
      description,
      xAxisKey = 'name',
      height = 300,
      showGrid = true,
      showLegend = true,
      className,
      tooltipFormatter,
      yAxisFormatter,
      xAxisFormatter,
      referenceLines,
      syncId,
    },
    ref
  ) => {
    const defaultColors = [
      '#6366f1',
      '#8b5cf6',
      '#10b981',
      '#f59e0b',
      '#ef4444',
      '#06b6d4',
      '#ec4899',
    ];

    return (
      <Card ref={ref} className={cn('overflow-hidden', className)}>
        {(title || description) && (
          <CardHeader className="pb-2">
            {title && <CardTitle className="text-lg">{title}</CardTitle>}
            {description && <CardDescription>{description}</CardDescription>}
          </CardHeader>
        )}
        <CardContent>
          <ResponsiveContainer width="100%" height={height}>
            <RechartsLineChart
              data={data}
              margin={{ top: 10, right: 30, left: 20, bottom: 5 }}
              syncId={syncId}
            >
              {showGrid && (
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="var(--border-subtle)"
                  opacity={0.3}
                />
              )}
              <XAxis
                dataKey={xAxisKey}
                tickFormatter={xAxisFormatter}
                stroke="var(--text-secondary)"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: 'var(--border-subtle)' }}
              />
              <YAxis
                tickFormatter={yAxisFormatter}
                stroke="var(--text-secondary)"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: 'var(--border-subtle)' }}
              />
              <Tooltip content={<CustomTooltip formatter={tooltipFormatter} />} />
              {showLegend && (
                <Legend
                  wrapperStyle={{ paddingTop: 10 }}
                  formatter={(value: string) => {
                    const seriesItem = series.find((s) => s.name === value);
                    return seriesItem?.hideInLegend ? '' : value;
                  }}
                />
              )}
              {referenceLines?.map((line, index) => (
                <ReferenceLine
                  key={index}
                  y={line.y}
                  label={line.label}
                  stroke={line.color}
                  strokeDasharray="5 5"
                />
              ))}
              {series.map((s, index) => (
                <Line
                  key={s.key}
                  type={s.type || 'monotone'}
                  dataKey={s.key}
                  name={s.name}
                  stroke={s.color || defaultColors[index % defaultColors.length]}
                  strokeWidth={s.strokeWidth || 2}
                  strokeDasharray={s.dotted ? '5 5' : undefined}
                  dot={s.showDots !== false ? { fill: s.color || defaultColors[index % defaultColors.length], strokeWidth: 0, r: 4 } : false}
                  activeDot={{ r: 6, stroke: s.color || defaultColors[index % defaultColors.length], strokeWidth: 2, fill: 'var(--bg-primary)' }}
                  isAnimationActive={false}
                  hide={s.hideInLegend}
                />
              ))}
            </RechartsLineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    );
  }
);
LineChartComponent.displayName = 'LineChart';

/**
 * Sparkline - Mini line chart for inline use
 */
interface SparklineProps {
  data: number[];
  color?: string;
  height?: number;
  width?: number;
  showArea?: boolean;
}

const Sparkline = React.forwardRef<SVGSVGElement, SparklineProps>(
  ({ data, color = '#6366f1', height = 40, width = 120, showArea = false }, ref) => {
    const chartData = data.map((value, index) => ({ name: index, value }));

    return (
      <svg ref={ref} width={width} height={height} viewBox={`0 0 ${width} ${height}`}>
        {showArea && (
          <defs>
            <linearGradient id="sparkline-gradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.2} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
        )}
        <ResponsiveContainer width={width} height={height}>
          <RechartsLineChart data={chartData} margin={{ top: 2, right: 2, left: 2, bottom: 2 }}>
            <Line
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={2}
              dot={false}
              isAnimationActive={false}
            />
          </RechartsLineChart>
        </ResponsiveContainer>
      </svg>
    );
  }
);
Sparkline.displayName = 'Sparkline';

export { LineChartComponent as LineChart, Sparkline };
