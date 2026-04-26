import { galleryApi, type GalleryFunction } from '@/api/composer';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Search,
  Filter,
  X,
  Star,
  GitFork,
  Heart,
  ExternalLink,
  Code2,
  Loader2,
  TrendingUp,
  Clock,
  Award,
  Sparkles,
  Grid3X3,
  List,
  Tag,
  FolderOpen,
  ChevronDown,
  SearchX,
  Flame,
  Zap,
  LayoutGrid,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

const RUNTIME_ICONS: Record<string, string> = {
  python: '🐍',
  nodejs: '🟢',
  go: '🐹',
  rust: '🦀',
  deno: '🦕',
  bun: '🥯',
  java: '☕',
  'csharp': '#️⃣',
  ruby: '💎',
  php: '🐘',
};

const CATEGORIES = [
  { id: 'data-processing', label: 'Data', icon: FolderOpen, color: 'bg-blue-500/20 text-blue-700 dark:text-blue-300' },
  { id: 'web-scraping', label: 'Web', icon: ExternalLink, color: 'bg-green-500/20 text-green-700 dark:text-green-300' },
  { id: 'api', label: 'API', icon: Code2, color: 'bg-purple-500/20 text-purple-700 dark:text-purple-300' },
  { id: 'ml', label: 'ML/AI', icon: Sparkles, color: 'bg-pink-500/20 text-pink-700 dark:text-pink-300' },
  { id: 'utility', label: 'Utils', icon: Tag, color: 'bg-gray-500/20 text-gray-700 dark:text-gray-300' },
  { id: 'automation', label: 'Auto', icon: Clock, color: 'bg-orange-500/20 text-orange-700 dark:text-orange-300' },
  { id: 'finance', label: 'Finance', icon: TrendingUp, color: 'bg-emerald-500/20 text-emerald-700 dark:text-emerald-300' },
];

const ITEMS_PER_PAGE = 12;
const DEBOUNCE_MS = 300;

// Featured/spotlight functions (top rated from first batch)
const getFeaturedFunctions = (functions: GalleryFunction[]) => {
  return [...functions]
    .sort((a, b) => (b?.trust_score || 0) - (a?.trust_score || 0))
    .slice(0, 6);
};

// Trending functions (by popularity)
const getTrendingFunctions = (functions: GalleryFunction[]) => {
  return [...functions]
    .sort((a, b) => (b?.popularity_score || 0) - (a?.popularity_score || 0))
    .slice(0, 8);
};

/**
 * Function Gallery Page - Browse, search, and remix public functions
 * Features curated sections, pagination, and smart filtering
 */
