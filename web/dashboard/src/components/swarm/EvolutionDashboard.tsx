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

// Types for Autonomous Operations (SEBG) Dashboard
interface SEBGModificationProposal {
  id: string;
  graph_id: string;
  change_type: 'add_node' | 'remove_node' | 'rewire_edge' | 'add_specialist' | 'optimize';
  target_node_id: string;
  target_node_name: string;
  expected_revenue_lift: number;
  expected_lift_pct: number;
  risk_score: number;
  status: 'pending' | 'approved' | 'rejected' | 'applied' | 'expired';
  approved_by?: string;
  created_at: string;
  updated_at?: string;
}

interface SEBGTenantConfig {
  tenant_id: string;
  autonomy_tier: 'manual' | 'assisted' | 'fully_autonomous';
  revenue_share_fee_pct: number;
  max_risk_score_auto_apply: number;
  is_active: boolean;
}

interface AutonomyTierOption {
  value: 'manual' | 'assisted' | 'fully_autonomous';
  label: string;
  description: string;
  badge: string;
}

// Autonomy tier configuration
const AUTONOMY_TIERS: AutonomyTierOption[] = [
  {
    value: 'manual',
    label: 'Manual',
    description: 'SEBG observes and recommends; you approve all changes',
    badge: 'bg-gray-500',
  },
  {
    value: 'assisted',
    label: 'Assisted',
    description: 'Low-risk changes auto-apply; high-risk changes need approval',
    badge: 'bg-blue-500',
  },
  {
    value: 'fully_autonomous',
    label: 'Fully Autonomous',
    description: 'SEBG operates without intervention — premium tier',
    badge: 'bg-green-500',
  },
];

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

  // SEBG state
  const [sebgProposals, setSebgProposals] = useState<SEBGModificationProposal[]>([]);
  const [sebgConfig, setSebgConfig] = useState<SEBGTenantConfig | null>(null);
  const [sebgLoading, setSebgLoading] = useState(false);
  const [selectedTier, setSelectedTier] = useState<'manual' | 'assisted' | 'fully_autonomous'>(
    'assisted'
  );
  const [roiSummary, setRoiSummary] = useState<{
    applied: number;
    pending: number;
    revenueLift: number;
  }>({
    applied: 0,
    pending: 0,
    revenueLift: 0,
  });

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

  // SEBG: fetch modification proposals and tenant config
  const fetchSebgData = useCallback(async () => {
    if (!agentId) return;
    setSebgLoading(true);
    try {
      const [configRes, proposalsRes, roiRes] = await Promise.allSettled([
        agentApi.getSEBGConfig(agentId),
        agentApi.listSEBGProposals(agentId, { status: 'pending' }),
        agentApi.getSEBGROI(agentId),
      ]);

      const cfg = configRes.status === 'fulfilled' ? configRes.value?.config : null;
      if (cfg) {
        setSebgConfig({
          tenant_id: cfg.tenant_id,
          autonomy_tier: cfg.autonomy_tier as 'manual' | 'assisted' | 'fully_autonomous',
          revenue_share_fee_pct: cfg.revenue_share_fee_pct,
          max_risk_score_auto_apply: cfg.max_risk_score_auto_apply,
          is_active: cfg.is_active,
        });
        setSelectedTier(cfg.autonomy_tier as 'manual' | 'assisted' | 'fully_autonomous');
      }

      const proposals = proposalsRes.status === 'fulfilled' ? proposalsRes.value?.proposals ?? [] : [];
      setSebgProposals(proposals);

      const roi = roiRes.status === 'fulfilled' ? roiRes.value?.roi : null;
      if (roi) {
        setRoiSummary({
          applied: roi.applied_count ?? 0,
          pending: roi.pending_count ?? 0,
          revenueLift: roi.revenue_lift_cents ?? 0,
        });
      }
    } catch (e) {
      // SEBG data is non-critical — don't block the dashboard on failure
    } finally {
      setSebgLoading(false);
    }
  }, [agentId]);

  // SEBG: handle one-click approve/reject
  const handleSebgDecision = async (proposalId: string, decision: 'approved' | 'rejected') => {
    try {
      await agentApi.decideSEBGProposal(agentId, proposalId, decision);
      setSebgProposals((prev) =>
        prev.map((p) => (p.id === proposalId ? { ...p, status: decision } : p))
      );
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Decision failed');
    }
  };

  // SEBG: handle autonomy tier change
  const handleTierChange = async (tier: 'manual' | 'assisted' | 'fully_autonomous') => {
    setSelectedTier(tier);
    try {
      await agentApi.updateSEBGTier(agentId, tier);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update tier');
    }
  };

  useEffect(() => {
    fetchData();
    fetchSebgData();
  }, [fetchData, fetchSebgData]);

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
      applied: { color: 'bg-green-600', icon: <CheckCircle className="h-3 w-3" /> },
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
            <Zap className="h-8 w-8 text-primary" />
            Autonomous Operations
          </h1>
          <p className="text-muted-foreground mt-1">
            Self-evolving backend graph — AI-optimized checkout and payments
          </p>
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
          <TabsTrigger value="autonomous">Autonomous Ops</TabsTrigger>
          <TabsTrigger value="trends">Performance Trends</TabsTrigger>
          <TabsTrigger value="failures">Failure Analysis</TabsTrigger>
        </TabsList>

        <TabsContent value="autonomous" className="space-y-4">
          {/* ROI Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Changes Applied</p>
                    <p className="text-2xl font-bold">{roiSummary.applied}</p>
                  </div>
                  <CheckCircle className="h-8 w-8 text-green-500" />
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Pending Review</p>
                    <p className="text-2xl font-bold">{roiSummary.pending}</p>
                  </div>
                  <Clock className="h-8 w-8 text-yellow-500" />
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Est. Revenue Lift</p>
                    <p className="text-2xl font-bold text-green-600 dark:text-green-400">
                      {roiSummary.revenueLift > 0
                        ? `+$${(roiSummary.revenueLift / 100).toFixed(2)}`
                        : '—'}
                    </p>
                  </div>
                  <TrendingUp className="h-8 w-8 text-emerald-500" />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Autonomy Tier Selector */}
          <Card>
            <CardHeader>
              <CardTitle>Autonomy Tier</CardTitle>
              <CardDescription>
                Controls how SEBG applies changes to your backend graph
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                {AUTONOMY_TIERS.map((tier) => (
                  <div
                    key={tier.value}
                    className={`p-4 border-2 rounded-lg cursor-pointer transition-all ${
                      selectedTier === tier.value
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                    onClick={() => handleTierChange(tier.value)}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-semibold">{tier.label}</span>
                      <Badge className={`${tier.badge} text-white`}>
                        {tier.value === 'fully_autonomous'
                          ? 'Premium'
                          : tier.value === 'assisted'
                            ? 'Recommended'
                            : 'Free'}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">{tier.description}</p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* SEBG Modification Proposals */}
          <Card>
            <CardHeader>
              <CardTitle>Graph Modifications</CardTitle>
              <CardDescription>
                AI-generated graph changes ranked by expected revenue impact
              </CardDescription>
            </CardHeader>
            <CardContent>
              {sebgLoading ? (
                <div className="flex items-center justify-center py-8">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary" />
                </div>
              ) : sebgProposals.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Brain className="h-12 w-12 mx-auto mb-3 opacity-30" />
                  <p>No modification proposals yet.</p>
                  <p className="text-sm">SEBG will analyze your graph and suggest improvements.</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {sebgProposals.map((proposal) => (
                    <div
                      key={proposal.id}
                      className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex items-center gap-4">
                        <div
                          className={`p-2 rounded-lg ${
                            proposal.risk_score < 0.2
                              ? 'bg-green-500/10'
                              : proposal.risk_score < 0.4
                                ? 'bg-yellow-500/10'
                                : 'bg-red-500/10'
                          }`}
                        >
                          <Zap
                            className={`h-5 w-5 ${
                              proposal.risk_score < 0.2
                                ? 'text-green-500'
                                : proposal.risk_score < 0.4
                                  ? 'text-yellow-500'
                                  : 'text-red-500'
                            }`}
                          />
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <p className="font-medium capitalize">
                              {proposal.change_type.replace('_', ' ')}
                            </p>
                            <Badge variant="outline" className="text-xs">
                              {proposal.risk_score < 0.2
                                ? 'Low Risk'
                                : proposal.risk_score < 0.4
                                  ? 'Medium Risk'
                                  : 'High Risk'}
                            </Badge>
                          </div>
                          <p className="text-sm text-muted-foreground mt-0.5">
                            {proposal.target_node_name
                              ? `Target: ${proposal.target_node_name}`
                              : `Target: ${proposal.target_node_id.slice(0, 8)}…`}
                          </p>
                          {proposal.expected_revenue_lift > 0 && (
                            <p className="text-sm text-green-600 dark:text-green-400 mt-0.5">
                              Expected lift: +${(proposal.expected_revenue_lift / 100).toFixed(2)}
                            </p>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        {proposal.status === 'pending' ? (
                          <>
                            <Button
                              size="sm"
                              variant="outline"
                              className="text-red-600 border-red-200 hover:bg-red-50 dark:hover:bg-red-900/20"
                              onClick={() => handleSebgDecision(proposal.id, 'rejected')}
                            >
                              <XCircle className="h-4 w-4 mr-1" />
                              Reject
                            </Button>
                            <Button
                              size="sm"
                              className="bg-green-600 hover:bg-green-700 text-white"
                              onClick={() => handleSebgDecision(proposal.id, 'approved')}
                            >
                              <CheckCircle className="h-4 w-4 mr-1" />
                              Approve
                            </Button>
                          </>
                        ) : (
                          getStatusBadge(proposal.status)
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

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
