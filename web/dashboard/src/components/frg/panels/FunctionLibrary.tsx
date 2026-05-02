/**
 * FunctionLibrary Sidebar (Left Sidebar)
 * Features: Drag functions into canvas, Search + filter, Categories
 */

import { useState, useCallback, useMemo } from 'react';
import { 
  Search, 
  Filter, 
  Grid3X3, 
  List, 
  Plus,
  Star,
  Clock,
  Zap,
  Database,
  Globe,
  FileText,
  Image,
  Music,
  Video,
  Code,
  Cpu,
  Sparkles,
  ChevronRight,
  MoreHorizontal,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

import { useFRGStore } from '@/stores/frgStore';
import type { FunctionCatalogItem } from '@/types/frg';
import { useFunctionCatalog, useRecentFunctions, type CatalogFilters } from '@/hooks/useFunctionCatalog';
import { Skeleton } from '@/components/ui/skeleton';

// Function categories with icons
const categories: { id: string; name: string; icon: React.ElementType }[] = [
  { id: 'all', name: 'All Functions', icon: Grid3X3 },
  { id: 'recent', name: 'Recently Used', icon: Clock },
  { id: 'favorites', name: 'Favorites', icon: Star },
  { id: 'api', name: 'API & HTTP', icon: Globe },
  { id: 'data', name: 'Data Processing', icon: Database },
  { id: 'text', name: 'Text & NLP', icon: FileText },
  { id: 'image', name: 'Image & Vision', icon: Image },
  { id: 'audio', name: 'Audio & Speech', icon: Music },
  { id: 'video', name: 'Video Processing', icon: Video },
  { id: 'code', name: 'Code & Dev', icon: Code },
  { id: 'ml', name: 'ML & AI', icon: Cpu },
];

// Category icon mapping
const categoryIcons: Record<string, React.ElementType> = {
  text: FileText,
  ml: Cpu,
  data: Database,
  api: Globe,
  image: Image,
  video: Video,
  audio: Music,
  code: Code,
  all: Grid3X3,
  recent: Clock,
  favorites: Star,
};

interface FunctionCardProps {
  fn: FunctionCatalogItem;
  viewMode: 'grid' | 'list';
}

function FunctionCard({ fn, viewMode }: FunctionCardProps) {
  const IconComponent = (categoryIcons[fn.category] || Code) as React.ComponentType<{className?: string}>;
  
  const onDragStart = (event: React.DragEvent) => {
    event.dataTransfer.setData('application/reactflow', 'functionNode');
    event.dataTransfer.setData('application/functionfly-function', JSON.stringify(fn));
    event.dataTransfer.effectAllowed = 'move';
  };

  if (viewMode === 'list') {
    return (
      <div
        draggable
        onDragStart={onDragStart}
        className={cn(
          "flex items-center gap-3 p-3 rounded-lg border cursor-move",
          "bg-[var(--bg-secondary)] border-[var(--border-subtle)]",
          "hover:border-brand-500 hover:shadow-md transition-all duration-200"
        )}
      >
        <div 
          className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
          style={{ backgroundColor: `${fn.color}20`, color: fn.color }}
        >
          <IconComponent className="w-5 h-5" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="font-medium text-sm text-[var(--text-primary)] truncate">
            {fn.name}
          </div>
          <div className="text-xs text-[var(--text-secondary)] truncate">
            {fn.author} • v{fn.version}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="secondary" className="text-[10px]">
            <Star className="w-3 h-3 mr-1 fill-yellow-500 text-yellow-500" />
            {fn.trustScore}
          </Badge>
        </div>
      </div>
    );
  }

  return (
    <div
      draggable
      onDragStart={onDragStart}
      className={cn(
        "p-4 rounded-xl border cursor-move",
        "bg-[var(--bg-secondary)] border-[var(--border-subtle)]",
        "hover:border-brand-500 hover:shadow-md hover:scale-[1.02] transition-all duration-200"
      )}
    >
      <div className="flex items-start gap-3 mb-3">
        <div 
          className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
          style={{ backgroundColor: `${fn.color}20`, color: fn.color }}
        >
          <IconComponent className="w-5 h-5" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="font-medium text-sm text-[var(--text-primary)] truncate">
            {fn.name}
          </div>
          <div className="text-[10px] text-[var(--text-secondary)]">
            {fn.author}
          </div>
        </div>
      </div>
      
      <p className="text-xs text-[var(--text-secondary)] line-clamp-2 mb-3">
        {fn.description}
      </p>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Badge variant="secondary" className="text-[10px]">
            <Star className="w-3 h-3 mr-1 fill-yellow-500 text-yellow-500" />
            {fn.trustScore}
          </Badge>
          <span className="text-[10px] text-[var(--text-muted)]">
            {fn.usageCount.toLocaleString()} uses
          </span>
        </div>
        <span className="text-[10px] text-[var(--text-muted)]">
          {fn.avgExecutionTimeMs}ms
        </span>
      </div>

      <div className="flex flex-wrap gap-1 mt-2">
        {fn.tags.slice(0, 3).map((tag) => (
          <Badge key={tag} variant="outline" className="text-[9px] px-1 py-0">
            {tag}
          </Badge>
        ))}
      </div>
    </div>
  );
}

export function FunctionLibrary() {
  const [search, setSearch] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [activeTab, setActiveTab] = useState<string>('browse');

  // Prepare filters for API call
  const filters: CatalogFilters = useMemo(() => {
    const baseFilters: CatalogFilters = {
      limit: 50,
      search: search.trim() || undefined,
    };

    // Only apply category filter for real categories (not special ones)
    if (selectedCategory !== 'all' && selectedCategory !== 'recent' && selectedCategory !== 'favorites') {
      baseFilters.category = selectedCategory;
    }

    return baseFilters;
  }, [selectedCategory, search]);

  // Fetch catalog functions from API
  const { data: catalogFunctions, isLoading: isLoadingCatalog, error: catalogError } = useFunctionCatalog(
    // Only fetch when not on "recent" tab (that uses a separate hook)
    activeTab === 'browse' && selectedCategory !== 'recent' ? filters : undefined
  );

  // Fetch recent functions when that category is selected
  const { data: recentFunctions, isLoading: isLoadingRecent } = useRecentFunctions(10);

  // Determine which functions to display
  const displayFunctions = useMemo((): FunctionCatalogItem[] => {
    // When "recent" is selected, show recent functions
    if (selectedCategory === 'recent') {
      return recentFunctions || [];
    }
    // Otherwise use catalog functions (already filtered by category/search via API)
    return catalogFunctions || [];
  }, [selectedCategory, catalogFunctions, recentFunctions]);

  // Client-side search filtering (API does basic search, but we also filter locally for instant results)
  const filteredFunctions = useMemo(() => {
    let filtered = displayFunctions;

    // For local search refinement when API search is not specific enough
    if (search) {
      const searchLower = search.toLowerCase();
      filtered = filtered.filter(fn =>
        fn.name.toLowerCase().includes(searchLower) ||
        fn.description.toLowerCase().includes(searchLower) ||
        fn.tags.some(tag => tag.toLowerCase().includes(searchLower)) ||
        fn.author.toLowerCase().includes(searchLower)
      );
    }

    return filtered;
  }, [displayFunctions, search]);

  const clearSearch = () => setSearch('');

  const isLoading = isLoadingCatalog || isLoadingRecent;

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-[var(--border-subtle)] space-y-3">
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search functions..."
              className="pl-8 h-9 text-sm"
            />
            {search && (
              <Button
                variant="ghost"
                size="icon"
                className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6"
                onClick={clearSearch}
              >
                <Plus className="w-3 h-3 rotate-45" />
              </Button>
            )}
          </div>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-9 w-9"
                  onClick={() => setViewMode(viewMode === 'grid' ? 'list' : 'grid')}
                >
                  {viewMode === 'grid' ? <List className="w-4 h-4" /> : <Grid3X3 className="w-4 h-4" />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {viewMode === 'grid' ? <p>List view</p> : <p>Grid view</p>}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        {/* Category Tabs */}
        <ScrollArea className="h-10 overflow-x-auto">
          <div className="flex gap-1 px-1">
            {categories.slice(0, 5).map((cat) => (
              <Button
                key={cat.id}
                variant={selectedCategory === cat.id ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setSelectedCategory(cat.id)}
                className={cn(
                  "text-xs h-7 whitespace-nowrap",
                  selectedCategory === cat.id && "bg-brand-500 text-white"
                )}
              >
                {(() => {
                  const Icon = cat.icon as any;
                  return <Icon className="w-3 h-3 mr-1" />;
                })()}
                {cat.name}
              </Button>
            ))}
          </div>
        </ScrollArea>
      </div>

      {/* Content */}
      <Tabs value={activeTab as any} onValueChange={setActiveTab as any} className="flex-1 flex flex-col h-full">
        <TabsList className="w-full rounded-none border-b border-[var(--border-subtle)] bg-transparent p-0 h-9 px-3">
          <TabsTrigger
            value={"browse" as any}
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            Browse
          </TabsTrigger>
          <TabsTrigger
            value={"templates" as any}
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            Templates
          </TabsTrigger>
          <TabsTrigger
            value={"my-fns" as any}
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            My Functions
          </TabsTrigger>
        </TabsList>

        <ScrollArea className="flex-1">
          <TabsContent value={"browse" as any} className="m-0 p-3 h-full">
            {/* Smart Suggestions */}
            <div className="mb-4">
              <div className="flex items-center gap-2 mb-2">
                <Sparkles className="w-4 h-4 text-brand-500" />
                <span className="text-xs font-medium text-[var(--text-primary)]">Smart Suggestions</span>
              </div>
              <div className="bg-brand-500/10 border border-brand-500/20 rounded-lg p-2 text-xs text-[var(--text-secondary)]">
                Based on your workflow, consider adding a text-summarize node
              </div>
            </div>

            {/* Error State */}
            {catalogError && (
              <div className="text-center py-6 mb-4 bg-red-500/10 border border-red-500/20 rounded-lg">
                <p className="text-sm text-red-400">Failed to load functions</p>
                <p className="text-xs text-red-400/70 mt-1">
                  {catalogError instanceof Error ? catalogError.message : 'Unknown error'}
                </p>
              </div>
            )}

            {/* Loading State */}
            {isLoading && (
              <div className={cn(
                "space-y-2",
                viewMode === 'grid' && "grid grid-cols-1 gap-3"
              )}>
                {Array.from({ length: 6 }).map((_, i) => (
                  <Skeleton
                    key={i}
                    className={cn(
                      "h-20 rounded-lg",
                      viewMode === 'grid' && "h-32"
                    )}
                  />
                ))}
              </div>
            )}

            {/* Functions */}
            {!isLoading && (
              <div className={cn(
                "space-y-2",
                viewMode === 'grid' && "grid grid-cols-1 gap-3"
              )}>
                {filteredFunctions.map((fn) => (
                  <FunctionCard key={fn.id} fn={fn} viewMode={viewMode} />
                ))}
              </div>
            )}

            {!isLoading && filteredFunctions.length === 0 && !catalogError && (
              <div className="text-center py-8">
                <div className="w-12 h-12 rounded-full bg-[var(--bg-tertiary)] flex items-center justify-center mx-auto mb-3">
                  <Search className="w-6 h-6 text-[var(--text-muted)]" />
                </div>
                <p className="text-sm text-[var(--text-secondary)]">No functions found</p>
                <p className="text-xs text-[var(--text-muted)]">
                  {search ? 'Try adjusting your search' : 'Check back later for new functions'}
                </p>
              </div>
            )}
          </TabsContent>

          <TabsContent value={"templates" as any} className="m-0 p-3 space-y-3">
            <Accordion type="multiple" className="space-y-2">
              <AccordionItem value="workflows" className="border-0">
                <AccordionTrigger className="text-sm py-2 hover:no-underline">
                  Workflow Templates
                </AccordionTrigger>
                <AccordionContent className="space-y-2">
                  {['Data Pipeline', 'AI Chatbot', 'Image Processing', 'Webhook Handler'].map((template) => (
                    <div 
                      key={template}
                      className="flex items-center gap-2 p-2 rounded-lg border border-[var(--border-subtle)] hover:border-brand-500 cursor-pointer"
                    >
                      <Zap className="w-4 h-4 text-brand-500" />
                      <span className="text-sm">{template}</span>
                    </div>
                  ))}
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          </TabsContent>

          <TabsContent value={"my-fns" as any} className="m-0 p-3">
            <div className="text-center py-8">
              <Button variant="outline">
                <Plus className="w-4 h-4 mr-2" />
                Create Function
              </Button>
            </div>
          </TabsContent>
        </ScrollArea>
      </Tabs>

      {/* Footer */}
      <div className="p-3 border-t border-[var(--border-subtle)]">
        <Button variant="outline" className="w-full">
          <Plus className="w-4 h-4 mr-2" />
          Import Function
        </Button>
      </div>
    </div>
  );
}

export default FunctionLibrary;
