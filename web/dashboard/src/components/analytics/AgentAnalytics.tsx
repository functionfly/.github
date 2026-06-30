'use client';

import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Activity,
  Clock,
  BarChart3,
  PieChart as PieChartIcon,
} from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { AreaChart, PieChart, BarChart } from '@/components/ui/charts';
import { agentApi, type AgentAnalytics as AgentAnalyticsType, type ExecutionRecord, type CostBreakdown, type AgentUsage } from '@/api/agent';

type TimeRange = '24h' | '7d' | '30d';

interface StatCardProps {
  title: string;
  value: string | number;
  change?: number;
  changeLabel?: string;
  icon: React.ElementType;
  trend?: 'up' | 'down' | 'neutral';
}

function StatCard({ title, value, change, changeLabel, icon: Icon, trend = 'neutral' }: StatCardProps) {
  const { t } = useTranslation();
  const IconComponent = Icon as React.ComponentType<{ className?: string }>;
  
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <p className="text-sm font-medium text-text-secondary">{title}</p>
            <p className="text-2xl font-bold mt-1">{value}</p>
            {change !== undefined && (
              <div className="flex items-center gap-1 mt-2">
                {trend === 'up' && <TrendingUp className="h-3 w-3 text-success" />}
                {trend === 'down' && <TrendingDown className="h-3 w-3 text-error" />}
                <span className={`text-xs font-medium ${
                  trend === 'up' ? 'text-success' : trend === 'down' ? 'text-error' : 'text-text-muted'
                }`}>
                  {change > 0 ? '+' : ''}{change.toFixed(1)}%
                </span>
                {changeLabel && (
                  <span className="text-xs text-text-muted">{changeLabel}</span>
                )}
              </div>
            )}
          </div>
          <span className="p-3 rounded-full bg-primary/10">
            <IconComponent className="h-5 w-5 text-brand" />
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

interface StatsGridProps {
  analytics: AgentAnalyticsType;
  usage: AgentUsage | null;
  isLoading: boolean;
}

function StatsGrid({ analytics, usage, isLoading }: StatsGridProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardContent className="p-6">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-8 w-16 mt-2" />
              <Skeleton className="h-3 w-12 mt-2" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  const successRate = analytics.successRate ?? 0;
  const successRateTrend = successRate >= 95 ? 'up' : successRate < 90 ? 'down' : 'neutral';

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        title="Total Executions"
        value={analytics.totalExecutions.toLocaleString()}
        icon={Activity}
      />
      <StatCard
        title="Success Rate"
        value={`${successRate.toFixed(1)}%`}
        trend={successRateTrend}
        change={0}
        changeLabel="vs last period"
        icon={TrendingUp}
      />
      <StatCard
        title="Avg Latency"
        value={`${((analytics.avgLatencyMs ?? 0) / 1000).toFixed(2)}s`}
        icon={Clock}
      />
      <StatCard
        title="Avg Cost"
        value={`$${(analytics.avgCostUsd ?? 0).toFixed(4)}`}
        icon={DollarSign}
      />
    </div>
  );
}

interface CostBreakdownChartProps {
  breakdown: CostBreakdown | null;
  isLoading: boolean;
}

function CostBreakdownChart({ breakdown, isLoading }: CostBreakdownChartProps) {
  if (isLoading) {
    return <Skeleton className="h-[300px] w-full" />;
  }

  if (!breakdown || breakdown.totalCost === 0) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center h-[300px]">
          <p className="text-text-muted">No cost data available</p>
        </CardContent>
      </Card>
    );
  }

  const pieData = Object.entries(breakdown.byFunction).map(([name, value]) => ({
    name,
    value,
  }));

  const barData = Object.entries(breakdown.byPeriod).map(([name, value]) => ({
    name,
    value,
  }));

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <PieChart
        data={pieData}
        title="Cost by Function"
        description="Distribution of execution costs across functions"
        donut
        donutInnerRadius={60}
        height={280}
        tooltipFormatter={(value) => [`$${Number(value).toFixed(4)}`, 'Cost']}
      />
      <BarChart
        data={barData}
        series={[{ key: 'value', name: 'Cost', color: 'var(--brand-500)' }]}
        title="Cost by Period"
        description="Cost breakdown over time"
        height={280}
        yAxisFormatter={(value) => `$${value.toFixed(2)}`}
      />
    </div>
  );
}

interface ExecutionsChartProps {
  executions: ExecutionRecord[];
  isLoading: boolean;
}

function ExecutionsChart({ executions, isLoading }: ExecutionsChartProps) {
  if (isLoading) {
    return <Skeleton className="h-[300px] w-full" />;
  }

  if (!executions || executions.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center h-[300px]">
          <p className="text-text-muted">No execution data available</p>
        </CardContent>
      </Card>
    );
  }

  const last7Days = [...Array(7)].map((_, i) => {
    const date = new Date();
    date.setDate(date.getDate() - (6 - i));
    return date.toISOString().split('T')[0];
  });

  const dailyData = last7Days.map((date) => {
    const dayExecutions = executions.filter((e) => {
      const execDate = new Date(e.timestamp).toISOString().split('T')[0];
      return execDate === date;
    });
    return {
      name: new Date(date).toLocaleDateString('en-US', { weekday: 'short' }),
      executions: dayExecutions.length,
      costs: dayExecutions.reduce((sum, e) => sum + e.costUsd, 0),
    };
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Execution Activity</CardTitle>
        <CardDescription>Executions and costs over the last 7 days</CardDescription>
      </CardHeader>
      <CardContent>
        <AreaChart
          data={dailyData}
          series={[
            { key: 'executions', name: 'Executions', color: '#6366f1' },
            { key: 'costs', name: 'Costs ($)', color: '#10b981', strokeWidth: 2 },
          ]}
          height={280}
          yAxisFormatter={(value) => value.toString()}
          xAxisFormatter={(value) => value}
          tooltipFormatter={(value, name) => {
            if (name === 'Costs ($)') {
              return [`$${Number(value).toFixed(4)}`, name];
            }
            return [String(value), name];
          }}
        />
      </CardContent>
    </Card>
  );
}

