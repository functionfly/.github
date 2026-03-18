import { agentApi } from '@/api/agent';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Activity,
  AlertTriangle,
  Brain,
  CheckCircle,
  ChevronRight,
  Clock,
  Sparkles,
  Target,
  TrendingDown,
  TrendingUp,
  XCircle,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';

// Types for Evolution Dashboard (UI)
interface EvolutionProposalUI {
  id: string;
  type:
    | 'spawn_specialist'
    | 'modify_policy'
    | 'adjust_timeout'
    | 'generate_function'
    | 'retire_child';
  status: 'pending' | 'approved' | 'rejected' | 'implemented' | 'expired';
  createdAt: string;
  description: string;
  impact?: {
    successRate?: number;
    latency?: number;
    cost?: number;
  };
}

interface PerformanceMetrics {
  totalExecutions: number;
  successRate: number;
  avgLatency: number;
  avgCost: number;
  failureCategories: Record<string, number>;
}

interface TrendData {
  date: string;
  successRate: number;
  latency: number;
  cost: number;
}

function mapApiProposalToUI(p: {
  id: string;
  proposal_type?: string;
  status: string;
  created_at?: string;
  updated_at?: string;
  proposal_data?: Record<string, unknown>;
}): EvolutionProposalUI {
  const type = (p.proposal_type ??
    (p as { proposalType?: string }).proposalType ??
    'modify_policy') as EvolutionProposalUI['type'];
  const status = (
    ['pending', 'approved', 'rejected', 'implemented', 'expired'].includes(p.status)
      ? p.status
      : 'pending'
  ) as EvolutionProposalUI['status'];
  const createdAt =
    p.created_at ?? (p as { createdAt?: string }).createdAt ?? new Date().toISOString();
  const data = p.proposal_data ?? (p as { proposalData?: Record<string, unknown> }).proposalData;
  const reason = (data?.reason as string) ?? type;
  const description =
    type === 'spawn_specialist'
      ? `Spawn specialist to improve ${reason.replace('_', ' ')}`
      : type === 'modify_policy'
        ? `Modify policy: ${reason.replace('_', ' ')}`
        : type === 'generate_function'
          ? `Generate function (${reason.replace('_', ' ')})`
          : type === 'retire_child'
            ? 'Retire underperforming child agent'
            : `Evolution: ${String(type).replace('_', ' ')}`;
  const impact: EvolutionProposalUI['impact'] = {};
  if (typeof data?.current_success === 'number' && typeof data?.target_success === 'number') {
    impact.successRate = (data.target_success as number) - (data.current_success as number);
  }
  if (typeof data?.current_latency === 'number' && typeof data?.target_latency === 'number') {
    impact.latency = (data.target_latency as number) - (data.current_latency as number);
  }
  return { id: p.id, type, status, createdAt, description, impact };
}

