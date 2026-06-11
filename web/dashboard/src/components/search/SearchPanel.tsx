import { useState } from 'react';
import { Search } from 'lucide-react';
import { SearchInput } from './SearchInput';
import { SearchFilters } from './SearchFilters';
import { SearchResults } from './SearchResults';
import { useExecuteSearchTool, useSearchTools } from '@/hooks/useAgentSearch';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui';

const SEARCH_TOOLS = [
  { value: 'search.web', label: 'Web' },
  { value: 'search.news', label: 'News' },
  { value: 'search.docs', label: 'Docs' },
  { value: 'search.company', label: 'Company' },
] as const;

export function SearchPanel() {
  const [selectedTool, setSelectedTool] = useState<string>('search.web');
  const [query, setQuery] = useState('');
  const [filters, setFilters] = useState<Record<string, unknown>>({});

  const executeSearch = useExecuteSearchTool();
  const { data: toolsData } = useSearchTools();

  const handleSearch = async (searchQuery: string) => {
    if (!searchQuery.trim()) return;

    executeSearch.mutate({
      toolName: selectedTool as 'search.web' | 'search.news' | 'search.docs' | 'search.company',
      parameters: { query: searchQuery, ...filters },
      enableCache: true,
    });
  };

  const result = executeSearch.data?.result;
  const results = Array.isArray(result) ? result : result ? [result] : [];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Search className="h-5 w-5" />
          Agent Search
        </CardTitle>
        <Tabs value={selectedTool} onValueChange={setSelectedTool} className="mt-4">
          <TabsList>
            {SEARCH_TOOLS.map((tool) => (
              <TabsTrigger key={tool.value} value={tool.value}>
                {tool.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex gap-2">
          <div className="flex-1">
            <SearchInput
              placeholder={`Search the web...`}
              onSearch={setQuery}
              initialValue={query}
            />
          </div>
          <SearchFilters
            toolName={selectedTool}
            filters={filters}
            onFiltersChange={setFilters}
          />
          <Button
            onClick={() => handleSearch(query)}
            disabled={!query || executeSearch.isPending}
          >
            {executeSearch.isPending ? 'Searching...' : 'Search'}
          </Button>
        </div>

        {executeSearch.isError && (
          <div className="text-sm text-destructive">
            Search failed: {executeSearch.error.message}
          </div>
        )}

        {executeSearch.isSuccess && (
          <SearchResults
            toolName={selectedTool as 'search.web' | 'search.news' | 'search.docs' | 'search.company'}
            results={results}
            loading={executeSearch.isPending}
          />
        )}

        {executeSearch.data && (
          <div className="text-xs text-muted-foreground flex gap-4">
            <span>{executeSearch.data.resultsCount} results</span>
            <span>{executeSearch.data.executionTimeMs}ms</span>
            <span>{executeSearch.data.creditsUsed} credits</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}