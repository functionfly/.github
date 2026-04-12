import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Loader2, DollarSign, Activity } from 'lucide-react';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  LineChart,
  Line,
  Legend,
} from 'recharts';
import { UsageGraph, ExecutionRateChart } from '@/components/dashboard';
import { formatCostUsd } from '@/api/usageAnalytics';
import { COLORS, USAGE_DAYS } from '../constants';
import type { UsageLimits } from '../types';

interface OverviewTabProps {
  // Loading states
  usageLoading: boolean;
  executionRateLoading: boolean;
  periodLoading: boolean;

  // Data
  totalUsage: number;
  usageGraphData: Array<{ time: string; value: number }>;
  executionRateData: Array<{ time: string; rate: number }>;
  dailyChartData: Array<{ date: string; fullDate: string; executions: number; cost: number }>;

  // Limits
  limits: UsageLimits;
}

export function OverviewTab({
  usageLoading,
  executionRateLoading,
  periodLoading,
  totalUsage,
  usageGraphData,
  executionRateData,
  dailyChartData,
  limits,
}: OverviewTabProps) {
  const { isUnlimited, requestLimit, usagePercent, isOverLimit } = limits;

  return (
    <div className="space-y-4">
      {/* Usage Progress */}
      {!isUnlimited && requestLimit > 0 && (
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base">Usage vs plan limit</CardTitle>
            <CardDescription>
              Monthly request allowance. Overage may incur charges depending on your plan.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {usageLoading ? (
              <div className="h-4 w-full rounded-full bg-bg-hover animate-pulse" />
            ) : (
              <div className="space-y-2">
                <div className="flex justify-between text-sm text-text-secondary">
                  <span>
                    {totalUsage.toLocaleString()} / {requestLimit.toLocaleString()} requests
                  </span>
                  <span>{usagePercent.toFixed(0)}%</span>
                </div>
                <Progress
                  value={usagePercent}
                  className={isOverLimit ? '[&>div]:bg-destructive' : undefined}
                />
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {usageLoading ? (
          <Card className="border-theme bg-card h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
          </Card>
        ) : (
          <UsageGraph
            data={usageGraphData}
            title={`Requests (last ${USAGE_DAYS} days)`}
            valueLabel="Requests"
          />
        )}
        {executionRateLoading ? (
          <Card className="border-theme bg-card h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
          </Card>
        ) : (
          <ExecutionRateChart
            data={executionRateData}
            title="Execution rate (last 7 days)"
            unit="exec/s"
          />
        )}
      </div>

      {/* Cost Trend */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <DollarSign className="h-4 w-4" />
            Daily Cost Trend
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[240px]">
            {periodLoading || dailyChartData.length === 0 ? (
              <div className="flex h-full items-center justify-center text-text-muted text-sm">
                {periodLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  'No cost data available'
                )}
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart
                  data={dailyChartData}
                  margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
                >
                  <defs>
                    <linearGradient id="colorCost" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor={COLORS.execution} stopOpacity={0.3} />
                      <stop offset="95%" stopColor={COLORS.execution} stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid
                    strokeDasharray="3 3"
                    stroke="var(--color-border-subtle)"
                  />
                  <XAxis
                    dataKey="date"
                    tick={{ fill: 'var(--color-text-muted)', fontSize: 11 }}
                  />
                  <YAxis
                    tick={{ fill: 'var(--color-text-muted)', fontSize: 11 }}
                    tickFormatter={(v) => `$${v}`}
                  />
                  <Tooltip
                    formatter={(value: number) => [formatCostUsd(value), 'Cost']}
                    contentStyle={{
                      backgroundColor: 'var(--color-bg-tertiary)',
                      border: '1px solid var(--color-border-default)',
                      borderRadius: '8px',
                    }}
                  />
                  <Area
                    type="monotone"
                    dataKey="cost"
                    stroke={COLORS.execution}
                    fillOpacity={1}
                    fill="url(#colorCost)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Execution Velocity */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Execution Velocity
          </CardTitle>
          <CardDescription>
            Daily execution volume with 7-day moving average
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-[240px]">
            {usageLoading || usageGraphData.length === 0 ? (
              <div className="flex h-full items-center justify-center text-text-muted text-sm">
                {usageLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  'No execution data available'
                )}
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart
                  data={usageGraphData}
                  margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
                >
                  <CartesianGrid
                    strokeDasharray="3 3"
                    stroke="var(--color-border-subtle)"
                  />
                  <XAxis
                    dataKey="time"
                    tick={{ fill: 'var(--color-text-muted)', fontSize: 11 }}
                  />
                  <YAxis
                    tick={{ fill: 'var(--color-text-muted)', fontSize: 11 }}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--color-bg-tertiary)',
                      border: '1px solid var(--color-border-default)',
                      borderRadius: '8px',
                    }}
                  />
                  <Legend wrapperStyle={{ paddingTop: '10px' }} />
                  <Line
                    type="monotone"
                    dataKey="value"
                    name="Executions"
                    stroke={COLORS.execution}
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4 }}
                  />
                  <Line
                    type="monotone"
                    dataKey="value"
                    name="7-day Trend"
                    stroke={COLORS.success}
                    strokeWidth={2}
                    strokeDasharray="5 5"
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
