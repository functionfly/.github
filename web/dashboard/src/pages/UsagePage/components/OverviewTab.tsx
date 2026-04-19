import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Loader2, DollarSign, Activity, Zap } from 'lucide-react';
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
        <Card className="border-theme bg-card card-brand-accent">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Zap className="h-4 w-4 text-text-brand" />
              <span className="v-text-gradient-flame">Usage vs plan limit</span>
            </CardTitle>
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
                <div className="relative">
                  <Progress
                    value={usagePercent}
                    className={`h-2 ${isOverLimit ? '[&>div]:bg-destructive' : '[&>div]:bg-gradient-to-r [&>div]:from-ff-flame [&>div]:to-ff-afterburner'}`}
                  />
                  <div className="absolute -right-1 -top-1 w-3 h-3 rounded-full bg-ff-flame shadow-glow-sm animate-pulse" />
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {usageLoading ? (
          <Card className="border-theme bg-card v-top-border-brand h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-brand" />
          </Card>
        ) : (
          <div className="v-top-border-brand">
            <UsageGraph
              data={usageGraphData}
              title={`Requests (last ${USAGE_DAYS} days)`}
              valueLabel="Requests"
            />
          </div>
        )}
        {executionRateLoading ? (
          <Card className="border-theme bg-card v-top-border-brand h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-brand" />
          </Card>
        ) : (
          <div className="v-top-border-brand">
            <ExecutionRateChart
              data={executionRateData}
              title="Execution rate (last 7 days)"
              unit="exec/s"
            />
          </div>
        )}
      </div>

      {/* Cost Trend */}
      <Card className="border-theme bg-card v-top-border-brand">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <div className="v-icon-brand w-8 h-8">
              <DollarSign className="h-4 w-4" />
            </div>
            <span className="v-text-brand">Daily Cost Trend</span>
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
                      <stop offset="5%" stopColor="#FF6B35" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#FF6B35" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="colorCostStroke" x1="0" y1="0" x2="1" y2="0">
                      <stop offset="0%" stopColor="#FF6B35" />
                      <stop offset="50%" stopColor="#FF4F5E" />
                      <stop offset="100%" stopColor="#00D4FF" />
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
                    stroke="url(#colorCostStroke)"
                    strokeWidth={3}
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
      <Card className="border-theme bg-card v-top-border-brand">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <div className="v-icon-brand w-8 h-8">
              <Activity className="h-4 w-4" />
            </div>
            <span className="v-text-brand">Execution Velocity</span>
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
                    stroke="#FF6B35"
                    strokeWidth={3}
                    dot={false}
                    activeDot={{ r: 5, fill: '#FF6B35', stroke: '#fff', strokeWidth: 2 }}
                  />
                  <Line
                    type="monotone"
                    dataKey="value"
                    name="7-day Trend"
                    stroke="#00D4FF"
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
