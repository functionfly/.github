import { useState, useEffect } from 'react';
import { Star, User } from 'lucide-react';
import { marketplaceUnifiedApi } from '@/api/marketplace-unified';
import { marketplaceApi } from '@/api/marketplace';
import type { MarketplaceItemType } from './types';

interface Review {
  id: string;
  rating: number;
  review?: string;
  tenant_id?: string;
  username?: string;
  user_name?: string;
  created_at: string;
}

interface ReviewsListProps {
  itemType: MarketplaceItemType;
  itemId: string;
  limit?: number;
  refreshKey?: number;
}

export function ReviewsList({ itemType, itemId, limit = 20, refreshKey }: ReviewsListProps) {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    const fetch = async () => {
      try {
        let data: Review[] = [];
        if (itemType === 'agent') {
          const resp = await marketplaceUnifiedApi.listAgentRatings(itemId, limit);
          data = resp.ratings ?? [];
        } else if (itemType === 'function') {
          const resp = await marketplaceUnifiedApi.listFunctionRatings(itemId, limit);
          data = resp.ratings ?? [];
        } else if (itemType === 'extension') {
          const resp = await marketplaceApi.listRatings(itemId, limit);
          data = resp.ratings ?? [];
        }
        if (!cancelled) setReviews(data);
      } catch {
        if (!cancelled) setReviews([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    fetch();
    return () => { cancelled = true; };
  }, [itemType, itemId, limit, refreshKey]);

  if (loading) {
    return <p style={{ fontSize: 13, color: 'var(--text-faint)', padding: 'var(--space-2) 0' }}>Loading reviews...</p>;
  }

  if (reviews.length === 0) {
    return <p style={{ fontSize: 13, color: 'var(--text-faint)', padding: 'var(--space-2) 0' }}>No reviews yet</p>;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
      {reviews.map((r) => (
        <div key={r.id} style={{ padding: 'var(--space-3)', background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-1)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)' }}>
              <User style={{ width: 14, height: 14, color: 'var(--text-faint)' }} />
              <span style={{ fontSize: 12, color: 'var(--text-dim)', fontFamily: 'var(--font-mono)' }}>
                {r.username || r.user_name || (r.tenant_id ? `${String(r.tenant_id).slice(0, 8)}...` : 'Anonymous')}
              </span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              {[1, 2, 3, 4, 5].map((s) => (
                <Star key={s} style={{ width: 12, height: 12, color: s <= r.rating ? '#eab308' : 'var(--text-faint)', fill: s <= r.rating ? '#eab308' : 'none' }} />
              ))}
            </div>
          </div>
          {r.review && <p style={{ fontSize: 13, color: 'var(--text)', lineHeight: 1.5, marginTop: 'var(--space-1)' }}>{r.review}</p>}
          <p style={{ fontSize: 11, color: 'var(--text-faint)', marginTop: 'var(--space-1)' }}>{new Date(r.created_at).toLocaleDateString()}</p>
        </div>
      ))}
    </div>
  );
}
