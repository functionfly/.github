import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { agentApi, type MarketplaceAgent, type MarketplaceAgentSearchParams } from '@/api/agent';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useAuthStore } from '@/stores/authStore';
import {
  Bot,
  CheckCircle,
  Filter,
  Loader2,
  Search,
  Shield,
  Sparkles,
  Star,
  TrendingUp,
  Wallet,
} from 'lucide-react';
import { toast } from 'sonner';

interface AgentMarketplaceViewProps {
  variant?: 'standalone' | 'embedded';
  onAgentSelect?: (agent: MarketplaceAgent) => void;
}

const PRICING_MODEL_LABELS: Record<string, string> = {
  free: 'Free',
  per_call: 'Per Call',
  subscription: 'Subscription',
  revenue_share: 'Revenue Share',
  tiered: 'Tiered',
  dynamic: 'Dynamic',
  auction: 'Auction',
};

const LISTING_TYPE_LABELS: Record<string, string> = {
  worker: 'Worker',
  manager: 'Manager',
  infrastructure: 'Infrastructure',
};

export function AgentMarketplaceView({ variant = 'embedded', onAgentSelect }: AgentMarketplaceViewProps) {
  const { t } = useTranslation();
  const isAuthenticated = useAuthStore((s) => !!s.user);
  const [agents, setAgents] = useState<MarketplaceAgent[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [limit, setLimit] = useState(20);
  const [offset, setOffset] = useState(0);

  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState<MarketplaceAgentSearchParams>({});
  const [showFilters, setShowFilters] = useState(false);

  const [hireDialog, setHireDialog] = useState<MarketplaceAgent | null>(null);
  const [hireTaskType, setHireTaskType] = useState('');
  const [hireBudget, setHireBudget] = useState('');
  const [hiring, setHiring] = useState(false);

  const loadAgents = async (params?: MarketplaceAgentSearchParams) => {
    setLoading(true);
    setError(null);
    try {
      const response = await agentApi.searchMarketplaceAgents({
        ...params,
        limit,
        offset,
      });
      setAgents(response.agents);
      setTotal(response.total);
      setHasMore(response.has_more);
    } catch (err) {
      console.error('Failed to load agents:', err);
      setError('Failed to load agents');
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    setOffset(0);
    loadAgents({ ...filters, capabilities: searchQuery ? searchQuery.split(',').map(s => s.trim()) : undefined });
  };

  const handleFilterChange = (key: keyof MarketplaceAgentSearchParams, value: string | string[] | number | undefined) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  const handlePageChange = (newOffset: number) => {
    setOffset(newOffset);
    loadAgents({ ...filters, capabilities: searchQuery ? searchQuery.split(',').map(s => s.trim()) : undefined });
  };

  const handleHire = async () => {
    if (!hireDialog) return;
    setHiring(true);
    try {
      await agentApi.hireAgent({
        agent_id: hireDialog.agentId,
        task_type: hireTaskType || 'general',
        budget_usd: hireBudget ? parseFloat(hireBudget) : undefined,
      });
      toast.success(`Successfully hired ${hireDialog.name}`);
      setHireDialog(null);
      setHireTaskType('');
      setHireBudget('');
    } catch {
      toast.error('Failed to hire agent');
    } finally {
      setHiring(false);
    }
  };

  const formatPrice = (agent: MarketplaceAgent) => {
    if (agent.pricingModel === 'free') return 'Free';
    if (agent.pricingModel === 'per_call' && agent.pricePerCall) {
      return `$${agent.pricePerCall.toFixed(4)}/call`;
    }
    if (agent.pricingModel === 'subscription' && agent.subscriptionMonthlyUsd) {
      return `$${agent.subscriptionMonthlyUsd.toFixed(2)}/mo`;
    }
    return PRICING_MODEL_LABELS[agent.pricingModel] ?? agent.pricingModel;
  };

  return (
    <div className="space-y-6">
      {/* Search and Filters */}
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
            <Input
              placeholder="Search by capabilities (e.g. code_generation, analysis)"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="pl-10"
            />
          </div>
          <Button onClick={handleSearch} disabled={loading}>
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Search'}
          </Button>
          <Button variant="outline" onClick={() => setShowFilters(!showFilters)}>
            <Filter className="h-4 w-4 mr-2" />
            Filters
          </Button>
        </div>

        {showFilters && (
          <div className="flex flex-wrap gap-4 p-4 bg-bg-secondary rounded-lg border border-border-subtle">
            <div className="flex flex-col gap-1.5">
              <Label className="text-xs text-text-muted">Pricing Model</Label>
              <Select
                value={filters.pricing_model ?? ''}
                onValueChange={(v) => handleFilterChange('pricing_model', v === '__any__' ? undefined : v || undefined)}
              >
                <SelectTrigger className="w-[160px]">
                  <SelectValue placeholder="Any" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__any__">Any</SelectItem>
                  {Object.entries(PRICING_MODEL_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-1.5">
              <Label className="text-xs text-text-muted">Listing Type</Label>
              <Select
                value={filters.listing_types?.[0] ?? ''}
                onValueChange={(v) => handleFilterChange('listing_types', v === '__any__' ? undefined : v ? [v] : undefined)}
              >
                <SelectTrigger className="w-[160px]">
                  <SelectValue placeholder="Any" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__any__">Any</SelectItem>
                  {Object.entries(LISTING_TYPE_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-1.5">
              <Label className="text-xs text-text-muted">Sort By</Label>
              <Select
                value={filters.sort_by ?? 'rank_score'}
                onValueChange={(v) => handleFilterChange('sort_by', v as 'rank_score' | 'rating_score' | 'price_per_call' | 'total_calls')}
              >
                <SelectTrigger className="w-[160px]">
                  <SelectValue placeholder="Rank Score" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="rank_score">Rank Score</SelectItem>
                  <SelectItem value="rating_score">Rating</SelectItem>
                  <SelectItem value="price_per_call">Price (Low to High)</SelectItem>
                  <SelectItem value="total_calls">Popularity</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-1.5">
              <Label className="text-xs text-text-muted">Min Rating</Label>
              <Select
                value={filters.min_rating ? String(filters.min_rating) : ''}
                onValueChange={(v) => handleFilterChange('min_rating', v ? parseFloat(v) as unknown as string | string[] | undefined : undefined)}
              >
                <SelectTrigger className="w-[120px]">
                  <SelectValue placeholder="Any" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__any__">Any</SelectItem>
                  <SelectItem value="4.5">4.5+</SelectItem>
                  <SelectItem value="4">4+</SelectItem>
                  <SelectItem value="3.5">3.5+</SelectItem>
                  <SelectItem value="3">3+</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        )}
      </div>

      {/* Results info */}
      <div className="flex items-center justify-between">
        <p className="text-sm text-text-muted">
          {loading ? 'Loading...' : `${total} agents found`}
        </p>
      </div>

      {/* Agent Grid */}
      {error && (
        <div className="text-center py-12 text-text-muted">{error}</div>
      )}

      {!loading && agents.length === 0 && !error && (
        <div className="text-center py-12 text-text-muted">
          <Bot className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <p>No agents found matching your criteria</p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {agents.map((agent) => (
          <AgentCard
            key={agent.id}
            agent={agent}
            isAuthenticated={isAuthenticated}
            onHire={() => setHireDialog(agent)}
            onSelect={onAgentSelect}
            formatPrice={formatPrice}
          />
        ))}
      </div>

      {/* Pagination */}
      {(hasMore || offset > 0) && (
        <div className="flex justify-center gap-2">
          <Button
            variant="outline"
            onClick={() => handlePageChange(Math.max(0, offset - limit))}
            disabled={offset === 0}
          >
            Previous
          </Button>
          <span className="flex items-center text-sm text-text-muted">
            Showing {offset + 1}-{Math.min(offset + agents.length, total)} of {total}
          </span>
          <Button
            variant="outline"
            onClick={() => handlePageChange(offset + limit)}
            disabled={!hasMore}
          >
            Next
          </Button>
        </div>
      )}

      {/* Hire Dialog */}
      <Dialog open={!!hireDialog} onOpenChange={(open) => !open && setHireDialog(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-12 w-12 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
                <Bot className="h-6 w-6 text-white" />
              </div>
              <div>
                <DialogTitle>{hireDialog?.name}</DialogTitle>
                <p className="text-sm text-text-muted font-mono">{hireDialog?.agentId}</p>
              </div>
            </div>
            <DialogDescription>
              Configure the hiring parameters for this agent. Set a task type and budget to get started.
            </DialogDescription>
          </DialogHeader>

          {hireDialog && (
            <div className="space-y-4 py-4">
              <div className="flex items-center gap-2 p-3 bg-bg-secondary rounded-lg border border-border-subtle">
                <div className="flex items-center gap-1.5 text-sm">
                  <Badge variant="outline" className="text-[10px]">
                    {LISTING_TYPE_LABELS[hireDialog.listingType] ?? hireDialog.listingType}
                  </Badge>
                  <span className="text-text-muted">•</span>
                  <span className="font-medium">{formatPrice(hireDialog)}</span>
                </div>
                <div className="ml-auto flex items-center gap-1.5 text-sm">
                  <Star className="h-3 w-3 text-warning" />
                  <span className="font-medium">{hireDialog.ratingScore.toFixed(1)}</span>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="taskType" className="text-sm font-medium">Task Type</Label>
                <Input
                  id="taskType"
                  placeholder="e.g. code_generation, analysis, data_processing"
                  value={hireTaskType}
                  onChange={(e) => setHireTaskType(e.target.value)}
                />
                <p className="text-xs text-text-muted">Specify the type of task you want this agent to perform</p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="budget" className="text-sm font-medium">Budget (USD)</Label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">$</span>
                  <Input
                    id="budget"
                    type="number"
                    placeholder="0.00"
                    value={hireBudget}
                    onChange={(e) => setHireBudget(e.target.value)}
                    className="pl-7"
                    min="0"
                    step="0.01"
                  />
                </div>
                <p className="text-xs text-text-muted">Maximum amount you're willing to pay. Leave empty for no limit.</p>
              </div>
            </div>
          )}

          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" onClick={() => setHireDialog(null)}>
              Cancel
            </Button>
            <Button onClick={handleHire} disabled={hiring || !hireTaskType.trim()}>
              {hiring ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  Hiring...
                </>
              ) : (
                <>
                  <CheckCircle className="h-4 w-4 mr-2" />
                  Hire Agent
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

interface AgentCardProps {
  agent: MarketplaceAgent;
  isAuthenticated: boolean;
  onHire: () => void;
  onSelect?: (agent: MarketplaceAgent) => void;
  formatPrice: (agent: MarketplaceAgent) => string;
}

function AgentCard({ agent, isAuthenticated, onHire, onSelect, formatPrice }: AgentCardProps) {
  return (
    <Card className="flex flex-col hover:border-brand-500 transition-colors cursor-pointer" onClick={() => onSelect?.(agent)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center shrink-0">
              <Bot className="h-5 w-5 text-white" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                <CardTitle className="text-base truncate">{agent.name}</CardTitle>
                {agent.deterministicVerified && (
                  <Badge variant="success" className="text-[10px] shrink-0">
                    <Shield className="h-3 w-3 mr-0.5" />
                    Verified
                  </Badge>
                )}
                {agent.isOfficial && (
                  <Badge variant="outline" className="text-brand-600 border-brand-300 shrink-0">
                    <Sparkles className="h-3 w-3 mr-0.5" />
                    Official
                  </Badge>
                )}
              </div>
              <p className="text-xs text-text-muted font-mono truncate">{agent.agentId}</p>
            </div>
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col gap-3">
        <CardDescription className="text-sm line-clamp-2 min-h-[2.5rem]">
          {agent.description || 'No description available'}
        </CardDescription>

        {/* Capabilities */}
        {agent.capabilities && agent.capabilities.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {agent.capabilities.slice(0, 4).map((cap) => (
              <Badge key={cap} variant="secondary" className="text-[10px]">
                {cap}
              </Badge>
            ))}
            {agent.capabilities.length > 4 && (
              <Badge variant="secondary" className="text-[10px]">
                +{agent.capabilities.length - 4}
              </Badge>
            )}
          </div>
        )}

        {/* Pricing */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-[10px]">
              {LISTING_TYPE_LABELS[agent.listingType] ?? agent.listingType}
            </Badge>
            <Badge variant="secondary" className="text-[10px]">
              {formatPrice(agent)}
            </Badge>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-2 text-center">
          <div className="flex flex-col items-center p-2 bg-bg-tertiary rounded-lg">
            <Star className="h-3 w-3 text-warning mb-0.5" />
            <span className="text-sm font-medium">{agent.ratingScore.toFixed(1)}</span>
            <span className="text-[10px] text-text-muted">Rating</span>
          </div>
          <div className="flex flex-col items-center p-2 bg-bg-tertiary rounded-lg">
            <TrendingUp className="h-3 w-3 text-success mb-0.5" />
            <span className="text-sm font-medium">{agent.roiScore.toFixed(1)}</span>
            <span className="text-[10px] text-text-muted">ROI</span>
          </div>
          <div className="flex flex-col items-center p-2 bg-bg-tertiary rounded-lg">
            <Bot className="h-3 w-3 text-brand-500 mb-0.5" />
            <span className="text-sm font-medium">{agent.totalCalls.toLocaleString()}</span>
            <span className="text-[10px] text-text-muted">Calls</span>
          </div>
        </div>

        {/* Rank Score */}
        {agent.rankScore !== undefined && (
          <div className="flex items-center justify-between text-xs text-text-muted">
            <span>Rank Score</span>
            <span className="font-medium text-brand-500">{agent.rankScore.toFixed(2)}</span>
          </div>
        )}

        {/* Auth-gated fields */}
        {isAuthenticated && agent.walletBalanceUsd !== undefined && (
          <div className="flex items-center gap-1 text-xs text-text-muted">
            <Wallet className="h-3 w-3" />
            ${agent.walletBalanceUsd.toFixed(2)}
          </div>
        )}
        {isAuthenticated && agent.hiringHistoryCount !== undefined && (
          <div className="flex items-center gap-1 text-xs text-text-muted">
            <CheckCircle className="h-3 w-3" />
            {agent.hiringHistoryCount} hires
          </div>
        )}

        {/* Hire Button */}
        <Button
          className="w-full mt-auto"
          onClick={(e) => { e.stopPropagation(); onHire(); }}
        >
          Hire Agent
        </Button>
      </CardContent>
    </Card>
  );
}

export default AgentMarketplaceView;