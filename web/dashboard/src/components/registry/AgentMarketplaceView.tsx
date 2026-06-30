import { useState } from 'react';
import { agentApi, type MarketplaceAgent, type MarketplaceAgentSearchParams } from '@/api/agent';
import { Chamber, SealedButton, FrameButton, StatusPill, Modal } from '@/components/containment';
import { useAuthStore } from '@/stores/authStore';
import {
  Bot, CheckCircle, Filter, Loader2, Search, Shield, Sparkles, Star, TrendingUp, Wallet,
} from 'lucide-react';
import { toast } from 'sonner';

interface AgentMarketplaceViewProps {
  variant?: 'standalone' | 'embedded';
  onAgentSelect?: (agent: MarketplaceAgent) => void;
}

const PRICING_MODEL_LABELS: Record<string, string> = {
  free: 'Free', per_call: 'Per Call', subscription: 'Subscription',
  revenue_share: 'Revenue Share', tiered: 'Tiered', dynamic: 'Dynamic', auction: 'Auction',
};

const LISTING_TYPE_LABELS: Record<string, string> = {
  worker: 'Worker', manager: 'Manager', infrastructure: 'Infrastructure',
};

export function AgentMarketplaceView({ variant = 'embedded', onAgentSelect }: AgentMarketplaceViewProps) {
  const isAuthenticated = useAuthStore((s) => !!s.user);
  const [agents, setAgents] = useState<MarketplaceAgent[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [limit] = useState(20);
  const [offset, setOffset] = useState(0);

  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState<MarketplaceAgentSearchParams>({});
  const [showFilters, setShowFilters] = useState(false);

  const [hireDialog, setHireDialog] = useState<MarketplaceAgent | null>(null);
  const [hireTaskType, setHireTaskType] = useState('');
  const [hireBudget, setHireBudget] = useState('');
  const [hiring, setHiring] = useState(false);

  const loadAgents = async (params?: MarketplaceAgentSearchParams) => {
    setLoading(true); setError(null);
    try {
      const response = await agentApi.searchMarketplaceAgents({ ...params, limit, offset });
      setAgents(response.agents); setTotal(response.total); setHasMore(response.has_more);
    } catch { setError('Failed to load agents'); } finally { setLoading(false); }
  };

  const handleSearch = () => { setOffset(0); loadAgents({ ...filters, capabilities: searchQuery ? searchQuery.split(',').map(s => s.trim()) : undefined }); };
  const handlePageChange = (newOffset: number) => { setOffset(newOffset); loadAgents({ ...filters, capabilities: searchQuery ? searchQuery.split(',').map(s => s.trim()) : undefined }); };

  const handleHire = async () => {
    if (!hireDialog) return; setHiring(true);
    try {
      await agentApi.hireAgent({ agent_id: hireDialog.agentId, task_type: hireTaskType || 'general', budget_usd: hireBudget ? parseFloat(hireBudget) : undefined });
      toast.success(`Successfully hired ${hireDialog.name}`); setHireDialog(null); setHireTaskType(''); setHireBudget('');
    } catch { toast.error('Failed to hire agent'); } finally { setHiring(false); }
  };

  const formatPrice = (agent: MarketplaceAgent) => {
    if (agent.pricingModel === 'free') return 'Free';
    if (agent.pricingModel === 'per_call' && agent.pricePerCall) return `$${agent.pricePerCall.toFixed(4)}/call`;
    if (agent.pricingModel === 'subscription' && agent.subscriptionMonthlyUsd) return `$${agent.subscriptionMonthlyUsd.toFixed(2)}/mo`;
    return PRICING_MODEL_LABELS[agent.pricingModel] ?? agent.pricingModel;
  };

  const inputStyle: React.CSSProperties = { width: '100%', padding: 'var(--space-3) var(--space-4)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', color: 'var(--text)', fontFamily: 'var(--font-body)', fontSize: 13, outline: 'none', transition: 'border-color var(--duration-fast) var(--ease-out)' };
  const selectStyle: React.CSSProperties = { ...inputStyle, cursor: 'pointer', appearance: 'none' as const };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
      {/* Search and Filters */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
        <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <Search style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', width: 14, height: 14, color: 'var(--text-faint)', pointerEvents: 'none' }} />
            <input placeholder="Search by capabilities (e.g. code_generation, analysis)" value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} onKeyDown={(e) => e.key === 'Enter' && handleSearch()} style={{ ...inputStyle, paddingLeft: 36 }} />
          </div>
          <SealedButton onClick={handleSearch} disabled={loading}>{loading ? <Loader2 style={{ width: 14, height: 14, animation: 'spin 1s linear infinite' }} /> : 'Search'}</SealedButton>
          <FrameButton onClick={() => setShowFilters(!showFilters)} iconLeft={<Filter style={{ width: 14, height: 14 }} />}>Filters</FrameButton>
        </div>

        {showFilters && (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-4)', padding: 'var(--space-4)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)' }}>
            {[
              { label: 'Pricing Model', value: filters.pricing_model ?? '', options: Object.entries(PRICING_MODEL_LABELS), onChange: (v: string) => setFilters(p => ({ ...p, pricing_model: v || undefined })) },
              { label: 'Listing Type', value: filters.listing_types?.[0] ?? '', options: Object.entries(LISTING_TYPE_LABELS), onChange: (v: string) => setFilters(p => ({ ...p, listing_types: v ? [v] : undefined })) },
              { label: 'Sort By', value: filters.sort_by ?? 'rank_score', options: [['rank_score', 'Rank Score'], ['rating_score', 'Rating'], ['price_per_call', 'Price'], ['total_calls', 'Popularity']], onChange: (v: string) => setFilters(p => ({ ...p, sort_by: v as MarketplaceAgentSearchParams['sort_by'] })) },
            ].map(({ label, value, options, onChange }) => (
              <div key={label} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                <span style={{ fontSize: 11, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{label}</span>
                <select value={value} onChange={(e) => onChange(e.target.value)} style={{ ...selectStyle, minWidth: 140 }}>
                  <option value="">Any</option>
                  {options.map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Results info */}
      <p style={{ fontSize: 13, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)' }}>{loading ? 'Loading...' : `${total} agents found`}</p>

      {/* Error */}
      {error && <Chamber><div style={{ textAlign: 'center', padding: 'var(--space-7) 0', color: 'var(--status-revoked)' }}>{error}</div></Chamber>}

      {/* Empty */}
      {!loading && agents.length === 0 && !error && (
        <Chamber>
          <div style={{ textAlign: 'center', padding: 'var(--space-7) 0' }}>
            <Bot style={{ width: 48, height: 48, color: 'var(--text-faint)', margin: '0 auto var(--space-4)' }} />
            <p style={{ color: 'var(--text-dim)' }}>No agents found matching your criteria</p>
          </div>
        </Chamber>
      )}

      {/* Agent Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 'var(--space-4)' }}>
        {agents.map((agent) => (
          <Chamber nested key={agent.id} style={{ cursor: 'pointer', display: 'flex', flexDirection: 'column' }} onClick={() => onAgentSelect?.(agent)}>
            {/* Header */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-3)' }}>
              <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: 'linear-gradient(135deg, rgba(143,255,208,0.1), rgba(159,216,255,0.1))', border: '1px solid var(--panel-edge)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Bot style={{ width: 18, height: 18, color: 'var(--status-ok)' }} />
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
                  <span style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{agent.name}</span>
                  {agent.deterministicVerified && <StatusPill status="live" label="Verified" />}
                  {agent.isOfficial && <StatusPill status="pending" label="Official" />}
                </div>
                <p style={{ fontSize: 11, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{agent.agentId}</p>
              </div>
            </div>

            {/* Description */}
            <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5, marginBottom: 'var(--space-3)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden', minHeight: '2.5rem' }}>{agent.description || 'No description available'}</p>

            {/* Capabilities */}
            {agent.capabilities && agent.capabilities.length > 0 && (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-1)', marginBottom: 'var(--space-3)' }}>
                {agent.capabilities.slice(0, 4).map((cap) => (
                  <span key={cap} style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)' }}>{cap}</span>
                ))}
                {agent.capabilities.length > 4 && <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', color: 'var(--text-faint)' }}>+{agent.capabilities.length - 4}</span>}
              </div>
            )}

            {/* Pricing */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-3)' }}>
              <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', border: '1px solid var(--panel-edge)', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase' }}>{LISTING_TYPE_LABELS[agent.listingType] ?? agent.listingType}</span>
              <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', color: 'var(--status-ok)', fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{formatPrice(agent)}</span>
            </div>

            {/* Stats */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 'var(--space-2)', textAlign: 'center', marginBottom: 'var(--space-3)' }}>
              {[
                { icon: Star, value: (agent.ratingScore ?? 0).toFixed(1), label: 'Rating', color: 'var(--status-pending)' },
                { icon: TrendingUp, value: (agent.roiScore ?? 0).toFixed(1), label: 'ROI', color: 'var(--status-ok)' },
                { icon: Bot, value: agent.totalCalls.toLocaleString(), label: 'Calls', color: 'var(--foil-a)' },
              ].map(({ icon: Icon, value, label, color }) => (
                <div key={label} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: 'var(--space-2)', background: 'var(--panel)', borderRadius: 'var(--radius)' }}>
                  <Icon style={{ width: 11, height: 11, color, marginBottom: 2 }} />
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 500, color: 'var(--text)' }}>{value}</span>
                  <span style={{ fontSize: 10, color: 'var(--text-faint)' }}>{label}</span>
                </div>
              ))}
            </div>

            {/* Rank Score */}
            {agent.rankScore !== undefined && (
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: 11, color: 'var(--text-faint)', marginBottom: 'var(--space-3)' }}>
                <span>Rank Score</span>
                <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--status-ok)' }}>{agent.rankScore.toFixed(2)}</span>
              </div>
            )}

            {/* Hire Button */}
            <div style={{ marginTop: 'auto', paddingTop: 'var(--space-3)' }}>
              <SealedButton style={{ width: '100%' }} onClick={(e) => { e.stopPropagation(); setHireDialog(agent); }}>Hire Agent</SealedButton>
            </div>
          </Chamber>
        ))}
      </div>

      {/* Pagination */}
      {(hasMore || offset > 0) && (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 'var(--space-3)' }}>
          <FrameButton size="sm" onClick={() => handlePageChange(Math.max(0, offset - limit))} disabled={offset === 0}>Previous</FrameButton>
          <span style={{ fontSize: 13, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)' }}>{offset + 1}–{Math.min(offset + agents.length, total)} of {total}</span>
          <FrameButton size="sm" onClick={() => handlePageChange(offset + limit)} disabled={!hasMore}>Next</FrameButton>
        </div>
      )}

      {/* Hire Modal */}
      <Modal open={!!hireDialog} onClose={() => setHireDialog(null)} title={hireDialog?.name ?? 'Hire Agent'}>
        {hireDialog && (
          <>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', padding: 'var(--space-3)', background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-4)' }}>
              <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', border: '1px solid var(--panel-edge)', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase' }}>{LISTING_TYPE_LABELS[hireDialog.listingType]}</span>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 600, color: 'var(--text)' }}>{formatPrice(hireDialog)}</span>
              <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 'var(--space-1)' }}>
                <Star style={{ width: 11, height: 11, color: 'var(--status-pending)' }} />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 500 }}>{(hireDialog.ratingScore ?? 0).toFixed(1)}</span>
              </div>
            </div>

            <div style={{ marginBottom: 'var(--space-4)' }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Task Type</label>
              <input placeholder="e.g. code_generation, analysis, data_processing" value={hireTaskType} onChange={(e) => setHireTaskType(e.target.value)} style={inputStyle} />
              <p style={{ fontSize: 11, color: 'var(--text-faint)', marginTop: 'var(--space-1)' }}>Specify the type of task you want this agent to perform</p>
            </div>

            <div style={{ marginBottom: 'var(--space-5)' }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Budget (USD)</label>
              <div style={{ position: 'relative' }}>
                <span style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-faint)', fontSize: 13 }}>$</span>
                <input type="number" placeholder="0.00" value={hireBudget} onChange={(e) => setHireBudget(e.target.value)} min="0" step="0.01" style={{ ...inputStyle, paddingLeft: 28 }} />
              </div>
              <p style={{ fontSize: 11, color: 'var(--text-faint)', marginTop: 'var(--space-1)' }}>Maximum amount you're willing to pay. Leave empty for no limit.</p>
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)' }}>
              <FrameButton onClick={() => setHireDialog(null)}>Cancel</FrameButton>
              <SealedButton onClick={handleHire} disabled={hiring || !hireTaskType.trim()} iconLeft={hiring ? <Loader2 style={{ width: 14, height: 14, animation: 'spin 1s linear infinite' }} /> : <CheckCircle style={{ width: 14, height: 14 }} />}>
                {hiring ? 'Hiring...' : 'Hire Agent'}
              </SealedButton>
            </div>
          </>
        )}
      </Modal>

      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}

export default AgentMarketplaceView;
