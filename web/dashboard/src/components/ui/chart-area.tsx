'use client';

import * as React from 'react';
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  ReferenceLine,
  ReferenceArea,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface AreaChartData {
  [key: string]: string | number;
}

export interface AreaSeries {
  key: string;
  name: string;
  color?: string;
  gradient?: { from: string; to: string };
  fillOpacity?: number;
  strokeWidth?: number;
  type?: 'monotone' | 'linear' | 'step' | 'natural';
}

export interface AreaChartProps {
  data: AreaChartData[];
  series: AreaSeries[];
  title?: string;
  description?: string;
  xAxisKey?: string;
  height?: number;
  showGrid?: boolean;
  showLegend?: boolean;
  showDots?: boolean;
  stacked?: boolean;
  className?: string;
  tooltipFormatter?: (value: any, name: string) => [string, string];
  yAxisFormatter?: (value: any) => string;
  xAxisFormatter?: (value: any) => string;
  referenceLines?: Array<{ y: number; label: string; color: string }>;
  referenceAreas?: Array<{ x1: string | number; x2: string | number; label: string; color: string }>;
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

const AreaChartComponent = React.forwardRef<HTMLDivElement, AreaChartProps>(
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
      showDots = false,
      stacked = false,
      className,
      tooltipFormatter,
      yAxisFormatter,
      xAxisFormatter,
      referenceLines,
      referenceAreas,
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

    const getGradientId = (key: string) => `area-gradient-${key}`;

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
            <RechartsAreaChart
              data={data}
              margin={{ top: 10, right: 30, left: 20, bottom: 5 }}
              syncId={syncId}
            >
              <defs>
                {series.map((s, index) => {
                  const color = s.gradient?.from || s.color || defaultColors[index % defaultColors.length];
                  const toColor = s.gradient?.to || color;
                  return (
                    <linearGradient key={s.key} id={getGradientId(s.key)} x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor={color} stopOpacity={s.fillOpacity || 0.3} />
                      <stop offset="95%" stopColor={toColor} stopOpacity={0} />
                    </linearGradient>
                  );
                })}
              </defs>
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
              {showLegend && <Legend wrapperStyle={{ paddingTop: 10 }} />}
              {referenceLines?.map((line, index) => (
                <ReferenceLine
                  key={index}
                  y={line.y}
                  label={line.label}
                  stroke={line.color}
                  strokeDasharray="5 5"
                />
              ))}
              {referenceAreas?.map((area, index) => (
                <ReferenceArea
                  key={index}
                  x1={area.x1}
                  x2={area.x2}
                  label={area.label}
                  fill={area.color}
                  opacity={0.1}
                />
              ))}
              {series.map((s, index) => {
                const color = s.color || defaultColors[index % defaultColors.length];
                return (
                  <Area
                    key={s.key}
                    type={s.type || 'monotone'}
                    dataKey={s.key}
                    name={s.name}
                    stroke={color}
                    fill={`url(#${getGradientId(s.key)})`}
                    strokeWidth={s.strokeWidth || 2}
                    dot={showDots ? { fill: color, strokeWidth: 0, r: 4 } : false}
                    activeDot={showDots ? { r: 6, stroke: color, strokeWidth: 2, fill: 'var(--bg-primary)' } : false}
                    stackId={stacked ? 'stack' : undefined}
                    isAnimationActive={false}
                  />
                );
              })}
            </RechartsAreaChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    );
  }
);
AreaChartComponent.displayName = 'AreaChart';

/**
 * SparkAreaChart - Mini area chart for inline use
 */
interface SparkAreaChartProps {
  data: number[];
  color?: string;
  height?: number;
  width?: number;
  showDots?: boolean;
}

const SparkAreaChart = React.forwardRef<SVGSVGElement, SparkAreaChartProps>(
  ({ data, color = '#6366f1', height = 60, width = 200, showDots = false }, ref) => {
    const chartData = data.map((value, index) => ({ name: index, value }));

    return (
      <svg ref={ref} width={width} height={height} viewBox={`0 0 ${width} ${height}`}>
        <defs>
          <linearGradient id="spark-gradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={color} stopOpacity={0.3} />
            <stop offset="95%" stopColor={color} stopOpacity={0} />
          </linearGradient>
        </defs>
        <ResponsiveContainer width={width} height={height}>
          <RechartsAreaChart data={chartData} margin={{ top: 5, right: 0, left: 0, bottom: 0 }}>
            <Area
              type="monotone"
              dataKey="value"
              stroke={color}
              fill="url(#spark-gradient)"
              strokeWidth={2}
              dot={showDots}
              isAnimationActive={false}
            />
          </RechartsAreaChart>
        </ResponsiveContainer>
      </svg>
    );
  }
);
SparkAreaChart.displayName = 'SparkAreaChart';

export { AreaChartComponent as AreaChart, SparkAreaChart };
