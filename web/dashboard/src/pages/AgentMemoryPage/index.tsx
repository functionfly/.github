import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Search, Brain, Trash2, RefreshCw, Filter, MoreVertical,
  Clock, Star, FileText,
} from "lucide-react";
import { Chamber, PageGrid, SealedButton, FrameButton } from "@/components/containment";
import {
  useAgentMemories, useDeleteAgentMemory,
  useSearchAgentMemories, useRebuildIndex,
} from "@/hooks/useAgentMemory";
import type { AgentMemory, AgentMemoryType, AgentMemorySearchRequest } from "@/types";
import './agent-memory-page.css';

const memoryTypeLabels: Record<AgentMemoryType, string> = {
  working: "Working", longterm: "Long-term", context: "Context", episodic: "Episodic",
};

function MemoryCard({ memory, onDelete, onView }: { memory: AgentMemory; onDelete: (id: string) => void; onView: (id: string) => void }) {
  const createdAt = new Date(memory.created_at);
  const expiresAt = memory.expires_at ? new Date(memory.expires_at) : null;

  return (
    <div className="agent-memory-card" onClick={() => onView(memory.id)}>
      <div className="agent-memory-card-header">
        <div className="agent-memory-card-info">
          <div className="agent-memory-card-badges">
            <span className={`agent-memory-badge ${memory.memory_type}`}>{memoryTypeLabels[memory.memory_type]}</span>
          </div>
          <p className="agent-memory-card-agent">Agent: {memory.agent_id}</p>
        </div>
        <button className="agent-memory-card-menu-btn" onClick={(e) => { e.stopPropagation(); onDelete(memory.id); }} title="Delete memory">
          <Trash2 style={{ width: 14, height: 14 }} />
        </button>
      </div>
      <div className="agent-memory-card-content">
        <p className={`agent-memory-card-text ${!memory.content ? 'empty' : ''}`}>{memory.content || "No content available"}</p>
      </div>
      <div className="agent-memory-card-footer">
        <div className="agent-memory-card-stats">
          <span className="agent-memory-card-stat star"><Star style={{ width: 11, height: 11 }} />{memory.importance_score.toFixed(2)}</span>
          <span className="agent-memory-card-stat"><FileText style={{ width: 11, height: 11 }} />{memory.access_count}</span>
        </div>
        <span className="agent-memory-card-stat"><Clock style={{ width: 11, height: 11 }} />{createdAt.toLocaleDateString()}</span>
      </div>
      {expiresAt && <p className="agent-memory-card-expiry">Expires: {expiresAt.toLocaleDateString()}</p>}
    </div>
  );
}

export function AgentMemoryPage() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");
  const [agentIdFilter, setAgentIdFilter] = useState("");
  const [memoryTypeFilter, setMemoryTypeFilter] = useState("all");
  const [isSearching, setIsSearching] = useState(false);
  const [searchResults, setSearchResults] = useState<AgentMemory[] | null>(null);

  const { data, isLoading, error } = useAgentMemories({
    agent_id: agentIdFilter || undefined,
    memory_type: memoryTypeFilter !== "all" ? (memoryTypeFilter as AgentMemoryType) : undefined,
  });

  const deleteMutation = useDeleteAgentMemory();
  const searchMutation = useSearchAgentMemories();
  const rebuildIndexMutation = useRebuildIndex();

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;
    setIsSearching(true);
    try {
      const result = await searchMutation.mutateAsync({
        query: searchQuery,
        agent_id: agentIdFilter || undefined,
        memory_type: memoryTypeFilter !== "all" ? (memoryTypeFilter as AgentMemoryType) : undefined,
        limit: 20,
      });
      setSearchResults(result.memories);
    } finally { setIsSearching(false); }
  };

  const handleDelete = (id: string) => { if (confirm("Are you sure you want to delete this memory?")) deleteMutation.mutate(id); };
  const memories = searchResults ?? data?.memories ?? [];
  const total = searchResults ? searchResults.length : data?.total ?? 0;

  return (
    <div className="agent-memory-page">
      <PageGrid />

      {/* Header */}
      <div className="agent-memory-header">
        <div className="agent-memory-header-title">
          <div className="agent-memory-header-icon"><Brain style={{ width: 20, height: 20 }} /></div>
          <h1>Agent Memory</h1>
        </div>
        <FrameButton size="sm" onClick={() => rebuildIndexMutation.mutate({})} disabled={rebuildIndexMutation.isPending}
          iconLeft={<RefreshCw style={{ width: 14, height: 14 }} />}>
          Rebuild Index
        </FrameButton>
      </div>

      {/* Search & Filter */}
      <div className="agent-memory-search-card">
        <div className="agent-memory-search-row">
          <div className="agent-memory-search-input-wrapper">
            <Search className="agent-memory-search-icon" />
            <input type="text" placeholder="Search memories..." value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              className="agent-memory-search-input" />
          </div>
          <input type="text" placeholder="Filter by Agent ID..." value={agentIdFilter}
            onChange={(e) => setAgentIdFilter(e.target.value)}
            className="agent-memory-agent-filter-input" />
          <select value={memoryTypeFilter} onChange={(e) => setMemoryTypeFilter(e.target.value)}
            className="agent-memory-type-select-trigger">
            <option value="all">All Types</option>
            <option value="working">Working Memory</option>
            <option value="longterm">Long-term Memory</option>
            <option value="context">Context</option>
            <option value="episodic">Episodic Memory</option>
          </select>
          {searchResults ? (
            <FrameButton size="sm" onClick={() => { setSearchQuery(""); setSearchResults(null); }}>Clear Search</FrameButton>
          ) : (
            <SealedButton size="sm" onClick={handleSearch} disabled={isSearching || !searchQuery.trim()}>
              {isSearching ? "Searching..." : "Search"}
            </SealedButton>
          )}
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="agent-memory-loading"><div className="agent-memory-loading-spinner" /></div>
      ) : error ? (
        <Chamber>
          <div className="agent-memory-error-state" style={{ background: 'transparent', border: 'none', padding: 0 }}>
            <p className="agent-memory-error-state-title">Failed to load memories</p>
            <p className="agent-memory-error-state-message">{error.message}</p>
          </div>
        </Chamber>
      ) : memories.length === 0 ? (
        <Chamber>
          <div style={{ textAlign: 'center', padding: 'var(--space-8) 0' }}>
            <div style={{ width: 64, height: 64, margin: '0 auto var(--space-4)', borderRadius: 'var(--radius-lg)', background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Brain style={{ width: 32, height: 32, color: 'var(--text-faint)' }} />
            </div>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>No memories found</h3>
            <p style={{ fontSize: 14, color: 'var(--text-dim)' }}>{searchResults ? "Try adjusting your search query" : "Create your first agent memory to get started"}</p>
          </div>
        </Chamber>
      ) : (
        <>
          <p className="agent-memory-stats">{searchResults ? "Search results" : "Total memories"}: {total}</p>
          <div className="agent-memory-grid">
            {memories.map((memory) => <MemoryCard key={memory.id} memory={memory} onDelete={handleDelete} onView={(id) => navigate(`/agent-memories/${id}`)} />)}
          </div>
        </>
      )}
    </div>
  );
}
