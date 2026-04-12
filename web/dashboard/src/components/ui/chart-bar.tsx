'use client';

import * as React from 'react';
import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  LabelList,
  Cell,
  ReferenceLine,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface BarChartData {
  [key: string]: string | number;
}

export interface BarSeries {
  key: string;
  name: string;
  color?: string;
  gradient?: { from: string; to: string };
  radius?: number;
}

export interface BarChartProps {
  data: BarChartData[];
  series: BarSeries[];
  title?: string;
  description?: string;
  xAxisKey?: string;
  height?: number;
  showGrid?: boolean;
  showLegend?: boolean;
  showLabels?: boolean;
  horizontal?: boolean;
  stacked?: boolean;
  className?: string;
  tooltipFormatter?: (value: any, name: string) => [string, string];
  yAxisFormatter?: (value: any) => string;
  xAxisFormatter?: (value: any) => string;
  referenceLines?: Array<{ y: number; label: string; color: string }>;
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
                style={{ backgroundColor: entry.color }}
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

const BarChartComponent = React.forwardRef<HTMLDivElement, BarChartProps>(
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
      showLabels = false,
      horizontal = false,
      stacked = false,
      className,
      tooltipFormatter,
      yAxisFormatter,
      xAxisFormatter,
      referenceLines,
    },
    ref
  ) => {
    const defaultColors = [
      'var(--brand-500)',
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
            <RechartsBarChart
              data={data}
              layout={horizontal ? 'vertical' : 'horizontal'}
              margin={{ top: 10, right: 30, left: 20, bottom: 5 }}
            >
              {showGrid && (
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="var(--border-subtle)"
                  opacity={0.3}
                  horizontal={!horizontal}
                  vertical={horizontal}
                />
              )}
              <XAxis
                type={horizontal ? 'number' : 'category'}
                dataKey={horizontal ? undefined : xAxisKey}
                tickFormatter={xAxisFormatter}
                stroke="var(--text-secondary)"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: 'var(--border-subtle)' }}
              />
              <YAxis
                type={horizontal ? 'category' : 'number'}
                dataKey={horizontal ? xAxisKey : undefined}
                tickFormatter={yAxisFormatter}
                stroke="var(--text-secondary)"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: 'var(--border-subtle)' }}
                width={horizontal ? 100 : undefined}
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
              {series.map((s, index) => (
                <Bar
                  key={s.key}
                  dataKey={s.key}
                  name={s.name}
                  fill={s.color || defaultColors[index % defaultColors.length]}
                  radius={[4, 4, 0, 0]}
                  stackId={stacked ? 'stack' : undefined}
                  isAnimationActive={false}
                >
                  {showLabels && (
                    <LabelList
                      position={horizontal ? 'right' : 'top'}
                      formatter={yAxisFormatter}
                    />
                  )}
                </Bar>
              ))}
            </RechartsBarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    );
  }
);
BarChartComponent.displayName = 'BarChart';

/**
 * SimpleBarChart - Simplified single-series bar chart
 */
interface SimpleBarChartProps extends Omit<BarChartProps, 'series'> {
  dataKey: string;
  color?: string;
  name?: string;
}

const SimpleBarChart = React.forwardRef<HTMLDivElement, SimpleBarChartProps>(
  ({ dataKey, color, name = 'Value', ...props }, ref) => {
    return (
      <BarChartComponent
        ref={ref}
        series={[{ key: dataKey, name, color }]}
        {...props}
      />
    );
  }
);
SimpleBarChart.displayName = 'SimpleBarChart';

export { BarChartComponent as BarChart, SimpleBarChart };
