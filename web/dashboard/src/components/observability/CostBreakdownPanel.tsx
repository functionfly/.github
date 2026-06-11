'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DollarSign, TrendingUp, TrendingDown } from 'lucide-react';

interface CostBreakdownPanelProps {
  runId: string | null;
}

interface CostStats {
  total_cost_usd: number;
  input_tokens: number;
  output_tokens: number;
  event_count: number;
  avg_latency_ms: number;
  cost_per_token: number;
}

export default function CostBreakdownPanel({ runId }: CostBreakdownPanelProps) {
  const [stats, setStats] = useState<CostStats | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!runId) return;

    const fetchStats = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/v1/agent-observability/runs/${runId}/stats`);
        if (response.ok) {
          const data = await response.json();
          setStats(data);
        }
      } catch (error) {
        console.error('Failed to fetch stats:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, [runId]);

  if (!runId) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Select a run to view cost breakdown
      </div>
    );
  }

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Cost</p>
                <p className="text-2xl font-bold">${stats?.total_cost_usd?.toFixed(6) || '0'}</p>
              </div>
              <DollarSign className="h-8 w-8 text-muted-foreground" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Input Tokens</p>
                <p className="text-2xl font-bold">{stats?.input_tokens?.toLocaleString() || 0}</p>
              </div>
              <TrendingUp className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Output Tokens</p>
                <p className="text-2xl font-bold">{stats?.output_tokens?.toLocaleString() || 0}</p>
              </div>
              <TrendingDown className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Cost/Token</p>
                <p className="text-2xl font-bold">${stats?.cost_per_token?.toFixed(6) || '0'}</p>
              </div>
              <DollarSign className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Token Usage</CardTitle>
            <CardDescription>Input vs Output breakdown</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm">Input Tokens</span>
                <span className="font-medium">{stats?.input_tokens?.toLocaleString() || 0}</span>
              </div>
              <div className="w-full bg-muted rounded-full h-2">
                <div
                  className="bg-blue-500 h-2 rounded-full"
                  style={{
                    width: `${stats?.input_tokens && stats?.output_tokens
                      ? (stats.input_tokens / (stats.input_tokens + stats.output_tokens)) * 100
                      : 50}%`,
                  }}
                />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Output Tokens</span>
                <span className="font-medium">{stats?.output_tokens?.toLocaleString() || 0}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Performance</CardTitle>
            <CardDescription>Execution metrics</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm">Avg Latency</span>
              <span className="font-medium">{stats?.avg_latency_ms?.toFixed(2) || '0'} ms</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Total Events</span>
              <span className="font-medium">{stats?.event_count?.toLocaleString() || 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Cost per Event</span>
              <span className="font-medium">
                ${stats?.event_count
                  ? (stats.total_cost_usd / stats.event_count).toFixed(6)
                  : '0'}
              </span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
