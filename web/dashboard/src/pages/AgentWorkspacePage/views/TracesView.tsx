import { agentApi } from '@/api/agent';
import { FrameButton, SealedButton, StatusPill } from '@/components/containment';
import { useQuery } from '@tanstack/react-query';
import { Activity, Clock, DollarSign, Hash, Search } from 'lucide-react';
import { useState } from 'react';

interface TracesViewProps {
  agentId: string;
  setRightContext: (ctx: { type: string; id: string } | null) => void;
}

export function TracesView({ agentId, setRightContext }: TracesViewProps) {
  const [search, setSearch] = useState('');
  const [timeRange, setTimeRange] = useState('24h');

  const { data: runsData, isLoading } = useQuery({
    queryKey: ['agent-runs', agentId, timeRange],
    queryFn: async () => {
      try {
        const res = await fetch(`/api/v1/agent-observability/runs?agent_id=${agentId}&limit=50`, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token') || ''}` },
        });
        return res.json();
      } catch {
        return { runs: [] };
      }
    },
    refetchInterval: 15000,
  });

  const runs = runsData?.runs ?? [];

  const filtered = runs.filter((run: any) => {
    if (!search) return true;
    return (run.id ?? '').toLowerCase().includes(search.toLowerCase()) ||
      (run.status ?? '').toLowerCase().includes(search.toLowerCase());
  });

  const stats = {
    total: runs.length,
    avgCost: runs.length > 0 ? runs.reduce((s: number, r: any) => s + (r.cost_usd ?? 0), 0) / runs.length : 0,
    avgDuration: runs.length > 0 ? runs.reduce((s: number, r: any) => s + (r.duration_ms ?? 0), 0) / runs.length : 0,
    errorRate: runs.length > 0 ? runs.filter((r: any) => r.status === 'error').length / runs.length : 0,
  };

  const timeRanges = ['1h', '6h', '24h', '7d'];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Traces</h2>
          <p className="aw-center__subtitle">Execution traces and observability runs</p>
        </div>
      </div>

      <div className="aw-stats">
        <div className="aw-stat">
          <p className="aw-stat__label">Total Runs</p>
          <p className="aw-stat__value">{stats.total}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Avg Cost</p>
          <p className="aw-stat__value">${stats.avgCost.toFixed(4)}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Avg Duration</p>
          <p className="aw-stat__value">{(stats.avgDuration / 1000).toFixed(1)}s</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Error Rate</p>
          <p className="aw-stat__value" style={{ color: stats.errorRate > 0.1 ? 'var(--status-revoked)' : undefined }}>
            {(stats.errorRate * 100).toFixed(1)}%
          </p>
        </div>
      </div>

      {/* Filters */}
      <div style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'center' }}>
        <div style={{ flex: 1, position: 'relative' }}>
          <Search size={14} style={{ position: 'absolute', left: 'var(--space-3)', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-faint)' }} />
          <input
            className="aw-card"
            style={{ width: '100%', padding: 'var(--space-2) var(--space-3) var(--space-2) var(--space-7)', fontFamily: 'var(--font-body)', fontSize: '13px', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', color: 'var(--text)' }}
            placeholder="Search traces..."
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
        <div style={{ display: 'flex', gap: 'var(--space-1)' }}>
          {timeRanges.map(tr => (
            <button
              key={tr}
              className={`aw-nav-item ${timeRange === tr ? 'aw-nav-item--active' : ''}`}
              style={{ padding: 'var(--space-1) var(--space-2)', fontSize: '12px' }}
              onClick={() => setTimeRange(tr)}
            >
              {tr}
            </button>
          ))}
        </div>
      </div>

      {/* Runs List */}
      {isLoading ? (
        <div className="aw-loading"><div className="aw-loading__spinner" /></div>
      ) : filtered.length === 0 ? (
        <div className="aw-empty">
          <Activity size={40} className="aw-empty__icon" />
          <span className="aw-empty__title">No traces found</span>
          <span className="aw-empty__desc">Traces will appear when your agent executes functions</span>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
          {filtered.map((run: any) => (
            <button
              key={run.id}
              className="aw-feed-item"
              style={{ cursor: 'pointer', textAlign: 'left', width: '100%' }}
              onClick={() => setRightContext({ type: 'trace', id: run.id })}
            >
              <span className={`aw-feed-item__dot ${run.status === 'error' ? 'aw-feed-item__dot--error' : 'aw-feed-item__dot--result'}`} />
              <div className="aw-feed-item__content">
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-1)' }}>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text)' }}>
                    {(run.id ?? '').slice(0, 12)}...
                  </span>
                  <StatusPill
                    status={run.status === 'error' ? 'revoked' : run.status === 'running' ? 'pending' : 'live'}
                    label={run.status}
                  />
                </div>
                <div className="aw-feed-item__meta">
                  <span><Hash size={10} /> {run.event_count ?? 0} events</span>
                  <span><DollarSign size={10} /> ${(run.cost_usd ?? 0).toFixed(4)}</span>
                  <span><Clock size={10} /> {((run.duration_ms ?? 0) / 1000).toFixed(1)}s</span>
                  {run.created_at && (
                    <span>{new Date(run.created_at).toLocaleString()}</span>
                  )}
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
