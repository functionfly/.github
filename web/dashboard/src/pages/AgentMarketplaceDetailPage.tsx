import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { agentApi, type MarketplaceAgent } from '@/api/agent';
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useAuthStore } from '@/stores/authStore';
import {
  ArrowLeft,
  Bot,
  CheckCircle,
  Loader2,
  Shield,
  Sparkles,
  Star,
  TrendingUp,
  Wallet,
} from 'lucide-react';
import { toast } from 'sonner';

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

export function AgentMarketplaceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const isAuthenticated = useAuthStore((s) => !!s.user);

  const [agent, setAgent] = useState<MarketplaceAgent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [hireDialog, setHireDialog] = useState(false);
  const [hireTaskType, setHireTaskType] = useState('');
  const [hireBudget, setHireBudget] = useState('');
  const [hiring, setHiring] = useState(false);

  const loadAgent = async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const response = await agentApi.searchMarketplaceAgents({ agent_id: id, limit: 1 });
      const found = response.agents.find((a) => a.id === id || a.agentId === id) ?? response.agents[0];
      if (found) {
        setAgent(found);
        return;
      }
    } catch {
      // marketplace search failed, fall through to agent API
    }

    try {
      const res = await agentApi.getAgent(id);
      const a = res.agent as Record<string, unknown>;
      setAgent({
        id: (a.id as string) ?? id,
        agentId: (a.agentId as string) ?? id,
        name: (a.name as string) ?? 'Unknown Agent',
        description: (a.description as string) ?? '',
        listingType: 'worker',
        pricingModel: 'free',
        ratingScore: 0,
        totalCalls: 0,
        roiScore: 0,
        rankScore: 0,
        walletBalanceUsd: 0,
      });
    } catch {
      setError('Agent not found');
    }
  };

  useEffect(() => {
    loadAgent().finally(() => setLoading(false));
  }, [id]);

  const handleHire = async () => {
    if (!agent) return;
    setHiring(true);
    try {
      await agentApi.hireAgent({
        agent_id: agent.agentId,
        task_type: hireTaskType || 'general',
        budget_usd: hireBudget ? parseFloat(hireBudget) : undefined,
      });
      toast.success(`Successfully hired ${agent.name}`);
      setHireDialog(false);
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

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-16 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <p className="text-sm">Loading agent...</p>
      </div>
    );
  }

  if (error || !agent) {
    return (
      <div className="space-y-4">
        <Link to="/marketplace/agents" className="flex items-center text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Marketplace
        </Link>
        <Card className="border-destructive/40">
          <CardHeader>
            <CardTitle>Agent Not Found</CardTitle>
            <CardDescription>{error ?? 'The agent you are looking for does not exist.'}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-2">
          <Link to="/marketplace/agents" className="flex items-center text-sm text-muted-foreground hover:text-foreground w-fit">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Marketplace
          </Link>
          <div className="flex items-center gap-3">
            <div className="h-14 w-14 rounded-xl bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
              <Bot className="h-7 w-7 text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-bold flex items-center gap-2">
                {agent.name}
                {agent.deterministicVerified && (
                  <Badge variant="success" className="text-[10px]">
                    <Shield className="h-3 w-3 mr-0.5" />
                    Verified
                  </Badge>
                )}
                {agent.isOfficial && (
                  <Badge variant="default" className="text-[10px]">
                    <Sparkles className="h-3 w-3 mr-0.5" />
                    Official
                  </Badge>
                )}
              </h1>
              <p className="font-mono text-xs text-muted-foreground">{agent.agentId}</p>
            </div>
          </div>
        </div>
        <Button onClick={() => setHireDialog(true)}>
          Hire Agent
        </Button>
      </div>

      {/* Badges */}
      <div className="flex flex-wrap gap-2">
        <Badge variant="outline">
          {LISTING_TYPE_LABELS[agent.listingType] ?? agent.listingType}
        </Badge>
        <Badge variant="secondary">
          {PRICING_MODEL_LABELS[agent.pricingModel] ?? agent.pricingModel}
        </Badge>
        <Badge variant="outline">
          {formatPrice(agent)}
        </Badge>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 mb-2">
              <Star className="h-4 w-4 text-warning" />
              <span className="text-sm text-muted-foreground">Rating</span>
            </div>
            <p className="text-2xl font-bold">{(agent.ratingScore ?? 0).toFixed(1)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="h-4 w-4 text-success" />
              <span className="text-sm text-muted-foreground">ROI Score</span>
            </div>
            <p className="text-2xl font-bold">{(agent.roiScore ?? 0).toFixed(1)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 mb-2">
              <Bot className="h-4 w-4 text-brand-500" />
              <span className="text-sm text-muted-foreground">Total Calls</span>
            </div>
            <p className="text-2xl font-bold">{agent.totalCalls.toLocaleString()}</p>
          </CardContent>
        </Card>
        {agent.rankScore !== undefined && (
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-2 mb-2">
                <Sparkles className="h-4 w-4 text-purple-500" />
                <span className="text-sm text-muted-foreground">Rank Score</span>
              </div>
              <p className="text-2xl font-bold">{agent.rankScore.toFixed(2)}</p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Tabs */}
      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="capabilities">Capabilities</TabsTrigger>
          <TabsTrigger value="pricing">Pricing</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>About</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                {agent.description || 'No description available.'}
              </p>
            </CardContent>
          </Card>

          {isAuthenticated && (agent.walletBalanceUsd !== undefined || agent.hiringHistoryCount !== undefined) && (
            <Card>
              <CardHeader>
                <CardTitle>Agent Stats</CardTitle>
              </CardHeader>
              <CardContent className="flex gap-6">
                {agent.walletBalanceUsd !== undefined && (
                  <div className="flex items-center gap-2">
                    <Wallet className="h-5 w-5 text-muted-foreground" />
                    <div>
                      <p className="text-sm text-muted-foreground">Wallet Balance</p>
                      <p className="text-lg font-medium">${agent.walletBalanceUsd.toFixed(2)}</p>
                    </div>
                  </div>
                )}
                {agent.hiringHistoryCount !== undefined && (
                  <div className="flex items-center gap-2">
                    <CheckCircle className="h-5 w-5 text-muted-foreground" />
                    <div>
                      <p className="text-sm text-muted-foreground">Total Hires</p>
                      <p className="text-lg font-medium">{agent.hiringHistoryCount}</p>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="capabilities" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Capabilities</CardTitle>
              <CardDescription>Skills and features this agent provides</CardDescription>
            </CardHeader>
            <CardContent>
              {agent.capabilities && agent.capabilities.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {agent.capabilities.map((cap) => (
                    <Badge key={cap} variant="secondary" className="text-sm px-3 py-1">
                      {cap}
                    </Badge>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No capabilities listed</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="pricing" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Pricing Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">Pricing Model</p>
                  <p className="font-medium">{PRICING_MODEL_LABELS[agent.pricingModel] ?? agent.pricingModel}</p>
                </div>
                {agent.pricePerCall && (
                  <div>
                    <p className="text-sm text-muted-foreground">Price per Call</p>
                    <p className="font-medium">${agent.pricePerCall.toFixed(4)}</p>
                  </div>
                )}
                {agent.subscriptionMonthlyUsd && (
                  <div>
                    <p className="text-sm text-muted-foreground">Monthly Subscription</p>
                    <p className="font-medium">${agent.subscriptionMonthlyUsd.toFixed(2)}</p>
                  </div>
                )}
                {agent.revenueSharePercent && (
                  <div>
                    <p className="text-sm text-muted-foreground">Revenue Share</p>
                    <p className="font-medium">{agent.revenueSharePercent}%</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Hire Dialog */}
      <Dialog open={hireDialog} onOpenChange={setHireDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Hire {agent.name}</DialogTitle>
            <DialogDescription>
              Configure the hiring parameters for this agent.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="taskType">Task Type</Label>
              <Input
                id="taskType"
                placeholder="e.g. code_generation, analysis"
                value={hireTaskType}
                onChange={(e) => setHireTaskType(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="budget">Budget (USD)</Label>
              <Input
                id="budget"
                type="number"
                placeholder="Optional"
                value={hireBudget}
                onChange={(e) => setHireBudget(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setHireDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleHire} disabled={hiring}>
              {hiring ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Hire Agent'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default AgentMarketplaceDetailPage;