/**
 * @functionfly/ui-marketplace
 * Function marketplace and registry components
 */

import * as React from "react";
import { cn, Badge } from "@functionfly/ui-core";
import {
  Search,
  Zap,
  Shield,
  Star,
  Eye,
  Code,
  Download,
  Heart,
  Share2,
  GitBranch,
  Clock,
  AlertCircle,
  CheckCircle,
  Loader2,
  Filter,
  SortAsc,
  Tag,
  Sparkles,
  TrendingUp,
  Layers,
  Play,
  AlertTriangle,
  RefreshCw,
} from "lucide-react";

// --- Types ---
export type PricingModel = "free" | "per_call" | "subscription" | "revenue_share";

export type FunctionCardVariant = "compact" | "expanded" | "analytics";

export interface FunctionAuthor {
  id: string;
  username: string;
  name: string;
  avatar?: string;
  profileUrl?: string;
}

export interface FunctionMetrics {
  executionCount: number;
  executionTrend?: number[];
  averageLatency?: number;
  errorRate?: number;
}

export interface FunctionRating {
  average: number;
  count: number;
  distribution?: Record<number, number>;
}

export interface FunctionCardData {
  id: string;
  name: string;
  description: string;
  author: FunctionAuthor;
  trustScore: number;
  metrics: FunctionMetrics;
  pricing: {
    model: PricingModel;
    pricePerCall?: number;
    currency?: string;
  };
  isVerified: boolean;
  isDeterministic: boolean;
  rating: FunctionRating;
  tags?: string[];
  category?: string;
  language?: string;
  lastUpdated?: string;
  version?: string;
  isFavorite?: boolean;
  isFeatured?: boolean;
}

export interface FunctionMarketplaceProps {
  functions: FunctionCardData[];
  onSelect?: (id: string) => void;
  onExecute?: (id: string) => void;
  onFavorite?: (id: string, isFavorite: boolean) => void;
  onShare?: (id: string) => void;
  searchQuery?: string;
  onSearchChange?: (query: string) => void;
  categoryFilter?: string;
  onCategoryChange?: (category: string) => void;
  isLoading?: boolean;
  className?: string;
}

export interface FunctionCardProps extends FunctionCardData {
  variant?: FunctionCardVariant;
  className?: string;
  onView?: (id: string) => void;
  onExecute?: (id: string) => void;
  onFavorite?: (id: string, isFavorite: boolean) => void;
  onShare?: (id: string) => void;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onAdminAction?: (id: string, action: string) => void;
}

export interface FunctionProfileProps {
  data: FunctionCardData;
  onExecute?: () => void;
  onFavorite?: () => void;
  onShare?: () => void;
}

export interface VerifiedBadgeProps {
  isVerified: boolean;
  trustScore: number;
  size?: "sm" | "md";
}

export interface TrustScoreMeterProps {
  score: number;
  size?: "sm" | "md" | "lg";
  showLabel?: boolean;
}

// --- Helpers ---
function getTrustTier(score: number): string {
  if (score >= 80) return "excellent";
  if (score >= 60) return "good";
  if (score >= 40) return "fair";
  return "poor";
}

function getTrustColor(score: number): string {
  if (score >= 80) return "#10b981";
  if (score >= 60) return "#f59e0b";
  if (score >= 40) return "#f97316";
  return "#ef4444";
}

function formatRating(avg: number): string {
  return avg.toFixed(1);
}

function renderStars(rating: number, count?: number): React.ReactElement {
  const full = Math.floor(rating);
  const half = rating % 1 >= 0.5 ? 1 : 0;
  const empty = 5 - full - half;
  return (
    <div className="flex items-center gap-0.5">
      {Array(full).fill(0).map((_, i) => (
        <Star key={`full-${i}`} className="size-3.5 fill-yellow-400 text-yellow-400" />
      ))}
      {half > 0 && <Star className="size-3.5 fill-yellow-400 text-yellow-400/50" />}
      {Array(empty).fill(0).map((_, i) => (
        <Star key={`empty-${i}`} className="size-3.5 fill-transparent text-gray-600" />
      ))}
      {count != null && <span className="text-[10px] text-text-muted ml-1">({count})</span>}
    </div>
  );
}

// --- FunctionCard ---
export function FunctionCard({
  variant = "compact",
  className,
  onView,
  onExecute,
  onFavorite,
  onShare,
  onEdit,
  onDelete,
  onAdminAction,
  ...data
}: FunctionCardProps) {
  if (variant === "expanded") {
    return (
      <div
        className={cn(
          "bg-bg-primary border border-border-subtle rounded-xl overflow-hidden transition-all duration-200",
          "hover:border-border-default hover:shadow-lg",
          className
        )}
      >
        {/* Header */}
        <div className="p-4">
          <div className="flex items-start gap-3">
            <div className="size-10 rounded-lg bg-bg-tertiary flex items-center justify-center text-xl shrink-0">
              {data.name[0]}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-text-primary truncate">{data.name}</h3>
                {data.isVerified && <VerifiedBadge isVerified size="sm" trustScore={data.trustScore} />}
                {data.pricing.model === "free" && (
                  <Badge variant="brand" size="sm">Free</Badge>
                )}
              </div>
              <p className="text-[11px] text-text-muted mt-0.5 line-clamp-2">{data.description}</p>
              <div className="flex items-center gap-2 mt-2 text-[10px] text-text-muted">
                <span className="flex items-center gap-1">
                  by <span className="font-medium text-text-primary">{data.author.name}</span>
                </span>
                {data.version && <span>v{data.version}</span>}
                {data.language && (
                  <Badge variant="ghost" size="sm">
                    {data.language}
                  </Badge>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Metrics */}
        <div className="px-4 py-3 bg-bg-secondary border-y border-border-subtle grid grid-cols-3 gap-4">
          <div className="text-center">
            <div className="text-sm font-bold text-text-primary">{data.metrics.executionCount.toLocaleString()}</div>
            <div className="text-[10px] text-text-muted">Executions</div>
          </div>
          <div className="text-center">
            <div className="text-sm font-bold text-text-primary">{data.metrics.averageLatency ?? "—"}ms</div>
            <div className="text-[10px] text-text-muted">Avg Latency</div>
          </div>
          <div className="text-center">
            <div className="flex items-center justify-center gap-1">
              <span className="text-sm font-bold text-text-primary">{formatRating(data.rating.average)}</span>
              <Star className="size-3.5 fill-yellow-400 text-yellow-400" />
            </div>
            <div className="text-[10px] text-text-muted">{data.rating.count} reviews</div>
          </div>
        </div>

        {/* Trust + Pricing */}
        <div className="px-4 py-3 flex items-center justify-between">
          <TrustScoreMeter score={data.trustScore} size="sm" />
          {data.pricing.model === "per_call" && (
            <div className="text-sm font-bold text-brand-500">${(data.pricing.pricePerCall ?? 0).toFixed(3)}/call</div>
          )}
        </div>

        {/* Actions */}
        <div className="px-4 py-3 flex items-center gap-1 border-t border-border-subtle">
          {onExecute && (
            <button
              onClick={() => onExecute(data.id)}
              className="flex-1 py-2 text-[11px] bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors flex items-center justify-center gap-1.5"
            >
              <Zap className="size-3.5" /> Execute
            </button>
          )}
          {onView && (
            <button
              onClick={() => onView(data.id)}
              className="py-2 px-3 text-[11px] bg-bg-tertiary hover:bg-bg-hover text-text-secondary rounded-lg transition-colors flex items-center gap-1.5 border border-border-subtle"
            >
              <Eye className="size-3.5" /> View
            </button>
          )}
          {onShare && (
            <button
              onClick={() => onShare(data.id)}
              className="p-2 text-text-muted hover:text-text-primary hover:bg-bg-hover rounded-lg transition-colors"
              title="Share"
            >
              <Share2 className="size-4" />
            </button>
          )}
          {onFavorite && (
            <button
              onClick={() => onFavorite(data.id, !data.isFavorite)}
              className={cn(
                "p-2 rounded-lg transition-colors",
                data.isFavorite
                  ? "text-red-400 hover:text-red-500 hover:bg-red-500/10"
                  : "text-text-muted hover:text-text-primary hover:bg-bg-hover"
              )}
              title={data.isFavorite ? "Remove from favorites" : "Add to favorites"}
            >
              <Heart className={cn("size-4", data.isFavorite && "fill-current")} />
            </button>
          )}
        </div>
      </div>
    );
  }

  // Compact variant
  return (
    <div
      className={cn(
        "bg-bg-primary border border-border-subtle rounded-xl p-4 transition-all duration-200 group",
        "hover:border-border-default hover:shadow-md hover:shadow-black/10",
        className
      )}
    >
      {/* Top bar */}
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <div className="size-8 rounded-lg bg-bg-tertiary flex items-center justify-center text-sm font-bold shrink-0">
            {data.name[0]}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <h3 className="text-sm font-semibold text-text-primary truncate">{data.name}</h3>
              {data.isVerified && <VerifiedBadge isVerified size="sm" trustScore={data.trustScore} />}
            </div>
            <div className="text-[11px] text-text-muted truncate">by {data.author.name}</div>
          </div>
        </div>
        <button
          onClick={() => onFavorite?.(data.id, !data.isFavorite)}
          className={cn(
            "shrink-0 p-1 rounded-md transition-colors",
            data.isFavorite
              ? "text-red-400 hover:text-red-500"
              : "text-text-muted hover:text-text-primary opacity-0 group-hover:opacity-100"
          )}
        >
          <Heart className={cn("size-4", data.isFavorite && "fill-current")} />
        </button>
      </div>

      {/* Description */}
      <p className="text-[11px] text-text-muted line-clamp-2 mb-3">{data.description}</p>

      {/* Metrics row */}
      <div className="flex items-center gap-3 mb-3 text-[10px] text-text-muted">
        <span className="flex items-center gap-1">
          <Eye className="size-3" />
          {data.metrics.executionCount.toLocaleString()}
        </span>
        <span className="flex items-center gap-1">
          <Clock className="size-3" />
          {data.metrics.averageLatency ?? "—"}ms
        </span>
        {data.pricing.model === "free" && (
          <Badge variant="brand" size="sm">Free</Badge>
        )}
        {data.pricing.model === "per_call" && (
          <span className="font-medium text-brand-500">${(data.pricing.pricePerCall ?? 0).toFixed(3)}/call</span>
        )}
      </div>

      {/* Trust + Rating */}
      <div className="flex items-center justify-between mb-3">
        <TrustScoreMeter score={data.trustScore} size="sm" />
        {renderStars(data.rating.average, data.rating.count)}
      </div>

      {/* Tags */}
      <div className="flex items-center gap-1 mb-3">
        {data.tags?.slice(0, 3).map((tag) => (
          <span key={tag} className="px-1.5 py-0.5 text-[9px] bg-bg-tertiary text-text-muted rounded capitalize">
            {tag}
          </span>
        ))}
        {data.category && (
          <span className="px-1.5 py-0.5 text-[9px] bg-brand-500/10 text-brand-400 rounded capitalize">
            {data.category}
          </span>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 pt-2 border-t border-border-subtle">
        {onExecute && (
          <button
            onClick={() => onExecute(data.id)}
            className="flex-1 py-1.5 text-[11px] bg-brand-500/10 hover:bg-brand-500/20 text-brand-400 rounded-lg transition-colors flex items-center justify-center gap-1"
          >
            <Zap className="size-3" /> Quick Execute
          </button>
        )}
        {onView && (
          <button
            onClick={() => onView(data.id)}
            className="flex items-center gap-1 px-3 py-1.5 text-[11px] bg-bg-tertiary hover:bg-bg-hover text-text-secondary rounded-lg transition-colors"
          >
            <Code className="size-3" />
            View
          </button>
        )}
        <button className="p-1.5 text-text-muted hover:text-text-primary hover:bg-bg-hover rounded-lg transition-colors">
          <GitBranch className="size-3" />
        </button>
      </div>
    </div>
  );
}

// --- VerifiedBadge ---
export function VerifiedBadge({ isVerified, trustScore, size = "md" }: VerifiedBadgeProps) {
  const sizeClasses = { sm: "px-1.5 py-0.5 text-[9px]", md: "px-2 py-0.5 text-[10px]" };

  if (!isVerified) return null;

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full font-medium",
        sizeClasses[size],
        trustScore >= 80
          ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
          : "bg-brand-500/20 text-brand-400 border border-brand-500/30"
      )}
    >
      <Shield className="size-3" />
      Verified
    </span>
  );
}

