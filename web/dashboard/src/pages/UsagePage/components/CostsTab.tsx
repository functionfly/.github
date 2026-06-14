import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, PieChart as PieChartIcon, Globe, BarChart3 } from 'lucide-react';
import {
  ResponsiveContainer,
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  Tooltip,
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Legend,
} from 'recharts';
import { formatCostUsd } from '@/api/usageAnalytics';
import { COLORS, REGION_COLORS } from '../constants';

interface CostsTabProps {
  // Loading states
  costSummaryLoading: boolean;
  functionCostsLoading: boolean;
  periodLoading: boolean;
  regionLoading: boolean;

  // Data
  costSummary: {
    total_cost_usd: number;
    previous_period_cost_usd?: number;
    cost_trend_percent?: number;
    total_executions: number;
    cost_breakdown?: {
      execution: number;
      compute: number;
      platform_fee: number;
      data_transfer: number;
    };
  } | null | undefined;

  costBreakdownData: Array<{ name: string; value: number; color: string }>;
  regionChartData: Array<{ name: string; executions: number; cost: number }>;
  functionChartData: Array<{ name: string; executions: number; cost: number; successRate: number }>;
  dailyChartData: Array<{ date: string; fullDate: string; executions: number; cost: number }>;
}

export function CostsTab({
  costSummaryLoading,
  functionCostsLoading,
  periodLoading,
  costSummary,
  costBreakdownData,
  regionChartData,
  functionChartData,
  dailyChartData,
}: CostsTabProps) {
  return (
    <div className="space-y-4">
      {/* Pie Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Cost Breakdown */}
        <Card className="border-theme">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <PieChartIcon className="h-4 w-4" />
              Cost by Category
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[240px]">
              {costBreakdownData.length === 0 ? (
                <div className="flex h-full items-center justify-center text-text-muted text-sm">
                  No cost data
                </div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <RechartsPieChart>
                    <Pie
                      data={costBreakdownData}
                      cx="50%"
                      cy="50%"
                      innerRadius={56}
                      outerRadius={80}
                      paddingAngle={2}
                      dataKey="value"
                      nameKey="name"
                      label={({ name, percent }) =>
                        `${name} ${(percent * 100).toFixed(0)}%`
                      }
                    >
                      {costBreakdownData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value: number) => [formatCostUsd(value), 'Cost']}
                    />
                  </RechartsPieChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Region Breakdown */}
        <Card className="border-theme">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Globe className="h-4 w-4" />
              Cost by Region
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[240px]">
              {regionChartData.length === 0 ? (
                <div className="flex h-full items-center justify-center text-text-muted text-sm">
                  No region data
                </div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <RechartsPieChart>
                    <Pie
                      data={regionChartData.slice(0, 5)}
                      cx="50%"
                      cy="50%"
                      innerRadius={56}
                      outerRadius={80}
                      paddingAngle={2}
                      dataKey="cost"
                      nameKey="name"
                      label={({ name }) => name}
                    >
                      {regionChartData.map((_, index) => (
                        <Cell
                          key={`cell-${index}`}
                          fill={REGION_COLORS[index % REGION_COLORS.length]}
                        />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value: number) => [formatCostUsd(value), 'Cost']}
                    />
                  </RechartsPieChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Period Comparison */}
        <Card className="border-theme">
          <CardHeader>
            <CardTitle className="text-base">Period Comparison</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Current Period</span>
                <span className="font-medium">
                  {formatCostUsd(costSummary?.total_cost_usd ?? 0)}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Previous Period</span>
                <span className="font-medium">
                  {formatCostUsd(costSummary?.previous_period_cost_usd ?? 0)}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Change</span>
                <span
                  className={`font-medium ${
                    (costSummary?.cost_trend_percent ?? 0) > 0
                      ? 'text-red-500'
                      : (costSummary?.cost_trend_percent ?? 0) < 0
                        ? 'text-emerald-500'
                        : ''
                  }`}
                >
                  {(costSummary?.cost_trend_percent ?? 0) > 0 ? '+' : ''}
                  {(costSummary?.cost_trend_percent ?? 0).toFixed(1)}%
                </span>
              </div>
              <div className="flex justify-between items-center pt-2 border-t">
                <span className="text-sm text-text-secondary">Avg Cost/Execution</span>
                <span className="font-medium">
                  {costSummary?.total_cost_usd && costSummary.total_executions > 0
                    ? formatCostUsd(costSummary.total_cost_usd / costSummary.total_executions)
                    : '$0.00'}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Cost Breakdown Over Time */}
      <Card className="border-theme">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Cost Breakdown Over Time
          </CardTitle>
          <CardDescription>Stacked view of cost categories by day</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-[280px]">
            {periodLoading || dailyChartData.length === 0 ? (
              <div className="flex h-full items-center justify-center text-text-muted text-sm">
                {periodLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  'No cost breakdown data available'
                )}
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <RechartsBarChart
                  data={dailyChartData}
                  margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
                >
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
                  <Legend wrapperStyle={{ paddingTop: '10px' }} />
                  <Bar
                    dataKey="cost"
                    name="Total Cost"
                    stackId="a"
                    fill={COLORS.execution}
                    radius={[4, 4, 0, 0]}
                  />
                  <Bar dataKey="cost" name="Compute" stackId="b" fill={COLORS.compute} />
                  <Bar
                    dataKey="cost"
                    name="Data Transfer"
                    stackId="b"
                    fill={COLORS.data_transfer}
                  />
                </RechartsBarChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Top Functions */}
      <Card className="border-theme">
        <CardHeader>
          <CardTitle className="text-base">Top Functions by Cost</CardTitle>
          <CardDescription>Most expensive functions in the selected period</CardDescription>
        </CardHeader>
        <CardContent>
          {functionCostsLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            </div>
          ) : functionChartData.length === 0 ? (
            <div className="text-center py-8 text-text-muted">No function data available</div>
          ) : (
            <div className="space-y-2">
              {functionChartData.map((fn, idx) => (
                <div
                  key={fn.name}
                  className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-sm text-text-muted w-6">#{idx + 1}</span>
                    <div>
                      <p className="font-medium text-sm">{fn.name}</p>
                      <p className="text-xs text-text-muted">
                        {fn.executions.toLocaleString()} executions
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-medium">{formatCostUsd(fn.cost)}</p>
                    <p
                      className={`text-xs ${
                        fn.successRate >= 95
                          ? 'text-emerald-500'
                          : fn.successRate >= 90
                            ? 'text-amber-500'
                            : 'text-red-500'
                      }`}
                    >
                      {fn.successRate.toFixed(1)}% success
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
