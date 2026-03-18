import type { RegistryFunction } from '@/api/registry';
import { registryApi } from '@/api/registry';
import { AIDiscoveryUrlCard } from '@/components/common/AIDiscoveryUrlCard';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { useQuery } from '@tanstack/react-query';
import { Bot, Code, ExternalLink, Search, Shield, Star, TrendingUp, Zap } from 'lucide-react';
import { useMemo, useState } from 'react';

// Types for Function Marketplace
interface FunctionListing {
  id: string;
  functionId: string;
  name: string;
  description: string;
  category: string;
  tags: string[];
  pricingModel: 'free' | 'per_call' | 'subscription' | 'revenue_share';
  pricePerCall?: number;
  subscriptionMonthlyUsd?: number;
  ratingScore: number;
  callVolume: number;
  deterministicVerified: boolean;
  agentGenerated: boolean;
  ownerAgentId?: string;
  reliabilityScore: number;
}

interface MarketplaceFilters {
  search: string;
  category: string;
  deterministicOnly: boolean;
  agentGeneratedOnly: boolean;
}

const categories = [
  'All',
  'data-processing',
  'utilities',
  'ai-ml',
  'integrations',
  'transformations',
  'validations',
  'analytics',
];

/** Map registry API response to marketplace FunctionListing */
function registryToListing(fn: RegistryFunction | Record<string, unknown>): FunctionListing {
  const author = String((fn as Record<string, unknown>).author ?? '');
  const name = String((fn as Record<string, unknown>).name ?? '');
  const title = (fn as Record<string, unknown>).title as string | undefined;
  const desc = (fn as Record<string, unknown>).description as string | undefined;
  const category = String((fn as Record<string, unknown>).category ?? '');
  const tags = Array.isArray((fn as Record<string, unknown>).tags)
    ? ((fn as Record<string, unknown>).tags as string[])
    : [];
  const pricePerCall = Number((fn as Record<string, unknown>).price_per_call ?? 0);
  const reliability = Number(
    (fn as Record<string, unknown>).reliability ??
      (fn as Record<string, unknown>).reliability_score ??
      0
  );
  const deterministic: boolean =
    typeof (fn as Record<string, unknown>).deterministic === 'boolean'
      ? ((fn as Record<string, unknown>).deterministic as boolean)
      : Number((fn as Record<string, unknown>).deterministic_score ?? 0) >= 0.9;
  const overallScore = Number(
    (fn as Record<string, unknown>).overall_score ??
      (fn as Record<string, unknown>).trust_score ??
      0
  );
  const popularityScore = Number((fn as Record<string, unknown>).popularity_score ?? 0);
  const id = String((fn as Record<string, unknown>).id ?? `${author}/${name}`);

  return {
    id,
    functionId: `${author}/${name}`,
    name: title || name,
    description: desc || '',
    category: category || 'utilities',
    tags,
    pricingModel: pricePerCall > 0 ? 'per_call' : 'free',
    pricePerCall: pricePerCall > 0 ? pricePerCall : undefined,
    ratingScore: overallScore <= 5 ? overallScore : Math.min(5, overallScore / 20),
    callVolume: Math.round(popularityScore * 1000) || 0,
    deterministicVerified: deterministic,
    agentGenerated: false,
    reliabilityScore: reliability <= 1 ? Math.round(reliability * 100) : Math.round(reliability),
  };
}

