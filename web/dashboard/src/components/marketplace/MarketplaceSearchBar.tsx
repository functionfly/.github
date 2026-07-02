import { Search, Bot, Puzzle, Code, LayoutGrid } from 'lucide-react';
import type { MarketplaceItemType } from './types';

interface MarketplaceSearchBarProps {
  query: string;
  selectedType: MarketplaceItemType | '';
  onQueryChange: (q: string) => void;
  onTypeChange: (type: MarketplaceItemType | '') => void;
  onSearch: () => void;
}

const typeChips: { value: MarketplaceItemType | ''; label: string; icon: React.ReactNode }[] = [
  { value: '', label: 'All', icon: <LayoutGrid style={{ width: 14, height: 14 }} /> },
  { value: 'agent', label: 'Agents', icon: <Bot style={{ width: 14, height: 14 }} /> },
  { value: 'extension', label: 'Extensions', icon: <Puzzle style={{ width: 14, height: 14 }} /> },
  { value: 'function', label: 'Functions', icon: <Code style={{ width: 14, height: 14 }} /> },
];

export function MarketplaceSearchBar({ query, selectedType, onQueryChange, onTypeChange, onSearch }: MarketplaceSearchBarProps) {
  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: 'var(--space-3) var(--space-4)',
    paddingLeft: 36,
    background: 'var(--panel-raised)',
    border: '1px solid var(--steel)',
    borderRadius: 'var(--radius)',
    color: 'var(--text)',
    fontFamily: 'var(--font-body)',
    fontSize: 13,
    outline: 'none',
    transition: 'border-color var(--duration-fast) var(--ease-out)',
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
      <div style={{ position: 'relative' }}>
        <Search style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', width: 14, height: 14, color: 'var(--text-faint)', pointerEvents: 'none' }} />
        <input
          placeholder="Search agents, extensions, and functions..."
          value={query}
          onChange={(e) => onQueryChange(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && onSearch()}
          style={inputStyle}
        />
      </div>

      <div style={{ display: 'flex', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
        {typeChips.map(({ value, label, icon }) => {
          const isActive = selectedType === value;
          return (
            <button
              key={value}
              onClick={() => onTypeChange(value)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 'var(--space-1)',
                padding: '6px 14px',
                borderRadius: 'var(--radius)',
                border: `1px solid ${isActive ? 'var(--brand-500)' : 'var(--panel-edge)'}`,
                background: isActive ? 'rgba(124,58,237,0.1)' : 'var(--panel)',
                color: isActive ? 'var(--brand-400, #a78bfa)' : 'var(--text-dim)',
                fontSize: 12,
                fontFamily: 'var(--font-body)',
                fontWeight: isActive ? 600 : 400,
                cursor: 'pointer',
                transition: 'all var(--duration-fast) var(--ease-out)',
                whiteSpace: 'nowrap',
              }}
            >
              {icon}
              {label}
            </button>
          );
        })}
      </div>
    </div>
  );
}
