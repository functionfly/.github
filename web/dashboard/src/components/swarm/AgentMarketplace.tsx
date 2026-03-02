import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { agentApi, type MarketplaceAgent } from '@/api/agent';
import {
  Search,
  Filter,
  Star,
  Users,
  Zap,
  Shield,
  DollarSign,
  TrendingUp,
  CheckCircle,
  ExternalLink,
  Loader2,
} from 'lucide-react';

// Types for Marketplace - using API type
export type { MarketplaceAgent };

interface MarketplaceFilters {
  search: string;
  listingType: string;
  pricingModel: string;
  minRating: number;
}

export function AgentMarketplace() {
  const [listings, setListings] = useState<MarketplaceAgent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<MarketplaceFilters>({
    search: '',
    listingType: '',
    pricingModel: '',
    minRating: 0
  });

  useEffect(() => {
    loadAgents();
  }, []);

  const loadAgents = async () => {
    try {
      setLoading(true);
      const response = await agentApi.searchMarketplaceAgents({
        limit: 50,
      });
      setListings(response.agents);
    } catch (err) {
      console.error('Failed to load marketplace agents:', err);
      setError('Failed to load agents. Using demo data.');
      // Fallback to demo data on error
      setListings([
        {
          id: '1',
          agentId: 'ai-researcher',
          name: 'AI Research Specialist',
          description: 'Advanced research agent capable of gathering and analyzing data from multiple sources.',
          listingType: 'worker',
          pricingModel: 'per_call',
          pricePerCall: 0.05,
          ratingScore: 4.8,
          totalCalls: 15420,
          roiScore: 92,
        },
        {
          id: '2',
          agentId: 'data-processor',
          name: 'Data Processing Agent',
          description: 'Efficiently processes and transforms data with built-in validation.',
          listingType: 'worker',
          pricingModel: 'subscription',
          subscriptionMonthlyUsd: 49.99,
          ratingScore: 4.6,
          totalCalls: 8930,
          roiScore: 88,
        },
        {
          id: '3',
          agentId: 'swarm-manager',
          name: 'Swarm Orchestrator',
          description: 'Manages and coordinates multiple agents for complex workflows.',
          listingType: 'manager',
          pricingModel: 'revenue_share',
          revenueSharePercent: 15,
          ratingScore: 4.9,
          totalCalls: 3210,
          roiScore: 95,
        },
        {
          id: '4',
          agentId: 'monitor-agent',
          name: 'System Monitor',
          description: '24/7 monitoring agent with alerting and reporting capabilities.',
          listingType: 'infrastructure',
          pricingModel: 'subscription',
          subscriptionMonthlyUsd: 99.99,
          ratingScore: 4.7,
          totalCalls: 12450,
          roiScore: 90,
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const filteredListings = listings.filter(listing => {
    if (filters.search && !listing.name.toLowerCase().includes(filters.search.toLowerCase())) {
      return false;
    }
    if (filters.listingType && listing.listingType !== filters.listingType) {
      return false;
    }
    if (filters.pricingModel && listing.pricingModel !== filters.pricingModel) {
      return false;
    }
    if (filters.minRating > 0 && listing.ratingScore < filters.minRating) {
      return false;
    }
    return true;
  });

  const getPriceDisplay = (listing: MarketplaceAgent) => {
    switch (listing.pricingModel) {
      case 'free':
        return 'Free';
      case 'per_call':
        return `$${listing.pricePerCall?.toFixed(2)}/call`;
      case 'subscription':
        return `$${listing.subscriptionMonthlyUsd}/mo`;
      case 'revenue_share':
        return `${listing.revenueSharePercent}% revenue share`;
    }
  };

  const getRoleBadge = (type: string) => {
    const colors = {
      worker: 'bg-blue-500',
      manager: 'bg-purple-500',
      infrastructure: 'bg-orange-500'
    };
    return colors[type as keyof typeof colors] || 'bg-gray-500';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold">Agent Marketplace</h1>
          <p className="text-muted-foreground mt-1">
            Discover and hire autonomous AI agents
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="bg-green-500/10 text-green-500">
            <CheckCircle className="h-3 w-3 mr-1" />
            {listings.length} Verified Agents
          </Badge>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search agents..."
                className="pl-10"
                value={filters.search}
                onChange={(e) => setFilters(f => ({ ...f, search: e.target.value }))}
              />
            </div>
            <select
              className="px-3 py-2 border rounded-md"
              value={filters.listingType}
              onChange={(e) => setFilters(f => ({ ...f, listingType: e.target.value }))}
            >
              <option value="">All Types</option>
              <option value="worker">Worker</option>
              <option value="manager">Manager</option>
              <option value="infrastructure">Infrastructure</option>
            </select>
            <select
              className="px-3 py-2 border rounded-md"
              value={filters.pricingModel}
              onChange={(e) => setFilters(f => ({ ...f, pricingModel: e.target.value }))}
            >
              <option value="">All Pricing</option>
              <option value="free">Free</option>
              <option value="per_call">Per Call</option>
              <option value="subscription">Subscription</option>
              <option value="revenue_share">Revenue Share</option>
            </select>
          </div>
        </CardContent>
      </Card>

      {/* Listings Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredListings.map((listing) => (
          <ListingCard
            key={listing.id}
            listing={listing}
            priceDisplay={getPriceDisplay(listing)}
            roleBadge={getRoleBadge(listing.listingType)}
          />
        ))}
      </div>

      {filteredListings.length === 0 && (
        <div className="text-center py-12">
          <Users className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">No agents found</h3>
          <p className="text-muted-foreground">Try adjusting your filters</p>
        </div>
      )}
    </div>
  );
}

function ListingCard({
  listing,
  priceDisplay,
  roleBadge
}: {
  listing: MarketplaceAgent;
  priceDisplay: string;
  roleBadge: string;
}) {
  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle className="text-lg">{listing.name}</CardTitle>
            <CardDescription className="line-clamp-2">
              {listing.description}
            </CardDescription>
          </div>
          {listing.deterministicVerified && (
            <Shield className="h-5 w-5 text-green-500" />
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2">
          <Badge className={roleBadge}>
            {listing.listingType}
          </Badge>
          <Badge variant="outline">
            {priceDisplay}
          </Badge>
        </div>

        <div className="grid grid-cols-3 gap-2 text-center">
          <div>
            <div className="flex items-center justify-center gap-1">
              <Star className="h-3 w-3 text-yellow-500" />
              <span className="text-sm font-medium">{listing.ratingScore}</span>
            </div>
            <p className="text-xs text-muted-foreground">Rating</p>
          </div>
          <div>
            <div className="flex items-center justify-center gap-1">
              <Zap className="h-3 w-3 text-blue-500" />
              <span className="text-sm font-medium">{listing.totalCalls.toLocaleString()}</span>
            </div>
            <p className="text-xs text-muted-foreground">Calls</p>
          </div>
          <div>
            <div className="flex items-center justify-center gap-1">
              <TrendingUp className="h-3 w-3 text-green-500" />
              <span className="text-sm font-medium">{listing.roiScore}%</span>
            </div>
            <p className="text-xs text-muted-foreground">ROI</p>
          </div>
        </div>

        {listing.trustScore !== undefined && (
          <div className="space-y-2">
            <div className="flex justify-between text-xs">
              <span>Trust Score</span>
              <span className="font-medium">{listing.trustScore}%</span>
            </div>
            <Progress value={listing.trustScore} className="h-2" />
          </div>
        )}

        <div className="flex flex-wrap gap-1">
          {listing.capabilities?.slice(0, 3).map((cap) => (
            <Badge key={cap} variant="secondary" className="text-xs">
              {cap}
            </Badge>
          ))}
          {listing.capabilities && listing.capabilities.length > 3 && (
            <Badge variant="secondary" className="text-xs">
              +{listing.capabilities.length - 3}
            </Badge>
          )}
        </div>
      </CardContent>
      <CardFooter>
        <Button className="w-full">
          Hire Agent
          <ExternalLink className="h-4 w-4 ml-2" />
        </Button>
      </CardFooter>
    </Card>
  );
}

export default AgentMarketplace;