export function FunctionGalleryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<'discover' | 'all' | 'trending'>('discover');
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [sortBy, setSortBy] = useState<'popular' | 'recent' | 'rating'>('popular');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedRuntime, setSelectedRuntime] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [selectedFunction, setSelectedFunction] = useState<GalleryFunction | null>(null);
  const [remixDialogOpen, setRemixDialogOpen] = useState(false);
  const [customization, setCustomization] = useState('');
  const [mounted, setMounted] = useState(false);

  // Remix cost and balance state
  const [remixCost, setRemixCost] = useState<number>(0.50);
  const [walletBalance, setWalletBalance] = useState<number>(0);
  const [canRemix, setCanRemix] = useState<boolean>(false);
  const [isOwnFunction, setIsOwnFunction] = useState<boolean>(false);

  const searchInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(searchQuery);
      setCurrentPage(1);
    }, DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Reset pagination when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [selectedCategory, selectedRuntime, sortBy, debouncedQuery]);

  // Fetch functions with pagination
  const { data: functionsData, isLoading: isLoadingFunctions, isError } = useQuery({
    queryKey: ['gallery', 'functions', currentPage, sortBy, selectedCategory, debouncedQuery],
    queryFn: async () => {
      const offset = (currentPage - 1) * ITEMS_PER_PAGE;
      
      // Use search endpoint if there's a query
      if (debouncedQuery.trim()) {
        return galleryApi.search({ 
          query: debouncedQuery, 
          sort_by: sortBy,
          limit: ITEMS_PER_PAGE,
          offset,
        });
      }
      
      // Otherwise use list with filters
      const response = await fetch(
        `/api/v1/functions?visibility=public&limit=${ITEMS_PER_PAGE}&offset=${offset}${selectedCategory !== 'all' ? `&category=${selectedCategory}` : ''}`
      );
      if (!response.ok) throw new Error('Failed to fetch');
      const json = await response.json();
      return {
        functions: json.functions || [],
        total_count: json.total || 0,
      };
    },
    staleTime: 2 * 60 * 1000, // 2 minutes
    placeholderData: (previousData) => previousData,
  });

  // Fetch first batch for featured/trending sections (only on discover tab)
  const { data: discoverData } = useQuery({
    queryKey: ['gallery', 'discover'],
    queryFn: async () => {
      const response = await fetch(`/api/v1/functions?visibility=public&limit=24&offset=0`);
      if (!response.ok) throw new Error('Failed to fetch');
      const json = await response.json();
      return json.functions || [];
    },
    staleTime: 5 * 60 * 1000,
    enabled: activeTab === 'discover',
  });

  const allFunctions = functionsData?.functions || [];
  const totalCount = functionsData?.total_count || 0;
  const totalPages = Math.ceil(totalCount / ITEMS_PER_PAGE);

  // Get unique runtimes from current data
  const availableRuntimes = useMemo(() => {
    const runtimes = new Set<string>();
    allFunctions.forEach((fn) => {
      if (fn?.runtime) runtimes.add(fn.runtime);
    });
    return ['all', ...Array.from(runtimes).sort()];
  }, [allFunctions]);

  // Filter functions by runtime (client-side only for now)
  const filteredFunctions = useMemo(() => {
    let functions = [...allFunctions];

    if (selectedRuntime !== 'all') {
      functions = functions.filter((fn) => fn?.runtime === selectedRuntime);
    }

    return functions;
  }, [allFunctions, selectedRuntime]);

  const featuredFunctions = discoverData ? getFeaturedFunctions(discoverData) : [];
  const trendingFunctions = discoverData ? getTrendingFunctions(discoverData) : [];

  const remixMutation = useMutation({
    mutationFn: (data: { author: string; name: string; new_name?: string; customization?: string }) =>
      galleryApi.remix(data.author, data.name, {
        new_name: data.new_name,
        customization: data.customization,
        private_function: true,
      }),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      const costMsg = data.cost_usd ? ` ($${data.cost_usd.toFixed(2)} charged)` : '';
      const balanceMsg = data.new_balance_usd !== undefined ? ` New balance: $${data.new_balance_usd.toFixed(2)}` : '';
      toast.success(`Remixed! Created "${data.new_name}"${costMsg}.${balanceMsg}`);
      setRemixDialogOpen(false);
      setCustomization('');
      // Refresh wallet balance
      if (selectedFunction) {
        galleryApi.getRemixCost(selectedFunction.author, selectedFunction.name)
          .then((costData) => setWalletBalance(costData.balance_usd));
      }
    },
    onError: (error: Error) => {
      toast.error(`Failed to remix: ${error.message}`);
    },
  });

  const likeMutation = useMutation({
    mutationFn: (data: { author: string; name: string }) => galleryApi.like(data.author, data.name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['gallery'] });
      toast.success('Liked!');
    },
  });

  // Fetch remix cost when opening dialog
  useEffect(() => {
    if (remixDialogOpen && selectedFunction) {
      galleryApi.getRemixCost(selectedFunction.author, selectedFunction.name)
        .then((data) => {
          setRemixCost(data.cost_usd);
          setWalletBalance(data.balance_usd);
          setCanRemix(data.can_remix || data.is_own_function);
          setIsOwnFunction(data.is_own_function);
        })
        .catch(() => {
          // Default to allowing remix if we can't check (fail open for UX)
          setCanRemix(true);
        });
    }
  }, [remixDialogOpen, selectedFunction]);

  const handleRemix = (fn: GalleryFunction) => {
    setSelectedFunction(fn);
    setRemixDialogOpen(true);
    setCustomization('');
  };

  const confirmRemix = () => {
    if (selectedFunction) {
      remixMutation.mutate({
        author: selectedFunction.author,
        name: selectedFunction.name,
        customization,
      });
    }
  };

  const clearFilters = useCallback(() => {
    setSearchQuery('');
    setDebouncedQuery('');
    setSelectedCategory('all');
    setSelectedRuntime('all');
    setSortBy('popular');
    setCurrentPage(1);
    searchInputRef.current?.focus();
  }, []);

  const hasActiveFilters = debouncedQuery || selectedCategory !== 'all' || selectedRuntime !== 'all';

  // Generate page numbers
  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const maxVisible = 5;

    if (totalPages <= maxVisible) {
      for (let i = 1; i <= totalPages; i++) pages.push(i);
    } else {
      if (currentPage <= 3) {
        pages.push(1, 2, 3, 4, '...', totalPages);
      } else if (currentPage >= totalPages - 2) {
        pages.push(1, '...', totalPages - 3, totalPages - 2, totalPages - 1, totalPages);
      } else {
        pages.push(1, '...', currentPage - 1, currentPage, currentPage + 1, '...', totalPages);
      }
    }
    return pages;
  };

  const FunctionCard = ({ fn, compact = false }: { fn: GalleryFunction; compact?: boolean }) => {
    if (!fn) return null;
    
    if (compact) {
      return (
        <Card className="group border-border/50 shadow-sm hover:shadow-md hover:border-primary/50 transition-all cursor-pointer overflow-hidden">
          <div 
            className="p-4 flex items-center gap-3"
            onClick={() => fn?.author && fn?.name && navigate(`/registry/${fn.author}/${fn.name}`)}
          >
            <span className="text-2xl shrink-0">{RUNTIME_ICONS[fn?.runtime] || '🔧'}</span>
            <div className="min-w-0 flex-1">
              <h4 className="font-medium truncate">{fn?.title || fn?.name || 'Unknown'}</h4>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>@{fn?.author || 'unknown'}</span>
                {fn?.category && (
                  <Badge variant="secondary" className="text-xs">
                    {fn.category}
                  </Badge>
                )}
              </div>
            </div>
            <div className="flex items-center gap-1 text-sm text-muted-foreground shrink-0">
              <Star className="h-3 w-3" />
              <span>{Math.round(fn?.trust_score || 0)}</span>
            </div>
          </div>
        </Card>
      );
    }

    return (
      <Card className={`group border-border/50 shadow-sm hover:shadow-md transition-all duration-300 ${mounted ? 'opacity-100' : 'opacity-0'}`}>
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0 flex-1">
              <span className="text-2xl shrink-0">{RUNTIME_ICONS[fn?.runtime] || '🔧'}</span>
              <div className="min-w-0 flex-1">
                <h3 className="font-semibold leading-tight truncate" title={fn?.title || fn?.name || 'Unknown'}>
                  {fn?.title || fn?.name || 'Unknown'}
                </h3>
                <p className="text-xs text-muted-foreground truncate">@{fn?.author || 'unknown'}</p>
              </div>
            </div>
            {fn?.category && (
              <Badge className={`shrink-0 ${CATEGORIES.find((c) => c.id === fn?.category)?.color || 'bg-gray-500/20 text-gray-700 dark:text-gray-300'}`}>
                {fn.category}
              </Badge>
            )}
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground line-clamp-2 h-10">
            {fn?.description || 'No description available'}
          </p>

          <div className="flex flex-wrap gap-1 h-6 overflow-hidden">
            {fn?.tags?.slice(0, 3).map((tag: string) => (
              <Badge key={tag} variant="secondary" className="text-xs">{tag}</Badge>
            ))}
            {fn?.tags && fn.tags.length > 3 && (
              <Badge variant="secondary" className="text-xs">+{fn.tags.length - 3}</Badge>
            )}
          </div>

          <div className="flex items-center justify-between text-sm">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Star className="h-4 w-4" />
                    <span>{Math.round(fn?.trust_score || 0)}</span>
                  </div>
                </TooltipTrigger>
                <TooltipContent><p>Trust Score</p></TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <GitFork className="h-4 w-4" />
                    <span>{fn?.remix_count || 0}</span>
                  </div>
                </TooltipTrigger>
                <TooltipContent><p>Remixes</p></TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Heart className="h-4 w-4" />
                    <span>{fn?.like_count || 0}</span>
                  </div>
                </TooltipTrigger>
                <TooltipContent><p>Likes</p></TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <span className="text-xs text-muted-foreground">
              {fn?.created_at ? new Date(fn.created_at).toLocaleDateString() : ''}
            </span>
          </div>

          <Separator />

          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              className="flex-1"
              onClick={() => fn?.author && fn?.name && navigate(`/registry/${fn.author}/${fn.name}`)}
            >
              <ExternalLink className="mr-2 h-4 w-4" />
              View
            </Button>
            <Button
              variant="default"
              size="sm"
              className="flex-1 bg-gradient-to-r from-violet-500 to-purple-600"
              onClick={() => fn && handleRemix(fn)}
            >
              <GitFork className="mr-2 h-4 w-4" />
              Remix
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => fn?.author && fn?.name && likeMutation.mutate({ author: fn.author, name: fn.name })}
            >
              <Heart className="h-4 w-4" />
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  };

  return (
    <div className="container mx-auto p-6 space-y-8">
      {/* Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="p-3 rounded-xl bg-gradient-to-br from-pink-500 to-rose-500 shadow-lg">
            <Sparkles className="h-8 w-8 text-white" />
          </div>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Function Gallery</h1>
            <p className="text-muted-foreground">
              Discover, remix, and share serverless functions
              {totalCount > 0 && <span className="ml-2 text-sm">({totalCount.toLocaleString()} functions available)</span>}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => navigate('/ai/composer')}>
            <Sparkles className="mr-2 h-4 w-4" />
            Create with AI
          </Button>
        </div>
      </div>

      {/* Main Tabs */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)} className="space-y-6">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <TabsList className="h-11">
            <TabsTrigger value="discover" className="gap-2">
              <LayoutGrid className="h-4 w-4" />
              Discover
            </TabsTrigger>
            <TabsTrigger value="trending" className="gap-2">
              <Flame className="h-4 w-4" />
              Trending
            </TabsTrigger>
            <TabsTrigger value="all" className="gap-2">
              <Grid3X3 className="h-4 w-4" />
              Browse All
            </TabsTrigger>
          </TabsList>

          {/* Search Bar */}
          <div className="flex items-center gap-2 flex-1 max-w-xl">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                ref={searchInputRef}
                placeholder="Search functions..."
                className="pl-10"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              {searchQuery && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-1 top-1/2 -translate-y-1/2 h-8 w-8"
                  onClick={() => { setSearchQuery(''); setDebouncedQuery(''); }}
                >
                  <X className="h-4 w-4" />
                </Button>
              )}
            </div>
            {searchQuery && (
              <span className="text-sm text-muted-foreground whitespace-nowrap">
                Searching...
              </span>
            )}
          </div>
        </div>

        {/* DISCOVER TAB */}
        <TabsContent value="discover" className="space-y-8">
          {/* Featured Section */}
          {featuredFunctions.length > 0 && (
            <section className="space-y-4">
              <div className="flex items-center gap-2">
                <Zap className="h-5 w-5 text-yellow-500" />
                <h2 className="text-xl font-semibold">Featured Functions</h2>
                <span className="text-sm text-muted-foreground">Hand-picked high quality functions</span>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {featuredFunctions.map((fn) => (
                  <FunctionCard key={fn?.id} fn={fn} />
                ))}
              </div>
            </section>
          )}

          {/* Trending Section */}
          {trendingFunctions.length > 0 && (
            <section className="space-y-4">
              <div className="flex items-center gap-2">
                <Flame className="h-5 w-5 text-orange-500" />
                <h2 className="text-xl font-semibold">Trending Now</h2>
                <span className="text-sm text-muted-foreground">Most popular this week</span>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-4 gap-3">
                {trendingFunctions.map((fn) => (
                  <FunctionCard key={fn?.id} fn={fn} compact />
                ))}
              </div>
              <div className="text-center">
                <Button variant="outline" onClick={() => setActiveTab('trending')}>
                  View All Trending
                  <ChevronDown className="ml-2 h-4 w-4 rotate-[-90deg]" />
                </Button>
              </div>
            </section>
          )}

          {/* Browse by Category */}
          <section className="space-y-4">
            <h2 className="text-xl font-semibold">Browse by Category</h2>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
              {CATEGORIES.map((cat) => {
                const Icon = cat.icon;
                return (
                  <Card
                    key={cat.id}
                    className="cursor-pointer hover:border-primary/50 hover:shadow-md transition-all group"
                    onClick={() => { setSelectedCategory(cat.id); setActiveTab('all'); }}
                  >
                    <CardContent className="p-4 flex flex-col items-center text-center gap-2">
                      <div className={`p-3 rounded-lg ${cat.color} group-hover:scale-110 transition-transform`}>
                        <Icon className="h-5 w-5" />
                      </div>
                      <span className="font-medium text-sm">{cat.label}</span>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          </section>
        </TabsContent>

        {/* TRENDING TAB */}
        <TabsContent value="trending" className="space-y-6">
          {isLoadingFunctions ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : filteredFunctions.length > 0 ? (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {filteredFunctions.map((fn) => (
                  <FunctionCard key={fn?.id} fn={fn} />
                ))}
              </div>
              
              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-2 pt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  {getPageNumbers().map((page, i) =>
                    page === '...' ? (
                      <span key={`ellipsis-${i}`} className="text-muted-foreground px-2">...</span>
                    ) : (
                      <Button
                        key={page}
                        variant={currentPage === page ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setCurrentPage(page as number)}
                      >
                        {page}
                      </Button>
                    )
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              )}
              
              <p className="text-center text-sm text-muted-foreground">
                Showing {(currentPage - 1) * ITEMS_PER_PAGE + 1} - {Math.min(currentPage * ITEMS_PER_PAGE, totalCount)} of {totalCount} functions
              </p>
            </>
          ) : (
            <Card className="border-border/50 shadow-sm">
              <CardContent className="flex flex-col items-center justify-center py-12">
                <SearchX className="h-12 w-12 text-muted-foreground/50 mb-4" />
                <h3 className="text-lg font-semibold mb-2">No functions found</h3>
                <Button variant="outline" onClick={clearFilters}>
                  Clear Filters
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* ALL/BROWSE TAB */}
        <TabsContent value="all" className="space-y-6">
          {/* Filters Bar */}
          <Card className="border-border/50">
            <CardContent className="pt-4 pb-4 space-y-4">
              <div className="flex flex-wrap items-center gap-4">
                {/* Category Select */}
                <div className="flex items-center gap-2">
                  <Filter className="h-4 w-4 text-muted-foreground" />
                  <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                    <SelectTrigger className="w-[180px]">
                      <SelectValue placeholder="Category" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All Categories</SelectItem>
                      {CATEGORIES.map((cat) => (
                        <SelectItem key={cat.id} value={cat.id}>{cat.label}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Runtime Select */}
                <Select value={selectedRuntime} onValueChange={setSelectedRuntime}>
                  <SelectTrigger className="w-[150px]">
                    <SelectValue placeholder="Runtime" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Runtimes</SelectItem>
                    {availableRuntimes.filter(r => r !== 'all').map((runtime) => (
                      <SelectItem key={runtime} value={runtime} className="capitalize">
                        {RUNTIME_ICONS[runtime] || ''} {runtime}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                {/* Sort Select */}
                <Select value={sortBy} onValueChange={(v) => setSortBy(v as typeof sortBy)}>
                  <SelectTrigger className="w-[150px]">
                    <SelectValue placeholder="Sort by" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="popular">
                      <TrendingUp className="mr-2 h-4 w-4 inline" />
                      Popular
                    </SelectItem>
                    <SelectItem value="recent">
                      <Clock className="mr-2 h-4 w-4 inline" />
                      Recent
                    </SelectItem>
                    <SelectItem value="rating">
                      <Award className="mr-2 h-4 w-4 inline" />
                      Rating
                    </SelectItem>
                  </SelectContent>
                </Select>

                {/* View Toggle */}
                <div className="flex items-center gap-1 border rounded-md p-1 ml-auto">
                  <Button
                    variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => setViewMode('grid')}
                  >
                    <Grid3X3 className="h-4 w-4" />
                  </Button>
                  <Button
                    variant={viewMode === 'list' ? 'secondary' : 'ghost'}
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => setViewMode('list')}
                  >
                    <List className="h-4 w-4" />
                  </Button>
                </div>

                {/* Clear Filters */}
                {hasActiveFilters && (
                  <Button variant="ghost" size="sm" onClick={clearFilters}>
                    <X className="mr-1 h-4 w-4" />
                    Clear
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Results */}
          {isLoadingFunctions ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : isError ? (
            <Card className="border-border/50 shadow-sm">
              <CardContent className="flex flex-col items-center justify-center py-12">
                <SearchX className="h-12 w-12 text-muted-foreground/50 mb-4" />
                <h3 className="text-lg font-semibold mb-2">Failed to load functions</h3>
                <Button variant="outline" onClick={() => queryClient.invalidateQueries({ queryKey: ['gallery'] })}>
                  Retry
                </Button>
              </CardContent>
            </Card>
          ) : filteredFunctions.length > 0 ? (
            <>
              <div className={viewMode === 'grid' ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4' : 'space-y-2'}>
                {filteredFunctions.map((fn) => (
                  viewMode === 'grid' ? (
                    <FunctionCard key={fn?.id} fn={fn} />
                  ) : (
                    // List View
                    <Card key={fn?.id} className="flex items-center p-4 gap-4">
                      <span className="text-xl">{RUNTIME_ICONS[fn?.runtime] || '🔧'}</span>
                      <div className="min-w-0 flex-1">
                        <h3 className="font-semibold truncate">{fn?.title || fn?.name || 'Unknown'}</h3>
                        <p className="text-xs text-muted-foreground truncate">
                          @{fn?.author || 'unknown'} {fn?.category && `• ${fn.category}`}
                        </p>
                      </div>
                      <p className="text-sm text-muted-foreground flex-1 truncate max-w-md hidden lg:block">
                        {fn?.description || 'No description'}
                      </p>
                      <div className="flex items-center gap-4 text-sm text-muted-foreground shrink-0">
                        <span className="flex items-center gap-1"><Star className="h-4 w-4" />{Math.round(fn?.trust_score || 0)}</span>
                        <span className="flex items-center gap-1"><GitFork className="h-4 w-4" />{fn?.remix_count || 0}</span>
                      </div>
                      <div className="flex gap-2 shrink-0">
                        <Button variant="outline" size="sm" onClick={() => fn?.author && fn?.name && navigate(`/registry/${fn.author}/${fn.name}`)}>View</Button>
                        <Button variant="default" size="sm" className="bg-gradient-to-r from-violet-500 to-purple-600" onClick={() => fn && handleRemix(fn)}>Remix</Button>
                      </div>
                    </Card>
                  )
                ))}
              </div>
              
              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-2 pt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  {getPageNumbers().map((page, i) =>
                    page === '...' ? (
                      <span key={`ellipsis-${i}`} className="text-muted-foreground px-2">...</span>
                    ) : (
                      <Button
                        key={page}
                        variant={currentPage === page ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setCurrentPage(page as number)}
                      >
                        {page}
                      </Button>
                    )
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              )}
              
              <p className="text-center text-sm text-muted-foreground">
                Showing {(currentPage - 1) * ITEMS_PER_PAGE + 1} - {Math.min(currentPage * ITEMS_PER_PAGE, totalCount)} of {totalCount} functions
              </p>
            </>
          ) : (
            <Card className="border-border/50 shadow-sm">
              <CardContent className="flex flex-col items-center justify-center py-12">
                <SearchX className="h-12 w-12 text-muted-foreground/50 mb-4" />
                <h3 className="text-lg font-semibold mb-2">No functions found</h3>
                <p className="text-muted-foreground text-center max-w-md mb-4">
                  {hasActiveFilters
                    ? `No functions match your current filters. Try clearing some filters.`
                    : 'No functions available yet. Be the first to publish one!'}
                </p>
                {hasActiveFilters && (
                  <Button variant="outline" onClick={clearFilters}>
                    <X className="mr-2 h-4 w-4" />
                    Clear all filters
                  </Button>
                )}
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>

      {/* Remix Dialog with Cost Info */}
      <Dialog open={remixDialogOpen} onOpenChange={setRemixDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Remix Function</DialogTitle>
            <DialogDescription>
              Create your own copy of <strong>{selectedFunction?.title || selectedFunction?.name}</strong> by @{selectedFunction?.author}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {/* Cost and Balance Info */}
            <div className="bg-muted rounded-lg p-4 space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Remix Cost:</span>
                <span className="font-semibold">${remixCost.toFixed(2)}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Your Balance:</span>
                <span className={walletBalance < remixCost ? "font-semibold text-red-500" : "font-semibold text-green-600"}>
                  ${walletBalance.toFixed(2)}
                </span>
              </div>
              <Separator />
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">Status:</span>
                {isOwnFunction ? (
                  <Badge variant="secondary">Free (your function)</Badge>
                ) : canRemix ? (
                  <Badge variant="default" className="bg-green-600">Ready to remix</Badge>
                ) : (
                  <Badge variant="destructive">Insufficient balance</Badge>
                )}
              </div>
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">Customizations (optional)</label>
              <textarea
                className="w-full min-h-[100px] rounded-md border bg-muted/50 p-3 text-sm resize-none"
                placeholder="e.g., Add error handling, change the output format, optimize for speed..."
                value={customization}
                onChange={(e) => setCustomization(e.target.value)}
                disabled={!canRemix && !isOwnFunction}
              />
            </div>

            {!canRemix && !isOwnFunction && (
              <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800 rounded-md p-3">
                <p className="text-sm text-red-600 dark:text-red-400">
                  You need ${(remixCost - walletBalance).toFixed(2)} more to remix this function.
                  Add funds to your wallet to continue.
                </p>
              </div>
            )}
          </div>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setRemixDialogOpen(false)}>Cancel</Button>
            {!canRemix && !isOwnFunction ? (
              <Button
                onClick={() => navigate('/wallet')}
                className="bg-gradient-to-r from-emerald-500 to-green-600"
              >
                <TrendingUp className="mr-2 h-4 w-4" />
                Add Funds
              </Button>
            ) : (
              <Button
                onClick={confirmRemix}
                disabled={remixMutation.isPending}
                className="bg-gradient-to-r from-violet-500 to-purple-600"
              >
                {remixMutation.isPending ? (
                  <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Remixing...</>
                ) : (
                  <><GitFork className="mr-2 h-4 w-4" />Remix for ${remixCost.toFixed(2)}</>
                )}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
