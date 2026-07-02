import { useState, useEffect, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { CheckCircle, Loader2, Star } from 'lucide-react';
import { MetaTags } from '@/components/seo/MetaTags';
import { PageGrid, Chamber, CornerBrace, SealedButton, FrameButton, Modal } from '@/components/containment';
import { MarketplaceSearchBar } from '@/components/marketplace/MarketplaceSearchBar';
import { MarketplaceGrid } from '@/components/marketplace/MarketplaceGrid';
import { ReviewsList } from '@/components/marketplace/ReviewsList';
import { marketplaceUnifiedApi } from '@/api/marketplace-unified';
import { marketplaceApi } from '@/api/marketplace';
import { agentApi } from '@/api/agent';
import { useAuthStore } from '@/stores/authStore';
import { toast } from 'sonner';
import type { UnifiedItem, MarketplaceItemType } from '@/components/marketplace/types';

export function MarketplacePage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => !!s.user);

  const [query, setQuery] = useState(searchParams.get('q') ?? '');
  const [selectedType, setSelectedType] = useState<MarketplaceItemType | ''>((searchParams.get('type') as MarketplaceItemType) ?? '');
  const [items, setItems] = useState<UnifiedItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [limit] = useState(20);
  const [installingId, setInstallingId] = useState<string | null>(null);

  const [hireDialog, setHireDialog] = useState<UnifiedItem | null>(null);
  const [hireTaskType, setHireTaskType] = useState('');
  const [hireBudget, setHireBudget] = useState('');
  const [hiring, setHiring] = useState(false);

  const [rateDialog, setRateDialog] = useState<UnifiedItem | null>(null);
  const [rateValue, setRateValue] = useState(5);
  const [rateReview, setRateReview] = useState('');
  const [rating, setRating] = useState(false);
  const [reviewRefreshKey, setReviewRefreshKey] = useState(0);

  const doSearch = useCallback(async (q: string, type: MarketplaceItemType | '', off: number) => {
    setLoading(true);
    setError(null);
    try {
      const resp = await marketplaceUnifiedApi.search({ q, type: type || undefined, limit, offset: off });
      setItems(resp.items ?? []);
      setTotal(resp.total);
      setHasMore(resp.has_more);
    } catch {
      setError('Failed to load marketplace results');
    } finally {
      setLoading(false);
    }
  }, [limit]);

  useEffect(() => {
    doSearch(query, selectedType, offset);
  }, [doSearch, query, selectedType, offset]);

  const updateURL = (q: string, type: MarketplaceItemType | '') => {
    const params = new URLSearchParams();
    if (q) params.set('q', q);
    if (type) params.set('type', type);
    setSearchParams(params, { replace: true });
  };

  const handleSearch = () => {
    setOffset(0);
    updateURL(query, selectedType);
  };

  const handleTypeChange = (type: MarketplaceItemType | '') => {
    setSelectedType(type);
    setOffset(0);
    updateURL(query, type);
  };

  const handlePageChange = (newOffset: number) => {
    setOffset(newOffset);
  };

  const handleHireAgent = (item: UnifiedItem) => {
    if (!isAuthenticated) { toast.error('Please sign in to hire agents'); return; }
    setHireDialog(item);
  };

  const doHire = async () => {
    if (!hireDialog) return;
    setHiring(true);
    try {
      await agentApi.hireAgent({ agent_id: hireDialog.id, task_type: hireTaskType || 'general', budget_usd: hireBudget ? parseFloat(hireBudget) : undefined });
      toast.success(`Successfully hired ${hireDialog.name}`);
      setHireDialog(null);
      setHireTaskType('');
      setHireBudget('');
    } catch {
      toast.error('Failed to hire agent');
    } finally {
      setHiring(false);
    }
  };

  const handleRateAgent = (item: UnifiedItem) => {
    if (!isAuthenticated) { toast.error('Please sign in to rate agents'); return; }
    setRateDialog(item);
    setRateValue(5);
    setRateReview('');
  };

  const doRate = async () => {
    if (!rateDialog) return;
    setRating(true);
    try {
      if (rateDialog.type === 'function') {
        await marketplaceUnifiedApi.rateFunction(rateDialog.id, rateValue, rateReview || undefined);
      } else if (rateDialog.type === 'extension') {
        await marketplaceApi.rate(rateDialog.id, rateValue, rateReview || undefined);
      } else {
        await marketplaceUnifiedApi.rateAgent(rateDialog.id, rateValue, rateReview || undefined);
      }
      toast.success(`Rated ${rateDialog.name}`);
      setReviewRefreshKey((k) => k + 1);
      doSearch(query, selectedType, offset);
    } catch {
      toast.error('Failed to submit rating');
    } finally {
      setRating(false);
    }
  };

  const handleRateFunction = (item: UnifiedItem) => {
    if (!isAuthenticated) { toast.error('Please sign in to rate functions'); return; }
    setRateDialog(item);
    setRateValue(5);
    setRateReview('');
  };

  const handleRateExtension = (item: UnifiedItem) => {
    if (!isAuthenticated) { toast.error('Please sign in to rate extensions'); return; }
    setRateDialog(item);
    setRateValue(5);
    setRateReview('');
  };

  const handleInstallExtension = async (item: UnifiedItem) => {
    if (!isAuthenticated) { toast.error('Please sign in to install extensions'); return; }
    setInstallingId(item.id);
    try {
      await marketplaceApi.install(item.id);
      toast.success(`Installed ${item.name}`);
    } catch {
      toast.error('Failed to install extension');
    } finally {
      setInstallingId(null);
    }
  };

  const handleDeployFunction = (item: UnifiedItem) => {
    navigate(`/functions/${item.id}`);
  };

  const handleViewItem = (item: UnifiedItem) => {
    switch (item.type) {
      case 'agent':
        navigate(`/marketplace/agents/${item.id}`);
        break;
      case 'extension':
        navigate(`/marketplace/extensions/${item.id}`);
        break;
      case 'function':
        navigate(`/functions/${item.id}`);
        break;
    }
  };

  const inputStyle: React.CSSProperties = { width: '100%', padding: 'var(--space-3) var(--space-4)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', color: 'var(--text)', fontFamily: 'var(--font-body)', fontSize: 13, outline: 'none' };

  return (
    <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <PageGrid />
      <MetaTags
        title="Marketplace | FunctionFly"
        description="Discover and deploy AI agents, extensions, and serverless functions."
        keywords={['marketplace', 'AI agents', 'extensions', 'functions', 'serverless']}
        url={`${window.location.origin}/marketplace`}
        type="website"
      />

      <div>
        <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, letterSpacing: '-0.005em', color: 'var(--text)' }}>Marketplace</h1>
        <p style={{ fontSize: 14, color: 'var(--text-dim)', marginTop: 'var(--space-2)' }}>
          Discover and deploy AI agents, extensions, and serverless functions.
        </p>
      </div>

      <MarketplaceSearchBar
        query={query}
        selectedType={selectedType}
        onQueryChange={setQuery}
        onTypeChange={handleTypeChange}
        onSearch={handleSearch}
      />

      <MarketplaceGrid
        items={items}
        loading={loading}
        error={error}
        total={total}
        hasMore={hasMore}
        offset={offset}
        limit={limit}
        installingId={installingId}
        onPageChange={handlePageChange}
        onHireAgent={handleHireAgent}
        onRateAgent={handleRateAgent}
        onInstallExtension={handleInstallExtension}
        onRateExtension={handleRateExtension}
        onDeployFunction={handleDeployFunction}
        onRateFunction={handleRateFunction}
        onViewItem={handleViewItem}
      />

      <Modal open={!!hireDialog} onClose={() => setHireDialog(null)} title={hireDialog?.name ?? 'Hire Agent'}>
        {hireDialog && (
          <>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', padding: 'var(--space-3)', background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-4)' }}>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 600, color: 'var(--text)' }}>{hireDialog.price}</span>
            </div>
            <div style={{ marginBottom: 'var(--space-4)' }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Task Type</label>
              <input placeholder="e.g. code_generation, analysis" value={hireTaskType} onChange={(e) => setHireTaskType(e.target.value)} style={inputStyle} />
            </div>
            <div style={{ marginBottom: 'var(--space-5)' }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Budget (USD)</label>
              <div style={{ position: 'relative' }}>
                <span style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-faint)', fontSize: 13 }}>$</span>
                <input type="number" placeholder="0.00" value={hireBudget} onChange={(e) => setHireBudget(e.target.value)} min="0" step="0.01" style={{ ...inputStyle, paddingLeft: 28 }} />
              </div>
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)' }}>
              <FrameButton onClick={() => setHireDialog(null)}>Cancel</FrameButton>
              <SealedButton onClick={doHire} disabled={hiring || !hireTaskType.trim()} iconLeft={hiring ? <Loader2 style={{ width: 14, height: 14, animation: 'spin 1s linear infinite' }} /> : <CheckCircle style={{ width: 14, height: 14 }} />}>
                {hiring ? 'Hiring...' : 'Hire Agent'}
              </SealedButton>
            </div>
          </>
        )}
      </Modal>

      <Modal open={!!rateDialog} onClose={() => setRateDialog(null)} title={`Rate ${rateDialog?.name ?? ''}`}>
        {rateDialog && (
          <>
            <div style={{ marginBottom: 'var(--space-4)' }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Your Rating</label>
              <div style={{ display: 'flex', gap: 'var(--space-1)' }}>
                {[1, 2, 3, 4, 5].map((s) => (
                  <button key={s} onClick={() => setRateValue(s)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 4 }}>
                    <Star style={{ width: 24, height: 24, color: s <= rateValue ? '#eab308' : 'var(--text-faint)', fill: s <= rateValue ? '#eab308' : 'none' }} />
                  </button>
                ))}
              </div>
            </div>
            <div style={{ marginBottom: 'var(--space-4)' }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Review (optional)</label>
              <textarea placeholder="Share your experience..." value={rateReview} onChange={(e) => setRateReview(e.target.value)} rows={3} style={{ ...inputStyle, resize: 'vertical' as const }} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)', marginBottom: 'var(--space-4)' }}>
              <FrameButton onClick={() => setRateDialog(null)}>Cancel</FrameButton>
              <SealedButton onClick={doRate} disabled={rating} iconLeft={rating ? <Loader2 style={{ width: 14, height: 14, animation: 'spin 1s linear infinite' }} /> : <Star style={{ width: 14, height: 14 }} />}>
                {rating ? 'Submitting...' : 'Submit Rating'}
              </SealedButton>
            </div>
            <div style={{ borderTop: '1px solid var(--panel-edge)', paddingTop: 'var(--space-4)' }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>Reviews</label>
              <ReviewsList itemType={rateDialog.type} itemId={rateDialog.id} refreshKey={reviewRefreshKey} />
            </div>
          </>
        )}
      </Modal>
    </div>
  );
}

export default MarketplacePage;
