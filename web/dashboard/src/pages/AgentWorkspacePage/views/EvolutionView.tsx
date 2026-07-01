import { agentApi } from '@/api/agent';
import { FrameButton, SealedButton, StatusPill } from '@/components/containment';
import { useQuery } from '@tanstack/react-query';
import { Award, BarChart3, CheckCircle, Clock, Sparkles, TrendingUp, XCircle } from 'lucide-react';
import { useCallback } from 'react';

interface EvolutionViewProps {
  agentId: string;
}

export function EvolutionView({ agentId }: EvolutionViewProps) {
  const { data: sebgConfig, isLoading: configLoading } = useQuery({
    queryKey: ['agent-sebg-config', agentId],
    queryFn: () => agentApi.getSEBGConfig(agentId),
    enabled: !!agentId,
  });

  const { data: proposalsData, refetch: refetchProposals } = useQuery({
    queryKey: ['agent-sebg-proposals', agentId],
    queryFn: () => agentApi.listSEBGProposals(agentId),
    enabled: !!agentId,
  });

  const { data: roiData } = useQuery({
    queryKey: ['agent-sebg-roi', agentId],
    queryFn: () => agentApi.getSEBGROI(agentId),
    enabled: !!agentId,
  });

  const { data: analyticsData } = useQuery({
    queryKey: ['agent-analytics', agentId],
    queryFn: () => agentApi.getAnalytics(agentId),
    enabled: !!agentId,
  });

  const config = sebgConfig?.config as any;
  const proposals = proposalsData?.proposals ?? [];
  const roi = roiData?.roi as any;
  const analytics = analyticsData?.analytics as any;

  const handleTierChange = useCallback(async (tier: string) => {
    await agentApi.updateSEBGTier(agentId, tier as any);
  }, [agentId]);

  const handleDecide = useCallback(async (proposalId: string, decision: 'approved' | 'rejected') => {
    await agentApi.decideSEBGProposal(agentId, proposalId, decision);
    refetchProposals();
  }, [agentId, refetchProposals]);

  const handleTriggerEvolve = useCallback(async () => {
    await agentApi.triggerSEBGEvolve(agentId);
    refetchProposals();
  }, [agentId, refetchProposals]);

  if (configLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  const tiers = ['manual', 'assisted', 'fully_autonomous'];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Evolution</h2>
          <p className="aw-center__subtitle">Self-evolution tracking and SEBG management</p>
        </div>
        <SealedButton size="sm" onClick={handleTriggerEvolve} iconLeft={<Sparkles size={12} />}>
          Trigger Evolution
        </SealedButton>
      </div>

      {/* Autonomy Tier */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title">Autonomy Tier</span>
        </div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
            {tiers.map(tier => (
              <button
                key={tier}
                className={`aw-nav-item ${(config?.tier ?? 'manual') === tier ? 'aw-nav-item--active' : ''}`}
                style={{ flex: 1, justifyContent: 'center', textTransform: 'capitalize' }}
                onClick={() => handleTierChange(tier)}
              >
                {tier.replace('_', ' ')}
              </button>
            ))}
          </div>
          <p style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--text-faint)', margin: 'var(--space-2) 0 0' }}>
            {config?.tier === 'fully_autonomous'
              ? 'Agent can propose and apply changes automatically.'
              : config?.tier === 'assisted'
              ? 'Agent proposes changes; you approve or reject.'
              : 'All changes require manual approval.'}
          </p>
        </div>
      </div>

      {/* ROI Metrics */}
      <div className="aw-stats">
        <div className="aw-stat">
          <p className="aw-stat__label">Cost Savings</p>
          <p className="aw-stat__value" style={{ color: 'var(--status-ok)' }}>
            ${(roi?.cost_savings_usd ?? 0).toFixed(2)}
          </p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Efficiency Gain</p>
          <p className="aw-stat__value">{(roi?.efficiency_gain_percent ?? 0).toFixed(1)}%</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Uptime</p>
          <p className="aw-stat__value">{(roi?.uptime_percent ?? 100).toFixed(1)}%</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Proposals</p>
          <p className="aw-stat__value">{proposals.length}</p>
        </div>
      </div>

      {/* Proposals */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <Award size={14} />
            Proposals ({proposals.length})
          </span>
        </div>
        <div className="aw-card__body">
          {proposals.length === 0 ? (
            <div className="aw-empty" style={{ padding: 'var(--space-4)' }}>
              <Sparkles size={32} className="aw-empty__icon" />
              <span className="aw-empty__title">No proposals</span>
              <span className="aw-empty__desc">Trigger an evolution to generate improvement proposals</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
              {proposals.map((proposal: any) => (
                <div key={proposal.id} className="aw-card" style={{ background: 'var(--panel)' }}>
                  <div className="aw-card__body">
                    <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 'var(--space-3)' }}>
                      <div style={{ flex: 1 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-1)' }}>
                          <span style={{
                            fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase',
                            padding: '2px 6px', borderRadius: 'var(--radius-sm)',
                            color: proposal.risk_level === 'high' ? 'var(--status-revoked)' : proposal.risk_level === 'medium' ? 'var(--status-pending)' : 'var(--status-ok)',
                            background: proposal.risk_level === 'high' ? 'rgba(255,107,107,0.08)' : proposal.risk_level === 'medium' ? 'rgba(232,196,104,0.08)' : 'rgba(143,255,208,0.08)',
                            border: `1px solid ${proposal.risk_level === 'high' ? 'rgba(255,107,107,0.3)' : proposal.risk_level === 'medium' ? 'rgba(232,196,104,0.3)' : 'rgba(143,255,208,0.3)'}`,
                          }}>
                            {proposal.risk_level ?? 'low'} risk
                          </span>
                          <StatusPill
                            status={proposal.status === 'approved' ? 'live' : proposal.status === 'rejected' ? 'revoked' : 'pending'}
                            label={proposal.status ?? 'pending'}
                          />
                        </div>
                        <p style={{ fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)', margin: 0 }}>
                          {proposal.description ?? proposal.summary ?? JSON.stringify(proposal).slice(0, 150)}
                        </p>
                      </div>
                      {(!proposal.status || proposal.status === 'pending') && (
                        <div style={{ display: 'flex', gap: 'var(--space-2)', flexShrink: 0 }}>
                          <SealedButton size="sm" onClick={() => handleDecide(proposal.id, 'approved')} iconLeft={<CheckCircle size={12} />}>
                            Approve
                          </SealedButton>
                          <FrameButton size="sm" onClick={() => handleDecide(proposal.id, 'rejected')} iconLeft={<XCircle size={12} />}>
                            Reject
                          </FrameButton>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Performance Trends */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <BarChart3 size={14} />
            Performance Trends
          </span>
        </div>
        <div className="aw-card__body">
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 'var(--space-4)' }}>
            <div>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Success Rate</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginTop: 'var(--space-1)' }}>
                <TrendingUp size={14} style={{ color: 'var(--status-ok)' }} />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '18px', fontWeight: 500, color: 'var(--text)' }}>
                  {analytics?.success_rate ? `${(analytics.success_rate * 100).toFixed(1)}%` : '—'}
                </span>
              </div>
            </div>
            <div>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Avg Latency</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginTop: 'var(--space-1)' }}>
                <Clock size={14} style={{ color: 'var(--text-dim)' }} />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '18px', fontWeight: 500, color: 'var(--text)' }}>
                  {analytics?.avg_latency_ms ? `${(analytics.avg_latency_ms / 1000).toFixed(1)}s` : '—'}
                </span>
              </div>
            </div>
            <div>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Total Executions</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginTop: 'var(--space-1)' }}>
                <BarChart3 size={14} style={{ color: 'var(--text-dim)' }} />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '18px', fontWeight: 500, color: 'var(--text)' }}>
                  {analytics?.total_executions ?? '—'}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