// --- TrustScoreMeter ---
export function TrustScoreMeter({ score, size = "md", showLabel = true }: TrustScoreMeterProps) {
  const tier = getTrustTier(score);
  const color = getTrustColor(score);

  const sizeClasses = {
    sm: { bar: "h-1.5 w-16", text: "text-[10px]" },
    md: { bar: "h-2 w-24", text: "text-xs" },
    lg: { bar: "h-3 w-32", text: "text-sm" },
  };

  return (
    <div className="flex items-center gap-2">
      <div className={sizeClasses[size].bar + " bg-bg-tertiary rounded-full overflow-hidden"}>
        <div
          className="h-full rounded-full transition-all duration-500"
          style={{
            width: `${score}%`,
            backgroundColor: color,
          }}
        />
      </div>
      {showLabel && (
        <span className={sizeClasses[size].text} style={{ color }}>
          {score}/100 · {tier}
        </span>
      )}
    </div>
  );
}

// --- FunctionMarketplace ---
export function FunctionMarketplace({
  functions,
  onSelect,
  onExecute,
  onFavorite,
  onShare,
  searchQuery,
  onSearchChange,
  categoryFilter,
  onCategoryChange,
  isLoading,
  className,
}: FunctionMarketplaceProps) {
  const allCategories = ["All", ...new Set(functions.map((f) => f.category).filter(Boolean))];

  const filtered = functions.filter((f) => {
    const matchesSearch = !searchQuery || f.name.toLowerCase().includes(searchQuery.toLowerCase()) || f.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = !categoryFilter || categoryFilter === "All" || f.category === categoryFilter;
    return matchesSearch && matchesCategory;
  });

  return (
    <div className={cn("space-y-4", className)}>
      {/* Toolbar */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-text-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => onSearchChange?.(e.target.value)}
            placeholder="Search functions..."
            className="w-full pl-9 pr-4 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500 transition-colors"
          />
        </div>
        <div className="flex gap-1">
          {allCategories.map((cat) => (
            <button
              key={cat}
              onClick={() => onCategoryChange?.(cat)}
              className={cn(
                "px-3 py-1.5 text-[11px] rounded-lg transition-colors",
                categoryFilter === cat || (!categoryFilter && cat === "All")
                  ? "bg-brand-500/20 text-brand-400 border border-brand-500/30"
                  : "bg-bg-secondary text-text-muted border border-border-subtle hover:border-border-default"
              )}
            >
              {cat}
            </button>
          ))}
        </div>
        <button className="p-2 text-text-muted hover:text-text-primary hover:bg-bg-tertiary rounded-lg transition-colors">
          <Filter className="size-4" />
        </button>
        <button className="p-2 text-text-muted hover:text-text-primary hover:bg-bg-tertiary rounded-lg transition-colors">
          <SortAsc className="size-4" />
        </button>
      </div>

      {/* Loading state */}
      {isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="bg-bg-secondary border border-border-subtle rounded-xl p-4 animate-pulse">
              <div className="flex items-center gap-3 mb-3">
                <div className="size-10 rounded-lg bg-bg-tertiary/50 animate-pulse" />
                <div className="space-y-2 flex-1">
                  <div className="h-3 bg-bg-tertiary/50 rounded w-1/2 animate-pulse" />
                  <div className="h-2 bg-bg-tertiary/50 rounded w-1/3 animate-pulse" />
                </div>
              </div>
              <div className="h-2 bg-bg-tertiary/50 rounded w-3/4 mb-2 animate-pulse" />
              <div className="h-2 bg-bg-tertiary/50 rounded w-1/2 animate-pulse" />
            </div>
          ))}
        </div>
      )}

      {/* Grid */}
      {!isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {filtered.length > 0 ? (
            filtered.map((fn) => (
              <FunctionCard
                key={fn.id}
                {...fn}
                variant="expanded"
                onView={onSelect}
                onExecute={onExecute}
                onFavorite={onFavorite}
                onShare={onShare}
              />
            ))
          ) : (
            <div className="col-span-full flex flex-col items-center justify-center py-16 text-text-muted">
              <Sparkles className="size-16 mb-4 opacity-30" />
              <h3 className="text-lg font-medium text-text-primary mb-1">No functions found</h3>
              <p className="text-sm">Try adjusting your filters or search terms</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

// --- InstallFunctionModal (simplified) ---
export function InstallFunctionModal({
  isOpen,
  onClose,
  onInstall,
  functionData,
}: {
  isOpen: boolean;
  onClose: () => void;
  onInstall: () => void;
  functionData?: FunctionCardData;
}) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div
        className="bg-bg-secondary border border-border-subtle rounded-2xl p-6 w-full max-w-md mx-4 shadow-2xl animate-in zoom-in-95"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-semibold text-text-primary mb-2">Install Function</h3>
        {functionData && (
          <>
            <p className="text-sm text-text-secondary mb-2">
              Install <strong className="text-text-primary">{functionData.name}</strong> by {functionData.author.name}?
            </p>
            <div className="flex items-center gap-2 mb-4">
              <VerifiedBadge isVerified={functionData.isVerified} trustScore={functionData.trustScore} />
              <span className="text-sm text-text-muted">Trust Score: {functionData.trustScore}/100</span>
            </div>
          </>
        )}
        <div className="flex gap-2">
          <button
            onClick={onClose}
            className="flex-1 py-2.5 bg-bg-primary border border-border-subtle rounded-lg text-text-primary hover:bg-bg-hover transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              onInstall();
              onClose();
            }}
            className="flex-1 py-2.5 bg-brand-500 hover:bg-brand-600 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <Download className="size-4" />
            Install
          </button>
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// Marketplace Economy: Revenue Analytics
// ============================================================================

export interface RevenueAnalyticsProps {
  totalRevenue: number;
  revenueGrowth: number; // percentage
  revenueByFunction: Array<{ name: string; revenue: number; percentage: number }>;
  revenueByPeriod: Array<{ period: string; amount: number; count: number }>;
  className?: string;
}

export function RevenueAnalytics({
  totalRevenue = 0,
  revenueGrowth = 0,
  revenueByFunction = [],
  revenueByPeriod = [],
  className,
}: RevenueAnalyticsProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <BarChart2 className="size-4 text-brand-400" /> Revenue Analytics
      </h4>

      {/* Summary cards */}
      <div className="grid grid-cols-2 gap-3">
        <div className="p-4 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Total Revenue</div>
          <div className="text-2xl font-bold text-brand-500">${totalRevenue.toLocaleString()}</div>
        </div>
        <div className="p-4 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Growth</div>
          <div className={cn(
            "text-2xl font-bold",
            revenueGrowth >= 0 ? "text-emerald-400" : "text-red-400"
          )}>
            {revenueGrowth >= 0 ? "+" : ""}{revenueGrowth.toFixed(1)}%
          </div>
        </div>
      </div>

      {/* By function */}
      <div>
        <div className="text-[10px] font-medium text-text-muted mb-2">Revenue by Function</div>
        <div className="space-y-1.5">
          {revenueByFunction.slice(0, 5).map((item) => (
            <div key={item.name} className="flex items-center gap-2">
              <span className="text-xs text-text-primary w-32 truncate">{item.name}</span>
              <div className="flex-1 h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full bg-brand-500"
                  style={{ width: `${item.percentage}%` }}
                />
              </div>
              <span className="text-[10px] text-text-muted w-16 text-right">${item.revenue.toFixed(0)}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Period chart (simple bar) */}
      <div>
        <div className="text-[10px] font-medium text-text-muted mb-2">Revenue by Period</div>
        <div className="flex items-end gap-1 h-16">
          {revenueByPeriod.map((period) => {
            const maxAmount = Math.max(...revenueByPeriod.map((p) => p.amount), 1);
            const height = (period.amount / maxAmount) * 100;
            return (
              <div key={period.period} className="flex-1 flex flex-col items-center gap-0.5">
                <div
                  className="w-full rounded-t-sm bg-brand-500/60 hover:bg-brand-500 transition-colors"
                  style={{ height: `${Math.max(height, 2)}%` }}
                  title={`${period.period}: $${period.amount.toFixed(0)}`}
                />
                <span className="text-[8px] text-text-muted">{period.period}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// Subscription Manager
// ============================================================================

export interface SubscriptionPlan {
  id: string;
  name: string;
  price: number;
  period: "monthly" | "yearly";
  features: string[];
  functionQuota: number;
  agentQuota: number;
  isPopular?: boolean;
}

export interface SubscriptionManagerProps {
  currentPlan: string;
  plans: SubscriptionPlan[];
  onPlanSelect?: (planId: string) => void;
  onUpgrade?: () => void;
  onCancel?: () => void;
  className?: string;
}

export function SubscriptionManager({
  currentPlan,
  plans,
  onPlanSelect,
  onUpgrade,
  onCancel,
  className,
}: SubscriptionManagerProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <CreditCard className="size-4 text-brand-400" /> Subscription
      </h4>

      <div className="space-y-3">
        {plans.map((plan) => {
          const isCurrent = plan.id === currentPlan;

          return (
            <div
              key={plan.id}
              className={cn(
                "p-4 rounded-lg border transition-all",
                isCurrent ? "border-brand-500 bg-brand-500/5" : "border-border-subtle hover:border-border-default",
                plan.isPopular && "border-brand-500/50"
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-text-primary">{plan.name}</span>
                  {plan.isPopular && (
                    <span className="text-[9px] px-1.5 py-0.5 bg-brand-500/20 text-brand-400 rounded-full">Popular</span>
                  )}
                  {isCurrent && (
                    <span className="text-[9px] px-1.5 py-0.5 bg-emerald-500/20 text-emerald-400 rounded-full">Current</span>
                  )}
                </div>
                <div className="text-right">
                  <span className="text-lg font-bold text-text-primary">${plan.price}</span>
                  <span className="text-[10px] text-text-muted">/{plan.period}</span>
                </div>
              </div>

              <div className="flex items-center gap-4 text-[10px] text-text-muted mb-3">
                <span>{plan.functionQuota} functions</span>
                <span>{plan.agentQuota} agents</span>
              </div>

              <div className="flex flex-wrap gap-1 mb-3">
                {plan.features.slice(0, 3).map((feature) => (
                  <span key={feature} className="text-[9px] px-1.5 py-0.5 bg-bg-tertiary text-text-muted rounded">
                    {feature}
                  </span>
                ))}
              </div>

              {!isCurrent && (
                <button
                  onClick={() => onPlanSelect?.(plan.id)}
                  className="w-full py-2 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors"
                >
                  {plan.price > 0 ? "Upgrade" : "Select"}
                </button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// Usage Billing Panel
// ============================================================================

export interface BillingEntry {
  id: string;
  description: string;
  quantity: number;
  unitPrice: number;
  total: number;
  timestamp: string;
  type: "function_call" | "agent_run" | "compute" | "storage" | "api_call";
}

export interface UsageBillingPanelProps {
  entries: BillingEntry[];
  totalThisPeriod: number;
  includedQuota: number;
  overageCost: number;
  className?: string;
}

export function UsageBillingPanel({
  entries = [],
  totalThisPeriod = 0,
  includedQuota = 0,
  overageCost = 0,
  className,
}: UsageBillingPanelProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Receipt className="size-4 text-brand-400" /> Usage & Billing
      </h4>

      {/* Summary */}
      <div className="grid grid-cols-3 gap-2">
        <div className="p-3 bg-bg-secondary rounded-lg text-center">
          <div className="text-sm font-bold text-text-primary">${totalThisPeriod.toFixed(2)}</div>
          <div className="text-[9px] text-text-muted">Total</div>
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg text-center">
          <div className="text-sm font-bold text-emerald-400">${includedQuota.toFixed(2)}</div>
          <div className="text-[9px] text-text-muted">Quota</div>
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg text-center">
          <div className="text-sm font-bold text-amber-400">${overageCost.toFixed(2)}</div>
          <div className="text-[9px] text-text-muted">Overage</div>
        </div>
      </div>

      {/* Entry list */}
      <div className="space-y-1 max-h-48 overflow-y-auto">
        {entries.map((entry) => {
          const typeColors = {
            function_call: "bg-blue-500/10",
            agent_run: "bg-purple-500/10",
            compute: "bg-amber-500/10",
            storage: "bg-emerald-500/10",
            api_call: "bg-gray-500/10",
          };

          return (
            <div key={entry.id} className={cn("p-2.5 rounded-lg border border-border-subtle", typeColors[entry.type])}>
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <div className="text-xs text-text-primary truncate">{entry.description}</div>
                  <div className="text-[9px] text-text-muted">{entry.quantity} × ${entry.unitPrice.toFixed(4)}</div>
                </div>
                <div className="text-xs font-medium text-text-primary">${entry.total.toFixed(4)}</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// Creator Profile
// ============================================================================

export interface CreatorProfileProps {
  creator: {
    id: string;
    username: string;
    name: string;
    avatar?: string;
    profileUrl?: string;
  };
  stats: {
    totalFunctions: number;
    totalDownloads: number;
    totalRevenue: number;
    averageRating: number;
  };
  onEdit?: () => void;
  className?: string;
}

export function CreatorProfile({
  creator,
  stats,
  onEdit,
  className,
}: CreatorProfileProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center gap-4">
        {creator.avatar ? (
          <img src={creator.avatar} className="size-16 rounded-full" alt={creator.name} />
        ) : (
          <div className="size-16 rounded-full bg-brand-500/20 flex items-center justify-center text-xl font-bold text-brand-400">
            {creator.name[0]}
          </div>
        )}
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h4 className="text-sm font-medium text-text-primary">{creator.name}</h4>
          </div>
          <p className="text-[11px] text-text-muted line-clamp-2">@{creator.username}</p>
        </div>
        <button onClick={onEdit} className="text-[10px] text-brand-400 hover:text-brand-300">
          Edit
        </button>
      </div>

      <div className="grid grid-cols-4 gap-2">
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-sm font-bold text-text-primary">{stats.totalFunctions}</div>
          <div className="text-[9px] text-text-muted">Functions</div>
        </div>
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-sm font-bold text-text-primary">{stats.totalDownloads.toLocaleString()}</div>
          <div className="text-[9px] text-text-muted">Installs</div>
        </div>
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-sm font-bold text-amber-400">{stats.averageRating.toFixed(1)}</div>
          <div className="text-[9px] text-text-muted">Rating</div>
        </div>
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-sm font-bold text-brand-500">${stats.totalRevenue.toFixed(0)}</div>
          <div className="text-[9px] text-text-muted">Revenue</div>
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// Marketplace Leaderboard
// ============================================================================

export interface LeaderboardEntry {
  rank: number;
  functionId: string;
  functionName: string;
  author: string;
  installs: number;
  revenue: number;
  rating: number;
  trending: "up" | "down" | "stable";
}

export interface MarketplaceLeaderboardProps {
  entries: LeaderboardEntry[];
  period: "daily" | "weekly" | "monthly" | "all";
  onEntryClick?: (functionId: string) => void;
  className?: string;
}

export function MarketplaceLeaderboard({
  entries = [],
  period = "weekly",
  onEntryClick,
  className,
}: MarketplaceLeaderboardProps) {
  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Trophy className="size-4 text-amber-400" /> Leaderboard
        </h4>
        <span className="text-[10px] px-2 py-0.5 bg-bg-tertiary text-text-muted rounded capitalize">{period}</span>
      </div>

      <div className="space-y-1">
        {entries.map((entry) => {
          const rankColors = {
            1: "text-amber-400 bg-amber-500/20",
            2: "text-gray-300 bg-gray-500/20",
            3: "text-amber-700 bg-amber-700/20",
          };
          const rankColor = rankColors[entry.rank as 1 | 2 | 3] || "text-text-muted bg-bg-tertiary";

          return (
            <div
              key={entry.functionId}
              onClick={() => onEntryClick?.(entry.functionId)}
              className="flex items-center gap-3 p-2.5 bg-bg-secondary rounded-lg border border-border-subtle hover:border-border-default cursor-pointer transition-colors"
            >
              <span className={cn(
                "size-6 rounded flex items-center justify-center text-xs font-bold",
                rankColor
              )}>
                {entry.rank}
              </span>
              <div className="flex-1 min-w-0">
                <div className="text-xs font-medium text-text-primary truncate">{entry.functionName}</div>
                <div className="text-[9px] text-text-muted">by {entry.author}</div>
              </div>
              <div className="text-right">
                <div className="text-xs font-medium text-text-primary">{entry.installs.toLocaleString()}</div>
                <div className="text-[9px] text-text-muted">installs</div>
              </div>
              <div className="text-right">
                <div className="text-xs font-medium text-brand-500">${entry.revenue.toFixed(0)}</div>
                <div className="text-[9px] text-text-muted">revenue</div>
              </div>
              <span className={cn(
                "text-[10px]",
                entry.trending === "up" ? "text-emerald-400" : entry.trending === "down" ? "text-red-400" : "text-text-muted"
              )}>
                {entry.trending === "up" ? "↑" : entry.trending === "down" ? "↓" : "→"}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// Function Royalties Panel
// ============================================================================

export interface RoyaltyEntry {
  functionId: string;
  functionName: string;
  period: string;
  calls: number;
  royaltyRate: number;
  earnings: number;
  paid: boolean;
}

export interface FunctionRoyaltiesPanelProps {
  royalties: RoyaltyEntry[];
  totalEarnings: number;
  pendingPayout: number;
  onRequestPayout?: () => void;
  className?: string;
}

export function FunctionRoyaltiesPanel({
  royalties = [],
  totalEarnings = 0,
  pendingPayout = 0,
  onRequestPayout,
  className,
}: FunctionRoyaltiesPanelProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <DollarSign className="size-4 text-brand-400" /> Royalties
      </h4>

      <div className="grid grid-cols-2 gap-3">
        <div className="p-4 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Total Earnings</div>
          <div className="text-xl font-bold text-brand-500">${totalEarnings.toFixed(2)}</div>
        </div>
        <div className="p-4 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Pending Payout</div>
          <div className="text-xl font-bold text-amber-400">${pendingPayout.toFixed(2)}</div>
        </div>
      </div>

      <button
        onClick={onRequestPayout}
        disabled={pendingPayout <= 0}
        className="w-full py-2.5 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Request Payout
      </button>

      <div className="space-y-2">
        {royalties.map((royalty) => (
          <div key={`${royalty.functionId}-${royalty.period}`} className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
            <div className="flex items-center justify-between mb-1">
              <span className="text-xs font-medium text-text-primary">{royalty.functionName}</span>
              <span className={cn(
                "text-[9px] px-1.5 py-0.5 rounded-full",
                royalty.paid ? "bg-emerald-500/20 text-emerald-400" : "bg-amber-500/20 text-amber-400"
              )}>
                {royalty.paid ? "Paid" : "Pending"}
              </span>
            </div>
            <div className="flex items-center justify-between text-[10px] text-text-muted">
              <span>{royalty.period} · {royalty.calls.toLocaleString()} calls · {royalty.royaltyRate * 100}% rate</span>
              <span className="font-medium text-brand-400">${royalty.earnings.toFixed(2)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// Asset Pricing Editor
// ============================================================================

export interface AssetPricingProps {
  functionId: string;
  currentPrice: number;
  pricingModel: "free" | "per_call" | "subscription" | "revenue_share";
  onPriceChange?: (price: number) => void;
  onModelChange?: (model: "free" | "per_call" | "subscription" | "revenue_share") => void;
  className?: string;
}

export function AssetPricingEditor({
  functionId,
  currentPrice,
  pricingModel,
  onPriceChange,
  onModelChange,
  className,
}: AssetPricingProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Tag className="size-4 text-brand-400" /> Asset Pricing
      </h4>

      {/* Pricing model selector */}
      <div className="flex gap-1 flex-wrap">
        {(["free", "per_call", "subscription", "revenue_share"] as const).map((model) => (
          <button
            key={model}
            onClick={() => onModelChange?.(model)}
            className={cn(
              "px-3 py-1.5 text-xs rounded-lg capitalize transition-colors",
              pricingModel === model
                ? "bg-brand-500/20 text-brand-400 font-medium"
                : "bg-bg-secondary text-text-muted hover:text-text-primary"
            )}
          >
            {model.replace("_", " ")}
          </button>
        ))}
      </div>

      {/* Price input */}
      {pricingModel !== "free" && (
        <div>
          <label className="block text-[10px] font-medium text-text-muted mb-1">Price (USD)</label>
          <div className="flex items-center gap-2">
            <span className="text-lg text-text-muted">$</span>
            <input
              type="number"
              value={currentPrice}
              onChange={(e) => onPriceChange?.(parseFloat(e.target.value) || 0)}
              step="0.01"
              min="0"
              className="flex-1 px-3 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
            />
            {pricingModel === "per_call" && <span className="text-[10px] text-text-muted">/call</span>}
            {pricingModel === "subscription" && <span className="text-[10px] text-text-muted">/month</span>}
            {pricingModel === "revenue_share" && <span className="text-[10px] text-text-muted">% rev share</span>}
          </div>
        </div>
      )}

      {/* Estimated revenue */}
      {pricingModel !== "free" && currentPrice > 0 && (
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Estimated Monthly Revenue</div>
          <div className="text-lg font-bold text-brand-500">
            ${(currentPrice * 100).toFixed(2)}
          </div>
          <div className="text-[9px] text-text-muted">Based on 100 calls/month</div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// LicenseManager
// ============================================================================

export type LicenseType = "mit" | "apache" | "gpl" | "proprietary" | "custom";

export interface LicenseManagerProps {
  functionId: string;
  currentLicense: LicenseType;
  customLicenseText?: string;
  onLicenseChange?: (license: LicenseType, customText?: string) => void;
  className?: string;
}

export function LicenseManager({
  functionId,
  currentLicense,
  customLicenseText,
  onLicenseChange,
  className,
}: LicenseManagerProps) {
  const licenseOptions: Array<{ type: LicenseType; name: string; description: string }> = [
    { type: "mit", name: "MIT", description: "Permissive, allows proprietary use" },
    { type: "apache", name: "Apache 2.0", description: "Permissive with patent grant" },
    { type: "gpl", name: "GPL 3.0", description: "Requires derivative works to be open source" },
    { type: "proprietary", name: "Proprietary", description: "No redistribution allowed" },
    { type: "custom", name: "Custom", description: "Write your own terms" },
  ];

  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Shield className="size-4 text-brand-400" /> License Manager
      </h4>

      <div className="space-y-2">
        {licenseOptions.map((option) => (
          <button
            key={option.type}
            onClick={() => onLicenseChange?.(option.type)}
            className={cn(
              "w-full flex items-center justify-between p-3 rounded-lg border transition-colors text-left",
              currentLicense === option.type
                ? "border-brand-500 bg-brand-500/5"
                : "border-border-subtle hover:border-border-default"
            )}
          >
            <div>
              <div className="text-sm font-medium text-text-primary">{option.name}</div>
              <div className="text-[10px] text-text-muted">{option.description}</div>
            </div>
            {currentLicense === option.type && (
              <CheckCircle className="size-4 text-brand-400" />
            )}
          </button>
        ))}
      </div>

      {currentLicense === "custom" && (
        <div>
          <label className="block text-[10px] font-medium text-text-muted mb-1">Custom License Text</label>
          <textarea
            value={customLicenseText || ""}
            onChange={(e) => onLicenseChange?.("custom", e.target.value)}
            className="w-full h-32 p-3 text-xs bg-bg-primary border border-border-subtle rounded-lg text-text-primary font-mono resize-none focus:outline-none focus:border-brand-500"
            placeholder="Enter your custom license terms..."
          />
        </div>
      )}

      <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
        <div className="text-[10px] text-text-muted mb-1">Function ID</div>
        <div className="text-xs font-mono text-text-primary">{functionId}</div>
      </div>
    </div>
  );
}

// ============================================================================
// MonetizationOptimizer
// ============================================================================

export interface PricingRecommendation {
  model: "free" | "per_call" | "subscription" | "revenue_share";
  price: number;
  confidence: number;
  reasoning: string;
  expectedRevenue: number;
}

export interface MonetizationOptimizerProps {
  functionId: string;
  currentMetrics: {
    executionCount: number;
    averageLatency: number;
    errorRate: number;
    userCount: number;
  };
  onApplyRecommendation?: (model: "free" | "per_call" | "subscription" | "revenue_share", price: number) => void;
  className?: string;
}

export function MonetizationOptimizer({
  functionId,
  currentMetrics,
  onApplyRecommendation,
  className,
}: MonetizationOptimizerProps) {
  const recommendations: PricingRecommendation[] = [
    {
      model: "per_call",
      price: 0.01,
      confidence: 0.92,
      reasoning: "High execution volume with low latency suggests pay-per-call model",
      expectedRevenue: currentMetrics.executionCount * 0.01 * 0.7,
    },
    {
      model: "subscription",
      price: 9.99,
      confidence: 0.78,
      reasoning: "Professional users prefer predictable pricing for production use",
      expectedRevenue: currentMetrics.userCount * 0.1 * 9.99,
    },
    {
      model: "revenue_share",
      price: 15,
      confidence: 0.65,
      reasoning: "High-value workflows benefit from usage-based revenue share",
      expectedRevenue: currentMetrics.executionCount * 0.001 * 15,
    },
  ];

  const bestRecommendation = recommendations.reduce((best, r) => 
    r.confidence > best.confidence ? r : best
  );

  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <TrendingUp className="size-4 text-brand-400" /> Monetization Optimizer
      </h4>

      {/* Current metrics */}
      <div className="grid grid-cols-2 gap-2">
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[9px] text-text-muted">Executions</div>
          <div className="text-lg font-bold text-text-primary">{currentMetrics.executionCount.toLocaleString()}</div>
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[9px] text-text-muted">Users</div>
          <div className="text-lg font-bold text-text-primary">{currentMetrics.userCount.toLocaleString()}</div>
        </div>
      </div>

      {/* AI Recommendation */}
      <div className="p-4 bg-gradient-to-br from-brand-500/10 to-brand-500/5 rounded-lg border border-brand-500/30">
        <div className="flex items-center gap-2 mb-2">
          <Sparkles className="size-4 text-brand-400" />
          <span className="text-xs font-medium text-brand-400">AI Recommendation</span>
        </div>
        <div className="text-sm font-bold text-text-primary capitalize">
          {bestRecommendation.model.replace("_", " ")} at ${bestRecommendation.price.toFixed(2)}
        </div>
        <div className="text-[10px] text-text-muted mt-1">
          Confidence: {(bestRecommendation.confidence * 100).toFixed(0)}%
        </div>
        <div className="text-[11px] text-text-secondary mt-2">
          {bestRecommendation.reasoning}
        </div>
        <div className="mt-3 flex items-center justify-between">
          <div>
            <div className="text-[9px] text-text-muted">Expected Revenue</div>
            <div className="text-lg font-bold text-success">
              ${bestRecommendation.expectedRevenue.toFixed(2)}/mo
            </div>
          </div>
          <button
            onClick={() => onApplyRecommendation?.(bestRecommendation.model, bestRecommendation.price)}
            className="px-3 py-1.5 text-xs bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors"
          >
            Apply
          </button>
        </div>
      </div>

      {/* Alternative recommendations */}
      <div className="space-y-2">
        <div className="text-[10px] text-text-muted uppercase tracking-wider">Other Options</div>
        {recommendations
          .filter((r) => r.model !== bestRecommendation.model)
          .map((rec) => (
            <div
              key={rec.model}
              className="p-3 bg-bg-secondary rounded-lg border border-border-subtle flex items-center justify-between"
            >
              <div>
                <div className="text-xs font-medium text-text-primary capitalize">
                  {rec.model.replace("_", " ")} (${rec.price.toFixed(2)})
                </div>
                <div className="text-[9px] text-text-muted">
                  Confidence: {(rec.confidence * 100).toFixed(0)}%
                </div>
              </div>
              <button
                onClick={() => onApplyRecommendation?.(rec.model, rec.price)}
                className="text-[10px] text-brand-400 hover:text-brand-300"
              >
                Try
              </button>
            </div>
          ))}
      </div>
    </div>
  );
}

// ============================================================================
// Helper icons (local)
// ============================================================================

function BarChart2({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <line x1="12" y1="20" x2="12" y2="10" />
      <line x1="18" y1="20" x2="18" y2="4" />
      <line x1="6" y1="20" x2="6" y2="16" />
    </svg>
  );
}

function CreditCard({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
      <line x1="1" y1="10" x2="23" y2="10" />
    </svg>
  );
}

function Receipt({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1Z" />
      <path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8" />
      <path d="M12 17.5v-11" />
    </svg>
  );
}

function Trophy({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6" />
      <path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18" />
      <path d="M4 22h16" />
      <path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22" />
      <path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22" />
      <path d="M18 2H6v7a6 6 0 0 0 12 0V2Z" />
    </svg>
  );
}

function DollarSign({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <line x1="12" y1="1" x2="12" y2="23" />
      <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
    </svg>
  );
}

function TagIcon({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M12 2H2v10l9.29 9.29c.94.94 2.48.94 3.42 0l6.58-6.58c.94-.94.94-2.48 0-3.42L12 2Z" />
      <path d="M7 7h.01" />
    </svg>
  );
}


// ============================================================================
// RegistrySearch - Search and filter functions in the registry
// ============================================================================

export interface RegistrySearchProps {
  onSearchChange?: (query: string) => void;
  onCategorySelect?: (category: string) => void;
  onSortChange?: (sort: "popular" | "recent" | "rating" | "price") => void;
  onFilterChange?: (filters: RegistryFilters) => void;
  searchQuery?: string;
  selectedCategory?: string;
  sortBy?: "popular" | "recent" | "rating" | "price";
  filters?: RegistryFilters;
  className?: string;
}

export interface RegistryFilters {
  isVerified?: boolean;
  isFree?: boolean;
  minTrustScore?: number;
  maxPrice?: number;
  language?: string;
  runtime?: string;
}

export function RegistrySearch({
  onSearchChange,
  onCategorySelect,
  onSortChange,
  onFilterChange,
  searchQuery = "",
  selectedCategory,
  sortBy = "popular",
  filters,
  className,
}: RegistrySearchProps) {
  const [localQuery, setLocalQuery] = React.useState(searchQuery);
  const [showFilters, setShowFilters] = React.useState(false);

  const categories = [
    { id: "all", label: "All Functions", count: 1247 },
    { id: "data-processing", label: "Data Processing", count: 342 },
    { id: "ai-ml", label: "AI & ML", count: 289 },
    { id: "automation", label: "Automation", count: 198 },
    { id: "integrations", label: "Integrations", count: 156 },
    { id: "utilities", label: "Utilities", count: 134 },
    { id: "analytics", label: "Analytics", count: 87 },
    { id: "security", label: "Security", count: 41 },
  ];

  const sortOptions = [
    { value: "popular", label: "Most Popular" },
    { value: "recent", label: "Recently Updated" },
    { value: "rating", label: "Highest Rated" },
    { value: "price", label: "Price: Low to High" },
  ] as const;

  const debouncedSearch = React.useMemo(
    () => debounce((value: string) => onSearchChange?.(value), 300),
    [onSearchChange]
  );

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setLocalQuery(e.target.value);
    debouncedSearch(e.target.value);
  };

  return (
    <div className={cn("space-y-3", className)}>
      {/* Search bar */}
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-text-muted" />
          <input
            type="text"
            value={localQuery}
            onChange={handleInputChange}
            placeholder="Search functions, categories, or tags..."
            className="w-full pl-10 pr-4 py-2.5 text-sm bg-bg-primary border border-border-subtle rounded-xl text-text-primary placeholder:text-text-dim focus:outline-none focus:border-brand-500 transition-colors"
          />
          {localQuery && (
            <button
              onClick={() => { setLocalQuery(""); onSearchChange?.(""); }}
              className="absolute right-3 top-1/2 -translate-y-1/2 p-1 hover:bg-bg-hover rounded"
            >
              <XCircle className="size-4 text-text-muted" />
            </button>
          )}
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={cn(
            "p-2.5 rounded-xl border transition-colors",
            showFilters
              ? "bg-brand-500/20 border-brand-500/30 text-brand-400"
              : "bg-bg-primary border-border-subtle text-text-muted hover:text-text-primary hover:border-border-default"
          )}
        >
          <Filter className="size-4" />
        </button>
      </div>

      {/* Categories */}
      <div className="flex items-center gap-2 overflow-x-auto pb-1 scrollbar-hide">
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => onCategorySelect?.(cat.id)}
            className={cn(
              "flex items-center gap-2 px-3 py-1.5 text-xs rounded-full border transition-all whitespace-nowrap",
              selectedCategory === cat.id || (!selectedCategory && cat.id === "all")
                ? "bg-brand-500/20 border-brand-500/30 text-brand-400"
                : "bg-bg-secondary border-border-subtle text-text-muted hover:text-text-primary hover:border-border-default"
            )}
          >
            {cat.label}
            <span className={cn(
              "px-1.5 py-0.5 text-[10px] rounded-full",
              selectedCategory === cat.id || (!selectedCategory && cat.id === "all")
                ? "bg-brand-500/30"
                : "bg-bg-tertiary"
            )}>
              {cat.count.toLocaleString()}
            </span>
          </button>
        ))}
      </div>

      {/* Sort + Active filters */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-xs text-text-muted">Sort by:</span>
          <div className="flex gap-1">
            {sortOptions.map((opt) => (
              <button
                key={opt.value}
                onClick={() => onSortChange?.(opt.value)}
                className={cn(
                  "px-2.5 py-1 text-[11px] rounded-lg transition-colors",
                  sortBy === opt.value
                    ? "bg-bg-tertiary text-text-primary"
                    : "text-text-muted hover:text-text-primary"
                )}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        {/* Active filter pills */}
        <div className="flex items-center gap-1">
          {filters?.isVerified && (
            <span className="flex items-center gap-1 px-2 py-0.5 text-[10px] bg-emerald-500/10 text-emerald-400 rounded-full border border-emerald-500/20">
              <Shield className="size-3" /> Verified
              <button onClick={() => onFilterChange?.({ ...filters, isVerified: false })} className="ml-0.5">
                <XCircle className="size-3" />
              </button>
            </span>
          )}
          {filters?.isFree && (
            <span className="flex items-center gap-1 px-2 py-0.5 text-[10px] bg-brand-500/10 text-brand-400 rounded-full border border-brand-500/20">
              Free
              <button onClick={() => onFilterChange?.({ ...filters, isFree: false })} className="ml-0.5">
                <XCircle className="size-3" />
              </button>
            </span>
          )}
          {filters?.minTrustScore && (
            <span className="flex items-center gap-1 px-2 py-0.5 text-[10px] bg-amber-500/10 text-amber-400 rounded-full border border-amber-500/20">
              Trust: {filters.minTrustScore}+
              <button onClick={() => onFilterChange?.({ ...filters, minTrustScore: undefined })} className="ml-0.5">
                <XCircle className="size-3" />
              </button>
            </span>
          )}
        </div>
      </div>

      {/* Expanded filters panel */}
      {showFilters && (
        <div className="p-4 bg-bg-secondary rounded-xl border border-border-subtle space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <label className="block text-[10px] font-medium text-text-muted mb-1.5">Min Trust Score</label>
              <input
                type="range"
                min="0"
                max="100"
                value={filters?.minTrustScore ?? 0}
                onChange={(e) => onFilterChange?.({ ...filters, minTrustScore: parseInt(e.target.value) })}
                className="w-full accent-brand-500"
              />
              <div className="text-[10px] text-text-muted text-right">{filters?.minTrustScore ?? 0}+</div>
            </div>
            <div>
              <label className="block text-[10px] font-medium text-text-muted mb-1.5">Max Price ($/call)</label>
              <input
                type="number"
                value={filters?.maxPrice ?? ""}
                onChange={(e) => onFilterChange?.({ ...filters, maxPrice: parseFloat(e.target.value) || undefined })}
                placeholder="No limit"
                className="w-full px-3 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
              />
            </div>
            <div>
              <label className="block text-[10px] font-medium text-text-muted mb-1.5">Language</label>
              <select
                value={filters?.language ?? ""}
                onChange={(e) => onFilterChange?.({ ...filters, language: e.target.value || undefined })}
                className="w-full px-3 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
              >
                <option value="">Any</option>
                <option value="typescript">TypeScript</option>
                <option value="python">Python</option>
                <option value="go">Go</option>
                <option value="rust">Rust</option>
                <option value="java">Java</option>
              </select>
            </div>
            <div>
              <label className="block text-[10px] font-medium text-text-muted mb-1.5">Runtime</label>
              <select
                value={filters?.runtime ?? ""}
                onChange={(e) => onFilterChange?.({ ...filters, runtime: e.target.value || undefined })}
                className="w-full px-3 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
              >
                <option value="">Any</option>
                <option value="nodejs">Node.js</option>
                <option value="python">Python</option>
                <option value="go">Go</option>
                <option value="rust">Rust</option>
                <option value="deno">Deno</option>
              </select>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={filters?.isVerified ?? false}
                onChange={(e) => onFilterChange?.({ ...filters, isVerified: e.target.checked })}
                className="accent-brand-500"
              />
              <span className="text-xs text-text-secondary">Verified only</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={filters?.isFree ?? false}
                onChange={(e) => onFilterChange?.({ ...filters, isFree: e.target.checked })}
                className="accent-brand-500"
              />
              <span className="text-xs text-text-secondary">Free only</span>
            </label>
            <button
              onClick={() => onFilterChange?.({})}
              className="ml-auto text-xs text-text-muted hover:text-text-primary"
            >
              Clear all filters
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// RegistryCategoryExplorer - Browse categories with visual navigation
// ============================================================================

export interface CategoryNode {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  count: number;
  subcategories?: CategoryNode[];
  trending?: boolean;
}

export interface RegistryCategoryExplorerProps {
  categories: CategoryNode[];
  onCategorySelect?: (categoryId: string) => void;
  selectedCategory?: string;
  className?: string;
}

export function RegistryCategoryExplorer({
  categories,
  onCategorySelect,
  selectedCategory,
  className,
}: RegistryCategoryExplorerProps) {
  const [expandedCategories, setExpandedCategories] = React.useState<Set<string>>(new Set());

  const toggleExpand = (id: string) => {
    setExpandedCategories((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const renderCategory = (category: CategoryNode, depth = 0) => {
    const hasChildren = category.subcategories && category.subcategories.length > 0;
    const isExpanded = expandedCategories.has(category.id);
    const isSelected = selectedCategory === category.id;

    return (
      <div key={category.id}>
        <button
          onClick={() => {
            if (hasChildren) toggleExpand(category.id);
            onCategorySelect?.(category.id);
          }}
          className={cn(
            "w-full flex items-center gap-3 p-3 rounded-lg border transition-all text-left",
            isSelected
              ? "bg-brand-500/10 border-brand-500/30"
              : "border-border-subtle hover:border-border-default hover:bg-bg-secondary"
          )}
          style={{ paddingLeft: `${depth * 20 + 12}px` }}
        >
          <div className={cn(
            "size-8 rounded-lg flex items-center justify-center shrink-0",
            isSelected ? "bg-brand-500/20" : "bg-bg-tertiary"
          )}>
            {category.icon}
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-text-primary">{category.name}</span>
              {category.trending && (
                <TrendingUp className="size-3 text-emerald-400" />
              )}
            </div>
            <p className="text-[11px] text-text-muted line-clamp-1">{category.description}</p>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-text-muted">{category.count.toLocaleString()}</span>
            {hasChildren && (
              <ChevronRight className={cn(
                "size-4 text-text-muted transition-transform",
                isExpanded && "rotate-90"
              )} />
            )}
          </div>
        </button>
        {hasChildren && isExpanded && (
          <div className="ml-4 border-l border-border-subtle">
            {category.subcategories!.map((sub) => renderCategory(sub, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className={cn("space-y-1", className)}>
      {categories.map((cat) => renderCategory(cat))}
    </div>
  );
}

// ============================================================================
// RuntimeCompatibilityMatrix - Check function compatibility with runtimes
// ============================================================================

export interface RuntimeVersion {
  runtime: string;
  version: string;
  isCompatible: boolean;
  minVersion?: string;
  notes?: string;
}

export interface FunctionCompatibility {
  functionId: string;
  functionName: string;
  runtimes: RuntimeVersion[];
}

export interface RuntimeCompatibilityMatrixProps {
  functions: FunctionCompatibility[];
  onFunctionSelect?: (functionId: string) => void;
  selectedFunctionId?: string;
  className?: string;
}

export function RuntimeCompatibilityMatrix({
  functions,
  onFunctionSelect,
  selectedFunctionId,
  className,
}: RuntimeCompatibilityMatrixProps) {
  const allRuntimes = React.useMemo(() => {
    const runtimeSet = new Set<string>();
    functions.forEach((f) => f.runtimes.forEach((r) => runtimeSet.add(r.runtime)));
    return Array.from(runtimeSet).sort();
  }, [functions]);

  const runtimeColors: Record<string, string> = {
    nodejs: "#339933",
    python: "#3776AB",
    go: "#00ADD8",
    rust: "#DEA584",
    deno: "#70FFAF",
    bun: "#FBF0C3",
    ruby: "#CC342D",
    java: "#007396",
  };

  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Layers className="size-4 text-brand-400" /> Runtime Compatibility
        </h4>
        <div className="flex items-center gap-2 text-[10px] text-text-muted">
          <span className="flex items-center gap-1"><CheckCircle className="size-3 text-emerald-400" /> Compatible</span>
          <span className="flex items-center gap-1"><XCircle className="size-3 text-red-400" /> Incompatible</span>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border-subtle">
              <th className="text-left py-2 px-3 text-text-muted font-medium">Function</th>
              {allRuntimes.map((rt) => (
                <th key={rt} className="py-2 px-2 text-center">
                  <div className="flex flex-col items-center gap-1">
                    <div
                      className="size-6 rounded flex items-center justify-center text-white text-[10px] font-bold"
                      style={{ backgroundColor: runtimeColors[rt] ?? "#6b7280" }}
                    >
                      {rt.slice(0, 2).toUpperCase()}
                    </div>
                    <span className="text-[9px] text-text-dim">{rt}</span>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {functions.map((fn) => (
              <tr
                key={fn.functionId}
                onClick={() => onFunctionSelect?.(fn.functionId)}
                className={cn(
                  "border-b border-border-subtle transition-colors cursor-pointer",
                  selectedFunctionId === fn.functionId
                    ? "bg-brand-500/5"
                    : "hover:bg-bg-secondary"
                )}
              >
                <td className="py-2 px-3">
                  <span className="font-medium text-text-primary">{fn.functionName}</span>
                </td>
                {allRuntimes.map((rt) => {
                  const compat = fn.runtimes.find((r) => r.runtime === rt);
                  return (
                    <td key={rt} className="py-2 px-2 text-center">
                      {compat ? (
                        compat.isCompatible ? (
                          <CheckCircle className="size-4 text-emerald-400 mx-auto" />
                        ) : (
                          <div className="relative group">
                            <XCircle className="size-4 text-red-400 mx-auto" />
                            {compat.notes && (
                              <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 bg-bg-instrument text-[10px] text-text-primary rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-10 border border-border-subtle">
                                {compat.notes}
                              </div>
                            )}
                          </div>
                        )
                      ) : (
                        <span className="text-text-dim">—</span>
                      )}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ============================================================================
// DependencyViewer - Visualize function dependencies
// ============================================================================

export interface DependencyNode {
  id: string;
  name: string;
  version: string;
  isExternal: boolean;
}

export interface DependencyEdge {
  isExternal?: boolean;
  from: string;
  to: string;
  type: "runtime" | "npm" | "go" | "pip" | "cargo";
}

export interface DependencyViewerProps {
  nodes: DependencyNode[];
  edges: DependencyEdge[];
  onNodeClick?: (nodeId: string) => void;
  selectedNodeId?: string;
  className?: string;
}

export function DependencyViewer({
  nodes,
  edges,
  onNodeClick,
  selectedNodeId,
  className,
}: DependencyViewerProps) {
  // Build adjacency for layout
  const adjacency = React.useMemo(() => {
    const adj = new Map<string, string[]>();
    nodes.forEach((n) => adj.set(n.id, []));
    edges.forEach((e) => {
      adj.get(e.from)?.push(e.to);
    });
    return adj;
  }, [nodes, edges]);

  // Simple hierarchical layout
  const layout = React.useMemo(() => {
    const positions = new Map<string, { x: number; y: number }>();
    const levels = new Map<string, number>();

    // BFS to assign levels
    const queue: string[] = nodes.filter((n) => n.isExternal).map((n) => n.id);
    queue.forEach((id) => levels.set(id, 0));

    while (queue.length > 0) {
      const current = queue.shift()!;
      const level = levels.get(current)!;
      adjacency.get(current)?.forEach((neighbor) => {
        if (!levels.has(neighbor)) {
          levels.set(neighbor, level + 1);
          queue.push(neighbor);
        }
      });
    }

    // Group by level
    const byLevel = new Map<number, string[]>();
    levels.forEach((level, id) => {
      if (!byLevel.has(level)) byLevel.set(level, []);
      byLevel.get(level)!.push(id);
    });

    // Assign positions
    byLevel.forEach((ids, level) => {
      ids.forEach((id, idx) => {
        positions.set(id, {
          x: 120 + level * 160,
          y: 40 + idx * 60 + (ids.length > 1 ? (ids.length - 1) * -30 : 0),
        });
      });
    });

    // Orphan nodes
    nodes.forEach((n) => {
      if (!positions.has(n.id)) {
        positions.set(n.id, { x: 400, y: 40 + nodes.indexOf(n) * 60 });
      }
    });

    return positions;
  }, [nodes, adjacency]);

  const edgeColors: Record<DependencyEdge["type"], string> = {
    runtime: "#f97316",
    npm: "#CB171E",
    go: "#00ADD8",
    pip: "#3776AB",
    cargo: "#DEA584",
  };

  const edgeLabels: Record<DependencyEdge["type"], string> = {
    runtime: "RT",
    npm: "npm",
    go: "go",
    pip: "pip",
    cargo: "cargo",
  };

  return (
    <div className={cn("relative", className)}>
      <svg viewBox="0 0 600 300" className="w-full h-64 bg-bg-secondary rounded-lg border border-border-subtle">
        {/* Edges */}
        {edges.map((edge, i) => {
          const fromPos = layout.get(edge.from);
          const toPos = layout.get(edge.to);
          if (!fromPos || !toPos) return null;

          const midX = (fromPos.x + toPos.x) / 2;
          const midY = (fromPos.y + toPos.y) / 2;

          return (
            <g key={i}>
              <line
                x1={fromPos.x}
                y1={fromPos.y}
                x2={toPos.x}
                y2={toPos.y}
                stroke={edgeColors[edge.type]}
                strokeWidth={1.5}
                strokeDasharray={edge.isExternal ? "4,2" : "none"}
                opacity={0.6}
              />
              <circle cx={midX} cy={midY} r={8} fill={edgeColors[edge.type]} />
              <text
                x={midX}
                y={midY + 3}
                textAnchor="middle"
                className="text-[8px] fill-white font-medium"
              >
                {edgeLabels[edge.type]}
              </text>
            </g>
          );
        })}

        {/* Nodes */}
        {nodes.map((node) => {
          const pos = layout.get(node.id);
          if (!pos) return null;
          const isSelected = selectedNodeId === node.id;

          return (
            <g
              key={node.id}
              onClick={() => onNodeClick?.(node.id)}
              className="cursor-pointer"
            >
              <rect
                x={pos.x - 50}
                y={pos.y - 16}
                width={100}
                height={32}
                rx={6}
                fill={isSelected ? "#ff6b35" : "#1a1a28"}
                stroke={isSelected ? "#ff8c42" : "#374151"}
                strokeWidth={isSelected ? 2 : 1}
              />
              <text x={pos.x} y={pos.y + 4} textAnchor="middle" className="text-[10px] fill-white font-medium">
                {node.name.length > 12 ? node.name.slice(0, 12) + "..." : node.name}
              </text>
              {node.isExternal && (
                <circle cx={pos.x + 40} cy={pos.y - 8} r={4} fill="#f97316" />
              )}
            </g>
          );
        })}
      </svg>

      {/* Legend */}
      <div className="flex items-center gap-4 mt-2 text-[10px] text-text-muted">
        <span className="flex items-center gap-1">
          <span className="size-3 rounded bg-[#f97316]" /> Runtime
        </span>
        <span className="flex items-center gap-1">
          <span className="size-3 rounded bg-[#CB171E]" /> npm
        </span>
        <span className="flex items-center gap-1">
          <span className="size-3 rounded bg-[#00ADD8]" /> go
        </span>
        <span className="flex items-center gap-1">
          <span className="size-3 rounded bg-[#3776AB]" /> pip
        </span>
        <span className="flex items-center gap-1">
          <span className="size-3 rounded bg-[#DEA584]" /> cargo
        </span>
      </div>
    </div>
  );
}

// ============================================================================
// FunctionSandbox - Interactive function testing environment
// ============================================================================

export interface SandboxInput {
  name: string;
  type: string;
  value: string;
  required: boolean;
}

export interface SandboxResult {
  success: boolean;
  output: string;
  executionTime: number;
  memoryUsed?: number;
  error?: string;
}

export interface FunctionSandboxProps {
  functionId: string;
  functionName: string;
  inputs: SandboxInput[];
  onRun?: (inputs: Record<string, string>) => void;
  className?: string;
}

export function FunctionSandbox({
  functionId,
  functionName,
  inputs,
  onRun,
  className,
}: FunctionSandboxProps) {
  const [inputValues, setInputValues] = React.useState<Record<string, string>>(
    inputs.reduce((acc, inp) => ({ ...acc, [inp.name]: inp.value }), {})
  );
  const [isRunning, setIsRunning] = React.useState(false);
  const [result, setResult] = React.useState<SandboxResult | null>(null);

  const handleRun = async () => {
    setIsRunning(true);
    setResult(null);
    onRun?.(inputValues);

    // Simulate execution
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setResult({
      success: Math.random() > 0.1,
      output: JSON.stringify({ message: "Function executed successfully", data: inputValues }, null, 2),
      executionTime: Math.random() * 200 + 50,
      memoryUsed: Math.random() * 50 + 10,
    });
    setIsRunning(false);
  };

  const handleInputChange = (name: string, value: string) => {
    setInputValues((prev) => ({ ...prev, [name]: value }));
  };

  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
            <Play className="size-4 text-brand-400" /> Sandbox: {functionName}
          </h4>
          <p className="text-[10px] text-text-muted">Test your function with custom inputs</p>
        </div>
        <button
          onClick={handleRun}
          disabled={isRunning}
          className="flex items-center gap-2 px-4 py-2 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors disabled:opacity-50"
        >
          {isRunning ? (
            <>
              <Loader2 className="size-3 animate-spin" /> Running...
            </>
          ) : (
            <>
              <Play className="size-3" /> Run Function
            </>
          )}
        </button>
      </div>

      {/* Inputs */}
      <div className="space-y-3">
        {inputs.map((input) => (
          <div key={input.name}>
            <label className="flex items-center gap-1 text-[11px] font-medium text-text-secondary mb-1">
              {input.name}
              {input.required && <span className="text-red-400">*</span>}
              <span className="text-[10px] text-text-dim ml-1">({input.type})</span>
            </label>
            <textarea
              value={inputValues[input.name] ?? ""}
              onChange={(e) => handleInputChange(input.name, e.target.value)}
              placeholder={`Enter ${input.type} value...`}
              rows={input.type === "object" ? 4 : 1}
              className="w-full px-3 py-2 text-xs bg-bg-primary border border-border-subtle rounded-lg text-text-primary font-mono resize-none focus:outline-none focus:border-brand-500 transition-colors"
            />
          </div>
        ))}
      </div>

      {/* Result */}
      {result && (
        <div className={cn(
          "p-4 rounded-lg border",
          result.success
            ? "bg-emerald-500/10 border-emerald-500/20"
            : "bg-red-500/10 border-red-500/20"
        )}>
          <div className="flex items-center justify-between mb-2">
            <span className={cn(
              "text-xs font-medium",
              result.success ? "text-emerald-400" : "text-red-400"
            )}>
              {result.success ? "Execution Successful" : "Execution Failed"}
            </span>
            <div className="flex items-center gap-3 text-[10px] text-text-muted">
              <span>{result.executionTime.toFixed(1)}ms</span>
              {result.memoryUsed && <span>{result.memoryUsed.toFixed(0)}MB</span>}
            </div>
          </div>
          <pre className="text-[11px] text-text-secondary font-mono whitespace-pre-wrap break-all bg-bg-primary/50 p-2 rounded">
            {result.success ? result.output : result.error}
          </pre>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// VersionDiffViewer - Compare two versions of a function
// ============================================================================

export interface VersionDiffLine {
  type: "added" | "removed" | "unchanged";
  content: string;
  lineNumber?: number;
}

export interface VersionInfo {
  version: string;
  timestamp: string;
  author: string;
  changelog?: string;
}

export interface VersionDiffViewerProps {
  oldVersion: VersionInfo;
  newVersion: VersionInfo;
  oldCode: string;
  newCode: string;
  onRestoreVersion?: (version: string) => void;
  className?: string;
}

export function VersionDiffViewer({
  oldVersion,
  newVersion,
  oldCode,
  newCode,
  onRestoreVersion,
  className,
}: VersionDiffViewerProps) {
  const [showDiff, setShowDiff] = React.useState(true);
  const [selectedVersion, setSelectedVersion] = React.useState<"old" | "new">("new");

  // Simple line-by-line diff
  const diff = React.useMemo(() => {
    const oldLines = oldCode.split("\n");
    const newLines = newCode.split("\n");
    const result: VersionDiffLine[] = [];

    // Very simple diff - just show both side by side with markers
    const maxLen = Math.max(oldLines.length, newLines.length);
    for (let i = 0; i < maxLen; i++) {
      const oldLine = oldLines[i];
      const newLine = newLines[i];

      if (oldLine === undefined) {
        result.push({ type: "added", content: newLine, lineNumber: i + 1 });
      } else if (newLine === undefined) {
        result.push({ type: "removed", content: oldLine, lineNumber: i + 1 });
      } else if (oldLine !== newLine) {
        result.push({ type: "removed", content: oldLine, lineNumber: i + 1 });
        result.push({ type: "added", content: newLine, lineNumber: i + 1 });
      } else {
        result.push({ type: "unchanged", content: oldLine, lineNumber: i + 1 });
      }
    }
    return result;
  }, [oldCode, newCode]);

  const stats = React.useMemo(() => {
    let added = 0, removed = 0;
    diff.forEach((d) => {
      if (d.type === "added") added++;
      if (d.type === "removed") removed++;
    });
    return { added, removed };
  }, [diff]);

  return (
    <div className={cn("space-y-4", className)}>
      {/* Version selector */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setSelectedVersion("old")}
            className={cn(
              "px-3 py-1.5 text-xs rounded-lg border transition-colors",
              selectedVersion === "old"
                ? "bg-bg-tertiary border-border-default text-text-primary"
                : "border-border-subtle text-text-muted hover:text-text-primary"
            )}
          >
            v{oldVersion.version}
          </button>
          <ChevronRight className="size-4 text-text-muted" />
          <button
            onClick={() => setSelectedVersion("new")}
            className={cn(
              "px-3 py-1.5 text-xs rounded-lg border transition-colors",
              selectedVersion === "new"
                ? "bg-bg-tertiary border-border-default text-text-primary"
                : "border-border-subtle text-text-muted hover:text-text-primary"
            )}
          >
            v{newVersion.version}
          </button>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-emerald-400">+{stats.added}</span>
          <span className="text-[10px] text-red-400">-{stats.removed}</span>
        </div>
      </div>

      {/* Version info cards */}
      <div className="grid grid-cols-2 gap-3">
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-medium text-text-primary">v{oldVersion.version}</span>
            <button
              onClick={() => onRestoreVersion?.(oldVersion.version)}
              className="text-[10px] text-brand-400 hover:text-brand-300"
            >
              Restore
            </button>
          </div>
          <div className="text-[10px] text-text-muted">
            {oldVersion.timestamp} · {oldVersion.author}
          </div>
          {oldVersion.changelog && (
            <p className="text-[10px] text-text-secondary mt-1 line-clamp-2">{oldVersion.changelog}</p>
          )}
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-medium text-text-primary">v{newVersion.version}</span>
            <span className="text-[10px] px-1.5 py-0.5 bg-brand-500/20 text-brand-400 rounded">Current</span>
          </div>
          <div className="text-[10px] text-text-muted">
            {newVersion.timestamp} · {newVersion.author}
          </div>
          {newVersion.changelog && (
            <p className="text-[10px] text-text-secondary mt-1 line-clamp-2">{newVersion.changelog}</p>
          )}
        </div>
      </div>

      {/* Diff toggle */}
      <div className="flex items-center gap-2">
        <button
          onClick={() => setShowDiff(!showDiff)}
          className="flex items-center gap-2 text-xs text-text-muted hover:text-text-primary"
        >
          {showDiff ? <Eye className="size-3" /> : <EyeOff className="size-3" />}
          {showDiff ? "Hide" : "Show"} changes
        </button>
      </div>

      {/* Diff view */}
      {showDiff && (
        <div className="border border-border-subtle rounded-lg overflow-hidden">
          <div className="flex">
            <div className="flex-1 p-3 bg-bg-secondary border-r border-border-subtle">
              <span className="text-[10px] font-medium text-text-muted">Previous</span>
            </div>
            <div className="flex-1 p-3 bg-bg-secondary">
              <span className="text-[10px] font-medium text-text-muted">Current</span>
            </div>
          </div>
          <div className="max-h-64 overflow-y-auto font-mono text-[11px]">
            {diff.map((line, i) => (
              <div
                key={i}
                className={cn(
                  "flex min-h-[20px]",
                  line.type === "added" && "bg-emerald-500/10",
                  line.type === "removed" && "bg-red-500/10"
                )}
              >
                <div className={cn(
                  "flex-1 px-3 py-0.5 border-r border-border-subtle",
                  line.type === "removed" ? "text-red-400" : "text-text-secondary"
                )}>
                  {line.type !== "added" && line.content}
                </div>
                <div className={cn(
                  "flex-1 px-3 py-0.5",
                  line.type === "added" ? "text-emerald-400" : "text-text-secondary"
                )}>
                  {line.type !== "removed" && line.content}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// FunctionUsageAnalytics - Detailed usage statistics for a function
// ============================================================================

export interface UsageDataPoint {
  timestamp: string;
  executions: number;
  errors: number;
  latency: number;
}

export interface FunctionUsageAnalyticsProps {
  functionId: string;
  functionName: string;
  usageData: UsageDataPoint[];
  totalExecutions: number;
  avgLatency: number;
  errorRate: number;
  topUsers?: Array<{ userId: string; name: string; calls: number }>;
  className?: string;
}

export function FunctionUsageAnalytics({
  functionId,
  functionName,
  usageData,
  totalExecutions,
  avgLatency,
  errorRate,
  topUsers = [],
  className,
}: FunctionUsageAnalyticsProps) {
  const [timeRange, setTimeRange] = React.useState<"24h" | "7d" | "30d">("7d");
  const [activeTab, setActiveTab] = React.useState<"overview" | "users" | "performance">("overview");

  const filteredData = React.useMemo(() => {
    const now = Date.now();
    const cutoff = timeRange === "24h" ? 24 * 60 * 60 * 1000 : timeRange === "7d" ? 7 * 24 * 60 * 60 * 1000 : 30 * 24 * 60 * 60 * 1000;
    return usageData.filter((d) => now - new Date(d.timestamp).getTime() < cutoff);
  }, [usageData, timeRange]);

  const maxExecutions = Math.max(...filteredData.map((d) => d.executions), 1);

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
            <BarChart2 className="size-4 text-brand-400" /> Usage: {functionName}
          </h4>
          <p className="text-[10px] text-text-muted">Function ID: {functionId}</p>
        </div>
        <div className="flex gap-1">
          {(["24h", "7d", "30d"] as const).map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={cn(
                "px-2.5 py-1 text-[11px] rounded-lg transition-colors",
                timeRange === range
                  ? "bg-brand-500/20 text-brand-400"
                  : "text-text-muted hover:text-text-primary"
              )}
            >
              {range}
            </button>
          ))}
        </div>
      </div>

      {/* Summary stats */}
      <div className="grid grid-cols-3 gap-3">
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Total Executions</div>
          <div className="text-lg font-bold text-text-primary">{totalExecutions.toLocaleString()}</div>
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Avg Latency</div>
          <div className="text-lg font-bold text-text-primary">{avgLatency.toFixed(0)}ms</div>
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Error Rate</div>
          <div className={cn(
            "text-lg font-bold",
            errorRate < 1 ? "text-emerald-400" : errorRate < 5 ? "text-amber-400" : "text-red-400"
          )}>
            {errorRate.toFixed(2)}%
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-border-subtle">
        {(["overview", "users", "performance"] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={cn(
              "px-3 py-2 text-xs capitalize transition-colors border-b-2 -mb-px",
              activeTab === tab
                ? "border-brand-500 text-brand-400"
                : "border-transparent text-text-muted hover:text-text-primary"
            )}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {activeTab === "overview" && (
        <div className="space-y-3">
          {/* Execution chart */}
          <div>
            <div className="text-[10px] font-medium text-text-muted mb-2">Executions over time</div>
            <div className="flex items-end gap-1 h-24">
              {filteredData.map((d, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-0.5">
                  <div
                    className="w-full rounded-t-sm bg-brand-500/60 hover:bg-brand-500 transition-colors"
                    style={{ height: `${(d.executions / maxExecutions) * 100}%` }}
                    title={`${d.executions} executions`}
                  />
                </div>
              ))}
            </div>
          </div>

          {/* Error rate trend */}
          <div>
            <div className="text-[10px] font-medium text-text-muted mb-2">Error rate</div>
            <div className="flex items-center gap-2">
              <div className="flex-1 h-2 bg-bg-tertiary rounded-full overflow-hidden">
                <div
                  className={cn(
                    "h-full rounded-full transition-all",
                    errorRate < 1 ? "bg-emerald-500" : errorRate < 5 ? "bg-amber-500" : "bg-red-500"
                  )}
                  style={{ width: `${Math.min(errorRate * 10, 100)}%` }}
                />
              </div>
              <span className="text-xs text-text-primary">{errorRate.toFixed(2)}%</span>
            </div>
          </div>
        </div>
      )}

      {activeTab === "users" && (
        <div className="space-y-2">
          <div className="text-[10px] font-medium text-text-muted mb-2">Top Users</div>
          {topUsers.length > 0 ? (
            topUsers.map((user) => (
              <div key={user.userId} className="flex items-center justify-between p-2 bg-bg-secondary rounded-lg border border-border-subtle">
                <div className="flex items-center gap-2">
                  <div className="size-6 rounded-full bg-bg-tertiary flex items-center justify-center text-[10px] font-medium text-text-primary">
                    {user.name[0]}
                  </div>
                  <span className="text-xs text-text-primary">{user.name}</span>
                </div>
                <span className="text-xs text-text-muted">{user.calls.toLocaleString()} calls</span>
              </div>
            ))
          ) : (
            <div className="text-center py-8 text-text-muted text-xs">No user data available</div>
          )}
        </div>
      )}

      {activeTab === "performance" && (
        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
              <div className="text-[10px] text-text-muted mb-1">P50 Latency</div>
              <div className="text-sm font-bold text-text-primary">{(avgLatency * 0.8).toFixed(0)}ms</div>
            </div>
            <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
              <div className="text-[10px] text-text-muted mb-1">P95 Latency</div>
              <div className="text-sm font-bold text-text-primary">{(avgLatency * 1.5).toFixed(0)}ms</div>
            </div>
            <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
              <div className="text-[10px] text-text-muted mb-1">P99 Latency</div>
              <div className="text-sm font-bold text-text-primary">{(avgLatency * 2.2).toFixed(0)}ms</div>
            </div>
            <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
              <div className="text-[10px] text-text-muted mb-1">Throughput</div>
              <div className="text-sm font-bold text-text-primary">{(totalExecutions / 7 / 24).toFixed(1)}/min</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// PublicWorkflowExplorer - Browse public workflows
// ============================================================================

export interface WorkflowTemplate {
  id: string;
  name: string;
  description: string;
  author: string;
  category: string;
  steps: number;
  executions: number;
  rating: number;
  tags: string[];
  isFeatured?: boolean;
}

export interface PublicWorkflowExplorerProps {
  workflows: WorkflowTemplate[];
  onSelect?: (workflowId: string) => void;
  onUse?: (workflowId: string) => void;
  className?: string;
}

export function PublicWorkflowExplorer({
  workflows,
  onSelect,
  onUse,
  className,
}: PublicWorkflowExplorerProps) {
  const [selectedCategory, setSelectedCategory] = React.useState<string>("all");
  const [searchQuery, setSearchQuery] = React.useState("");

  const categories = React.useMemo(() => {
    const cats = new Set(workflows.map((w) => w.category));
    return ["all", ...Array.from(cats)];
  }, [workflows]);

  const filtered = workflows.filter((w) => {
    const matchesSearch = !searchQuery ||
      w.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      w.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = selectedCategory === "all" || w.category === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  return (
    <div className={cn("space-y-4", className)}>
      {/* Search + filter */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-text-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search workflows..."
            className="w-full pl-10 pr-4 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500 transition-colors"
          />
        </div>
        <select
          value={selectedCategory}
          onChange={(e) => setSelectedCategory(e.target.value)}
          className="px-3 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
        >
          {categories.map((cat) => (
            <option key={cat} value={cat}>{cat === "all" ? "All Categories" : cat}</option>
          ))}
        </select>
      </div>

      {/* Workflow grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {filtered.map((workflow) => (
          <div
            key={workflow.id}
            className="p-4 bg-bg-primary border border-border-subtle rounded-xl hover:border-border-default transition-colors"
          >
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-2">
                <div className="size-8 rounded-lg bg-gradient-to-br from-brand-500/20 to-brand-500/5 flex items-center justify-center">
                  <GitBranch className="size-4 text-brand-400" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-medium text-text-primary">{workflow.name}</h3>
                    {workflow.isFeatured && (
                      <Sparkles className="size-3 text-amber-400" />
                    )}
                  </div>
                  <p className="text-[10px] text-text-muted">by {workflow.author}</p>
                </div>
              </div>
              <div className="flex items-center gap-1">
                <Star className="size-3 fill-yellow-400 text-yellow-400" />
                <span className="text-[10px] text-text-primary">{workflow.rating.toFixed(1)}</span>
              </div>
            </div>

            <p className="text-[11px] text-text-secondary line-clamp-2 mb-3">{workflow.description}</p>

            <div className="flex items-center gap-2 mb-3">
              {workflow.tags.slice(0, 3).map((tag) => (
                <span key={tag} className="px-1.5 py-0.5 text-[9px] bg-bg-tertiary text-text-muted rounded">
                  {tag}
                </span>
              ))}
            </div>

            <div className="flex items-center justify-between text-[10px] text-text-muted mb-3">
              <span>{workflow.steps} steps</span>
              <span>{workflow.executions.toLocaleString()} executions</span>
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => onSelect?.(workflow.id)}
                className="flex-1 py-1.5 text-[11px] bg-bg-tertiary hover:bg-bg-hover text-text-secondary rounded-lg transition-colors"
              >
                View
              </button>
              <button
                onClick={() => onUse?.(workflow.id)}
                className="flex-1 py-1.5 text-[11px] bg-brand-500/10 hover:bg-brand-500/20 text-brand-400 rounded-lg transition-colors flex items-center justify-center gap-1"
              >
                <Zap className="size-3" /> Use
              </button>
            </div>
          </div>
        ))}
      </div>

      {filtered.length === 0 && (
        <div className="text-center py-12 text-text-muted">
          <GitBranch className="size-12 mx-auto mb-3 opacity-30" />
          <p className="text-sm">No workflows found</p>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// WorkflowTemplateGallery - Curated workflow templates
// ============================================================================

export interface WorkflowTemplateGalleryProps {
  categories: Array<{
    id: string;
    name: string;
    templates: WorkflowTemplate[];
  }>;
  onSelect?: (workflowId: string) => void;
  onUse?: (workflowId: string) => void;
  className?: string;
}

export function WorkflowTemplateGallery({
  categories,
  onSelect,
  onUse,
  className,
}: WorkflowTemplateGalleryProps) {
  const [expandedCategory, setExpandedCategory] = React.useState<string>(categories[0]?.id ?? "");

  return (
    <div className={cn("space-y-4", className)}>
      {categories.map((category) => (
        <div key={category.id} className="space-y-3">
          <button
            onClick={() => setExpandedCategory(expandedCategory === category.id ? "" : category.id)}
            className="flex items-center justify-between w-full p-3 bg-bg-secondary rounded-lg border border-border-subtle hover:border-border-default transition-colors"
          >
            <span className="text-sm font-medium text-text-primary">{category.name}</span>
            <div className="flex items-center gap-2">
              <span className="text-xs text-text-muted">{category.templates.length} templates</span>
              <ChevronRight className={cn(
                "size-4 text-text-muted transition-transform",
                expandedCategory === category.id && "rotate-90"
              )} />
            </div>
          </button>

          {expandedCategory === category.id && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 pl-2">
              {category.templates.map((template) => (
                <div
                  key={template.id}
                  className="p-4 bg-bg-primary border border-border-subtle rounded-xl hover:border-brand-500/30 transition-colors group"
                >
                  <div className="flex items-center gap-2 mb-2">
                    <div className="size-8 rounded-lg bg-brand-500/10 flex items-center justify-center">
                      <GitBranch className="size-4 text-brand-400" />
                    </div>
                    <div>
                      <h4 className="text-sm font-medium text-text-primary">{template.name}</h4>
                      <p className="text-[10px] text-text-muted">by {template.author}</p>
                    </div>
                  </div>
                  <p className="text-[11px] text-text-secondary line-clamp-2 mb-3">{template.description}</p>
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] text-text-muted">{template.executions.toLocaleString()} uses</span>
                    <button
                      onClick={() => onUse?.(template.id)}
                      className="text-[10px] text-brand-400 hover:text-brand-300 opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      Use template →
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

// ============================================================================
// PackageIntegrityViewer - Verify function package integrity
// ============================================================================

export interface IntegrityCheck {
  id: string;
  name: string;
  status: "passed" | "warning" | "failed";
  description: string;
  details?: string;
}

export interface PackageIntegrityViewerProps {
  functionId: string;
  functionName: string;
  version: string;
  checks: IntegrityCheck[];
  onReverify?: () => void;
  className?: string;
}

export function PackageIntegrityViewer({
  functionId,
  functionName,
  version,
  checks,
  onReverify,
  className,
}: PackageIntegrityViewerProps) {
  const [isVerifying, setIsVerifying] = React.useState(false);

  const handleVerify = async () => {
    setIsVerifying(true);
    await new Promise((resolve) => setTimeout(resolve, 2000));
    setIsVerifying(false);
    onReverify?.();
  };

  const passedCount = checks.filter((c) => c.status === "passed").length;
  const warningCount = checks.filter((c) => c.status === "warning").length;
  const failedCount = checks.filter((c) => c.status === "failed").length;
  const overallStatus = failedCount > 0 ? "failed" : warningCount > 0 ? "warning" : "passed";

  const statusColors = {
    passed: "text-emerald-400 bg-emerald-500/10",
    warning: "text-amber-400 bg-amber-500/10",
    failed: "text-red-400 bg-red-500/10",
  };

  const statusIcons = {
    passed: <CheckCircle className="size-4 text-emerald-400" />,
    warning: <AlertTriangle className="size-4 text-amber-400" />,
    failed: <XCircle className="size-4 text-red-400" />,
  };

  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
            <Shield className="size-4 text-brand-400" /> Package Integrity
          </h4>
          <p className="text-[10px] text-text-muted">{functionName} v{version}</p>
        </div>
        <button
          onClick={handleVerify}
          disabled={isVerifying}
          className="flex items-center gap-2 px-3 py-1.5 text-xs bg-bg-tertiary hover:bg-bg-hover text-text-secondary rounded-lg border border-border-subtle transition-colors disabled:opacity-50"
        >
          {isVerifying ? (
            <>
              <Loader2 className="size-3 animate-spin" /> Verifying...
            </>
          ) : (
            <>
              <RefreshCw className="size-3" /> Verify
            </>
          )}
        </button>
      </div>

      {/* Overall status */}
      <div className={cn("p-4 rounded-lg border", statusColors[overallStatus])}>
        <div className="flex items-center gap-3">
          {statusIcons[overallStatus]}
          <div>
            <div className="text-sm font-medium capitalize">{overallStatus}</div>
            <div className="text-[10px] opacity-70">
              {passedCount} passed · {warningCount} warnings · {failedCount} failed
            </div>
          </div>
        </div>
      </div>

      {/* Checks */}
      <div className="space-y-2">
        {checks.map((check) => (
          <div
            key={check.id}
            className={cn(
              "p-3 rounded-lg border",
              check.status === "passed" && "border-emerald-500/20 bg-emerald-500/5",
              check.status === "warning" && "border-amber-500/20 bg-amber-500/5",
              check.status === "failed" && "border-red-500/20 bg-red-500/5"
            )}
          >
            <div className="flex items-start gap-2">
              <div className="mt-0.5">{statusIcons[check.status]}</div>
              <div className="flex-1">
                <div className="text-xs font-medium text-text-primary">{check.name}</div>
                <div className="text-[11px] text-text-secondary">{check.description}</div>
                {check.details && (
                  <div className="mt-1 text-[10px] text-text-muted font-mono bg-bg-primary/50 p-2 rounded">
                    {check.details}
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// FunctionPermissionAudit - Audit function permissions and access
// ============================================================================

export interface PermissionEntry {
  id: string;
  principal: string;
  principalType: "user" | "team" | "api_key" | "public";
  permissions: string[];
  grantedAt: string;
  expiresAt?: string;
  isActive: boolean;
}

export interface FunctionPermissionAuditProps {
  functionId: string;
  functionName: string;
  permissions: PermissionEntry[];
  onRevoke?: (permissionId: string) => void;
  onAdd?: () => void;
  className?: string;
}

export function FunctionPermissionAudit({
  functionId,
  functionName,
  permissions,
  onRevoke,
  onAdd,
  className,
}: FunctionPermissionAuditProps) {
  const [activeOnly, setActiveOnly] = React.useState(true);

  const filteredPermissions = permissions.filter((p) => !activeOnly || p.isActive);

  const principalTypeColors = {
    user: "bg-blue-500/10 text-blue-400",
    team: "bg-purple-500/10 text-purple-400",
    api_key: "bg-amber-500/10 text-amber-400",
    public: "bg-gray-500/10 text-gray-400",
  };

  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
            <Shield className="size-4 text-brand-400" /> Permission Audit
          </h4>
          <p className="text-[10px] text-text-muted">{functionName}</p>
        </div>
        <button
          onClick={onAdd}
          className="flex items-center gap-1 px-3 py-1.5 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors"
        >
          <Plus className="size-3" /> Add Permission
        </button>
      </div>

      {/* Filter */}
      <label className="flex items-center gap-2 cursor-pointer">
        <input
          type="checkbox"
          checked={activeOnly}
          onChange={(e) => setActiveOnly(e.target.checked)}
          className="accent-brand-500"
        />
        <span className="text-xs text-text-secondary">Show active only</span>
      </label>

      {/* Permission list */}
      <div className="space-y-2">
        {filteredPermissions.map((perm) => (
          <div
            key={perm.id}
            className="p-3 bg-bg-secondary rounded-lg border border-border-subtle"
          >
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-2">
                <div className="size-8 rounded-lg bg-bg-tertiary flex items-center justify-center text-xs font-medium text-text-primary">
                  {perm.principal[0].toUpperCase()}
                </div>
                <div>
                  <div className="text-xs font-medium text-text-primary">{perm.principal}</div>
                  <span className={cn(
                    "text-[9px] px-1.5 py-0.5 rounded capitalize",
                    principalTypeColors[perm.principalType]
                  )}>
                    {perm.principalType.replace("_", " ")}
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {perm.isActive ? (
                  <span className="text-[10px] px-1.5 py-0.5 bg-emerald-500/10 text-emerald-400 rounded">Active</span>
                ) : (
                  <span className="text-[10px] px-1.5 py-0.5 bg-gray-500/10 text-gray-400 rounded">Inactive</span>
                )}
                {perm.expiresAt && (
                  <span className="text-[10px] text-text-muted">Expires: {perm.expiresAt}</span>
                )}
              </div>
            </div>

            <div className="flex flex-wrap gap-1 mb-2">
              {perm.permissions.map((p) => (
                <span key={p} className="px-1.5 py-0.5 text-[9px] bg-bg-tertiary text-text-muted rounded">
                  {p}
                </span>
              ))}
            </div>

            <div className="flex items-center justify-between text-[10px] text-text-muted">
              <span>Granted: {perm.grantedAt}</span>
              {perm.isActive && onRevoke && (
                <button
                  onClick={() => onRevoke(perm.id)}
                  className="text-red-400 hover:text-red-300"
                >
                  Revoke
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// FunctionChangelog - Version history and changelog
// ============================================================================

export interface ChangelogEntry {
  version: string;
  date: string;
  author: string;
  type: "major" | "minor" | "patch" | "security";
  changes: string[];
}

export interface FunctionChangelogProps {
  functionId: string;
  functionName: string;
  entries: ChangelogEntry[];
  onVersionSelect?: (version: string) => void;
  className?: string;
}

export function FunctionChangelog({
  functionId,
  functionName,
  entries,
  onVersionSelect,
  className,
}: FunctionChangelogProps) {
  const [selectedVersion, setSelectedVersion] = React.useState<string>(entries[0]?.version ?? "");

  const typeColors = {
    major: "bg-red-500/10 text-red-400 border-red-500/20",
    minor: "bg-blue-500/10 text-blue-400 border-blue-500/20",
    patch: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    security: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  };

  const typeIcons = {
    major: <AlertTriangle className="size-3" />,
    minor: <Sparkles className="size-3" />,
    patch: <CheckCircle className="size-3" />,
    security: <Shield className="size-3" />,
  };

  return (
    <div className={cn("space-y-4", className)}>
      <div>
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <History className="size-4 text-brand-400" /> Changelog
        </h4>
        <p className="text-[10px] text-text-muted">{functionName}</p>
      </div>

      {/* Version selector */}
      <div className="flex gap-1 overflow-x-auto pb-1">
        {entries.map((entry) => (
          <button
            key={entry.version}
            onClick={() => {
              setSelectedVersion(entry.version);
              onVersionSelect?.(entry.version);
            }}
            className={cn(
              "px-3 py-1.5 text-xs rounded-lg border whitespace-nowrap transition-colors",
              selectedVersion === entry.version
                ? "bg-brand-500/20 border-brand-500/30 text-brand-400"
                : "border-border-subtle text-text-muted hover:text-text-primary hover:border-border-default"
            )}
          >
            v{entry.version}
          </button>
        ))}
      </div>

      {/* Selected entry details */}
      {entries.find((e) => e.version === selectedVersion) && (
        <div className="p-4 bg-bg-secondary rounded-lg border border-border-subtle space-y-3">
          {(() => {
            const entry = entries.find((e) => e.version === selectedVersion)!;
            return (
              <>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className={cn(
                      "flex items-center gap-1 px-2 py-0.5 text-[10px] rounded-full border capitalize",
                      typeColors[entry.type]
                    )}>
                      {typeIcons[entry.type]} {entry.type}
                    </span>
                    <span className="text-sm font-bold text-text-primary">v{entry.version}</span>
                  </div>
                  <span className="text-[10px] text-text-muted">{entry.date}</span>
                </div>
                <div className="text-[10px] text-text-muted">by {entry.author}</div>
                <div className="space-y-1">
                  {entry.changes.map((change, i) => (
                    <div key={i} className="flex items-start gap-2 text-xs text-text-secondary">
                      <span className="text-brand-400 mt-0.5">•</span>
                      {change}
                    </div>
                  ))}
                </div>
              </>
            );
          })()}
        </div>
      )}
    </div>
  );
}

// ============================================================================
// FunctionBenchmarkPanel - Performance benchmarking
// ============================================================================

export interface BenchmarkResult {
  name: string;
  iterations: number;
  avgMs: number;
  p50Ms: number;
  p95Ms: number;
  p99Ms: number;
  minMs: number;
  maxMs: number;
  stdDev: number;
  isBaseline?: boolean;
}

export interface FunctionBenchmarkPanelProps {
  functionId: string;
  functionName: string;
  benchmarks: BenchmarkResult[];
  onRunBenchmark?: () => void;
  onSetBaseline?: (benchmarkName: string) => void;
  className?: string;
}

export function FunctionBenchmarkPanel({
  functionId,
  functionName,
  benchmarks,
  onRunBenchmark,
  onSetBaseline,
  className,
}: FunctionBenchmarkPanelProps) {
  const [isRunning, setIsRunning] = React.useState(false);
  const [selectedBenchmark, setSelectedBenchmark] = React.useState<string>(benchmarks[0]?.name ?? "");

  const handleRun = async () => {
    setIsRunning(true);
    await new Promise((resolve) => setTimeout(resolve, 3000));
    setIsRunning(false);
    onRunBenchmark?.();
  };

  const selected = benchmarks.find((b) => b.name === selectedBenchmark);

  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
            <Gauge className="size-4 text-brand-400" /> Benchmarks
          </h4>
          <p className="text-[10px] text-text-muted">{functionName}</p>
        </div>
        <button
          onClick={handleRun}
          disabled={isRunning}
          className="flex items-center gap-2 px-3 py-1.5 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors disabled:opacity-50"
        >
          {isRunning ? (
            <>
              <Loader2 className="size-3 animate-spin" /> Running...
            </>
          ) : (
            <>
              <Play className="size-3" /> Run Benchmarks
            </>
          )}
        </button>
      </div>

      {/* Benchmark selector */}
      <div className="flex gap-1 overflow-x-auto pb-1">
        {benchmarks.map((b) => (
          <button
            key={b.name}
            onClick={() => setSelectedBenchmark(b.name)}
            className={cn(
              "px-3 py-1.5 text-xs rounded-lg border whitespace-nowrap transition-colors",
              selectedBenchmark === b.name
                ? "bg-bg-tertiary border-border-default text-text-primary"
                : "border-border-subtle text-text-muted hover:text-text-primary"
            )}
          >
            {b.name}
            {b.isBaseline && (
              <span className="ml-1 text-[9px] text-brand-400">(baseline)</span>
            )}
          </button>
        ))}
      </div>

      {/* Results */}
      {selected && (
        <div className="space-y-3">
          {/* Latency bars */}
          <div className="space-y-2">
            {[
              { label: "Average", value: selected.avgMs, color: "bg-brand-500" },
              { label: "P50", value: selected.p50Ms, color: "bg-emerald-500" },
              { label: "P95", value: selected.p95Ms, color: "bg-amber-500" },
              { label: "P99", value: selected.p99Ms, color: "bg-red-500" },
              { label: "Min", value: selected.minMs, color: "bg-blue-500" },
              { label: "Max", value: selected.maxMs, color: "bg-purple-500" },
            ].map((metric) => {
              const max = selected.maxMs * 1.2;
              return (
                <div key={metric.label} className="flex items-center gap-2">
                  <span className="w-12 text-[10px] text-text-muted">{metric.label}</span>
                  <div className="flex-1 h-3 bg-bg-tertiary rounded-full overflow-hidden">
                    <div
                      className={cn("h-full rounded-full transition-all", metric.color)}
                      style={{ width: `${(metric.value / max) * 100}%` }}
                    />
                  </div>
                  <span className="w-16 text-[10px] text-text-primary text-right font-mono">
                    {metric.value.toFixed(2)}ms
                  </span>
                </div>
              );
            })}
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-2">
            <div className="p-2 bg-bg-secondary rounded-lg text-center">
              <div className="text-[10px] text-text-muted">Iterations</div>
              <div className="text-sm font-bold text-text-primary">{selected.iterations.toLocaleString()}</div>
            </div>
            <div className="p-2 bg-bg-secondary rounded-lg text-center">
              <div className="text-[10px] text-text-muted">Std Dev</div>
              <div className="text-sm font-bold text-text-primary">{selected.stdDev.toFixed(2)}ms</div>
            </div>
            <div className="p-2 bg-bg-secondary rounded-lg text-center">
              <div className="text-[10px] text-text-muted">Variance</div>
              <div className="text-sm font-bold text-text-primary">{((selected.stdDev / selected.avgMs) * 100).toFixed(1)}%</div>
            </div>
          </div>

          {/* Set baseline button */}
          {onSetBaseline && !selected.isBaseline && (
            <button
              onClick={() => onSetBaseline(selected.name)}
              className="w-full py-2 text-xs text-brand-400 hover:text-brand-300 border border-brand-500/20 hover:border-brand-500/30 rounded-lg transition-colors"
            >
              Set as baseline
            </button>
          )}
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Additional icons needed
// ============================================================================

function XCircle({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <circle cx="12" cy="12" r="10" />
      <line x1="15" y1="9" x2="9" y2="15" />
      <line x1="9" y1="9" x2="15" y2="15" />
    </svg>
  );
}

function ChevronRight({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <polyline points="9 18 15 12 9 6" />
    </svg>
  );
}

function EyeOff({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
      <line x1="1" y1="1" x2="23" y2="23" />
    </svg>
  );
}

function Plus({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </svg>
  );
}

function History({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M3 3v5h5" />
      <path d="M3.05 13A9 9 0 1 0 6 5.3L3 8" />
      <path d="M12 7v5l4 2" />
    </svg>
  );
}

function Gauge({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M12 2a10 10 0 1 0 10 10H12V2z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

// --- Helper function for debouncing ---
function debounce<T extends (...args: any[]) => any>(fn: T, ms: number): T {
  let timeoutId: ReturnType<typeof setTimeout>;
  return ((...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), ms);
  }) as T;
}

export {};