export function EvolutionDashboard({ agentId }: { agentId: string }) {
  const [proposals, setProposals] = useState<EvolutionProposalUI[]>([]);
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [trends, setTrends] = useState<TrendData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [analyzing, setAnalyzing] = useState(false);

  const fetchData = useCallback(async () => {
    if (!agentId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [analyticsRes, executionsRes] = await Promise.allSettled([
        agentApi.getAnalytics(agentId, {
          since: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        }),
        agentApi.listExecutions(agentId, { limit: 200 }),
      ]);

      const analytics =
        analyticsRes.status === 'fulfilled' && analyticsRes.value?.analytics
          ? analyticsRes.value.analytics
          : null;
      const a = analytics ? (analytics as unknown as Record<string, unknown>) : null;
      const totalExecutions = Number(a?.totalExecutions ?? a?.total_executions ?? 0);
      const successRate = Number(a?.successRate ?? a?.success_rate ?? 0);
      const avgLatency = Number(a?.avgLatencyMs ?? a?.avg_latency_ms ?? 0);
      const avgCost = Number(a?.avgCostUsd ?? a?.avg_cost_usd ?? 0);

      let failureCategories: Record<string, number> = {};
      if (executionsRes.status === 'fulfilled' && executionsRes.value?.executions) {
        const execs = executionsRes.value.executions as { outcome?: string; errorCode?: string }[];
        for (const e of execs) {
          if (e.outcome === 'failure' || e.errorCode) {
            const key = (e.errorCode ?? e.outcome ?? 'unknown').replace(/\s+/g, '_').toLowerCase();
            failureCategories[key] = (failureCategories[key] ?? 0) + 1;
          }
        }
      }
      if (Object.keys(failureCategories).length === 0) {
        failureCategories = { unknown: 0 };
      }

      setMetrics({
        totalExecutions,
        successRate,
        avgLatency,
        avgCost,
        failureCategories,
      });
      setTrends([]);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load evolution data');
      setMetrics(null);
    } finally {
      setLoading(false);
    }
  }, [agentId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleAnalyzePropose = async () => {
    if (!agentId) return;
    setAnalyzing(true);
    setError(null);
    try {
      const res = await agentApi.proposeEvolution(agentId);
      const proposal = res?.proposal;
      const analysis = res?.analysis as Record<string, unknown> | undefined;
      if (proposal) {
        const ui = mapApiProposalToUI(proposal as Parameters<typeof mapApiProposalToUI>[0]);
        setProposals((prev) => [ui, ...prev]);
      }
      if (analysis && metrics) {
        setMetrics((m) =>
          m
            ? {
                ...m,
                totalExecutions: Number(
                  analysis.totalExecutions ?? analysis.total_executions ?? m.totalExecutions
                ),
                successRate: Number(analysis.successRate ?? analysis.success_rate ?? m.successRate),
                avgLatency: Number(
                  analysis.avgLatencyMs ?? analysis.avg_latency_ms ?? m.avgLatency
                ),
                avgCost: Number(analysis.avgCostUSD ?? analysis.avg_cost_usd ?? m.avgCost),
                failureCategories:
                  (analysis.failureCategories ?? analysis.failure_categories) != null
                    ? ((analysis.failureCategories ?? analysis.failure_categories) as Record<
                        string,
                        number
                      >)
                    : m.failureCategories,
              }
            : m
        );
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Analyze & propose failed');
    } finally {
      setAnalyzing(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const config = {
      pending: { color: 'bg-yellow-500', icon: <Clock className="h-3 w-3" /> },
      approved: { color: 'bg-blue-500', icon: <CheckCircle className="h-3 w-3" /> },
      rejected: { color: 'bg-red-500', icon: <XCircle className="h-3 w-3" /> },
      implemented: { color: 'bg-green-500', icon: <CheckCircle className="h-3 w-3" /> },
      expired: { color: 'bg-gray-500', icon: <Clock className="h-3 w-3" /> },
    };
    const c = config[status as keyof typeof config];
    return (
      <Badge className={`${c.color} text-white`}>
        {c.icon}
        <span className="ml-1 capitalize">{status}</span>
      </Badge>
    );
  };

  const getTypeLabel = (type: string) => {
    const labels = {
      spawn_specialist: 'Spawn Specialist',
      modify_policy: 'Modify Policy',
      adjust_timeout: 'Adjust Timeout',
      generate_function: 'Generate Function',
      retire_child: 'Retire Child',
    };
    return labels[type as keyof typeof labels] || type;
  };

  const getTypeIcon = (type: string) => {
    const icons = {
      spawn_specialist: <Zap className="h-4 w-4" />,
      modify_policy: <Target className="h-4 w-4" />,
      adjust_timeout: <Clock className="h-4 w-4" />,
      generate_function: <Sparkles className="h-4 w-4" />,
      retire_child: <TrendingDown className="h-4 w-4" />,
    };
    return icons[type as keyof typeof icons] || <Brain className="h-4 w-4" />;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-destructive">
        <p className="font-medium">Failed to load evolution data</p>
        <p className="text-sm mt-1">{error}</p>
        <Button variant="outline" size="sm" className="mt-3" onClick={() => fetchData()}>
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Brain className="h-8 w-8" />
            Evolution Center
          </h1>
          <p className="text-muted-foreground mt-1">Agent learning and autonomous improvement</p>
        </div>
        <Button onClick={handleAnalyzePropose} disabled={analyzing}>
          {analyzing ? (
            <span className="animate-spin mr-2 inline-block h-4 w-4 rounded-full border-2 border-primary border-t-transparent" />
          ) : (
            <Sparkles className="h-4 w-4 mr-2" />
          )}
          Analyze & Propose
        </Button>
      </div>

      {/* Performance Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Executions</p>
                <p className="text-2xl font-bold">{metrics?.totalExecutions.toLocaleString()}</p>
              </div>
              <Activity className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Success Rate</p>
                <p className="text-2xl font-bold">{metrics?.successRate}%</p>
              </div>
              <TrendingUp className="h-8 w-8 text-green-500" />
            </div>
            <Progress value={metrics?.successRate ?? 0} className="mt-2" />
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg Latency</p>
                <p className="text-2xl font-bold">{metrics?.avgLatency ?? 0}ms</p>
              </div>
              <Clock className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg Cost</p>
                <p className="text-2xl font-bold">${(metrics?.avgCost ?? 0).toFixed(3)}</p>
              </div>
              <Zap className="h-8 w-8 text-yellow-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="proposals" className="space-y-4">
        <TabsList>
          <TabsTrigger value="proposals">Proposals</TabsTrigger>
          <TabsTrigger value="trends">Performance Trends</TabsTrigger>
          <TabsTrigger value="failures">Failure Analysis</TabsTrigger>
        </TabsList>

        <TabsContent value="proposals" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Evolution Proposals</CardTitle>
              <CardDescription>
                AI-generated suggestions to improve agent performance
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {proposals.map((proposal) => (
                  <div
                    key={proposal.id}
                    className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex items-center gap-4">
                      <div className="p-2 bg-primary/10 rounded-lg">
                        {getTypeIcon(proposal.type)}
                      </div>
                      <div>
                        <p className="font-medium">{getTypeLabel(proposal.type)}</p>
                        <p className="text-sm text-muted-foreground">{proposal.description}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <Clock className="h-3 w-3 text-muted-foreground" />
                          <span className="text-xs text-muted-foreground">
                            {new Date(proposal.createdAt).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      {proposal.impact && (
                        <div className="text-right text-sm">
                          {proposal.impact.successRate && (
                            <span
                              className={
                                proposal.impact.successRate > 0 ? 'text-green-500' : 'text-red-500'
                              }
                            >
                              {proposal.impact.successRate > 0 ? '+' : ''}
                              {proposal.impact.successRate}% success
                            </span>
                          )}
                          {proposal.impact.latency && (
                            <span className="ml-2 text-blue-500">
                              {proposal.impact.latency}ms latency
                            </span>
                          )}
                        </div>
                      )}
                      {getStatusBadge(proposal.status)}
                      <Button variant="ghost" size="sm">
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="trends">
          <Card>
            <CardHeader>
              <CardTitle>Performance Trends</CardTitle>
              <CardDescription>
                Agent performance over time (trend data when available)
              </CardDescription>
            </CardHeader>
            <CardContent>
              {trends.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">
                  No trend data yet. Use &quot;Analyze & Propose&quot; or run more executions to
                  build history.
                </p>
              ) : (
                <>
                  <div className="h-64 flex items-end justify-between gap-2">
                    {trends.map((day) => (
                      <div key={day.date} className="flex-1 flex flex-col items-center gap-2">
                        <div
                          className="w-full bg-blue-500 rounded-t"
                          style={{ height: `${Math.min(100, day.successRate)}%` }}
                        />
                        <span className="text-xs text-muted-foreground">
                          {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}
                        </span>
                      </div>
                    ))}
                  </div>
                  <div className="flex justify-center gap-6 mt-4">
                    <div className="flex items-center gap-2">
                      <div className="h-3 w-3 bg-blue-500 rounded" />
                      <span className="text-sm">Success Rate</span>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="failures">
          <Card>
            <CardHeader>
              <CardTitle>Failure Analysis</CardTitle>
              <CardDescription>Breakdown of failure categories</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {Object.entries(metrics?.failureCategories ?? {}).map(([category, count]) => {
                  const total = Object.values(metrics?.failureCategories ?? {}).reduce(
                    (a, b) => a + b,
                    0
                  );
                  const pct = total > 0 ? (count / total) * 100 : 0;
                  return (
                    <div key={category} className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <AlertTriangle className="h-4 w-4 text-yellow-500" />
                        <span className="capitalize">{category.replace('_', ' ')}</span>
                      </div>
                      <div className="flex items-center gap-4">
                        <Progress value={pct} className="w-32" />
                        <span className="font-medium">{count}</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default EvolutionDashboard;