export function FunctionMarketplace() {
  const [filters, setFilters] = useState<MarketplaceFilters>({
    search: '',
    category: '',
    deterministicOnly: false,
    agentGeneratedOnly: false,
  });

  const {
    data: registryResponse,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['registry-marketplace', filters.search, filters.category],
    queryFn: async () => {
      if (filters.search.trim()) {
        const res = await registryApi.searchFunctions(
          filters.search.trim(),
          filters.category || undefined,
          100
        );
        return res;
      }
      return registryApi.getFunctions({
        visibility: 'public',
        limit: 100,
        offset: 0,
        ...(filters.category ? { category: filters.category } : {}),
      });
    },
    staleTime: 60 * 1000,
  });

  const listings = useMemo(() => {
    const raw = registryResponse?.functions ?? [];
    return Array.isArray(raw) ? raw.map(registryToListing) : [];
  }, [registryResponse]);

  const filteredListings = useMemo(
    () =>
      listings.filter((fn) => {
        if (filters.deterministicOnly && !fn.deterministicVerified) return false;
        if (filters.agentGeneratedOnly && !fn.agentGenerated) return false;
        return true;
      }),
    [listings, filters.deterministicOnly, filters.agentGeneratedOnly]
  );

  const getPriceDisplay = (fn: FunctionListing) => {
    switch (fn.pricingModel) {
      case 'free':
        return <Badge className="bg-green-500">Free</Badge>;
      case 'per_call':
        return <Badge variant="outline">${fn.pricePerCall?.toFixed(3)}/call</Badge>;
      case 'subscription':
        return <Badge variant="outline">${fn.subscriptionMonthlyUsd}/mo</Badge>;
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-6 text-center">
        <p className="text-sm text-destructive">Failed to load marketplace. Try again later.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold">Function Marketplace</h1>
          <p className="text-muted-foreground mt-1">
            Discover AI-generated and human-made functions
          </p>
        </div>
        <div className="flex items-center gap-2">
          {listings.filter((l) => l.deterministicVerified).length > 0 && (
            <Badge variant="outline" className="bg-green-500/10 text-green-500">
              <Shield className="h-3 w-3 mr-1" />
              Deterministic Verified
            </Badge>
          )}
          {listings.filter((l) => l.agentGenerated).length > 0 && (
            <Badge variant="outline" className="bg-purple-500/10 text-purple-500">
              <Bot className="h-3 w-3 mr-1" />
              AI Generated
            </Badge>
          )}
        </div>
      </div>

      {/* AI discovery endpoint for LLMs */}
      <AIDiscoveryUrlCard />

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search functions..."
                className="pl-10"
                value={filters.search}
                onChange={(e) => setFilters((f) => ({ ...f, search: e.target.value }))}
              />
            </div>
            <select
              className="px-3 py-2 border rounded-md"
              value={filters.category}
              onChange={(e) => setFilters((f) => ({ ...f, category: e.target.value }))}
            >
              {categories.map((cat) => (
                <option key={cat} value={cat === 'All' ? '' : cat}>
                  {cat === 'All' ? 'All Categories' : cat}
                </option>
              ))}
            </select>
            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={filters.deterministicOnly}
                  onChange={(e) =>
                    setFilters((f) => ({ ...f, deterministicOnly: e.target.checked }))
                  }
                  className="rounded border-gray-300"
                />
                <Shield className="h-4 w-4" />
                Deterministic
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={filters.agentGeneratedOnly}
                  onChange={(e) =>
                    setFilters((f) => ({ ...f, agentGeneratedOnly: e.target.checked }))
                  }
                  className="rounded border-gray-300"
                />
                <Bot className="h-4 w-4" />
                AI Generated
              </label>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Listings Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredListings.map((fn) => (
          <FunctionCard key={fn.id} fn={fn} priceDisplay={getPriceDisplay(fn)} />
        ))}
      </div>

      {filteredListings.length === 0 && (
        <div className="text-center py-12">
          <Code className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">No functions found</h3>
          <p className="text-muted-foreground">Try adjusting your filters</p>
        </div>
      )}
    </div>
  );
}

function FunctionCard({
  fn,
  priceDisplay,
}: {
  fn: FunctionListing;
  priceDisplay: React.ReactNode;
}) {
  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle className="text-lg flex items-center gap-2">
              {fn.name}
              {fn.agentGenerated && <Bot className="h-4 w-4 text-purple-500" />}
            </CardTitle>
            <CardDescription className="line-clamp-2">{fn.description}</CardDescription>
          </div>
          {fn.deterministicVerified && <Shield className="h-5 w-5 text-green-500" />}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2">
          {priceDisplay}
          <Badge variant="secondary">{fn.category}</Badge>
        </div>

        <div className="grid grid-cols-3 gap-2 text-center">
          <div>
            <div className="flex items-center justify-center gap-1">
              <Star className="h-3 w-3 text-yellow-500" />
              <span className="text-sm font-medium">{fn.ratingScore}</span>
            </div>
            <p className="text-xs text-muted-foreground">Rating</p>
          </div>
          <div>
            <div className="flex items-center justify-center gap-1">
              <Zap className="h-3 w-3 text-blue-500" />
              <span className="text-sm font-medium">{(fn.callVolume / 1000).toFixed(1)}k</span>
            </div>
            <p className="text-xs text-muted-foreground">Calls</p>
          </div>
          <div>
            <div className="flex items-center justify-center gap-1">
              <TrendingUp className="h-3 w-3 text-green-500" />
              <span className="text-sm font-medium">{fn.reliabilityScore}%</span>
            </div>
            <p className="text-xs text-muted-foreground">Reliability</p>
          </div>
        </div>

        {fn.agentGenerated && fn.ownerAgentId && (
          <div className="text-xs text-muted-foreground flex items-center gap-1">
            <Bot className="h-3 w-3" />
            Generated by {fn.ownerAgentId}
          </div>
        )}

        <div className="flex flex-wrap gap-1">
          {fn.tags.slice(0, 4).map((tag) => (
            <Badge key={tag} variant="secondary" className="text-xs">
              {tag}
            </Badge>
          ))}
        </div>
      </CardContent>
      <CardFooter>
        <Button className="w-full" variant={fn.pricingModel === 'free' ? 'default' : 'outline'}>
          {fn.pricingModel === 'free' ? 'Use Function' : 'Purchase'}
          <ExternalLink className="h-4 w-4 ml-2" />
        </Button>
      </CardFooter>
    </Card>
  );
}

export default FunctionMarketplace;
