import { Code, Star, Zap } from 'lucide-react';
import { Chamber, SealedButton, FrameButton } from '@/components/containment';
import type { UnifiedItem } from './types';

interface FunctionCardProps {
  item: UnifiedItem;
  onDeploy: (item: UnifiedItem) => void;
  onRate?: (item: UnifiedItem) => void;
  onView?: (item: UnifiedItem) => void;
}

export function FunctionCard({ item, onDeploy, onRate, onView }: FunctionCardProps) {
  const runtime = (item.metadata.runtime as string) ?? '';

  return (
    <Chamber nested style={{ display: 'flex', flexDirection: 'column', cursor: onView ? 'pointer' : undefined }} onClick={() => onView?.(item)}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-3)' }}>
        <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: 'linear-gradient(135deg, rgba(96,165,250,0.1), rgba(59,130,246,0.1))', border: '1px solid rgba(96,165,250,0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
          <Code style={{ width: 18, height: 18, color: '#60a5fa' }} />
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
            <span style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{item.name}</span>
            {item.verified && <span style={{ fontSize: 10, padding: '2px 6px', borderRadius: 'var(--radius-sm)', background: 'rgba(34,197,94,0.1)', border: '1px solid rgba(34,197,94,0.2)', color: '#22c55e', fontFamily: 'var(--font-mono)' }}>Verified</span>}
          </div>
          <p style={{ fontSize: 11, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{item.id}</p>
        </div>
      </div>

      <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5, marginBottom: 'var(--space-3)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden', minHeight: '2.5rem' }}>{item.description || 'No description available'}</p>

      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap', marginBottom: 'var(--space-3)' }}>
        {runtime && <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'rgba(96,165,250,0.1)', border: '1px solid rgba(96,165,250,0.2)', color: '#60a5fa', fontFamily: 'var(--font-mono)' }}>{runtime}</span>}
        <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 'var(--radius-sm)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', color: 'var(--status-ok)', fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{item.price}</span>
        <span style={{ display: 'flex', alignItems: 'center', gap: 3, fontSize: 11, color: 'var(--text-faint)' }}>
          <Zap style={{ width: 11, height: 11 }} /> {item.install_count.toLocaleString()} calls
        </span>
        <span style={{ display: 'flex', alignItems: 'center', gap: 3, fontSize: 11, color: '#eab308' }}>
          <Star style={{ width: 11, height: 11, fill: 'currentColor' }} /> {item.rating.toFixed(1)}
        </span>
      </div>

      <div style={{ marginTop: 'auto', paddingTop: 'var(--space-3)', display: 'flex', gap: 'var(--space-2)' }}>
        <SealedButton style={{ flex: 1 }} onClick={(e) => { e.stopPropagation(); onDeploy(item); }}>Deploy</SealedButton>
        {onRate && <FrameButton onClick={(e) => { e.stopPropagation(); onRate(item); }} iconLeft={<Star style={{ width: 12, height: 12 }} />}>Rate</FrameButton>}
      </div>
    </Chamber>
  );
}
