/**
 * @functionfly/ui-marketplace-economy
 * Marketplace Economy Components - Creator payouts, licensing, royalties, and billing
 */

// ============================================================================
// Shared Types
// ============================================================================

export type RevenueType = 'subscription' | 'one-time' | 'royalty' | 'tier';
export type BillingCycle = 'monthly' | 'quarterly' | 'annual';
export type LicenseType = 'open' | 'restricted' | 'commercial';
export type TrendDirection = 'up' | 'down' | 'stable';
export type SubscriptionStatus = 'active' | 'cancelled' | 'past_due' | 'trialing';

// ============================================================================
// CreatorEconomy
// ============================================================================

export interface CreatorEarnings {
  totalRevenue: number;
  subscriptionRevenue: number;
  oneTimeRevenue: number;
  royaltyRevenue: number;
  pendingPayout: number;
  currency: string;
}

export interface CreatorEconomyProps {
  earnings: CreatorEarnings;
  monthlyTrend: number;
  activeSubscribers: number;
  topFunctionId?: string | null;
  onViewDetails?: () => void;
  className?: string;
}

// ============================================================================
// RevenueAnalytics
// ============================================================================

export interface RevenueDataPoint {
  timestamp: number;
  revenue: number;
  subscription: number;
  oneTime: number;
  royalty: number;
}

export interface RevenueAnalyticsProps {
  data: RevenueDataPoint[];
  totalRevenue: number;
  period: '7d' | '30d' | '90d' | '1y';
  onPeriodChange?: (period: '7d' | '30d' | '90d' | '1y') => void;
  onDataPointSelect?: (point: RevenueDataPoint) => void;
  className?: string;
}

// ============================================================================
// SubscriptionManager
// ============================================================================

export interface Subscription {
  id: string;
  customerId: string;
  customerName: string;
  customerEmail: string;
  plan: string;
  status: SubscriptionStatus;
  billingCycle: BillingCycle;
  amount: number;
  currency: string;
  currentPeriodStart: number;
  currentPeriodEnd: number;
  cancelAtPeriodEnd: boolean;
}

export interface SubscriptionManagerProps {
  subscriptions: Subscription[];
  activeCount: number;
  cancelledCount: number;
  pastDueCount: number;
  onSubscriptionSelect?: (subscription: Subscription) => void;
  onSubscriptionCancel?: (subscriptionId: string) => void;
  onSubscriptionRenew?: (subscriptionId: string) => void;
  className?: string;
}

// ============================================================================
// UsageBillingPanel
// ============================================================================

export interface UsageMetric {
  name: string;
  unit: string;
  used: number;
  limit: number;
  cost: number;
}

export interface UsageBillingPanelProps {
  metrics: UsageMetric[];
  totalCost: number;
  billingCycle: BillingCycle;
  currency: string;
  onUpgradeClick?: () => void;
  onMetricClick?: (metric: UsageMetric) => void;
  className?: string;
}

// ============================================================================
// LicenseManager
// ============================================================================

export interface License {
  id: string;
  key: string;
  type: LicenseType;
  functionId: string;
  functionName: string;
  purchaserId: string;
  purchaserName: string;
  issuedAt: number;
  expiresAt?: number | null;
  maxActivations?: number | null;
  activationCount: number;
  revoked: boolean;
}

export interface LicenseManagerProps {
  licenses: License[];
  totalActive: number;
  totalRevoked: number;
  onLicenseSelect?: (license: License) => void;
  onLicenseRevoke?: (licenseId: string) => void;
  onLicenseGenerate?: () => void;
  className?: string;
}

// ============================================================================
// CreatorProfile
// ============================================================================

export interface CreatorStats {
  totalSales: number;
  totalRevenue: number;
  averageRating: number;
  totalReviews: number;
  functionsPublished: number;
  followers: number;
}

export interface PayoutInfo {
  nextPayoutDate: number;
  nextPayoutAmount: number;
  currency: string;
  payoutMethod: string;
  minimumPayout: number;
}

