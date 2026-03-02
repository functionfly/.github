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
