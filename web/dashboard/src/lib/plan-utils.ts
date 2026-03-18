import { PLANS } from './constants';

export type PlanTier = 'free' | 'starter' | 'professional' | 'enterprise';

export const PLAN_HIERARCHY: Record<PlanTier, number> = {
  free: 0,
  starter: 1,
  professional: 2,
  enterprise: 3,
};

/**
 * Check if user is on enterprise plan
 */
export const isEnterprise = (plan?: string): boolean =>
  plan?.toLowerCase() === 'enterprise';

/**
 * Check if user has at least the specified plan tier
 */
export const hasMinPlan = (userPlan: string, minPlan: PlanTier): boolean => {
  const userTier = PLAN_HIERARCHY[userPlan.toLowerCase() as PlanTier] ?? 0;
  const requiredTier = PLAN_HIERARCHY[minPlan];
  return userTier >= requiredTier;
};

/**
 * Get plan display name with enterprise styling
 */
export const getPlanDisplayName = (plan?: string): string => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.name || plan || 'Unknown';
};

/**
 * Get plan limits for the user's tier
 */
export const getPlanLimits = (plan?: string) => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.limits;
};

/**
 * Get the maximum number of secrets allowed for the user's tier
 * Returns Infinity for enterprise, 0 for free tier (no secrets)
 */
export const getSecretsLimit = (plan?: string): number => {
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  // Handle Infinity case for enterprise
  if (limits.secrets === Infinity) return 10000;
  return limits.secrets ?? 0;
};

/**
 * Get the maximum number of tokens per secret allowed for the user's tier
 * Returns 0 for free tier (no tokens)
 */
export const getTokensPerSecretLimit = (plan?: string): number => {
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  // Handle Infinity case for enterprise
  if (limits.tokensPerSecret === Infinity) return 100;
  return limits.tokensPerSecret ?? 0;
};

/**
 * Get the maximum number of custom domains allowed for the user's tier
 * Returns 0 for free tier (no custom domains)
 */
export const getCustomDomainsLimit = (plan?: string): number => {
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { customDomains?: number }).customDomains;
  if (raw === undefined) return 0;
  return raw === Infinity ? 10000 : raw;
};

/**
 * Check if the user can add more custom domains (has feature and under limit)
 */
export const canAddCustomDomain = (
  plan: string | undefined,
  currentCount: number
): boolean => {
  if (!hasFeature(plan, 'CUSTOM_DOMAINS')) return false;
  const limit = getCustomDomainsLimit(plan);
  if (limit === 0) return false;
  return currentCount < limit;
};

/**
 * Format custom domains usage for display (e.g. "2 of 5" or "Unlimited")
 */
export const formatCustomDomainsRemaining = (
  currentCount: number,
  plan?: string
): string => {
  const limit = getCustomDomainsLimit(plan);
  if (limit === 0) return 'Not available on your plan';
  if (limit >= 10000) return `${currentCount} (unlimited)`;
  return `${currentCount} of ${limit}`;
};

/**
 * Check if the user can create secrets based on their tier
 * Returns true if they have at least some secrets allowed
 */
export const canCreateSecrets = (plan?: string): boolean => {
  return getSecretsLimit(plan) > 0;
};

/**
 * Check if the user can create tokens based on their tier
 * Returns true if they have at least some tokens allowed
 */
export const canCreateTokens = (plan?: string): boolean => {
  return getTokensPerSecretLimit(plan) > 0;
};

/**
 * Format secrets remaining for display
 * Shows "X of Y used" or "X remaining" or "Unlimited"
 */
export const formatSecretsRemaining = (
  currentCount: number,
  plan?: string
): string => {
  const limit = getSecretsLimit(plan);

  if (limit === Infinity || limit >= 10000) {
    return `${currentCount} secrets (unlimited)`;
  }

  if (limit === 0) {
    return "Secrets not available on your plan";
  }

  const remaining = limit - currentCount;
  return `${currentCount} of ${limit} secrets used (${remaining} remaining)`;
};

/**
 * Feature availability by plan tier
 */
export const FEATURES: Record<string, readonly PlanTier[]> = {
  ADVANCED_ANALYTICS: ['professional', 'enterprise'],
  CUSTOM_DASHBOARDS: ['enterprise'],
  AUDIT_LOGS: ['enterprise'],
  SLA_DASHBOARD: ['enterprise'],
  DEDICATED_SUPPORT: ['enterprise'],
  EXPORT_REPORTS: ['enterprise'],
  TEAM_MANAGEMENT: ['professional', 'enterprise'],
  API_ACCESS: ['starter', 'professional', 'enterprise'],
  WEBHOOKS: ['professional', 'enterprise'],
  CUSTOM_DOMAINS: ['starter', 'professional', 'enterprise'],
  UNLIMITED_FUNCTIONS: ['enterprise'],
  UNLIMITED_PROVIDERS: ['enterprise'],
  PRIORITY_SUPPORT: ['professional', 'enterprise'],
} as const;

export type FeatureKey = keyof typeof FEATURES;

/**
 * Check if a feature is available for the plan
 */
export const hasFeature = (plan: string | undefined, feature: FeatureKey): boolean => {
  if (!plan) return false;
  const allowedPlans = FEATURES[feature];
  return allowedPlans.includes(plan.toLowerCase() as PlanTier);
};

/**
 * Get all features available for a plan
 */
export const getAvailableFeatures = (plan?: string): FeatureKey[] => {
  if (!plan) return [];
  return (Object.keys(FEATURES) as FeatureKey[]).filter((feature) =>
    hasFeature(plan, feature)
  );
};

/**
 * Check if plan has unlimited resources
 */
export const hasUnlimitedResources = (plan?: string): boolean =>
  isEnterprise(plan);

/**
 * Get plan color for UI theming
 */
export const getPlanColor = (plan?: string): string => {
  switch (plan?.toLowerCase()) {
    case 'enterprise':
      return 'amber';
    case 'professional':
      return 'indigo';
    case 'starter':
      return 'emerald';
    case 'free':
    default:
      return 'slate';
  }
};

/**
 * Format plan name for display with proper capitalization
 */
export const formatPlanName = (plan?: string): string => {
  if (!plan) return 'Unknown';
  return plan.charAt(0).toUpperCase() + plan.slice(1).toLowerCase();
};