export interface CreatorProfileProps {
  creatorId: string;
  creatorName: string;
  creatorAvatar?: string | null;
  stats: CreatorStats;
  payoutInfo: PayoutInfo;
  onEditProfile?: () => void;
  onViewPayoutHistory?: () => void;
  className?: string;
}

// ============================================================================
// MarketplaceLeaderboard
// ============================================================================

export interface LeaderboardEntry {
  rank: number;
  creatorId: string;
  creatorName: string;
  creatorAvatar?: string | null;
  functionId: string;
  functionName: string;
  sales: number;
  revenue: number;
  rating: number;
  trend: TrendDirection;
}

export interface MarketplaceLeaderboardProps {
  entries: LeaderboardEntry[];
  category: 'creators' | 'functions' | 'revenue';
  timeRange: '7d' | '30d' | '90d' | 'all';
  onEntrySelect?: (entry: LeaderboardEntry) => void;
  className?: string;
}

// ============================================================================
// FunctionRoyaltiesPanel
// ============================================================================

export interface RoyaltyRecord {
  id: string;
  functionId: string;
  functionName: string;
  licensee: string;
  licenseType: LicenseType;
  royaltyPercentage: number;
  saleAmount: number;
  royaltyAmount: number;
  currency: string;
  saleDate: number;
  paidOut: boolean;
}

export interface FunctionRoyaltiesPanelProps {
  royalties: RoyaltyRecord[];
  totalEarned: number;
  totalPending: number;
  currency: string;
  onRoyaltySelect?: (royalty: RoyaltyRecord) => void;
  onClaimPending?: () => void;
  className?: string;
}

// ============================================================================
// AssetPricingEditor
// ============================================================================

export interface PricingTier {
  id: string;
  name: string;
  price: number;
  currency: string;
  features: string[];
  popular?: boolean;
}

export interface AssetPricingEditorProps {
  functionId: string;
  functionName: string;
  currentPricing: PricingTier[];
  suggestedPrices?: PricingTier[];
  competitorPriceRange?: { min: number; max: number };
  onPriceUpdate?: (tiers: PricingTier[]) => void;
  onSave?: () => void;
  className?: string;
}

// ============================================================================
// SalesConversionAnalytics
// ============================================================================

export interface ConversionFunnelStep {
  stage: string;
  count: number;
  dropOffRate: number;
  conversionRate: number;
}

export interface SalesConversionAnalyticsProps {
  funnelSteps: ConversionFunnelStep[];
  totalVisitors: number;
  totalPurchases: number;
  overallConversionRate: number;
  averageOrderValue: number;
  onStageClick?: (step: ConversionFunnelStep) => void;
  className?: string;
}

// ============================================================================
// MonetizationOptimizer
// ============================================================================

export interface OptimizationSuggestion {
  id: string;
  type: 'pricing' | 'listing' | 'marketing' | 'retention';
  title: string;
  description: string;
  potentialImpact: number;
  effort: 'low' | 'medium' | 'high';
  priority: 'high' | 'medium' | 'low';
  actionable: boolean;
}

export interface MonetizationOptimizerProps {
  suggestions: OptimizationSuggestion[];
  currentRevenue: number;
  projectedRevenue?: number;
  onSuggestionApply?: (suggestion: OptimizationSuggestion) => void;
  onSuggestionDismiss?: (suggestionId: string) => void;
  className?: string;
}

// ============================================================================
// MarketplaceTrendRadar
// ============================================================================

export interface TrendItem {
  id: string;
  name: string;
  category: string;
  momentum: number;
  direction: TrendDirection;
  relatedFunctions: string[];
  avgPrice: number;
  searchVolume: number;
}

export interface MarketplaceTrendRadarProps {
  trends: TrendItem[];
  timeRange: '7d' | '30d' | '90d';
  selectedCategory?: string | null;
  onTrendSelect?: (trend: TrendItem) => void;
  onCategoryChange?: (category: string) => void;
  className?: string;
}
