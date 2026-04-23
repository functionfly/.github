import { agentApi, type MarketplaceAgent } from '@/api/agent';
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ROUTES } from '@/lib/constants';
import {
  AlertCircle,
  Award,
  Check,
  CheckCircle,
  ChevronDown,
  Copy,
  ExternalLink,
  RefreshCw,
  Search,
  Shield,
  Star,
  TrendingUp,
  Users,
  Zap,
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

export type { MarketplaceAgent };

// ─── Constants ────────────────────────────────────────────────────────────────

const OFFICIAL_AGENT_IDS = new Set([
  'proofsmith',
  'policymint',
  'marginpilot',
  'schemasheriff',
  'patchpulse',
  'runbookweaver',
]);

const CAPABILITY_LABELS: Record<string, string> = {
  determinism_audit: 'Determinism Audit',
  replay_runs: 'Replay Runs',
  variance_report: 'Variance Reports',
  trust_badging: 'Trust Badging',
  trust_policy_design: 'Policy Design',
  capability_deny: 'Capability Deny',
  policy_hashing: 'Policy Hashing',
  compliance_mode: 'Compliance Mode',
  budget_aware_planning: 'Budget Planning',
  function_substitution: 'Function Substitution',
  spend_cap_enforcement: 'Spend Cap',
  roi_reporting: 'ROI Reporting',
  schema_inference: 'Schema Inference',
  validator_generation: 'Validator Gen',
  idempotency_guard: 'Idempotency Guard',
  publish_functions: 'Publish Functions',
  canary_checks: 'Canary Checks',
  diff_detection: 'Diff Detection',
  adapter_patch: 'Adapter Patch',
  version_bump: 'Version Bump',
  incident_triage: 'Incident Triage',
  runbook_generation: 'Runbook Gen',
  verification_checks: 'Verification',
  postmortem_pack: 'Postmortem Pack',
};

type SortOption = 'featured' | 'top_rated' | 'most_used' | 'roi';

// ─── Component ────────────────────────────────────────────────────────────────

interface Filters {
  search: string;
  listingType: string;
  pricingModel: string;
  sort: SortOption;
}

const PAGE_SIZE = 9;

export function AgentMarketplace() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [listings, setListings] = useState<MarketplaceAgent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedAgent, setSelectedAgent] = useState<MarketplaceAgent | null>(null);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState<Filters>({
    search: '',
    listingType: '',
    pricingModel: '',
    sort: 'featured',
  });

  useEffect(() => {
    loadAgents();
  }, []);

  const loadAgents = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await agentApi.searchMarketplaceAgents({ limit: 100 });
      setListings(response.agents);
    } catch (err) {
      console.error('Failed to load marketplace agents:', err);
      setError(t('agentMarket.errorLoad'));
    } finally {
      setLoading(false);
    }
  };

  // ── Filtering ───────────────────────────────────────────────────────────────

  const filtered = useMemo(() => {
    let result = listings.filter((l) => {
      if (
        filters.search &&
        !l.name.toLowerCase().includes(filters.search.toLowerCase()) &&
        !l.description.toLowerCase().includes(filters.search.toLowerCase())
      )
        return false;
      if (filters.listingType && l.listingType !== filters.listingType) return false;
      if (filters.pricingModel && l.pricingModel !== filters.pricingModel) return false;
      return true;
    });

    // Sort
    result = [...result].sort((a, b) => {
      switch (filters.sort) {
        case 'featured':
          // Official agents first, then by rating
          const aOff = OFFICIAL_AGENT_IDS.has(a.agentId) ? 1 : 0;
          const bOff = OFFICIAL_AGENT_IDS.has(b.agentId) ? 1 : 0;
          if (bOff !== aOff) return bOff - aOff;
          return b.ratingScore - a.ratingScore;
        case 'top_rated':
          return b.ratingScore - a.ratingScore;
        case 'most_used':
          return b.totalCalls - a.totalCalls;
        case 'roi':
          return b.roiScore - a.roiScore;
        default:
          return 0;
      }
    });

    return result;
  }, [listings, filters]);

  const paged = filtered.slice(0, page * PAGE_SIZE);
  const hasMore = paged.length < filtered.length;

  // ── Helpers ─────────────────────────────────────────────────────────────────

  const getPriceDisplay = (listing: MarketplaceAgent): string => {
    switch (listing.pricingModel) {
      case 'free':
        return 'Free';
      case 'per_call':
        return `$${listing.pricePerCall?.toFixed(2)}/call`;
      case 'subscription':
        return `$${listing.subscriptionMonthlyUsd}/mo`;
      case 'revenue_share':
        return `${listing.revenueSharePercent}% rev share`;
      default:
        return '';
    }
  };

  const getRoleColor = (type: string) =>
    ({ worker: 'bg-blue-500', manager: 'bg-purple-500', infrastructure: 'bg-orange-500' })[type] ??
    'bg-gray-500';

  // ── Render ──────────────────────────────────────────────────────────────────

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center p-16 gap-3">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        <p className="text-sm text-muted-foreground">{t('common.loading')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold">{t('agentMarket.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('agentMarket.subtitle')}</p>
        </div>
        <div className="flex items-center gap-2">
          {listings.length > 0 && (
            <Badge variant="outline" className="bg-green-500/10 text-green-500">
              <CheckCircle className="h-3 w-3 mr-1" />
              {t('agentMarket.activeAgents', { count: listings.length })}
            </Badge>
          )}
          <Badge variant="outline" className="bg-brand-500/10 text-brand-500">
            <Award className="h-3 w-3 mr-1" />
            {t('agentMarket.official', { count: OFFICIAL_AGENT_IDS.size })}
          </Badge>
        </div>
      </div>

      {/* Error banner */}
      {error && (
        <div className="flex items-start gap-3 rounded-lg border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
          <div className="flex-1">{error}</div>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 shrink-0 text-destructive hover:text-destructive"
            onClick={loadAgents}
          >
            <RefreshCw className="h-3 w-3 mr-1" />
            {t('common.retry')}
          </Button>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col md:flex-row gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={t('agentMarket.searchPlaceholder')}
                className="pl-10"
                value={filters.search}
                onChange={(e) => {
                  setPage(1);
                  setFilters((f) => ({ ...f, search: e.target.value }));
                }}
              />
            </div>

            <Select
              value={filters.listingType || 'all'}
              onValueChange={(v) => {
                setPage(1);
                setFilters((f) => ({ ...f, listingType: v === 'all' ? '' : v }));
              }}
            >
              <SelectTrigger className="w-full md:w-[160px]">
                <SelectValue placeholder={t('agentMarket.allTypes')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('agentMarket.allTypes')}</SelectItem>
                <SelectItem value="worker">{t('agentMarket.worker')}</SelectItem>
                <SelectItem value="manager">{t('agentMarket.manager')}</SelectItem>
                <SelectItem value="infrastructure">{t('agentMarket.infrastructure')}</SelectItem>
              </SelectContent>
            </Select>

            <Select
              value={filters.pricingModel || 'all'}
              onValueChange={(v) => {
                setPage(1);
                setFilters((f) => ({ ...f, pricingModel: v === 'all' ? '' : v }));
              }}
            >
              <SelectTrigger className="w-full md:w-[160px]">
                <SelectValue placeholder={t('agentMarket.allPricing')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('agentMarket.allPricing')}</SelectItem>
                <SelectItem value="free">{t('agentMarket.free')}</SelectItem>
                <SelectItem value="per_call">{t('agentMarket.perCall')}</SelectItem>
                <SelectItem value="subscription">{t('agentMarket.subscription')}</SelectItem>
                <SelectItem value="revenue_share">{t('agentMarket.revenueShare')}</SelectItem>
              </SelectContent>
            </Select>

            <Select
              value={filters.sort}
              onValueChange={(v) => {
                setPage(1);
                setFilters((f) => ({ ...f, sort: v as SortOption }));
              }}
            >
              <SelectTrigger className="w-full md:w-[160px]">
                <SelectValue placeholder={t('agentMarket.sortBy')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="featured">{t('agentMarket.featured')}</SelectItem>
                <SelectItem value="top_rated">{t('agentMarket.topRated')}</SelectItem>
                <SelectItem value="most_used">{t('agentMarket.mostUsed')}</SelectItem>
                <SelectItem value="roi">{t('agentMarket.bestRoi')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Grid */}
      {paged.length > 0 ? (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {paged.map((listing) => (
              <ListingCard
                key={listing.id}
                listing={listing}
                priceDisplay={getPriceDisplay(listing)}
                roleColor={getRoleColor(listing.listingType)}
                onHire={() => setSelectedAgent(listing)}
              />
            ))}
          </div>

          {hasMore && (
            <div className="flex justify-center pt-2">
              <Button variant="outline" onClick={() => setPage((p) => p + 1)}>
                <ChevronDown className="h-4 w-4 mr-2" />
                {t('agentMarket.loadMore', { count: filtered.length - paged.length })}
              </Button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-16">
          <Users className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">{t('agentMarket.noAgentsFound')}</h3>
          <p className="text-muted-foreground mt-1">
            {filters.search || filters.listingType || filters.pricingModel
              ? t('agentMarket.adjustFilters')
              : t('agentMarket.seedScript')}
          </p>
          {(filters.search || filters.listingType || filters.pricingModel) && (
            <Button
              variant="outline"
              className="mt-4"
              onClick={() =>
                setFilters({ search: '', listingType: '', pricingModel: '', sort: 'featured' })
              }
            >
              {t('agentMarket.clearFilters')}
            </Button>
          )}
        </div>
      )}

      {/* Hire dialog */}
      <HireAgentDialog
        agent={selectedAgent}
        priceDisplay={selectedAgent ? getPriceDisplay(selectedAgent) : ''}
        onClose={() => setSelectedAgent(null)}
        onGoToAgents={() => {
          setSelectedAgent(null);
          navigate(ROUTES.AGENTS);
        }}
        onGoToSDK={() => {
          setSelectedAgent(null);
          navigate(ROUTES.SDK_INTEGRATIONS);
        }}
      />
    </div>
  );
}

// ─── ListingCard ──────────────────────────────────────────────────────────────

function StarRating({ score }: { score: number }) {
  return (
    <div className="flex items-center gap-0.5">
      {[1, 2, 3, 4, 5].map((s) => (
        <Star
          key={s}
          className={`h-3 w-3 ${s <= Math.round(score) ? 'text-yellow-400 fill-yellow-400' : 'text-muted-foreground/30'}`}
        />
      ))}
      <span className="text-xs font-medium ml-1">{score.toFixed(1)}</span>
    </div>
  );
}

function StatusDot({ status }: { status?: string }) {
  const isActive = !status || status === 'active';
  return (
    <span
      className={`inline-block h-2 w-2 rounded-full ${isActive ? 'bg-green-500' : 'bg-red-400'}`}
      title={isActive ? 'Active' : status}
    />
  );
}

function ListingCard({
  listing,
  priceDisplay,
  roleColor,
  onHire,
}: {
  listing: MarketplaceAgent;
  priceDisplay: string;
  roleColor: string;
  onHire: () => void;
}) {
  const { t } = useTranslation();
  const isOfficial = OFFICIAL_AGENT_IDS.has(listing.agentId);
  const caps = listing.capabilities ?? [];

  return (
    <Card className="hover:shadow-lg transition-shadow flex flex-col relative overflow-hidden">
      {/* Official banner strip */}
      {isOfficial && (
        <div className="absolute top-0 right-0 bg-brand-500 text-white text-[10px] font-bold px-2 py-0.5 rounded-bl-md flex items-center gap-1">
          <Award className="h-2.5 w-2.5" />
          {t('agentMarket.officialBadge')}
        </div>
      )}

      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="space-y-1 flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <StatusDot status={(listing as MarketplaceAgent & { status?: string }).status} />
              <CardTitle className="text-base leading-tight truncate">{listing.name}</CardTitle>
              {listing.deterministicVerified && (
                <span title="Deterministic Verified">
                  <Shield className="h-4 w-4 text-green-500 shrink-0" />
                </span>
              )}
            </div>
            <CardDescription className="line-clamp-2 text-xs">
              {listing.description}
            </CardDescription>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-3 flex-1">
        {/* Badges row */}
        <div className="flex items-center flex-wrap gap-1.5">
          <Badge className={`${roleColor} text-white text-xs capitalize`}>
            {listing.listingType}
          </Badge>
          <Badge variant="outline" className="text-xs">
            {priceDisplay}
          </Badge>
        </div>

        {/* Star rating */}
        <StarRating score={listing.ratingScore} />

        {/* Metrics */}
        <div className="grid grid-cols-3 gap-2 text-center">
          <div>
            <div className="flex items-center justify-center gap-1">
              <Zap className="h-3 w-3 text-blue-500" />
              <span className="text-xs font-semibold">{listing.totalCalls.toLocaleString()}</span>
            </div>
            <p className="text-[10px] text-muted-foreground">{t('agentMarket.calls')}</p>
          </div>
          <div>
            <div className="flex items-center justify-center gap-1">
              <TrendingUp className="h-3 w-3 text-green-500" />
              <span className="text-xs font-semibold">{listing.roiScore}%</span>
            </div>
            <p className="text-[10px] text-muted-foreground">{t('agentMarket.roi')}</p>
          </div>
          <div>
            <div className="flex items-center justify-center gap-1">
              <Shield className="h-3 w-3 text-purple-500" />
              <span className="text-xs font-semibold">
                {listing.trustScore ?? '—'}
                {listing.trustScore ? '%' : ''}
              </span>
            </div>
            <p className="text-[10px] text-muted-foreground">{t('agentMarket.trust')}</p>
          </div>
        </div>

        {/* Trust score bar */}
        {listing.trustScore !== undefined && (
          <Progress
            value={listing.trustScore}
            className="h-1.5"
            indicatorClassName="bg-emerald-500"
          />
        )}

        {/* Capabilities */}
        {caps.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {caps.slice(0, 3).map((cap) => (
              <Badge key={cap} variant="secondary" className="text-[10px] px-1.5 py-0">
                {CAPABILITY_LABELS[cap] ?? cap.replace(/_/g, ' ')}
              </Badge>
            ))}
            {caps.length > 3 && (
              <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
                +{caps.length - 3}
              </Badge>
            )}
          </div>
        )}
      </CardContent>

      <CardFooter className="pt-3">
        <Button className="w-full" size="sm" onClick={onHire}>
          {t('agentMarket.hireAgent')}
          <ExternalLink className="h-3.5 w-3.5 ml-2" />
        </Button>
      </CardFooter>
    </Card>
  );
}

// ─── HireAgentDialog ──────────────────────────────────────────────────────────

function HireAgentDialog({
  agent,
  priceDisplay,
  onClose,
  onGoToAgents,
  onGoToSDK,
}: {
  agent: MarketplaceAgent | null;
  priceDisplay: string;
  onClose: () => void;
  onGoToAgents: () => void;
  onGoToSDK: () => void;
}) {
  const { t } = useTranslation();
  const [copiedSnippet, setCopiedSnippet] = useState(false);
  const [copiedId, setCopiedId] = useState(false);

  if (!agent) return null;

  const isOfficial = OFFICIAL_AGENT_IDS.has(agent.agentId);
  const caps = agent.capabilities ?? [];

  const snippet = `from flypy import AgentClient

client = AgentClient(
    api_base="https://api.functionfly.com",
    api_key="YOUR_API_KEY",
)

result = client.execute(
    agent_id="${agent.agentId}",
    input={"task": "describe your task here"},
)
print(result)`;

  const copy = async (text: string, which: 'snippet' | 'id') => {
    await navigator.clipboard.writeText(text);
    if (which === 'snippet') {
      setCopiedSnippet(true);
      toast.success(t('agentMarket.codeCopied'));
      setTimeout(() => setCopiedSnippet(false), 2000);
    } else {
      setCopiedId(true);
      toast.success(t('agentMarket.agentIdCopied'));
      setTimeout(() => setCopiedId(false), 2000);
    }
  };

  const roleColor =
    agent.listingType === 'worker'
      ? 'bg-blue-500'
      : agent.listingType === 'manager'
        ? 'bg-purple-500'
        : 'bg-orange-500';

  return (
    <Dialog
      open={!!agent}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {isOfficial && <Award className="h-4 w-4 text-brand-500" />}
            {agent.name}
            {agent.deterministicVerified && <Shield className="h-4 w-4 text-green-500" />}
          </DialogTitle>
          <DialogDescription>{agent.description}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Badges */}
          <div className="flex flex-wrap items-center gap-2">
            <Badge className={`${roleColor} text-white capitalize`}>{agent.listingType}</Badge>
            <Badge variant="outline">{priceDisplay}</Badge>
            {isOfficial && (
              <Badge
                variant="outline"
                className="bg-brand-500/10 text-brand-500 border-brand-500/30"
              >
                <Award className="h-3 w-3 mr-1" />
                {t('agentMarket.officialByFunctionFly')}
              </Badge>
            )}
            {agent.deterministicVerified && (
              <Badge
                variant="outline"
                className="bg-green-500/10 text-green-600 border-green-500/30"
              >
                <CheckCircle className="h-3 w-3 mr-1" />
                {t('agentMarket.deterministic')}
              </Badge>
            )}
          </div>

          {/* Rating */}
          <div className="flex items-center gap-3">
            <StarRating score={agent.ratingScore} />
            <span className="text-xs text-muted-foreground">·</span>
            <span className="text-xs text-muted-foreground">
              {agent.totalCalls.toLocaleString()} calls
            </span>
            <span className="text-xs text-muted-foreground">·</span>
            <span className="text-xs text-muted-foreground">{agent.roiScore}% ROI</span>
          </div>

          {/* Trust bar */}
          {agent.trustScore !== undefined && (
            <div className="space-y-1.5">
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Trust Score</span>
                <span className="font-medium">{agent.trustScore}%</span>
              </div>
              <Progress
                value={agent.trustScore}
                className="h-2"
                indicatorClassName="bg-emerald-500"
              />
            </div>
          )}

          {/* Capabilities */}
          {caps.length > 0 && (
            <div className="space-y-1.5">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                Capabilities
              </p>
              <div className="flex flex-wrap gap-1">
                {caps.map((cap) => (
                  <Badge key={cap} variant="secondary" className="text-xs">
                    {CAPABILITY_LABELS[cap] ?? cap.replace(/_/g, ' ')}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* Agent ID */}
          <div className="space-y-1.5">
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              {t('agentMarket.agentId')}
            </p>
            <div className="flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
              <code className="flex-1 text-sm font-mono">{agent.agentId}</code>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0 shrink-0"
                onClick={() => copy(agent.agentId, 'id')}
              >
                {copiedId ? (
                  <Check className="h-3 w-3 text-green-500" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
              </Button>
            </div>
          </div>

          {/* Code snippet */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                {t('agentMarket.quickStartPython')}
              </p>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 gap-1 text-xs"
                onClick={() => copy(snippet, 'snippet')}
              >
                {copiedSnippet ? (
                  <Check className="h-3 w-3 text-green-500" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
                {copiedSnippet ? t('common.copied') : t('common.copy')}
              </Button>
            </div>
            <pre className="rounded-md border bg-muted/50 p-3 text-xs font-mono overflow-x-auto leading-relaxed">
              {snippet}
            </pre>
          </div>

          {/* Next steps */}
          <div className="rounded-md border bg-muted/30 px-4 py-3 space-y-1">
            <p className="text-xs font-semibold">{t('agentMarket.nextSteps')}</p>
            <ol className="text-xs text-muted-foreground space-y-1 list-decimal list-inside">
              <li>
                {t('agentMarket.nextStep1', { 1: 'strong' })}
              </li>
              <li>
                {t('agentMarket.nextStep2', { 1: 'code' })}
              </li>
              <li>{t('agentMarket.nextStep3')}</li>
            </ol>
          </div>
        </div>

        <DialogFooter className="flex-col sm:flex-row gap-2">
          <Button variant="outline" onClick={onClose} className="sm:flex-none">
            {t('agentMarket.close')}
          </Button>
          <Button variant="outline" onClick={onGoToSDK} className="sm:flex-none">
            {t('agentMarket.sdkDocs')}
          </Button>
          <Button onClick={onGoToAgents} className="flex-1 sm:flex-none">
            {t('agentMarket.goToMyAgents')}
            <ExternalLink className="h-4 w-4 ml-2" />
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default AgentMarketplace;
