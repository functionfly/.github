import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Search,
  Brain,
  Trash2,
  RefreshCw,
  Filter,
  Plus,
  MoreVertical,
  Clock,
  Star,
  FileText,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useAgentMemories,
  useDeleteAgentMemory,
  useSearchAgentMemories,
  useRebuildIndex,
} from "@/hooks/useAgentMemory";
import type { AgentMemory, AgentMemoryType, AgentMemorySearchRequest } from "@/types";

import './agent-memory-page.css';

const memoryTypeColors: Record<AgentMemoryType, string> = {
  working: "bg-blue-500/10 border-blue-500/20 text-blue-700",
  longterm: "bg-green-500/10 border-green-500/20 text-green-700",
  context: "bg-purple-500/10 border-purple-500/20 text-purple-700",
  episodic: "bg-orange-500/10 border-orange-500/20 text-orange-700",
};

const memoryTypeLabels: Record<AgentMemoryType, string> = {
  working: "Working Memory",
  longterm: "Long-term Memory",
  context: "Context",
  episodic: "Episodic Memory",
};

function MemoryCard({
  memory,
  onDelete,
  onView,
}: {
  memory: AgentMemory;
  onDelete: (id: string) => void;
  onView: (id: string) => void;
}) {
  const createdAt = new Date(memory.created_at);
  const expiresAt = memory.expires_at ? new Date(memory.expires_at) : null;

  return (
    <div className="agent-memory-card" onClick={() => onView(memory.id)}>
      {/* Header */}
      <div className="agent-memory-card-header">
        <div className="agent-memory-card-info">
          <div className="agent-memory-card-badges">
            <span className={`agent-memory-badge ${memory.memory_type}`}>
              {memoryTypeLabels[memory.memory_type]}
            </span>
          </div>
          <p className="agent-memory-card-agent">
            Agent: {memory.agent_id}
          </p>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
            <button className="agent-memory-card-menu-btn">
              <MoreVertical className="h-4 w-4" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              className="text-destructive"
              onClick={(e) => {
                e.stopPropagation();
                onDelete(memory.id);
              }}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Content */}
      <div className="agent-memory-card-content">
        <p className={`agent-memory-card-text ${!memory.content ? 'empty' : ''}`}>
          {memory.content || "No content available"}
        </p>
      </div>

      {/* Footer */}
      <div className="agent-memory-card-footer">
        <div className="agent-memory-card-stats">
          <span className="agent-memory-card-stat star">
            <Star className="h-3 w-3" />
            {memory.importance_score.toFixed(2)}
          </span>
          <span className="agent-memory-card-stat">
            <FileText className="h-3 w-3" />
            {memory.access_count}
          </span>
        </div>
        <span className="agent-memory-card-stat">
          <Clock className="h-3 w-3" />
          {createdAt.toLocaleDateString()}
        </span>
      </div>

      {expiresAt && (
        <p className="agent-memory-card-expiry">
          Expires: {expiresAt.toLocaleDateString()}
        </p>
      )}
    </div>
  );
}

export function AgentMemoryPage() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");
  const [agentIdFilter, setAgentIdFilter] = useState<string>("");
  const [memoryTypeFilter, setMemoryTypeFilter] = useState<string>("all");
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
      const searchData: AgentMemorySearchRequest = {
        query: searchQuery,
        agent_id: agentIdFilter || undefined,
        memory_type: memoryTypeFilter !== "all" ? (memoryTypeFilter as AgentMemoryType) : undefined,
        limit: 20,
      };
      const result = await searchMutation.mutateAsync(searchData);
      setSearchResults(result.memories);
    } finally {
      setIsSearching(false);
    }
  };

  const handleClearSearch = () => {
    setSearchQuery("");
    setSearchResults(null);
  };

  const handleDelete = (id: string) => {
    if (confirm("Are you sure you want to delete this memory?")) {
      deleteMutation.mutate(id);
    }
  };

  const handleRebuildIndex = () => {
    rebuildIndexMutation.mutate({});
  };

  const handleView = (id: string) => {
    navigate(`/agent-memories/${id}`);
  };

  const memories = searchResults ?? data?.memories ?? [];
  const total = searchResults ? searchResults.length : data?.total ?? 0;

  return (
    <div className="agent-memory-page container mx-auto">
      {/* Header */}
      <div className="agent-memory-header">
        <div className="agent-memory-header-title">
          <div className="agent-memory-header-icon">
            <Brain className="h-8 w-8" />
          </div>
          <h1>Agent Memory</h1>
        </div>
        <button
          onClick={handleRebuildIndex}
          disabled={rebuildIndexMutation.isPending}
          className={`agent-memory-rebuild-btn ${rebuildIndexMutation.isPending ? 'spinning' : ''}`}
        >
          <RefreshCw className={`h-4 w-4 ${rebuildIndexMutation.isPending ? 'animate-spin' : ''}`} />
          Rebuild Index
        </button>
      </div>

      {/* Search & Filter Card */}
      <div className="agent-memory-search-card">
        <div className="agent-memory-search-row">
          <div className="agent-memory-search-input-wrapper">
            <Search className="agent-memory-search-icon" />
            <input
              type="text"
              placeholder="Search memories..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              className="agent-memory-search-input"
            />
          </div>
          <input
            type="text"
            placeholder="Filter by Agent ID..."
            value={agentIdFilter}
            onChange={(e) => setAgentIdFilter(e.target.value)}
            className="agent-memory-agent-filter-input"
          />
          <Select value={memoryTypeFilter} onValueChange={setMemoryTypeFilter}>
            <SelectTrigger className="agent-memory-type-select-trigger">
              <Filter className="h-4 w-4" />
              <SelectValue placeholder="Memory Type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Types</SelectItem>
              <SelectItem value="working">Working Memory</SelectItem>
              <SelectItem value="longterm">Long-term Memory</SelectItem>
              <SelectItem value="context">Context</SelectItem>
              <SelectItem value="episodic">Episodic Memory</SelectItem>
            </SelectContent>
          </Select>
          {searchResults ? (
            <button onClick={handleClearSearch} className="agent-memory-action-btn outline">
              Clear Search
            </button>
          ) : (
            <button
              onClick={handleSearch}
              disabled={isSearching || !searchQuery.trim()}
              className="agent-memory-action-btn primary"
            >
              {isSearching ? "Searching..." : "Search"}
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="agent-memory-loading">
          <div className="agent-memory-loading-spinner" />
        </div>
      ) : error ? (
        <div className="agent-memory-error-state">
          <p className="agent-memory-error-state-title">Failed to load memories</p>
          <p className="agent-memory-error-state-message">{error.message}</p>
        </div>
      ) : memories.length === 0 ? (
        <div className="agent-memory-empty-state">
          <div className="agent-memory-empty-state-icon">
            <Brain className="h-12 w-12" />
          </div>
          <p className="agent-memory-empty-state-title">No memories found</p>
          <p className="agent-memory-empty-state-description">
            {searchResults
              ? "Try adjusting your search query"
              : "Create your first agent memory to get started"}
          </p>
        </div>
      ) : (
        <>
          <p className="agent-memory-stats">
            {searchResults ? "Search results" : "Total memories"}: {total}
          </p>
          <div className="agent-memory-grid">
            {memories.map((memory) => (
              <MemoryCard
                key={memory.id}
                memory={memory}
                onDelete={handleDelete}
                onView={handleView}
              />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
