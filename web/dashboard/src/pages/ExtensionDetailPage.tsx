import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Download, Shield, Star, Loader2 } from 'lucide-react';
import { MetaTags } from '@/components/seo/MetaTags';
import { PageGrid, Chamber, SealedButton, FrameButton, Modal } from '@/components/containment';
import { ReviewsList } from '@/components/marketplace/ReviewsList';
import { marketplaceApi, type Extension } from '@/api/marketplace';
import { useAuthStore } from '@/stores/authStore';
import { toast } from 'sonner';

export function ExtensionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const isAuthenticated = useAuthStore((s) => !!s.user);
  const [extension, setExtension] = useState<Extension | null>(null);
  const [loading, setLoading] = useState(true);
  const [installing, setInstalling] = useState(false);

  const [reviewDialog, setReviewDialog] = useState(false);
  const [reviewRating, setReviewRating] = useState(5);
  const [reviewText, setReviewText] = useState('');
  const [submittingReview, setSubmittingReview] = useState(false);
  const [reviewRefreshKey, setReviewRefreshKey] = useState(0);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    marketplaceApi.get(id)
      .then((resp) => setExtension(resp.extension))
      .catch(() => toast.error('Failed to load extension'))
      .finally(() => setLoading(false));
  }, [id]);

  const handleInstall = async () => {
    if (!id) return;
    setInstalling(true);
    try {
      await marketplaceApi.install(id);
      toast.success(`Installed ${extension?.name}`);
    } catch {
      toast.error('Failed to install extension');
    } finally {
      setInstalling(false);
    }
  };

  const handleSubmitReview = async () => {
    if (!id) return;
    setSubmittingReview(true);
    try {
      await marketplaceApi.rate(id, reviewRating, reviewText || undefined);
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

  const inputStyle: React.CSSProperties = { width: '100%', padding: 'var(--space-3) var(--space-4)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', color: 'var(--text)', fontFamily: 'var(--font-body)', fontSize: 13, outline: 'none' };

  if (loading) {
    return (
      <div style={{ maxWidth: 900, margin: '0 auto', padding: 'var(--space-7)' }}>
        <p style={{ color: 'var(--text-faint)' }}>Loading...</p>
      </div>
    );
  }

  if (!extension) {
    return (
      <div style={{ maxWidth: 900, margin: '0 auto', padding: 'var(--space-7)' }}>
        <p style={{ color: 'var(--status-revoked)' }}>Extension not found</p>
        <Link to="/marketplace?type=extension" style={{ color: 'var(--brand-400)', fontSize: 13, marginTop: 'var(--space-2)', display: 'inline-block' }}>Back to Marketplace</Link>
      </div>
    );
  }

  const formatNumber = (n: number) => n >= 1000 ? (n / 1000).toFixed(1) + 'k' : n.toString();

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <PageGrid />
      <MetaTags
        title={`${extension.name} | FunctionFly Marketplace`}
        description={extension.description}
        url={`${window.location.origin}/marketplace/extensions/${id}`}
        type="website"
      />

      <Link to="/marketplace?type=extension" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)', color: 'var(--text-dim)', fontSize: 13, textDecoration: 'none' }}>
        <ArrowLeft style={{ width: 14, height: 14 }} /> Back to Marketplace
      </Link>

      <Chamber>
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-4)' }}>
          <div style={{ width: 64, height: 64, borderRadius: 'var(--radius-lg)', background: 'linear-gradient(135deg, rgba(167,139,250,0.15), rgba(139,92,246,0.15))', border: '1px solid rgba(167,139,250,0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 700, color: '#a78bfa', flexShrink: 0 }}>
            {extension.name.charAt(0)}
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-1)' }}>
              <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 28, fontWeight: 700, color: 'var(--text)' }}>{extension.name}</h1>
              {extension.verified && <Shield style={{ width: 20, height: 20, color: '#22c55e' }} />}
            </div>
            <p style={{ fontSize: 13, color: 'var(--text-faint)', marginBottom: 'var(--space-2)' }}>by {extension.creator_id}</p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              {extension.category && <span style={{ fontSize: 11, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'rgba(167,139,250,0.1)', border: '1px solid rgba(167,139,250,0.2)', color: '#a78bfa', fontFamily: 'var(--font-mono)' }}>{extension.category}</span>}
              <span style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 12, color: 'var(--text-faint)' }}>
                <Download style={{ width: 12, height: 12 }} /> {formatNumber(extension.install_count)}
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 12, color: '#eab308' }}>
                <Star style={{ width: 12, height: 12, fill: 'currentColor' }} /> {extension.rating_average.toFixed(1)} ({extension.rating_count})
              </span>
              {extension.trust_score > 0 && <span style={{ fontSize: 12, color: 'var(--text-faint)' }}>Trust: {extension.trust_score}%</span>}
              <span style={{ fontSize: 12, color: 'var(--text-faint)' }}>v{extension.version}</span>
            </div>
          </div>
          <SealedButton onClick={handleInstall} disabled={installing} iconLeft={<Download style={{ width: 14, height: 14 }} />}>
            {installing ? 'Installing...' : 'Install'}
          </SealedButton>
        </div>

        {extension.description && (
          <p style={{ marginTop: 'var(--space-5)', fontSize: 14, color: 'var(--text-dim)', lineHeight: 1.7 }}>{extension.description}</p>
        )}

        {extension.tags && extension.tags.length > 0 && (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-2)', marginTop: 'var(--space-4)' }}>
            {extension.tags.map((tag) => (
              <span key={tag} style={{ fontSize: 11, padding: '3px 10px', borderRadius: 'var(--radius-sm)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)' }}>{tag}</span>
            ))}
          </div>
        )}
      </Chamber>

      <Chamber>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-4)' }}>
          <h2 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 600, color: 'var(--text)' }}>Reviews</h2>
          {isAuthenticated && (
            <SealedButton onClick={() => setReviewDialog(true)} iconLeft={<Star style={{ width: 14, height: 14 }} />}>Write Review</SealedButton>
          )}
        </div>
        <ReviewsList itemType="extension" itemId={id ?? ''} refreshKey={reviewRefreshKey} />
      </Chamber>

      <Modal open={reviewDialog} onClose={() => setReviewDialog(false)} title={`Review ${extension?.name ?? ''}`}>
        <div style={{ marginBottom: 'var(--space-4)' }}>
          <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Rating</label>
          <div style={{ display: 'flex', gap: 'var(--space-1)' }}>
            {[1, 2, 3, 4, 5].map((s) => (
              <button key={s} onClick={() => setReviewRating(s)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 4 }}>
                <Star style={{ width: 24, height: 24, color: s <= reviewRating ? '#eab308' : 'var(--text-faint)', fill: s <= reviewRating ? '#eab308' : 'none' }} />
              </button>
            ))}
          </div>
        </div>
        <div style={{ marginBottom: 'var(--space-4)' }}>
          <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Review</label>
          <textarea placeholder="Share your experience..." value={reviewText} onChange={(e) => setReviewText(e.target.value)} rows={3} style={{ ...inputStyle, resize: 'vertical' as const }} />
        </div>
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)' }}>
          <FrameButton onClick={() => setReviewDialog(false)}>Cancel</FrameButton>
          <SealedButton onClick={handleSubmitReview} disabled={submittingReview} iconLeft={submittingReview ? <Loader2 style={{ width: 14, height: 14, animation: 'spin 1s linear infinite' }} /> : <Star style={{ width: 14, height: 14 }} />}>
            {submittingReview ? 'Submitting...' : 'Submit Review'}
          </SealedButton>
        </div>
      </Modal>
    </div>
  );
}

export default ExtensionDetailPage;
