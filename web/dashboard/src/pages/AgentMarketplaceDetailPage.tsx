import { agentApi, type MarketplaceAgent } from '@/api/agent';
import { marketplaceUnifiedApi } from '@/api/marketplace-unified';
import {
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Modal,
} from '@/components/containment';
import { ReviewsList } from '@/components/marketplace/ReviewsList';
import { usePageTitle } from '@/hooks/usePageTitle';
import { useAuthStore } from '@/stores/authStore';
import {
  ArrowLeft,
  Bot,
  CheckCircle,
  Loader2,
  Shield,
  Sparkles,
  Star,
  TrendingUp,
  Wallet,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { toast } from 'sonner';

const PRICING_MODEL_LABELS: Record<string, string> = {
  free: 'Free',
  per_call: 'Per Call',
  subscription: 'Subscription',
  revenue_share: 'Revenue Share',
  tiered: 'Tiered',
  dynamic: 'Dynamic',
  auction: 'Auction',
};

const LISTING_TYPE_LABELS: Record<string, string> = {
  worker: 'Worker',
  manager: 'Manager',
  infrastructure: 'Infrastructure',
};

export function AgentMarketplaceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const isAuthenticated = useAuthStore((s) => !!s.user);

  const [agent, setAgent] = useState<MarketplaceAgent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('overview');

  const [hireDialog, setHireDialog] = useState(false);
  const [hireTaskType, setHireTaskType] = useState('');
  const [hireBudget, setHireBudget] = useState('');
  const [hiring, setHiring] = useState(false);

  const [reviewDialog, setReviewDialog] = useState(false);
  const [reviewRating, setReviewRating] = useState(5);
  const [reviewText, setReviewText] = useState('');
  const [submittingReview, setSubmittingReview] = useState(false);
  const [reviewRefreshKey, setReviewRefreshKey] = useState(0);

  usePageTitle(agent ? `Marketplace / ${agent.name}` : 'Marketplace');

  const loadAgent = async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const response = await agentApi.searchMarketplaceAgents({ agent_id: id, limit: 1 });
      const found =
        response.agents.find((a) => a.id === id || a.agentId === id) ?? response.agents[0];
      if (found) {
        setAgent(found);
        return;
      }
    } catch {
      /* fall through */
    }

    try {
      const res = await agentApi.getAgent(id);
      const a = res.agent as Record<string, unknown>;
      setAgent({
        id: (a.id as string) ?? id,
        agentId: (a.agentId as string) ?? id,
        name: (a.name as string) ?? 'Unknown Agent',
        description: (a.description as string) ?? '',
        listingType: 'worker',
        pricingModel: 'free',
        ratingScore: 0,
        totalCalls: 0,
        roiScore: 0,
        rankScore: 0,
        walletBalanceUsd: 0,
      });
    } catch {
      setError('Agent not found');
    }
  };

  useEffect(() => {
    loadAgent().finally(() => setLoading(false));
  }, [id]);

  const handleHire = async () => {
    if (!agent) return;
    setHiring(true);
    try {
      await agentApi.hireAgent({
        agent_id: agent.agentId,
        task_type: hireTaskType || 'general',
        budget_usd: hireBudget ? parseFloat(hireBudget) : undefined,
      });
      toast.success(`Successfully hired ${agent.name}`);
      setHireDialog(false);
      setHireTaskType('');
      setHireBudget('');
    } catch {
      toast.error('Failed to hire agent');
    } finally {
      setHiring(false);
    }
  };

  const handleSubmitReview = async () => {
    if (!agent) return;
    setSubmittingReview(true);
    try {
      await marketplaceUnifiedApi.rateAgent(agent.agentId, reviewRating, reviewText || undefined);
      toast.success('Review submitted');
      setReviewDialog(false);
      setReviewText('');
      setReviewRating(5);
      setReviewRefreshKey((k) => k + 1);
    } catch {
      toast.error('Failed to submit review');
    } finally {
      setSubmittingReview(false);
    }
  };

  const formatPrice = (agent: MarketplaceAgent) => {
    if (agent.pricingModel === 'free') return 'Free';
    if (agent.pricingModel === 'per_call' && agent.pricePerCall)
      return `$${agent.pricePerCall.toFixed(4)}/call`;
    if (agent.pricingModel === 'subscription' && agent.subscriptionMonthlyUsd)
      return `$${agent.subscriptionMonthlyUsd.toFixed(2)}/mo`;
    return PRICING_MODEL_LABELS[agent.pricingModel] ?? agent.pricingModel;
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 'var(--space-3)', padding: 'var(--space-9) 0', color: 'var(--text-faint)' }}>
        <Loader2 style={{ width: 32, height: 32, animation: 'spin 1s linear infinite' }} />
        <p style={{ fontSize: 13 }}>Loading agent...</p>
      </div>
    );
  }

  if (error || !agent) {
    return (
      <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
        <Link to="/marketplace?type=agents" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', fontSize: 13, color: 'var(--text-faint)', textDecoration: 'none', width: 'fit-content' }}>
          <ArrowLeft style={{ width: 14, height: 14 }} />
          Back to Marketplace
        </Link>
        <Chamber>
          <h2 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--status-revoked)', marginBottom: 'var(--space-2)' }}>Agent Not Found</h2>
          <p style={{ color: 'var(--text-dim)' }}>{error ?? 'The agent you are looking for does not exist.'}</p>
        </Chamber>
      </div>
    );
  }

  const detailTabs = [
    { value: 'overview', label: 'Overview' },
    { value: 'capabilities', label: 'Capabilities' },
    { value: 'pricing', label: 'Pricing' },
    { value: 'reviews', label: 'Reviews' },
  ];

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: 'var(--space-3) var(--space-4)',
    background: 'var(--panel-raised)',
    border: '1px solid var(--steel)',
    borderRadius: 'var(--radius)',
    color: 'var(--text)',
    fontFamily: 'var(--font-body)',
    fontSize: 13,
    outline: 'none',
  };

  return (
    <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      {/* Back link */}
      <Link to="/marketplace?type=agents" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', fontSize: 13, color: 'var(--text-faint)', textDecoration: 'none', width: 'fit-content' }}>
        <ArrowLeft style={{ width: 14, height: 14 }} />
        Back to Marketplace
      </Link>

      {/* Hero Chamber */}
      <Chamber ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="AGENT" secondary={agent.agentId} position="top-right" />

        {/* Header */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)', marginBottom: 'var(--space-5)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
            <div style={{ width: 56, height: 56, borderRadius: 'var(--radius-lg)', background: 'rgba(143,255,208,0.08)', border: '1px solid rgba(143,255,208,0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
              <Bot style={{ width: 28, height: 28, color: 'var(--status-ok)' }} />
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, letterSpacing: '-0.005em', color: 'var(--text)', display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
                {agent.name}
                {agent.deterministicVerified && (
                  <StatusPill status="live" label="Verified" />
                )}
                {agent.isOfficial && (
                  <StatusPill status="pending" label="Official" />
                )}
              </h1>
              <p style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-faint)', marginTop: 'var(--space-1)' }}>{agent.agentId}</p>
            </div>
          </div>

          {/* Badges */}
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-2)' }}>
            <span style={{ display: 'inline-flex', alignItems: 'center', padding: '3px var(--space-3)', borderRadius: 'var(--radius-sm)', border: '1px solid var(--panel-edge)', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, letterSpacing: '0.04em', textTransform: 'uppercase', color: 'var(--text-dim)', background: 'var(--panel-raised)' }}>
              {LISTING_TYPE_LABELS[agent.listingType] ?? agent.listingType}
            </span>
            <span style={{ display: 'inline-flex', alignItems: 'center', padding: '3px var(--space-3)', borderRadius: 'var(--radius-sm)', border: '1px solid rgba(232,196,104,0.3)', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, letterSpacing: '0.04em', textTransform: 'uppercase', color: 'var(--status-pending)', background: 'rgba(232,196,104,0.06)' }}>
              {PRICING_MODEL_LABELS[agent.pricingModel] ?? agent.pricingModel}
            </span>
            <span style={{ display: 'inline-flex', alignItems: 'center', padding: '3px var(--space-3)', borderRadius: 'var(--radius-sm)', border: '1px solid var(--panel-edge)', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, color: 'var(--status-ok)', background: 'rgba(143,255,208,0.06)' }}>
              {formatPrice(agent)}
            </span>
          </div>

          {/* Actions */}
          <div style={{ display: 'flex', gap: 'var(--space-3)' }}>
            <SealedButton onClick={() => setHireDialog(true)} iconLeft={<Bot style={{ width: 14, height: 14 }} />}>
              Hire Agent
            </SealedButton>
            {isAuthenticated && (
              <FrameButton onClick={() => setReviewDialog(true)} iconLeft={<Star style={{ width: 14, height: 14 }} />}>
                Write Review
              </FrameButton>
            )}
          </div>
        </div>

        {/* Stats GaugeStrip */}
        <GaugeStrip>
          <Gauge data={{ value: (agent.ratingScore ?? 0).toFixed(1), label: 'Rating' }} isFirst />
          <Gauge data={{ value: (agent.roiScore ?? 0).toFixed(1), label: 'ROI Score' }} />
          <Gauge data={{ value: agent.totalCalls.toLocaleString(), label: 'Total Calls' }} />
          {agent.rankScore !== undefined && (
            <Gauge data={{ value: agent.rankScore.toFixed(2), label: 'Rank' }} />
          )}
        </GaugeStrip>
      </Chamber>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 'var(--space-1)', padding: 'var(--space-1)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', width: 'fit-content' }}>
        {detailTabs.map(({ value, label }) => {
          const isActive = activeTab === value;
          return (
            <button
              key={value}
              role="tab"
              aria-selected={isActive}
              onClick={() => setActiveTab(value)}
              style={{
                padding: 'var(--space-2) var(--space-4)',
                fontSize: 13,
                fontWeight: isActive ? 500 : 400,
                fontFamily: 'var(--font-body)',
                color: isActive ? 'var(--text)' : 'var(--text-faint)',
                background: isActive ? 'var(--panel-raised)' : 'transparent',
                borderRadius: 'var(--radius-sm)',
                border: 'none',
                cursor: 'pointer',
                boxShadow: isActive ? '0 1px 3px rgba(0,0,0,0.2)' : 'none',
                transition: 'all var(--duration-fast) var(--ease-out)',
                whiteSpace: 'nowrap',
              }}
            >
              {label}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
        {activeTab === 'overview' && (
          <>
            <Chamber nested>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>About</h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, fontSize: 15 }}>
                {agent.description || 'No description available.'}
              </p>
            </Chamber>

            {isAuthenticated &&
              (agent.walletBalanceUsd !== undefined ||
                agent.hiringHistoryCount !== undefined) && (
                <Chamber nested>
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>Agent Stats</h3>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-7)' }}>
                    {agent.walletBalanceUsd !== undefined && (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                        <Wallet style={{ width: 20, height: 20, color: 'var(--status-ok)' }} />
                        <div>
                          <p style={{ fontSize: 13, color: 'var(--text-faint)' }}>Wallet Balance</p>
                          <p style={{ fontSize: 17, fontWeight: 500, color: 'var(--text)' }}>${agent.walletBalanceUsd.toFixed(2)}</p>
                        </div>
                      </div>
                    )}
                    {agent.hiringHistoryCount !== undefined && (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                        <CheckCircle style={{ width: 20, height: 20, color: 'var(--status-ok)' }} />
                        <div>
                          <p style={{ fontSize: 13, color: 'var(--text-faint)' }}>Total Hires</p>
                          <p style={{ fontSize: 17, fontWeight: 500, color: 'var(--text)' }}>{agent.hiringHistoryCount}</p>
                        </div>
                      </div>
                    )}
                  </div>
                </Chamber>
              )}
          </>
        )}

        {activeTab === 'capabilities' && (
          <Chamber nested>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Capabilities</h3>
            <p style={{ color: 'var(--text-dim)', fontSize: 13, marginBottom: 'var(--space-4)' }}>Skills and features this agent provides</p>
            {agent.capabilities && agent.capabilities.length > 0 ? (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-2)' }}>
                {agent.capabilities.map((cap) => (
                  <span key={cap} style={{ display: 'inline-flex', alignItems: 'center', padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius-sm)', background: 'var(--panel)', color: 'var(--text-dim)', border: '1px solid var(--panel-edge)', fontFamily: 'var(--font-mono)', fontSize: 12 }}>
                    {cap}
                  </span>
                ))}
              </div>
            ) : (
              <p style={{ color: 'var(--text-faint)', fontSize: 13 }}>No capabilities listed</p>
            )}
          </Chamber>
        )}

        {activeTab === 'pricing' && (
          <Chamber nested>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>Pricing Details</h3>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 'var(--space-4) var(--space-6)' }}>
              {[
                ['Pricing Model', PRICING_MODEL_LABELS[agent.pricingModel] ?? agent.pricingModel],
                ...(agent.pricePerCall ? [['Price per Call', `$${agent.pricePerCall.toFixed(4)}`]] : []),
                ...(agent.subscriptionMonthlyUsd ? [['Monthly Subscription', `$${agent.subscriptionMonthlyUsd.toFixed(2)}`]] : []),
                ...(agent.revenueSharePercent ? [['Revenue Share', `${agent.revenueSharePercent}%`]] : []),
              ].map(([label, value]) => (
                <div key={label}>
                  <p style={{ fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)', marginBottom: 'var(--space-1)' }}>{label}</p>
                  <p style={{ fontWeight: 500, color: 'var(--text)', fontSize: 14 }}>{value}</p>
                </div>
              ))}
            </div>
          </Chamber>
        )}

        {activeTab === 'reviews' && (
          <Chamber nested>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-4)' }}>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)' }}>Reviews</h3>
              {isAuthenticated && (
                <FrameButton size="sm" onClick={() => setReviewDialog(true)} iconLeft={<Star style={{ width: 12, height: 12 }} />}>
                  Write Review
                </FrameButton>
              )}
            </div>
            <ReviewsList itemType="agent" itemId={agent.agentId} refreshKey={reviewRefreshKey} />
          </Chamber>
        )}
      </div>

      {/* Hire Dialog */}
      <Modal open={hireDialog} onClose={() => setHireDialog(false)} title={`Hire ${agent.name}`}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          <div>
            <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Task Type</label>
            <input
              placeholder="e.g. code_generation, analysis"
              value={hireTaskType}
              onChange={(e) => setHireTaskType(e.target.value)}
              style={inputStyle}
            />
          </div>
          <div>
            <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Budget (USD)</label>
            <input
              type="number"
              placeholder="Optional"
              value={hireBudget}
              onChange={(e) => setHireBudget(e.target.value)}
              style={inputStyle}
            />
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)', paddingTop: 'var(--space-2)' }}>
            <FrameButton onClick={() => setHireDialog(false)}>Cancel</FrameButton>
            <SealedButton onClick={handleHire} disabled={hiring || !hireTaskType.trim()} iconLeft={hiring ? <Loader2 style={{ width: 14, height: 14, animation: 'spin 1s linear infinite' }} /> : <Bot style={{ width: 14, height: 14 }} />}>
              {hiring ? 'Hiring...' : 'Hire Agent'}
            </SealedButton>
          </div>
        </div>
      </Modal>

      {/* Review Dialog */}
      <Modal open={reviewDialog} onClose={() => setReviewDialog(false)} title={`Review ${agent.name}`}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          <div>
            <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Rating</label>
            <div style={{ display: 'flex', gap: 'var(--space-1)' }}>
              {[1, 2, 3, 4, 5].map((s) => (
                <button key={s} onClick={() => setReviewRating(s)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 4 }}>
                  <Star style={{ width: 24, height: 24, color: s <= reviewRating ? '#eab308' : 'var(--text-faint)', fill: s <= reviewRating ? '#eab308' : 'none' }} />
                </button>
              ))}
            </div>
          </div>
          <div>
            <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Review</label>
            <textarea
              placeholder="Share your experience..."
              value={reviewText}
              onChange={(e) => setReviewText(e.target.value)}
              rows={4}
              style={{ ...inputStyle, resize: 'vertical' as const }}
            />
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)', paddingTop: 'var(--space-2)' }}>
            <FrameButton onClick={() => setReviewDialog(false)}>Cancel</FrameButton>
            <SealedButton onClick={handleSubmitReview} disabled={submittingReview} iconLeft={submittingReview ? <Loader2 style={{ width: 14, height: 14, animation: 'spin 1s linear infinite' }} /> : <Star style={{ width: 14, height: 14 }} />}>
              {submittingReview ? 'Submitting...' : 'Submit Review'}
            </SealedButton>
          </div>
        </div>
      </Modal>
    </div>
  );
}

export default AgentMarketplaceDetailPage;
