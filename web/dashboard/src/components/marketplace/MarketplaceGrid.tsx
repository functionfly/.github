import { Bot, Loader2 } from 'lucide-react';
import { Chamber, FrameButton } from '@/components/containment';
import { AgentCard } from './AgentCard';
import { ExtensionCard } from './ExtensionCard';
import { FunctionCard } from './FunctionCard';
import type { UnifiedItem, MarketplaceItemType } from './types';

interface MarketplaceGridProps {
  items: UnifiedItem[];
  loading: boolean;
  error: string | null;
  total: number;
  hasMore: boolean;
  offset: number;
  limit: number;
  installingId: string | null;
  onPageChange: (newOffset: number) => void;
  onHireAgent: (item: UnifiedItem) => void;
  onRateAgent?: (item: UnifiedItem) => void;
  onInstallExtension: (item: UnifiedItem) => void;
  onRateExtension?: (item: UnifiedItem) => void;
  onDeployFunction: (item: UnifiedItem) => void;
  onRateFunction?: (item: UnifiedItem) => void;
  onViewItem?: (item: UnifiedItem) => void;
}

export function MarketplaceGrid({
  items, loading, error, total, hasMore, offset, limit,
  installingId, onPageChange, onHireAgent, onRateAgent, onInstallExtension, onRateExtension, onDeployFunction, onRateFunction, onViewItem,
}: MarketplaceGridProps) {
  if (error) {
    return <Chamber><div style={{ textAlign: 'center', padding: 'var(--space-7) 0', color: 'var(--status-revoked)' }}>{error}</div></Chamber>;
  }

  if (!loading && items.length === 0) {
    return (
      <Chamber>
        <div style={{ textAlign: 'center', padding: 'var(--space-7) 0' }}>
          <Bot style={{ width: 48, height: 48, color: 'var(--text-faint)', margin: '0 auto var(--space-4)' }} />
          <p style={{ color: 'var(--text-dim)' }}>No results found</p>
          <p style={{ color: 'var(--text-faint)', fontSize: 13, marginTop: 'var(--space-1)' }}>Try adjusting your search or filters</p>
        </div>
      </Chamber>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
      <p style={{ fontSize: 13, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)' }}>
        {loading ? 'Loading...' : `${total} results`}
      </p>

      {loading && (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 'var(--space-7) 0' }}>
          <Loader2 style={{ width: 24, height: 24, color: 'var(--text-faint)', animation: 'spin 1s linear infinite' }} />
        </div>
      )}

      {!loading && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 'var(--space-4)' }}>
          {items.map((item) => {
            switch (item.type) {
              case 'agent':
                return <AgentCard key={`agent-${item.id}`} item={item} onHire={onHireAgent} onRate={onRateAgent} onView={onViewItem} />;
              case 'extension':
                return <ExtensionCard key={`ext-${item.id}`} item={item} onInstall={onInstallExtension} installing={installingId === item.id} onRate={onRateExtension} onView={onViewItem} />;
              case 'function':
                return <FunctionCard key={`func-${item.id}`} item={item} onDeploy={onDeployFunction} onRate={onRateFunction} onView={onViewItem} />;
              default:
                return null;
            }
          })}
        </div>
      )}

      {(hasMore || offset > 0) && (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 'var(--space-3)' }}>
          <FrameButton size="sm" onClick={() => onPageChange(Math.max(0, offset - limit))} disabled={offset === 0}>Previous</FrameButton>
          <span style={{ fontSize: 13, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)' }}>
            {offset + 1}–{Math.min(offset + items.length, total)} of {total}
          </span>
          <FrameButton size="sm" onClick={() => onPageChange(offset + limit)} disabled={!hasMore}>Next</FrameButton>
        </div>
      )}

      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}
