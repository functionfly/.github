import { useAgentMemories } from '@/hooks/useAgentMemory';
import { Database, Search, Star } from 'lucide-react';
import { useMemo, useState } from 'react';

interface MemoryViewProps {
  agentId: string;
}

const TYPE_COLORS: Record<string, string> = {
  working: 'var(--foil-a)',
  longterm: 'var(--status-ok)',
  context: 'var(--foil-b)',
  episodic: 'var(--status-pending)',
};

export function MemoryView({ agentId }: MemoryViewProps) {
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState<string | null>(null);

  const { data: memoriesData, isLoading } = useAgentMemories({ agent_id: agentId, limit: 200 });
  const memories = useMemo(() => {
    const data = memoriesData as any;
    return data?.memories ?? data?.items ?? [];
  }, [memoriesData]);

  const filtered = useMemo(() => {
    let result = memories;
    if (typeFilter) {
      result = result.filter((m: any) => m.memory_type === typeFilter);
    }
    if (search) {
      const q = search.toLowerCase();
      result = result.filter((m: any) =>
        (m.content ?? '').toLowerCase().includes(q) ||
        (m.summary ?? '').toLowerCase().includes(q)
      );
    }
    return result;
  }, [memories, typeFilter, search]);

  const typeCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    memories.forEach((m: any) => {
      const t = m.memory_type ?? 'unknown';
      counts[t] = (counts[t] ?? 0) + 1;
    });
    return counts;
  }, [memories]);

  const types = ['working', 'longterm', 'context', 'episodic'];

  if (isLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Memory</h2>
          <p className="aw-center__subtitle">Agent memory store and semantic search</p>
        </div>
      </div>

      {/* Stats */}
      <div className="aw-stats">
        <div className="aw-stat">
          <p className="aw-stat__label">Total</p>
          <p className="aw-stat__value">{memories.length}</p>
        </div>
        {types.map(t => (
          <div key={t} className="aw-stat">
            <p className="aw-stat__label">{t}</p>
            <p className="aw-stat__value">{typeCounts[t] ?? 0}</p>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'center' }}>
        <div style={{ flex: 1, position: 'relative' }}>
          <Search size={14} style={{ position: 'absolute', left: 'var(--space-3)', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-faint)' }} />
          <input
            style={{ width: '100%', padding: 'var(--space-2) var(--space-3) var(--space-2) var(--space-7)', fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }}
            placeholder="Search memories..."
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
        <div style={{ display: 'flex', gap: 'var(--space-1)' }}>
          <button
            className={`aw-nav-item ${!typeFilter ? 'aw-nav-item--active' : ''}`}
            style={{ padding: 'var(--space-1) var(--space-2)', fontSize: '11px' }}
            onClick={() => setTypeFilter(null)}
          >
            All
          </button>
          {types.map(t => (
            <button
              key={t}
              className={`aw-nav-item ${typeFilter === t ? 'aw-nav-item--active' : ''}`}
              style={{ padding: 'var(--space-1) var(--space-2)', fontSize: '11px', textTransform: 'capitalize' }}
              onClick={() => setTypeFilter(typeFilter === t ? null : t)}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      {/* Memory List */}
      {filtered.length === 0 ? (
        <div className="aw-empty">
          <Database size={40} className="aw-empty__icon" />
          <span className="aw-empty__title">No memories found</span>
          <span className="aw-empty__desc">
            {search ? 'Try a different search term' : 'Memories are created as the agent learns from executions'}
          </span>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
          {filtered.map((memory: any) => (
            <div key={memory.id} className="aw-card" style={{ cursor: 'pointer' }}>
              <div className="aw-card__body" style={{ padding: 'var(--space-3) var(--space-4)' }}>
                <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 'var(--space-3)' }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-1)' }}>
                      <span style={{
                        fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase',
                        padding: '2px 6px', borderRadius: 'var(--radius-sm)',
                        color: TYPE_COLORS[memory.memory_type] ?? 'var(--text-dim)',
                        background: `${TYPE_COLORS[memory.memory_type] ?? 'var(--text-dim)'}11`,
                        border: `1px solid ${TYPE_COLORS[memory.memory_type] ?? 'var(--text-dim)'}33`,
                      }}>
                        {memory.memory_type ?? 'unknown'}
                      </span>
                      {memory.importance_score !== undefined && (
                        <span style={{ display: 'flex', alignItems: 'center', gap: '2px', fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-faint)' }}>
                          <Star size={10} />
                          {(memory.importance_score * 100).toFixed(0)}%
                        </span>
                      )}
                    </div>
                    <p style={{ fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)', margin: 0, lineHeight: 1.5, overflow: 'hidden', textOverflow: 'ellipsis', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' }}>
                      {memory.content ?? memory.summary ?? 'No content'}
                    </p>
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 'var(--space-1)', flexShrink: 0 }}>
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-faint)' }}>
                      Accessed: {memory.access_count ?? 0}
                    </span>
                    {memory.created_at && (
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-faint)' }}>
                        {new Date(memory.created_at).toLocaleDateString()}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
