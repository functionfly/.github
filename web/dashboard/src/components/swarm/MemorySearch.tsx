import { useState, useCallback, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
  Brain, 
  Search,
  Loader2,
  Clock,
  Star,
  CheckCircle,
  X,
  Filter
} from 'lucide-react';
import { agentApi } from '@/api/agent';
import { toast } from 'sonner';

interface Memory {
  id: string;
  agent_id: string;
  memory_type: 'execution' | 'insight' | 'pattern' | 'optimization';
  content: Record<string, unknown>;
  importance: number;
  is_learned: boolean;
  source?: string;
  created_at: string;
  expires_at?: string;
}

interface MemoryStats {
  total_memories: number;
  by_type: Record<string, number>;
  learned_count: number;
  unlearned_count: number;
  average_importance: number;
}

interface MemorySearchProps {
  agentId: string;
  agentName: string;
}

const typeIcons = {
  execution: Clock,
  insight: Brain,
  pattern: Filter,
  optimization: CheckCircle,
};

const typeColors = {
  execution: 'bg-blue-100 text-blue-800 border-blue-200',
  insight: 'bg-purple-100 text-purple-800 border-purple-200',
  pattern: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  optimization: 'bg-green-100 text-green-800 border-green-200',
};

export function MemorySearch({ agentId, agentName }: MemorySearchProps) {
  const [memories, setMemories] = useState<Memory[]>([]);
  const [stats, setStats] = useState<MemoryStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searching, setSearching] = useState(false);
  const [selectedType, setSelectedType] = useState<string | null>(null);
  const [minImportance, setMinImportance] = useState(0);
  const [page, setPage] = useState(0);
  const pageSize = 20;

  const loadMemories = useCallback(async () => {
    setLoading(true);
    try {
      const { memories: data, count } = await agentApi.searchMemories(agentId, {
        q: searchQuery,
        limit: pageSize,
      });
      
      setMemories(data);
      
      // Calculate stats from data
      const typeCounts: Record<string, number> = {};
      let learned = 0;
      let unlearned = 0;
      let totalImportance = 0;
      
      data.forEach(m => {
        typeCounts[m.memory_type] = (typeCounts[m.memory_type] || 0) + 1;
        if (m.is_learned) learned++;
        else unlearned++;
        totalImportance += m.importance;
      });
      
      setStats({
        total_memories: count,
        by_type: typeCounts,
        learned_count: learned,
        unlearned_count: unlearned,
        average_importance: data.length > 0 ? totalImportance / data.length : 0,
      });
    } catch (err) {
      console.error('Failed to load memories:', err);
      toast.error('Failed to load memories');
    } finally {
      setLoading(false);
    }
  }, [agentId, searchQuery, page]);

  useEffect(() => {
    loadMemories();
  }, [loadMemories]);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    setSearching(true);
    setPage(0);
    await loadMemories();
    setSearching(false);
  };

  const filteredMemories = memories.filter(m => {
    if (selectedType && m.memory_type !== selectedType) return false;
    if (m.importance < minImportance) return false;
    return true;
  });

  const formatContent = (content: Record<string, unknown>) => {
    const str = JSON.stringify(content);
    if (str.length > 200) {
      return str.slice(0, 200) + '...';
    }
    return str;
  };

  const getImportanceStars = (importance: number) => {
    return '★'.repeat(Math.round(importance * 5)) + '☆'.repeat(5 - Math.round(importance * 5));
  };

  return (
    <Card className="h-[700px] flex flex-col">
      <CardHeader className="pb-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Brain className="h-5 w-5" />
              Memory Search
            </CardTitle>
            <CardDescription>
              Search {agentName}'s learned memories and insights
            </CardDescription>
          </div>
          {stats && (
            <div className="flex items-center gap-3">
              <div className="text-right">
                <p className="text-2xl font-bold">{stats.total_memories}</p>
                <p className="text-xs text-muted-foreground">memories</p>
              </div>
              <div className="text-right border-l pl-3">
                <p className="text-lg font-semibold">{stats.learned_count}</p>
                <p className="text-xs text-muted-foreground">learned</p>
              </div>
            </div>
          )}
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col min-h-0 space-y-4">
        {/* Search */}
        <form onSubmit={handleSearch} className="flex gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search memories by content..."
              className="pl-9"
            />
            {searchQuery && (
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6"
                onClick={() => setSearchQuery('')}
              >
                <X className="h-3 w-3" />
              </Button>
            )}
          </div>
          <Button type="submit" disabled={searching}>
            {searching ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
          </Button>
        </form>

        {/* Filters */}
        <div className="flex items-center gap-2 flex-wrap">
          <Filter className="h-4 w-4 text-muted-foreground" />
          <Badge 
            variant={selectedType === null ? 'default' : 'outline'}
            className="cursor-pointer"
            onClick={() => setSelectedType(null)}
          >
            All
          </Badge>
          {Object.entries(typeIcons).map(([type, Icon]) => (
            <Badge
              key={type}
              variant={selectedType === type ? 'default' : 'outline'}
              className="cursor-pointer capitalize flex items-center gap-1"
              onClick={() => setSelectedType(selectedType === type ? null : type)}
            >
              <Icon className="h-3 w-3" />
              {type}
              {stats?.by_type[type] && (
                <span className="ml-1 text-xs opacity-70">({stats.by_type[type]})</span>
              )}
            </Badge>
          ))}
          <div className="ml-auto flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Min importance:</span>
            <input
              type="range"
              min="0"
              max="1"
              step="0.1"
              value={minImportance}
              onChange={(e) => setMinImportance(parseFloat(e.target.value))}
              className="w-24"
            />
            <span className="text-xs w-8">{(minImportance * 100).toFixed(0)}%</span>
          </div>
        </div>

        {/* Results */}
        <ScrollArea className="flex-1 -mx-2 px-2">
          {loading && memories.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : filteredMemories.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Brain className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No memories found</p>
              <p className="text-sm mt-1">Try adjusting your filters</p>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredMemories.map((memory) => {
                const Icon = typeIcons[memory.memory_type];
                const typeColorClass = typeColors[memory.memory_type];
                
                return (
                  <div
                    key={memory.id}
                    className="p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                  >
                    <div className="flex items-start gap-3">
                      <div className={`p-2 rounded-lg ${typeColorClass.split(' ')[0]}`}>
                        <Icon className="h-4 w-4" />
                      </div>
                      
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap mb-1">
                          <Badge variant="outline" className={`text-xs capitalize ${typeColorClass}`}>
                            {memory.memory_type}
                          </Badge>
                          
                          {memory.is_learned && (
                            <Badge variant="secondary" className="text-xs">
                              <CheckCircle className="h-3 w-3 mr-1" />
                              Learned
                            </Badge>
                          )}
                          
                          <span className="text-xs text-muted-foreground flex items-center gap-1">
                            <Star className="h-3 w-3" />
                            {getImportanceStars(memory.importance)}
                          </span>
                        </div>
                        
                        <div className="bg-muted rounded p-2 font-mono text-xs overflow-x-auto">
                          <code>{formatContent(memory.content)}</code>
                        </div>
                        
                        <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {new Date(memory.created_at).toLocaleString()}
                          </span>
                          {memory.source && (
                            <span>Source: {memory.source}</span>
                          )}
                          {memory.expires_at && (
                            <span>Expires: {new Date(memory.expires_at).toLocaleDateString()}</span>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </ScrollArea>

        {/* Pagination */}
        {filteredMemories.length > 0 && (
          <div className="flex items-center justify-between pt-2 border-t flex-shrink-0">
            <p className="text-xs text-muted-foreground">
              Showing {filteredMemories.length} of {stats?.total_memories || '?'} memories
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page === 0}
                onClick={() => setPage(p => Math.max(0, p - 1))}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={filteredMemories.length < pageSize}
                onClick={() => setPage(p => p + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default MemorySearch;
