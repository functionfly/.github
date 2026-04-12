'use client';

import * as React from 'react';
import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface PieChartData {
  name: string;
  value: number;
  color?: string;
}

export interface PieChartProps {
  data: PieChartData[];
  title?: string;
  description?: string;
  height?: number;
  donut?: boolean;
  donutInnerRadius?: number;
  showLegend?: boolean;
  showLabels?: boolean;
  className?: string;
  tooltipFormatter?: (value: number, name: string) => [string, string];
  labelFormatter?: (entry: any) => string;
  interactive?: boolean;
  colors?: string[];
}

const defaultColors = [
  '#6366f1', // indigo-500
  '#8b5cf6', // violet-500
  '#10b981', // emerald-500
  '#f59e0b', // amber-500
  '#ef4444', // red-500
  '#06b6d4', // cyan-500
  '#ec4899', // pink-500
  '#64748b', // slate-500
  '#84cc16', // lime-500
  '#14b8a6', // teal-500
];

const CustomTooltip = ({ active, payload, formatter }: any) => {
  if (active && payload && payload.length) {
    const data = payload[0];
    const [formattedValue, formattedName] = formatter
      ? formatter(data.value, data.name)
      : [data.value, data.name];

    return (
      <div className="rounded-lg border border-border-default bg-card p-3 shadow-lg">
        <div className="flex items-center gap-2">
          <div
            className="h-3 w-3 rounded-full"
            style={{ backgroundColor: data.payload.color || data.color }}
          />
          <span className="font-medium text-text-primary">{formattedName}</span>
        </div>
        <div className="mt-1 text-sm text-text-secondary">{formattedValue}</div>
      </div>
    );
  }
  return null;
};



const PieChartComponent = React.forwardRef<HTMLDivElement, PieChartProps>(
  (
    {
      data,
      title,
      description,
      height = 300,
      donut = false,
      donutInnerRadius = 60,
      showLegend = true,
      showLabels = false,
      className,
      tooltipFormatter,
      labelFormatter,
      colors = defaultColors,
    },
    ref
  ) => {
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
            <RechartsPieChart>
              <Pie
                data={data}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={showLabels ? labelFormatter || (({ name }) => name) : undefined}
                outerRadius={donut ? 80 : 100}
                innerRadius={donut ? donutInnerRadius : 0}
                fill="#8884d8"
                dataKey="value"
                isAnimationActive={false}
              >
                {data.map((entry, index) => (
                  <Cell
                    key={`cell-${index}`}
                    fill={entry.color || colors[index % colors.length]}
                    stroke="var(--card)"
                    strokeWidth={2}
                  />
                ))}
              </Pie>
              <Tooltip content={<CustomTooltip formatter={tooltipFormatter} />} />
              {showLegend && (
                <Legend
                  verticalAlign="bottom"
                  height={36}
                  wrapperStyle={{ paddingTop: 20 }}
                />
              )}
            </RechartsPieChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    );
  }
);
PieChartComponent.displayName = 'PieChart';

export { PieChartComponent as PieChart };
export type { PieChartData as PieChartDataType };
