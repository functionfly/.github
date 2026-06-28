import { agentApi } from '@/api/agent';
import { Chamber, SealedButton, FrameButton, StatusPill } from '@/components/containment';
import {
  Activity, AlertTriangle, Brain, CheckCircle, ChevronRight, Clock,
  Sparkles, Target, TrendingDown, TrendingUp, XCircle, Zap,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';

interface SEBGModificationProposal {
  id: string; graph_id: string;
  change_type: 'add_node' | 'remove_node' | 'rewire_edge' | 'add_specialist' | 'optimize';
  target_node_id: string; target_node_name: string;
  expected_revenue_lift: number; expected_lift_pct: number; risk_score: number;
  status: 'pending' | 'approved' | 'rejected' | 'applied' | 'expired';
  approved_by?: string; created_at: string; updated_at?: string;
}

interface EvolutionProposalUI {
  id: string;
  type: 'spawn_specialist' | 'modify_policy' | 'adjust_timeout' | 'generate_function' | 'retire_child';
  status: 'pending' | 'approved' | 'rejected' | 'implemented' | 'expired';
  createdAt: string; description: string;
  impact?: { successRate?: number; latency?: number; cost?: number };
}

interface PerformanceMetrics {
  totalExecutions: number; successRate: number; avgLatency: number; avgCost: number;
  failureCategories: Record<string, number>;
}

const AUTONOMY_TIERS = [
  { value: 'manual' as const, label: 'Manual', description: 'SEBG observes and recommends; you approve all changes', badge: 'Free' },
  { value: 'assisted' as const, label: 'Assisted', description: 'Low-risk changes auto-apply; high-risk changes need approval', badge: 'Recommended' },
  { value: 'fully_autonomous' as const, label: 'Fully Autonomous', description: 'SEBG operates without intervention — premium tier', badge: 'Premium' },
];

function mapApiProposalToUI(p: { id: string; proposal_type?: string; status: string; created_at?: string; updated_at?: string; proposal_data?: Record<string, unknown> }): EvolutionProposalUI {
  const type = (p.proposal_type ?? 'modify_policy') as EvolutionProposalUI['type'];
  const status = (['pending', 'approved', 'rejected', 'implemented', 'expired'].includes(p.status) ? p.status : 'pending') as EvolutionProposalUI['status'];
  const createdAt = p.created_at ?? new Date().toISOString();
  const data = p.proposal_data;
  const reason = (data?.reason as string) ?? type;
  const description = type === 'spawn_specialist' ? `Spawn specialist to improve ${reason.replace('_', ' ')}` : type === 'modify_policy' ? `Modify policy: ${reason.replace('_', ' ')}` : type === 'generate_function' ? `Generate function (${reason.replace('_', ' ')})` : type === 'retire_child' ? 'Retire underperforming child agent' : `Evolution: ${String(type).replace('_', ' ')}`;
  const impact: EvolutionProposalUI['impact'] = {};
  if (typeof data?.current_success === 'number' && typeof data?.target_success === 'number') impact.successRate = (data.target_success as number) - (data.current_success as number);
  if (typeof data?.current_latency === 'number' && typeof data?.target_latency === 'number') impact.latency = (data.target_latency as number) - (data.current_latency as number);
  return { id: p.id, type, status, createdAt, description, impact };
}

const typeLabels: Record<string, string> = { spawn_specialist: 'Spawn Specialist', modify_policy: 'Modify Policy', adjust_timeout: 'Adjust Timeout', generate_function: 'Generate Function', retire_child: 'Retire Child' };
const typeIcons: Record<string, React.ReactNode> = { spawn_specialist: <Zap style={{ width: 14, height: 14 }} />, modify_policy: <Target style={{ width: 14, height: 14 }} />, adjust_timeout: <Clock style={{ width: 14, height: 14 }} />, generate_function: <Sparkles style={{ width: 14, height: 14 }} />, retire_child: <TrendingDown style={{ width: 14, height: 14 }} /> };

const statusToPill = (s: string): 'live' | 'pending' | 'revoked' => {
  if (s === 'approved' || s === 'applied' || s === 'implemented') return 'live';
  if (s === 'pending') return 'pending';
  return 'revoked';
};

const inputStyle: React.CSSProperties = { width: '100%', padding: 'var(--space-3) var(--space-4)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', color: 'var(--text)', fontFamily: 'var(--font-body)', fontSize: 13, outline: 'none' };

export function EvolutionDashboard({ agentId }: { agentId: string }) {
  const [proposals, setProposals] = useState<EvolutionProposalUI[]>([]);
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [analyzing, setAnalyzing] = useState(false);
  const [activeTab, setActiveTab] = useState('proposals');

  const [sebgProposals, setSebgProposals] = useState<SEBGModificationProposal[]>([]);
  const [sebgConfig, setSebgConfig] = useState<{ tenant_id: string; autonomy_tier: string; revenue_share_fee_pct: number; max_risk_score_auto_apply: number; is_active: boolean } | null>(null);
  const [sebgLoading, setSebgLoading] = useState(false);
  const [selectedTier, setSelectedTier] = useState<'manual' | 'assisted' | 'fully_autonomous'>('assisted');
  const [roiSummary, setRoiSummary] = useState({ applied: 0, pending: 0, revenueLift: 0 });

  const fetchData = useCallback(async () => {
    if (!agentId) { setLoading(false); return; }
    setLoading(true); setError(null);
    try {
      const [analyticsRes, executionsRes] = await Promise.allSettled([
        agentApi.getAnalytics(agentId, { since: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString() }),
        agentApi.listExecutions(agentId, { limit: 200 }),
      ]);
      const analytics = analyticsRes.status === 'fulfilled' && analyticsRes.value?.analytics ? analyticsRes.value.analytics : null;
      const a = analytics ? (analytics as unknown as Record<string, unknown>) : null;
      const totalExecutions = Number(a?.totalExecutions ?? a?.total_executions ?? 0);
      const successRate = Number(a?.successRate ?? a?.success_rate ?? 0);
      const avgLatency = Number(a?.avgLatencyMs ?? a?.avg_latency_ms ?? 0);
      const avgCost = Number(a?.avgCostUsd ?? a?.avg_cost_usd ?? 0);
      let failureCategories: Record<string, number> = {};
      if (executionsRes.status === 'fulfilled' && executionsRes.value?.executions) {
        for (const e of executionsRes.value.executions as { outcome?: string; errorCode?: string }[]) {
          if (e.outcome === 'failure' || e.errorCode) { const key = (e.errorCode ?? e.outcome ?? 'unknown').replace(/\s+/g, '_').toLowerCase(); failureCategories[key] = (failureCategories[key] ?? 0) + 1; }
        }
      }
      if (Object.keys(failureCategories).length === 0) failureCategories = { unknown: 0 };
      setMetrics({ totalExecutions, successRate, avgLatency, avgCost, failureCategories });
    } catch (e) { setError(e instanceof Error ? e.message : 'Failed to load evolution data'); setMetrics(null); } finally { setLoading(false); }
  }, [agentId]);

  const fetchSebgData = useCallback(async () => {
    if (!agentId) return; setSebgLoading(true);
    try {
      const [configRes, proposalsRes, roiRes] = await Promise.allSettled([agentApi.getSEBGConfig(agentId), agentApi.listSEBGProposals(agentId, { status: 'pending' }), agentApi.getSEBGROI(agentId)]);
      const cfg = configRes.status === 'fulfilled' ? configRes.value?.config : null;
      if (cfg) { setSebgConfig({ tenant_id: cfg.tenant_id, autonomy_tier: cfg.autonomy_tier, revenue_share_fee_pct: cfg.revenue_share_fee_pct, max_risk_score_auto_apply: cfg.max_risk_score_auto_apply, is_active: cfg.is_active }); setSelectedTier(cfg.autonomy_tier as 'manual' | 'assisted' | 'fully_autonomous'); }
      setSebgProposals(proposalsRes.status === 'fulfilled' ? proposalsRes.value?.proposals ?? [] : []);
      const roi = roiRes.status === 'fulfilled' ? roiRes.value?.roi : null;
      if (roi) setRoiSummary({ applied: roi.applied_count ?? 0, pending: roi.pending_count ?? 0, revenueLift: roi.revenue_lift_cents ?? 0 });
    } catch { /* non-critical */ } finally { setSebgLoading(false); }
  }, [agentId]);

  const handleSebgDecision = async (proposalId: string, decision: 'approved' | 'rejected') => {
    try { await agentApi.decideSEBGProposal(agentId, proposalId, decision); setSebgProposals(prev => prev.map(p => p.id === proposalId ? { ...p, status: decision } : p)); } catch (e) { setError(e instanceof Error ? e.message : 'Decision failed'); }
  };

  const handleTierChange = async (tier: 'manual' | 'assisted' | 'fully_autonomous') => { setSelectedTier(tier); try { await agentApi.updateSEBGTier(agentId, tier); } catch (e) { setError(e instanceof Error ? e.message : 'Failed to update tier'); } };

  useEffect(() => { fetchData(); fetchSebgData(); }, [fetchData, fetchSebgData]);

  const handleAnalyzePropose = async () => {
    if (!agentId) return; setAnalyzing(true); setError(null);
    try {
      const res = await agentApi.proposeEvolution(agentId);
      if (res?.proposal) setProposals(prev => [mapApiProposalToUI(res.proposal as Parameters<typeof mapApiProposalToUI>[0]), ...prev]);
      const analysis = res?.analysis as Record<string, unknown> | undefined;
      if (analysis && metrics) setMetrics(m => m ? { ...m, totalExecutions: Number(analysis.totalExecutions ?? analysis.total_executions ?? m.totalExecutions), successRate: Number(analysis.successRate ?? analysis.success_rate ?? m.successRate), avgLatency: Number(analysis.avgLatencyMs ?? analysis.avg_latency_ms ?? m.avgLatency), avgCost: Number(analysis.avgCostUSD ?? analysis.avg_cost_usd ?? m.avgCost), failureCategories: (analysis.failureCategories ?? analysis.failure_categories) != null ? ((analysis.failureCategories ?? analysis.failure_categories) as Record<string, number>) : m.failureCategories } : m);
    } catch (e) { setError(e instanceof Error ? e.message : 'Analyze & propose failed'); } finally { setAnalyzing(false); }
  };

  const StatCard = ({ label, value, icon, color }: { label: string; value: string; icon: React.ReactNode; color: string }) => (
    <Chamber nested>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <p style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)', marginBottom: 'var(--space-1)' }}>{label}</p>
          <p style={{ fontFamily: 'var(--font-mono)', fontSize: 26, fontWeight: 700, color: 'var(--text)' }}>{value}</p>
        </div>
        <div style={{ color }}>{icon}</div>
      </div>
    </Chamber>
  );

  const ProgressBar = ({ value, color }: { value: number; color: string }) => (
    <div style={{ height: 6, background: 'var(--panel)', borderRadius: 'var(--radius-sm)', overflow: 'hidden', marginTop: 'var(--space-2)' }}>
      <div style={{ height: '100%', width: `${Math.min(100, value)}%`, background: color, borderRadius: 'var(--radius-sm)', transition: 'width var(--duration-base) var(--ease-out)' }} />
    </div>
  );

  const TabBtn = ({ value, label }: { value: string; label: string }) => (
    <button onClick={() => setActiveTab(value)} style={{ position: 'relative', padding: 'var(--space-3) var(--space-4)', fontFamily: 'var(--font-body)', fontSize: 14, fontWeight: 500, color: activeTab === value ? 'var(--status-ok)' : 'var(--text-dim)', background: 'transparent', border: 'none', borderBottom: `2px solid ${activeTab === value ? 'var(--status-ok)' : 'transparent'}`, cursor: 'pointer', whiteSpace: 'nowrap', transition: 'all var(--duration-fast) var(--ease-out)' }}>{label}</button>
  );

  if (loading) return <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 'var(--space-8)' }}><div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)' }}>Loading evolution data...</div></div>;

  if (error) return (
    <Chamber>
      <div style={{ textAlign: 'center', padding: 'var(--space-6) 0' }}>
        <p style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 600, color: 'var(--status-revoked)', marginBottom: 'var(--space-2)' }}>Failed to load evolution data</p>
        <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-4)' }}>{error}</p>
        <FrameButton size="sm" onClick={() => fetchData()}>Retry</FrameButton>
      </div>
    </Chamber>
  );

  return (
    <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 'var(--space-4)' }}>
        <div>
          <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, letterSpacing: '-0.005em', color: 'var(--text)', display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
            <Zap style={{ width: 32, height: 32, color: 'var(--status-ok)' }} />Autonomous Operations
          </h1>
          <p style={{ fontSize: 14, color: 'var(--text-dim)', marginTop: 'var(--space-2)' }}>Self-evolving backend graph — AI-optimized checkout and payments</p>
        </div>
        <SealedButton onClick={handleAnalyzePropose} disabled={analyzing} iconLeft={analyzing ? <span style={{ width: 14, height: 14, border: '2px solid var(--text-faint)', borderTopColor: 'var(--status-ok)', borderRadius: '50%', animation: 'spin 1s linear infinite' }} /> : <Sparkles style={{ width: 14, height: 14 }} />}>Analyze & Propose</SealedButton>
      </div>

      {/* Performance Overview */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-4)' }}>
        <StatCard label="Total Executions" value={metrics?.totalExecutions.toLocaleString() ?? '0'} icon={<Activity style={{ width: 28, height: 28 }} />} color="var(--foil-a)" />
        <StatCard label="Success Rate" value={`${metrics?.successRate ?? 0}%`} icon={<TrendingUp style={{ width: 28, height: 28 }} />} color="var(--status-ok)" />
        <StatCard label="Avg Latency" value={`${metrics?.avgLatency ?? 0}ms`} icon={<Clock style={{ width: 28, height: 28 }} />} color="var(--foil-b)" />
        <StatCard label="Avg Cost" value={`$${(metrics?.avgCost ?? 0).toFixed(3)}`} icon={<Zap style={{ width: 28, height: 28 }} />} color="var(--status-pending)" />
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 0, borderBottom: '1px solid var(--panel-edge)' }}>
        <TabBtn value="proposals" label="Proposals" />
        <TabBtn value="autonomous" label="Autonomous Ops" />
        <TabBtn value="trends" label="Performance Trends" />
        <TabBtn value="failures" label="Failure Analysis" />
      </div>

      {/* Proposals Tab */}
      {activeTab === 'proposals' && (
        <Chamber nested>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)', marginBottom: 'var(--space-4)' }}>Evolution Proposals</div>
          {proposals.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 'var(--space-7) 0', color: 'var(--text-dim)' }}>No proposals yet. Click "Analyze & Propose" to generate AI suggestions.</div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
              {proposals.map(p => (
                <div key={p.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: 'var(--space-4)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', transition: 'border-color var(--duration-fast) var(--ease-out)' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
                    <div style={{ width: 36, height: 36, borderRadius: 'var(--radius)', background: 'rgba(143,255,208,0.06)', border: '1px solid rgba(143,255,208,0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--status-ok)' }}>{typeIcons[p.type] ?? <Brain style={{ width: 14, height: 14 }} />}</div>
                    <div>
                      <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--text)' }}>{typeLabels[p.type] ?? p.type}</p>
                      <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>{p.description}</p>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)', marginTop: 'var(--space-1)' }}><Clock style={{ width: 11, height: 11, color: 'var(--text-faint)' }} /><span style={{ fontSize: 11, color: 'var(--text-faint)' }}>{new Date(p.createdAt).toLocaleDateString()}</span></div>
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                    {p.impact && <div style={{ textAlign: 'right', fontSize: 12 }}>{p.impact.successRate != null && <span style={{ color: p.impact.successRate > 0 ? 'var(--status-ok)' : 'var(--status-revoked)' }}>{p.impact.successRate > 0 ? '+' : ''}{p.impact.successRate}% success</span>}{p.impact.latency != null && <span style={{ marginLeft: 'var(--space-2)', color: 'var(--foil-a)' }}>{p.impact.latency}ms</span>}</div>}
                    <StatusPill status={statusToPill(p.status)} label={p.status} />
                  </div>
                </div>
              ))}
            </div>
          )}
        </Chamber>
      )}

      {/* Autonomous Ops Tab */}
      {activeTab === 'autonomous' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-4)' }}>
            <StatCard label="Changes Applied" value={String(roiSummary.applied)} icon={<CheckCircle style={{ width: 28, height: 28 }} />} color="var(--status-ok)" />
            <StatCard label="Pending Review" value={String(roiSummary.pending)} icon={<Clock style={{ width: 28, height: 28 }} />} color="var(--status-pending)" />
            <StatCard label="Est. Revenue Lift" value={roiSummary.revenueLift > 0 ? `+$${(roiSummary.revenueLift / 100).toFixed(2)}` : '—'} icon={<TrendingUp style={{ width: 28, height: 28 }} />} color="var(--status-ok)" />
          </div>

          {/* Autonomy Tier */}
          <Chamber nested>
            <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)', marginBottom: 'var(--space-2)' }}>Autonomy Tier</div>
            <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-4)' }}>Controls how SEBG applies changes to your backend graph</p>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-3)' }}>
              {AUTONOMY_TIERS.map(tier => (
                <div key={tier.value} onClick={() => handleTierChange(tier.value)} style={{ padding: 'var(--space-4)', border: `2px solid ${selectedTier === tier.value ? 'var(--status-ok)' : 'var(--panel-edge)'}`, borderRadius: 'var(--radius)', background: selectedTier === tier.value ? 'rgba(143,255,208,0.03)' : 'var(--panel)', cursor: 'pointer', transition: 'all var(--duration-fast) var(--ease-out)' }}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-2)' }}>
                    <span style={{ fontWeight: 600, color: 'var(--text)' }}>{tier.label}</span>
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: 10, textTransform: 'uppercase', letterSpacing: '0.04em', padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: selectedTier === tier.value ? 'rgba(143,255,208,0.1)' : 'var(--panel-raised)', border: `1px solid ${selectedTier === tier.value ? 'rgba(143,255,208,0.3)' : 'var(--panel-edge)'}`, color: selectedTier === tier.value ? 'var(--status-ok)' : 'var(--text-faint)' }}>{tier.badge}</span>
                  </div>
                  <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>{tier.description}</p>
                </div>
              ))}
            </div>
          </Chamber>

          {/* SEBG Proposals */}
          <Chamber nested>
            <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)', marginBottom: 'var(--space-2)' }}>Graph Modifications</div>
            <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-4)' }}>AI-generated graph changes ranked by expected revenue impact</p>
            {sebgLoading ? <div style={{ textAlign: 'center', padding: 'var(--space-7) 0', color: 'var(--text-faint)' }}>Loading...</div>
            : sebgProposals.length === 0 ? (
              <div style={{ textAlign: 'center', padding: 'var(--space-7) 0' }}>
                <Brain style={{ width: 48, height: 48, color: 'var(--text-faint)', margin: '0 auto var(--space-3)' }} />
                <p style={{ color: 'var(--text-dim)' }}>No modification proposals yet.</p>
                <p style={{ fontSize: 13, color: 'var(--text-faint)' }}>SEBG will analyze your graph and suggest improvements.</p>
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
                {sebgProposals.map(p => {
                  const riskColor = p.risk_score < 0.2 ? 'var(--status-ok)' : p.risk_score < 0.4 ? 'var(--status-pending)' : 'var(--status-revoked)';
                  const riskLabel = p.risk_score < 0.2 ? 'Low Risk' : p.risk_score < 0.4 ? 'Medium Risk' : 'High Risk';
                  return (
                    <div key={p.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: 'var(--space-4)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
                        <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: `${riskColor}11`, border: `1px solid ${riskColor}33`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Zap style={{ width: 18, height: 18, color: riskColor }} /></div>
                        <div>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                            <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--text)', textTransform: 'capitalize' }}>{p.change_type.replace('_', ' ')}</p>
                            <span style={{ fontFamily: 'var(--font-mono)', fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', border: `1px solid ${riskColor}33`, color: riskColor, textTransform: 'uppercase', letterSpacing: '0.04em' }}>{riskLabel}</span>
                          </div>
                          <p style={{ fontSize: 12, color: 'var(--text-dim)', marginTop: 2 }}>{p.target_node_name ? `Target: ${p.target_node_name}` : `Target: ${p.target_node_id.slice(0, 8)}…`}</p>
                          {p.expected_revenue_lift > 0 && <p style={{ fontSize: 12, color: 'var(--status-ok)', marginTop: 2 }}>Expected lift: +${(p.expected_revenue_lift / 100).toFixed(2)}</p>}
                        </div>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                        {p.status === 'pending' ? (
                          <>
                            <FrameButton size="sm" onClick={() => handleSebgDecision(p.id, 'rejected')} iconLeft={<XCircle style={{ width: 12, height: 12 }} />}>Reject</FrameButton>
                            <SealedButton size="sm" onClick={() => handleSebgDecision(p.id, 'approved')} iconLeft={<CheckCircle style={{ width: 12, height: 12 }} />}>Approve</SealedButton>
                          </>
                        ) : <StatusPill status={statusToPill(p.status)} label={p.status} />}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </Chamber>
        </div>
      )}

      {/* Trends Tab */}
      {activeTab === 'trends' && (
        <Chamber nested>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)', marginBottom: 'var(--space-4)' }}>Performance Trends</div>
          <div style={{ textAlign: 'center', padding: 'var(--space-7) 0', color: 'var(--text-dim)', fontSize: 13 }}>No trend data yet. Use "Analyze & Propose" or run more executions to build history.</div>
        </Chamber>
      )}

      {/* Failures Tab */}
      {activeTab === 'failures' && (
        <Chamber nested>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)', marginBottom: 'var(--space-4)' }}>Failure Analysis</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            {Object.entries(metrics?.failureCategories ?? {}).map(([category, count]) => {
              const total = Object.values(metrics?.failureCategories ?? {}).reduce((a, b) => a + b, 0);
              const pct = total > 0 ? (count / total) * 100 : 0;
              return (
                <div key={category}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-1)' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}><AlertTriangle style={{ width: 14, height: 14, color: 'var(--status-pending)' }} /><span style={{ fontSize: 13, color: 'var(--text)', textTransform: 'capitalize' }}>{category.replace('_', ' ')}</span></div>
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 600, color: 'var(--text)' }}>{count}</span>
                  </div>
                  <ProgressBar value={pct} color="var(--status-pending)" />
                </div>
              );
            })}
          </div>
        </Chamber>
      )}

      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}

export default EvolutionDashboard;
