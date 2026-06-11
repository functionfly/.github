/**
 * MCP Center - Charts Grid Component
 * Analytics charts for MCP usage data
 */

import { Loader2 } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AreaChart } from '@/components/ui/chart-area';
import { PieChart } from '@/components/ui/chart-pie';
import { BarChart } from '@/components/ui/chart-bar';
import type { MCPAnalytics } from '../types';

interface ChartsGridProps {
  analytics: MCPAnalytics | undefined;
  isLoading: boolean;
}

export function ChartsGrid({ analytics, isLoading }: ChartsGridProps) {
  if (isLoading) {
    return (
      <div className="mcp-charts-grid">
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} className="mcp-chart-card">
            <CardHeader>
              <div className="h-5 w-32 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-[250px] bg-muted/50 animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (!analytics) {
    return (
      <div className="mcp-charts-grid">
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} className="mcp-chart-card">
            <CardHeader>
              <CardTitle className="mcp-chart-title">No data available</CardTitle>
            </CardHeader>
            <CardContent className="flex items-center justify-center h-[250px]">
              <p className="text-sm text-muted-foreground">
                MCP analytics will appear here once you have active connections.
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="mcp-charts-grid">
      {/* Calls Over Time - Area Chart */}
      <Card className="mcp-chart-card">
        <CardHeader className="mcp-chart-header">
          <CardTitle className="mcp-chart-title">MCP Calls Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mcp-chart-container">
            <AreaChart
              data={analytics.calls_over_time.map((d) => ({
                time: d.time,
                value: d.count,
              }))}
              xAxisKey="time"
              series={[{ key: 'value', name: 'Calls', color: 'var(--brand-500)' }]}
              showLegend={false}
            />
          </div>
        </CardContent>
      </Card>

      {/* Client Breakdown - Pie Chart */}
      <Card className="mcp-chart-card">
        <CardHeader className="mcp-chart-header">
          <CardTitle className="mcp-chart-title">Client Breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mcp-chart-container">
            <PieChart
              data={analytics.client_breakdown.map((d) => ({
                name: d.client,
                value: d.count,
              }))}
              colors={['#6366f1', '#8b5cf6', '#a855f7', '#d946ef', '#ec4899']}
            />
          </div>
        </CardContent>
      </Card>

      {/* Top Functions - Horizontal Bar Chart */}
      <Card className="mcp-chart-card">
        <CardHeader className="mcp-chart-header">
          <CardTitle className="mcp-chart-title">Top Functions via MCP</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mcp-chart-container">
            <BarChart
              data={analytics.top_functions.map((d) => ({
                label: `${d.author}/${d.name}`,
                value: d.calls,
              }))}
              layout="horizontal"
              colors={['var(--brand-500)']}
            />
          </div>
        </CardContent>
      </Card>

      {/* Transport Usage - Bar Chart */}
      <Card className="mcp-chart-card">
        <CardHeader className="mcp-chart-header">
          <CardTitle className="mcp-chart-title">Transport Usage</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mcp-chart-container">
            <BarChart
              data={analytics.transport_usage.map((d) => ({
                label: d.transport,
                value: d.count,
              }))}
              layout="vertical"
              colors={['var(--brand-500)', 'var(--amber-500)']}
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// Insights section below charts
interface InsightsSectionProps {
  insights: { type: 'success' | 'info' | 'warning' | 'trending'; message: string }[];
}

export function MCPInsightsSection({ insights }: InsightsSectionProps) {
  if (!insights || insights.length === 0) {
    return null;
  }

  const getTypeStyles = (type: string) => {
    switch (type) {
      case 'success':
        return 'border-emerald-500/30 bg-emerald-500/10';
      case 'info':
        return 'border-brand-500/30 bg-brand-500/10';
      case 'warning':
        return 'border-amber-500/30 bg-amber-500/10';
      case 'trending':
        return 'border-purple-500/30 bg-purple-500/10';
      default:
        return 'border-border bg-muted/50';
    }
  };

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium text-text-secondary">Insights & Recommendations</h3>
      <div className="space-y-2">
        {insights.map((insight, i) => (
          <div key={i} className={`p-4 rounded-lg border ${getTypeStyles(insight.type)}`}>
            <p className="text-sm text-text-primary">{insight.message}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
