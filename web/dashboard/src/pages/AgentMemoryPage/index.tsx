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
    <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => onView(memory.id)}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <Badge className={memoryTypeColors[memory.memory_type]}>
                {memoryTypeLabels[memory.memory_type]}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground font-mono truncate">
              Agent: {memory.agent_id}
            </p>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="h-4 w-4" />
              </Button>
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
      </CardHeader>
      <CardContent className="pt-0">
        <p className="text-sm line-clamp-3 mb-3">
          {memory.content || "No content available"}
        </p>
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-1">
              <Star className="h-3 w-3" />
              {memory.importance_score.toFixed(2)}
            </span>
            <span className="flex items-center gap-1">
              <FileText className="h-3 w-3" />
              {memory.access_count}
            </span>
          </div>
          <span className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {createdAt.toLocaleDateString()}
          </span>
        </div>
        {expiresAt && (
          <p className="text-xs text-muted-foreground mt-2">
            Expires: {expiresAt.toLocaleDateString()}
          </p>
        )}
      </CardContent>
    </Card>
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
    <div className="container mx-auto py-8">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Brain className="h-8 w-8" />
          <h1 className="text-2xl font-bold">Agent Memory</h1>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleRebuildIndex}
          disabled={rebuildIndexMutation.isPending}
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${rebuildIndexMutation.isPending ? "animate-spin" : ""}`} />
          Rebuild Index
        </Button>
      </div>

      <Card className="mb-6">
        <CardContent className="pt-6">
          <div className="flex gap-4 flex-wrap">
            <div className="flex-1 min-w-[200px]">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search memories..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                  className="pl-9"
                />
              </div>
            </div>
            <Input
              placeholder="Filter by Agent ID..."
              value={agentIdFilter}
              onChange={(e) => setAgentIdFilter(e.target.value)}
              className="w-[200px]"
            />
            <Select value={memoryTypeFilter} onValueChange={setMemoryTypeFilter}>
              <SelectTrigger className="w-[180px]">
                <Filter className="h-4 w-4 mr-2" />
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
              <Button variant="outline" onClick={handleClearSearch}>
                Clear Search
              </Button>
            ) : (
              <Button onClick={handleSearch} disabled={isSearching || !searchQuery.trim()}>
                {isSearching ? "Searching..." : "Search"}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      ) : error ? (
        <div className="text-center py-12 text-destructive">
          <p>Failed to load memories</p>
          <p className="text-sm text-muted-foreground">{error.message}</p>
        </div>
      ) : memories.length === 0 ? (
        <div className="text-center py-12">
          <Brain className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-lg font-medium">No memories found</p>
          <p className="text-muted-foreground">
            {searchResults
              ? "Try adjusting your search query"
              : "Create your first agent memory to get started"}
          </p>
        </div>
      ) : (
        <>
          <p className="text-sm text-muted-foreground mb-4">
            {searchResults ? "Search results" : "Total memories"}: {total}
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
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
