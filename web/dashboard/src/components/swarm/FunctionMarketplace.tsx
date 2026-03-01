import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { 
  Search, 
  Code, 
  Star, 
  Zap,
  Shield,
  DollarSign,
  TrendingUp,
  CheckCircle,
  ExternalLink,
  Bot,
  GitCommit
} from 'lucide-react';

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
  'analytics'
];

export function FunctionMarketplace() {
  const [listings, setListings] = useState<FunctionListing[]>([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<MarketplaceFilters>({
    search: '',
    category: '',
    deterministicOnly: false,
    agentGeneratedOnly: false
  });

  useEffect(() => {
    // Mock data - in production fetch from API
    setTimeout(() => {
      setListings([
        {
          id: '1',
          functionId: 'json-transform',
          name: 'JSON Transformer',
          description: 'Transform and reshape JSON data structures with powerful mapping capabilities.',
          category: 'transformations',
          tags: ['json', 'transform', 'data'],
          pricingModel: 'per_call',
          pricePerCall: 0.001,
          ratingScore: 4.9,
          callVolume: 125430,
          deterministicVerified: true,
          agentGenerated: true,
          reliabilityScore: 99
        },
        {
          id: '2',
          functionId: 'csv-parser',
          name: 'Smart CSV Parser',
          description: 'Parse and validate CSV data with automatic type inference and error handling.',
          category: 'data-processing',
          tags: ['csv', 'parser', 'data'],
          pricingModel: 'free',
          ratingScore: 4.7,
          callVolume: 89230,
          deterministicVerified: true,
          agentGenerated: false,
          reliabilityScore: 97
        },
        {
          id: '3',
          functionId: 'sentiment-analysis',
          name: 'AI Sentiment Analyzer',
          description: 'Analyze text sentiment using advanced NLP models. Supports multiple languages.',
          category: 'ai-ml',
          tags: ['ai', 'nlp', 'sentiment'],
          pricingModel: 'per_call',
          pricePerCall: 0.005,
          ratingScore: 4.8,
          callVolume: 45620,
          deterministicVerified: false,
          agentGenerated: true,
          ownerAgentId: 'nlp-specialist',
          reliabilityScore: 94
        },
        {
          id: '4',
          functionId: 'url-shortener',
          name: 'URL Shortener',
          description: 'Create short, trackable URLs with analytics and custom aliases.',
          category: 'utilities',
          tags: ['url', 'shortener', 'link'],
          pricingModel: 'subscription',
          subscriptionMonthlyUsd: 9.99,
          ratingScore: 4.5,
          callVolume: 67890,
          deterministicVerified: true,
          agentGenerated: false,
          reliabilityScore: 98
        },
        {
          id: '5',
          functionId: 'email-validator',
          name: 'Email Validator',
          description: 'Validate email addresses with MX record checking and disposable email detection.',
          category: 'validations',
          tags: ['email', 'validation', 'verification'],
          pricingModel: 'per_call',
          pricePerCall: 0.002,
          ratingScore: 4.6,
          callVolume: 112340,
          deterministicVerified: true,
          agentGenerated: true,
          reliabilityScore: 96
        }
      ]);
      setLoading(false);
    }, 800);
  }, []);

  const filteredListings = listings.filter(fn => {
    if (filters.search && !fn.name.toLowerCase().includes(filters.search.toLowerCase())) {
      return false;
    }
    if (filters.category && fn.category !== filters.category) {
      return false;
    }
    if (filters.deterministicOnly && !fn.deterministicVerified) {
      return false;
    }
    if (filters.agentGeneratedOnly && !fn.agentGenerated) {
      return false;
    }
    return true;
  });

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
          <h1 className="text-3xl font-bold">Function Marketplace</h1>
          <p className="text-muted-foreground mt-1">
            Discover AI-generated and human-made functions
          </p>
        </div>
        <div className="flex items-center gap-2">
          {listings.filter(l => l.deterministicVerified).length > 0 && (
            <Badge variant="outline" className="bg-green-500/10 text-green-500">
              <Shield className="h-3 w-3 mr-1" />
              Deterministic Verified
            </Badge>
          )}
          {listings.filter(l => l.agentGenerated).length > 0 && (
            <Badge variant="outline" className="bg-purple-500/10 text-purple-500">
              <Bot className="h-3 w-3 mr-1" />
              AI Generated
            </Badge>
          )}
        </div>
      </div>

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
                onChange={(e) => setFilters(f => ({ ...f, search: e.target.value }))}
              />
            </div>
            <select 
              className="px-3 py-2 border rounded-md"
              value={filters.category}
              onChange={(e) => setFilters(f => ({ ...f, category: e.target.value }))}
            >
              {categories.map(cat => (
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
                  onChange={(e) => setFilters(f => ({ ...f, deterministicOnly: e.target.checked }))}
                  className="rounded border-gray-300"
                />
                <Shield className="h-4 w-4" />
                Deterministic
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input 
                  type="checkbox"
                  checked={filters.agentGeneratedOnly}
                  onChange={(e) => setFilters(f => ({ ...f, agentGeneratedOnly: e.target.checked }))}
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
          <FunctionCard 
            key={fn.id} 
            fn={fn}
            priceDisplay={getPriceDisplay(fn)}
          />
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
  priceDisplay
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
              {fn.agentGenerated && (
                <Bot className="h-4 w-4 text-purple-500" />
              )}
            </CardTitle>
            <CardDescription className="line-clamp-2">
              {fn.description}
            </CardDescription>
          </div>
          {fn.deterministicVerified && (
            <Shield className="h-5 w-5 text-green-500" />
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2">
          {priceDisplay}
          <Badge variant="secondary">
            {fn.category}
          </Badge>
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
