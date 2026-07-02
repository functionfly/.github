import { Bot, Star, TrendingUp } from 'lucide-react';
import { Chamber, SealedButton, FrameButton, StatusPill } from '@/components/containment';
import type { UnifiedItem } from './types';

const LISTING_TYPE_LABELS: Record<string, string> = {
  worker: 'Worker', manager: 'Manager', infrastructure: 'Infrastructure',
};

interface AgentCardProps {
  item: UnifiedItem;
  onHire: (item: UnifiedItem) => void;
  onRate?: (item: UnifiedItem) => void;
  onView?: (item: UnifiedItem) => void;
}

export function AgentCard({ item, onHire, onRate, onView }: AgentCardProps) {
  const listingType = (item.metadata.listing_type as string) ?? '';
  const roiScore = (item.metadata.roi_score as number) ?? 0;

  return (
    <Chamber nested style={{ cursor: onView ? 'pointer' : undefined, display: 'flex', flexDirection: 'column' }} onClick={() => onView?.(item)}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-3)' }}>
        <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: 'linear-gradient(135deg, rgba(143,255,208,0.1), rgba(159,216,255,0.1))', border: '1px solid var(--panel-edge)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
          <Bot style={{ width: 18, height: 18, color: 'var(--status-ok)' }} />
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
            <span style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{item.name}</span>
            {item.verified && <StatusPill status="live" label="Verified" />}
          </div>
          <p style={{ fontSize: 11, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{item.id}</p>
        </div>
      </div>

      <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5, marginBottom: 'var(--space-3)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden', minHeight: '2.5rem' }}>{item.description || 'No description available'}</p>

      {item.tags.length > 0 && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-1)', marginBottom: 'var(--space-3)' }}>
          {item.tags.slice(0, 4).map((tag) => (
            <span key={tag} style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)' }}>{tag}</span>
          ))}
          {item.tags.length > 4 && <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', color: 'var(--text-faint)' }}>+{item.tags.length - 4}</span>}
        </div>
      )}

      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-3)' }}>
        {listingType && <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', border: '1px solid var(--panel-edge)', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase' }}>{LISTING_TYPE_LABELS[listingType] ?? listingType}</span>}
        <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', color: 'var(--status-ok)', fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{item.price}</span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 'var(--space-2)', textAlign: 'center', marginBottom: 'var(--space-3)' }}>
        {[
          { icon: Star, value: item.rating.toFixed(1), label: 'Rating', color: 'var(--status-pending)' },
          { icon: TrendingUp, value: roiScore.toFixed(1), label: 'ROI', color: 'var(--status-ok)' },
          { icon: Bot, value: item.install_count.toLocaleString(), label: 'Calls', color: 'var(--foil-a)' },
        ].map(({ icon: Icon, value, label, color }) => (
          <div key={label} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: 'var(--space-2)', background: 'var(--panel)', borderRadius: 'var(--radius)' }}>
            <Icon style={{ width: 11, height: 11, color, marginBottom: 2 }} />
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 500, color: 'var(--text)' }}>{value}</span>
            <span style={{ fontSize: 10, color: 'var(--text-faint)' }}>{label}</span>
          </div>
        ))}
      </div>

      <div style={{ marginTop: 'auto', paddingTop: 'var(--space-3)', display: 'flex', gap: 'var(--space-2)' }}>
        <SealedButton style={{ flex: 1 }} onClick={(e) => { e.stopPropagation(); onHire(item); }}>Hire Agent</SealedButton>
        {onRate && <FrameButton onClick={(e) => { e.stopPropagation(); onRate(item); }} iconLeft={<Star style={{ width: 12, height: 12 }} />}>Rate</FrameButton>}
      </div>
    </Chamber>
  );
}
