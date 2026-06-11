/**
 * MCP Center - Analytics Tab
 * MCP-specific metrics and insights
 */

import { useState } from 'react';
import { BarChart3, Loader2 } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { MCPMetricsCard, ChartsGrid, MCPInsightsSection } from '../components';
import { useMCPAnalytics } from '../hooks';
import { TIME_RANGES } from '../constants';
import type { TimeRange } from '../types';

export function AnalyticsTab() {
  const [timeRange, setTimeRange] = useState<TimeRange>('30d');
  const { analytics, kpis, insights, isLoading } = useMCPAnalytics(timeRange);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
            <BarChart3 className="h-5 w-5 text-brand-500" />
            MCP Analytics
          </h2>
          <p className="text-sm text-text-secondary mt-1">
            Track MCP usage metrics and client activity
          </p>
        </div>

        {/* Time Range Selector */}
        <div className="mcp-time-selector">
          {TIME_RANGES.map((range) => (
            <Button
              key={range.value}
              variant={timeRange === range.value ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setTimeRange(range.value)}
              className={
                timeRange === range.value ? 'mcp-time-selector-btn active' : 'mcp-time-selector-btn'
              }
            >
              {range.value}
            </Button>
          ))}
        </div>
      </div>

      {/* KPI Cards */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="border-theme bg-card">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-text-secondary">
                  <Loader2 className="h-4 w-4 animate-spin" />
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-8 w-20 bg-muted animate-pulse rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : kpis ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <MCPMetricsCard
            title="Total MCP Calls"
            value={kpis.totalCalls.value.toLocaleString()}
            change={kpis.totalCalls.change}
            subtitle="vs previous period"
            icon={<BarChart3 className="h-5 w-5" />}
            iconColor="text-brand-500"
          />
          <MCPMetricsCard
            title="Unique Clients"
            value={kpis.uniqueClients.value}
            change={kpis.uniqueClients.change}
            subtitle="vs previous period"
            icon={<BarChart3 className="h-5 w-5" />}
            iconColor="text-purple-500"
          />
          <MCPMetricsCard
            title="Avg Latency"
            value={`${kpis.avgLatency.value}ms`}
            change={kpis.avgLatency.change}
            subtitle="vs previous period"
            icon={<BarChart3 className="h-5 w-5" />}
            iconColor="text-amber-500"
            lowerIsBetter
          />
          <MCPMetricsCard
            title="Success Rate"
            value={`${kpis.successRate.value.toFixed(1)}%`}
            change={kpis.successRate.change}
            subtitle="vs previous period"
            icon={<BarChart3 className="h-5 w-5" />}
            iconColor="text-emerald-500"
            higherIsBetter
          />
        </div>
      ) : null}

      {/* Charts Grid */}
      <ChartsGrid analytics={analytics} isLoading={isLoading} />

      {/* Insights */}
      <MCPInsightsSection insights={insights} />

      {/* Export Section */}
      <Card className="mcp-export-section">
        <CardHeader>
          <CardTitle className="text-base">Export Analytics</CardTitle>
          <CardDescription>Download MCP analytics data for external analysis</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-3">
            <Button variant="outline" className="mcp-export-btn">
              Export as CSV
            </Button>
            <Button variant="outline" className="mcp-export-btn">
              Export as JSON
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