interface RecentExecutionsProps {
  executions: ExecutionRecord[];
  isLoading: boolean;
}

function RecentExecutions({ executions, isLoading }: RecentExecutionsProps) {
  if (isLoading) {
    return (
      <div className="space-y-2">
        {[...Array(5)].map((_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    );
  }

  if (!executions || executions.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center h-[200px]">
          <p className="text-text-muted">No recent executions</p>
        </CardContent>
      </Card>
    );
  }

  const recentExecutions = executions.slice(0, 10);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Executions</CardTitle>
        <CardDescription>Latest function executions</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {recentExecutions.map((execution) => (
            <div
              key={execution.id}
              className="flex items-center justify-between p-3 rounded-lg border border-border-subtle hover:bg-secondary/50 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-primary/10">
                  <BarChart3 className="h-4 w-4 text-brand" />
                </div>
                <div>
                  <p className="font-medium text-sm">
                    {execution.functionName}
                    <span className="text-text-muted ml-1">
                      by {execution.functionAuthor}
                    </span>
                  </p>
                  <p className="text-xs text-text-muted">
                    {new Date(execution.timestamp).toLocaleString()}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="text-right">
                  <p className="text-sm font-medium">${(execution.costUsd ?? 0).toFixed(4)}</p>
                  <p className="text-xs text-text-muted">
                    {((execution.latencyMs ?? 0) / 1000).toFixed(2)}s
                  </p>
                </div>
                <Badge
                  variant={execution.outcome === 'success' ? 'success' : 'error'}
                >
                  {execution.outcome}
                </Badge>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

interface TimeRangeSelectorProps {
  value: TimeRange;
  onChange: (range: TimeRange) => void;
}

function TimeRangeSelector({ value, onChange }: TimeRangeSelectorProps) {
  return (
    <div className="flex items-center gap-1 p-1 rounded-lg bg-secondary border border-border-subtle">
      {(['24h', '7d', '30d'] as TimeRange[]).map((range) => (
        <Button
          key={range}
          variant={value === range ? 'default' : 'ghost'}
          size="sm"
          onClick={() => onChange(range)}
          className="h-8"
        >
          {range}
        </Button>
      ))}
    </div>
  );
}

interface AgentAnalyticsProps {
  agentId: string;
  className?: string;
}

export function AgentAnalyticsComponent({ agentId, className }: AgentAnalyticsProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>('7d');

  const sinceMap: Record<TimeRange, string> = {
    '24h': new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    '7d': new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    '30d': new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
  };

  const { data: analyticsData, isLoading: analyticsLoading } = useQuery({
    queryKey: ['agent-analytics', agentId, timeRange],
    queryFn: () => agentApi.getAnalytics(agentId, { since: sinceMap[timeRange] }),
    enabled: !!agentId,
  });

  const { data: executionsData, isLoading: executionsLoading } = useQuery({
    queryKey: ['agent-executions', agentId, timeRange],
    queryFn: () => agentApi.listExecutions(agentId, { limit: 100 }),
    enabled: !!agentId,
  });

  const { data: costData, isLoading: costLoading } = useQuery({
    queryKey: ['agent-cost', agentId],
    queryFn: () => agentApi.getCostBreakdown(agentId),
    enabled: !!agentId,
  });

  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ['agent-usage', agentId],
    queryFn: () => agentApi.getUsage(agentId),
    enabled: !!agentId,
  });

  const analytics = analyticsData?.analytics;
  const executions = executionsData?.executions ?? [];
  const breakdown = costData?.breakdown;
  const usage = usageData?.usage ?? null;

  return (
    <div className={className}>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold">Agent Analytics</h2>
          <p className="text-text-muted mt-1">
            Execution statistics and cost breakdown
          </p>
        </div>
        <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
      </div>

      <StatsGrid
        analytics={analytics ?? { totalExecutions: 0, successRate: 0, avgLatencyMs: 0, avgCostUsd: 0, period: '' }}
        usage={usage}
        isLoading={analyticsLoading}
      />

      <div className="mt-6">
        <Tabs defaultValue="overview" className="w-full">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="costs">Costs</TabsTrigger>
            <TabsTrigger value="executions">Executions</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="mt-4">
            <div className="space-y-6">
              <ExecutionsChart executions={executions} isLoading={executionsLoading} />
            </div>
          </TabsContent>

          <TabsContent value="costs" className="mt-4">
            <div className="space-y-6">
              <CostBreakdownChart breakdown={breakdown ?? null} isLoading={costLoading} />
            </div>
          </TabsContent>

          <TabsContent value="executions" className="mt-4">
            <div className="space-y-6">
              <RecentExecutions executions={executions} isLoading={executionsLoading} />
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

export default AgentAnalyticsComponent;
