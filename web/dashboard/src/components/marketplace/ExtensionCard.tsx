import { Download, Shield, Star } from 'lucide-react';
import { Chamber, SealedButton, FrameButton } from '@/components/containment';
import type { UnifiedItem } from './types';

interface ExtensionCardProps {
  item: UnifiedItem;
  onInstall: (item: UnifiedItem) => void;
  installing: boolean;
  onRate?: (item: UnifiedItem) => void;
  onView?: (item: UnifiedItem) => void;
}

export function ExtensionCard({ item, onInstall, installing, onRate, onView }: ExtensionCardProps) {
  const category = (item.metadata.category as string) ?? '';
  const version = (item.metadata.version as string) ?? '';
  const trustScore = (item.metadata.trust_score as number) ?? 0;
  const authorId = (item.metadata.author_id as string) ?? '';

  const formatNumber = (n: number) => {
    if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
    return n.toString();
  };

  return (
    <Chamber nested style={{ display: 'flex', flexDirection: 'column', cursor: onView ? 'pointer' : undefined }} onClick={() => onView?.(item)}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-3)' }}>
        <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: 'linear-gradient(135deg, rgba(167,139,250,0.15), rgba(139,92,246,0.15))', border: '1px solid rgba(167,139,250,0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0, fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 700, color: '#a78bfa' }}>
          {item.name.charAt(0)}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
            <span style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{item.name}</span>
            {item.verified && <Shield style={{ width: 14, height: 14, color: '#22c55e' }} />}
          </div>
          {authorId && <p style={{ fontSize: 11, color: 'var(--text-faint)' }}>by {authorId}</p>}
        </div>
      </div>

      <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5, marginBottom: 'var(--space-3)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden', minHeight: '2.5rem' }}>{item.description || 'No description available'}</p>

      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap', marginBottom: 'var(--space-3)' }}>
        {category && <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'rgba(167,139,250,0.1)', border: '1px solid rgba(167,139,250,0.2)', color: '#a78bfa', fontFamily: 'var(--font-mono)' }}>{category}</span>}
        <span style={{ display: 'flex', alignItems: 'center', gap: 3, fontSize: 11, color: 'var(--text-faint)' }}>
          <Download style={{ width: 11, height: 11 }} /> {formatNumber(item.install_count)}
        </span>
        <span style={{ display: 'flex', alignItems: 'center', gap: 3, fontSize: 11, color: '#eab308' }}>
          <Star style={{ width: 11, height: 11, fill: 'currentColor' }} /> {item.rating.toFixed(1)}
        </span>
        {trustScore > 0 && <span style={{ fontSize: 11, color: 'var(--text-faint)' }}>Trust: {trustScore}%</span>}
        {version && <span style={{ fontSize: 11, color: 'var(--text-faint)', marginLeft: 'auto' }}>v{version}</span>}
      </div>

      <div style={{ marginTop: 'auto', paddingTop: 'var(--space-3)', display: 'flex', gap: 'var(--space-2)' }}>
        <SealedButton style={{ flex: 1 }} onClick={(e) => { e.stopPropagation(); onInstall(item); }} disabled={installing}>
          {installing ? 'Installing...' : 'Install'}
        </SealedButton>
        {onRate && <FrameButton onClick={(e) => { e.stopPropagation(); onRate(item); }} iconLeft={<Star style={{ width: 12, height: 12 }} />}>Rate</FrameButton>}
      </div>
    </Chamber>
  );
}
