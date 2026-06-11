/**
 * AgentCardBrowser — Grid of agent cards with search and filters.
 */

import { useState, useEffect } from 'react';
import { Search, ExternalLink, Shield, Loader2 } from 'lucide-react';
import { ProtocolBadge } from '@/components/common/protocol';
import type { AgentCard, AgentCardListResponse } from '../types';

export function AgentCardBrowser() {
  const [cards, setCards] = useState<AgentCard[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchCards();
  }, []);

  async function fetchCards() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch('/api/v1/a2a/agents/cards');
      if (!res.ok) throw new Error('Failed to fetch agent cards');
      const data: AgentCardListResponse = await res.json();
      setCards(data.cards || []);
      setTotal(data.total || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }

  const filtered = cards.filter(
    (card) =>
      card.name.toLowerCase().includes(search.toLowerCase()) ||
      card.description?.toLowerCase().includes(search.toLowerCase()) ||
      card.id.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="w-6 h-6 animate-spin text-text-secondary" />
        <span className="ml-2 text-text-secondary">Loading agent cards...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-20">
        <p className="text-red-400 mb-4">{error}</p>
        <button
          onClick={fetchCards}
          className="px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-secondary" />
        <input
          type="text"
          placeholder="Search agent cards..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-10 pr-4 py-2.5 bg-bg-secondary border border-border-subtle rounded-xl text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
        />
      </div>

      {/* Stats */}
      <div className="flex items-center gap-4 text-sm text-text-secondary">
        <span>{filtered.length} agents</span>
        <span>·</span>
        <span>{total} total</span>
        <span>·</span>
        <ProtocolBadge protocol="a2a" />
      </div>

      {/* Grid */}
      {filtered.length === 0 ? (
        <div className="text-center py-20">
          <Users className="w-12 h-12 text-text-secondary mx-auto mb-4" />
          <p className="text-text-secondary">No agent cards found</p>
          <p className="text-text-secondary text-sm mt-1">
            Publish an agent card to get started with A2A
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((card) => (
            <AgentCardTile key={card.id} card={card} />
          ))}
        </div>
      )}
    </div>
  );
}

function AgentCardTile({ card }: { card: AgentCard }) {
  return (
    <div className="group p-4 bg-card border border-border-subtle rounded-xl hover:border-brand-500/30 hover:shadow-lg transition-all duration-200">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="font-semibold text-text-primary group-hover:text-brand-400 transition-colors">
            {card.name}
          </h3>
          <p className="text-xs text-text-secondary font-mono">{card.id}</p>
        </div>
        <div className="flex items-center gap-1.5">
          <Shield className="w-3.5 h-3.5 text-emerald-400" />
          <span className="text-xs font-medium text-emerald-400">
            {card.trust_score.toFixed(1)}
          </span>
        </div>
      </div>

      <p className="text-sm text-text-secondary line-clamp-2 mb-3">
        {card.description || 'No description'}
      </p>

      <div className="flex flex-wrap gap-1.5 mb-3">
        {card.capabilities?.map((cap) => (
          <span
            key={cap}
            className="px-2 py-0.5 text-xs bg-zinc-800 text-zinc-300 rounded-full"
          >
            {cap}
          </span>
        ))}
      </div>

      <div className="flex items-center justify-between text-xs text-text-secondary">
        <span>v{card.version} · A2A {card.protocol_version}</span>
        {card.url && (
          <a
            href={card.url}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1 hover:text-brand-400 transition-colors"
          >
            <ExternalLink className="w-3 h-3" />
            endpoint
          </a>
        )}
      </div>

      {card.skills && card.skills.length > 0 && (
        <div className="mt-3 pt-3 border-t border-border-subtle">
          <p className="text-xs text-text-secondary mb-1.5">Skills</p>
          <div className="flex flex-wrap gap-1">
            {card.skills.map((skill) => (
              <span
                key={skill.id}
                className="px-2 py-0.5 text-xs bg-emerald-500/10 text-emerald-300 rounded-full border border-emerald-500/20"
              >
                {skill.id}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// Re-export for use in parent
import { Users } from 'lucide-react';
